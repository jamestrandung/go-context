package memoize

import (
	"errors"
)

var (
	ErrPanicExecutingMemoizedFn = errors.New("panic executing memoizedFn")
	ErrCacheAlreadyDestroyed    = errors.New("cache already destroyed, cannot be used anymore")
	ErrMemoizedFnCannotBeNil    = errors.New("memoizedFn cannot be nil")
)
