package memoize

import "context"

type contextKey struct{}

var memoizeStoreKey = contextKey{}

type DestroyFn func()

// WithCache returns a new context.Context that holds a reference to
// a cache for memoized functions. This is meant to be a request-level
// cache that will automatically get garbage-collected at the end of
// an API request when the context itself is garbage-collected.
//
// Note: the return DestroyFn must be deferred to minimize memory leaks.
func WithCache(ctx context.Context) (context.Context, DestroyFn) {
	c := newCache()
	return context.WithValue(ctx, memoizeStoreKey, c), c.destroy
}

// extractCache looks for the iCache stored in this context and
// returns it. If it doesn't exist, a no-op cache will be returned
// instead. All functions executed via this no-op cache will not
// be memoized.
func extractCache(ctx context.Context) iCache {
	val := ctx.Value(memoizeStoreKey)
	if c, ok := val.(iCache); ok {
		return c
	}

	return noMemoizeCache{}
}

// Execute guarantees that the given memoizedFn will be invoked only
// once regardless of how many times Execute gets called with the same
// executionKey. All callers will receive the same result and error as
// the result of this call.
//
// Note 1: this promise can only be kept if the given context has been
// initialized using WithCache before calling Execute. The last return
// value indicates whether this function call was memoized or not.
//
// Note 2: The provided key must be comparable and should not be of type
// string or any other built-in type to avoid collisions between packages
// using this context. Callers of Execute should define their own types
// for keys.
func Execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn func(context.Context) (interface{}, error),
) (result interface{}, err error, isMemoized bool) {
	c := extractCache(ctx)
	return c.execute(ctx, executionKey, memoizedFn)
}
