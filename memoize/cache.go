package memoize

import (
	"context"
	"reflect"
	"sync/atomic"
)

// iCache represents a cache for memoized functions.
type iCache interface {
	// execute guarantees that the given memoizedFn will be invoked only
	// once regardless of how many times Execute gets called with the same
	// executionKey. All callers will receive the same result and error as
	// the result of this call.
	execute(
		ctx context.Context,
		executionKey interface{},
		memoizedFn Function,
	) (interface{}, error, bool)
}

type noMemoizeCache struct{}

func (c noMemoizeCache) preExecuteCheck(memoizedFn Function) error {
	if memoizedFn == nil {
		return ErrMemoizedFnCannotBeNil
	}

	return nil
}

func (c noMemoizeCache) execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn Function,
) (interface{}, error, bool) {
	preExecuteCheckErr := c.preExecuteCheck(memoizedFn)
	if preExecuteCheckErr != nil {
		return nil, preExecuteCheckErr, false
	}

	result, err := execute(ctx, memoizedFn)
	return result, err, false
}

type cache struct {
	isDestroyed int32
	store       *store
}

func newCache() *cache {
	c := &cache{
		store: newStore(),
	}

	return c
}

// destroy clears existing items in this cache and mark it as destroyed.
// Subsequent calls to Execute will return ErrCacheAlreadyDestroyed.
func (c *cache) destroy() {
	atomic.StoreInt32(&c.isDestroyed, 1)
	c.store = nil
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
		result, err := execute(ctx, memoizedFn)
		return result, err, false
	}

	result, err := c.store.promise(executionKey, memoizedFn).get(ctx)
	return result, err, true
}
