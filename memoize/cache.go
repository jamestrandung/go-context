package memoize

import (
	"context"
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

    result, err := doExecute(ctx, memoizedFn)
    return result, err, false
}
