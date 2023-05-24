package memoize

import (
	"context"
	"sync/atomic"
)

// iCache represents a cache for memoized functions.
type iCache interface {
	// destroy clears existing items in this cache and mark it as destroyed.
	// Subsequent calls to execute will return ErrCacheAlreadyDestroyed.
	destroy()
	// take will put the given entries into this cache. The key of such
	// entries should be the executionKey that would be used to call
	// execute. The value should be the Outcome that you want to map to
	// this executionKey.
	take(entries map[interface{}]Outcome)
	// execute guarantees that the given memoizedFn will be invoked only
	// once regardless of how many times Execute gets called with the same
	// executionKey. All callers will receive the same result and error as
	// the result of this call.
	execute(
		ctx context.Context,
		executionKey interface{},
		memoizedFn Function,
	) (Outcome, Extra)
	// findPromises returns all promise that were memoized under the given
	// executionKey type at the time findPromises was called.
	//
	// Note: if executionKey is nil, all promises will be returned.
	findPromises(executionKey interface{}) map[interface{}]*promise
}

type noMemoizeCache struct {
	isDestroyed int64
}

func (c *noMemoizeCache) destroy() {
	atomic.StoreInt64(&c.isDestroyed, 1)
}

func (c *noMemoizeCache) take(entries map[interface{}]Outcome) {
	// do nothing
}

func (c *noMemoizeCache) execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn Function,
) (Outcome, Extra) {
	if atomic.LoadInt64(&c.isDestroyed) == 1 {
		return Outcome{
				Value: nil,
				Err:   ErrCacheAlreadyDestroyed,
			}, Extra{
				IsMemoized: false,
				IsExecuted: false,
			}
	}

	if memoizedFn == nil {
		return Outcome{
				Value: nil,
				Err:   ErrMemoizedFnCannotBeNil,
			}, Extra{
				IsMemoized: false,
				IsExecuted: false,
			}
	}

	result, err := doExecute(ctx, memoizedFn)
	return Outcome{
			Value: result,
			Err:   err,
		}, Extra{
			IsMemoized: false,
			IsExecuted: true,
		}
}

func (c *noMemoizeCache) findPromises(executionKey interface{}) map[interface{}]*promise {
	return nil
}
