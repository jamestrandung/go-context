package memoize

import (
	"context"
	"fmt"
	"github.com/jamestrandung/go-context/cext"
	"github.com/pkg/errors"
	"reflect"
	"runtime/debug"
	"runtime/trace"
	"sync"
	"sync/atomic"
)

// Function is the type of function that can be memoized.
//
// The argument must not materially affect the result of the function in ways
// that are not captured by the promise's key, since if promise.get is called
// twice concurrently, with the same (implicit) key but different arguments,
// the Function is called once and only once but its result must be suitable
// for both callers.
type Function func(ctx context.Context) (interface{}, error)

type outcome struct {
	value interface{}
	err   error
}

// A promise represents the future result of a call to a function.
type promise struct {
	debug string // for observability

	// the rootCtx that was used to initialize a cache and would provide
	// the cancelling signal for our execution
	rootCtx context.Context
	// executed is set when execution starts so that the memoized function
	// does not get executed more than once.
	executed int32
	// finished is set when execution completes so that future callers can
	// use outcome immediately instead of waiting on done.
	finished bool
	// done is closed when execution completes to unblock concurrent waiters.
	done chan struct{}
	// the function that will be used to populate the outcome.
	function Function
	// outcome is set when execution completes.
	outcome outcome
}

// newPromise returns a promise for the future result of calling the
// specified function.
//
// The debug string is used to classify promises in logs and metrics.
// It should be drawn from a small set.
func newPromise(debug string, rootCtx context.Context, function Function) *promise {
	if function == nil {
		panic("nil function")
	}

	return &promise{
		debug:    debug,
		rootCtx:  rootCtx,
		done:     make(chan struct{}),
		function: function,
	}
}

// get returns the value associated with a promise.
//
// All calls to promise.get on a given promise return the same result
// but the function is called (to completion) at most once.
//
// - If the underlying function has not been invoked, it will be.
// - If ctx is cancelled, get returns (nil, context.Canceled).
func (p *promise) get(ctx context.Context) (interface{}, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if atomic.CompareAndSwapInt32(&p.executed, 0, 1) {
		return p.run(ctx)
	}

	if p.finished {
		return p.outcome.value, p.outcome.err
	}

	return p.wait(ctx)
}

// run starts p.function and returns the result.
func (p *promise) run(ctx context.Context) (interface{}, error) {
	// To prevent one child goroutines from cancelling the execution of the memoized
	// function that is still meaningful to other goroutines, we will delegate the
	// value retrieving responsibility to the input context while letting the root
	// context handle cancelling signals.
	//
	// This makes sense because the root context that was used to initialize a cache
	// should be the parent of all child contexts, including the input context. If
	// the root context get cancelled, all child contexts must be cancelled as well.
	delegatingCtx := cext.Delegate(p.rootCtx, ctx)

	go func() {
		trace.WithRegion(
			delegatingCtx, fmt.Sprintf("promise.run %s", p.debug), func() {
				v, err := doExecute(delegatingCtx, p.function)

				p.outcome = outcome{
					value: v,
					err:   err,
				}
				p.function = nil // aid GC
				p.finished = true
				close(p.done)
			},
		)
	}()

	return p.wait(ctx)
}

// wait waits for the value to be computed, or ctx to be cancelled. p.mu must be locked.
func (p *promise) wait(ctx context.Context) (interface{}, error) {
	select {
	case <-p.done:
		return p.outcome.value, p.outcome.err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// A cache maps arbitrary keys to promises.
type cache struct {
	rootCtx     context.Context
	isDestroyed int32
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

// destroy clears existing items in this cache and mark it as destroyed.
// Subsequent calls to Execute will return ErrCacheAlreadyDestroyed.
func (c *cache) destroy() {
	atomic.StoreInt32(&c.isDestroyed, 1)
	c.promises = nil
}

// promise returns a promise for the future result of calling the given function.
// Calls to promise with the same key return the same promise.
func (c *cache) promise(executionKey interface{}, function Function) *promise {
	c.promisesMu.Lock()
	defer c.promisesMu.Unlock()

	p, ok := c.promises[executionKey]
	if !ok {
		p = c.createPromise(executionKey, function)
	}

	return p
}

func (c *cache) createPromise(executionKey interface{}, function Function) *promise {
	p := newPromise(reflect.TypeOf(executionKey).String(), c.rootCtx, function)
	if c.promises == nil {
		c.promises = make(map[interface{}]*promise)
	}

	c.promises[executionKey] = p

	return p
}

func (c *cache) execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn Function,
) (interface{}, error, bool) {
	err := c.preExecuteCheck(memoizedFn)
	if err != nil {
		return nil, err, false
	}

	if executionKey == nil || !reflect.TypeOf(executionKey).Comparable() {
		result, err := doExecute(ctx, memoizedFn)
		return result, err, false
	}

	result, err := c.promise(executionKey, memoizedFn).get(ctx)
	return result, err, true
}

func (c *cache) preExecuteCheck(memoizedFn Function) error {
	if memoizedFn == nil {
		return ErrMemoizedFnCannotBeNil
	}

	if atomic.LoadInt32(&c.isDestroyed) != 0 {
		return ErrCacheAlreadyDestroyed
	}

	return nil
}

func doExecute(ctx context.Context, memoizedFn Function) (result interface{}, err error) {
	// Convert panics into standard errors for clients to handle gracefully
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = errors.Wrap(ErrPanicExecutingMemoizedFn, fmt.Sprintf("%v \n %v", r, debug.Stack()))
		}
	}()

	result, err = memoizedFn(ctx)
	return
}
