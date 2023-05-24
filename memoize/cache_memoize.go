package memoize

import (
	"context"
	"fmt"
	"github.com/jamestrandung/go-context/helper"
	"github.com/pkg/errors"
	"reflect"
	"runtime/debug"
	"sync"
)

// A cache maps arbitrary keys to promises.
type cache struct {
	rootCtx     context.Context
	isDestroyed bool
	promisesMu  sync.Mutex
	promises    map[interface{}]*promise
}

// newCache creates a new cache.
func newCache(rootCtx context.Context) *cache {
	return &cache{
		rootCtx:  rootCtx,
		promises: make(map[interface{}]*promise),
	}
}

func (c *cache) destroy() {
	c.promisesMu.Lock()
	defer c.promisesMu.Unlock()

	c.isDestroyed = true
	c.promises = nil
}

func (c *cache) take(entries map[interface{}]Outcome) {
	c.promisesMu.Lock()
	defer c.promisesMu.Unlock()

	if c.isDestroyed {
		return
	}

	if c.promises == nil {
		c.promises = make(map[interface{}]*promise)
	}

	for executionKey, outcome := range entries {
		if executionKey == nil {
			continue
		}

		p := completedPromise(c.extractExecutionKeyType(executionKey), outcome)
		c.promises[executionKey] = p
	}
}

func (c *cache) execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn Function,
) (Outcome, Extra) {
	if memoizedFn == nil {
		return Outcome{
				Value: nil,
				Err:   ErrMemoizedFnCannotBeNil,
			}, Extra{
				IsMemoized: false,
				IsExecuted: false,
			}
	}

	if !helper.IsComparable(executionKey) {
		result, err := doExecute(ctx, memoizedFn)
		return Outcome{
				Value: result,
				Err:   err,
			}, Extra{
				IsMemoized: false,
				IsExecuted: true,
			}
	}

	p, err := c.promise(executionKey, memoizedFn)
	if err != nil {
		return Outcome{
				Value: nil,
				Err:   err,
			}, Extra{
				IsMemoized: false,
				IsExecuted: false,
			}
	}

	return p.get(ctx), Extra{
		IsMemoized: true,
		IsExecuted: p.isExecuted(),
	}
}

// promise returns a promise for the future result of calling the given function.
// Calls to promise with the same key return the same promise.
func (c *cache) promise(executionKey interface{}, function Function) (*promise, error) {
	c.promisesMu.Lock()
	defer c.promisesMu.Unlock()

	if c.isDestroyed {
		return nil, ErrCacheAlreadyDestroyed
	}

	p, ok := c.promises[executionKey]
	if !ok {
		return c.createPromise(executionKey, function), nil
	}

	return p, nil
}

func (c *cache) createPromise(executionKey interface{}, function Function) *promise {
	p := newPromise(c.extractExecutionKeyType(executionKey), c.rootCtx, function)
	if c.promises == nil {
		c.promises = make(map[interface{}]*promise)
	}

	c.promises[executionKey] = p

	return p
}

func (c *cache) findPromises(executionKey interface{}) map[interface{}]*promise {
	returnAll := false
	if executionKey == nil {
		returnAll = true
	}

	c.promisesMu.Lock()
	defer c.promisesMu.Unlock()

	if c.isDestroyed {
		return nil
	}

	executionKeyType := func() string {
		if returnAll {
			return ""
		}

		return c.extractExecutionKeyType(executionKey)
	}()

	m := make(map[interface{}]*promise)
	for key, p := range c.promises {
		if !returnAll && p.executionKeyType != executionKeyType {
			continue
		}

		m[key] = p
	}

	return m
}

func (c *cache) extractExecutionKeyType(executionKey interface{}) string {
	return reflect.TypeOf(executionKey).String()
}

func doExecute(ctx context.Context, memoizedFn Function) (result interface{}, err error) {
	// Convert panics into standard errors for clients to handle gracefully
	defer func() {
		if r := recover(); r != nil {
			result = nil

			stackTrace := func() string {
				stack := debug.Stack()
				if len(stack) == 0 {
					return ""
				}

				return string(stack)
			}()

			err = errors.Wrap(ErrPanicExecutingMemoizedFn, fmt.Sprintf("%v \n %v", r, stackTrace))
		}
	}()

	result, err = memoizedFn(ctx)
	return
}
