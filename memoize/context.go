package memoize

import (
	"context"
)

type contextKey struct{}

var memoizeStoreKey = contextKey{}

type DestroyFn func()

// WithCache returns a new context.Context that holds a reference to
// a cache for memoized functions. This is meant to be a request-level
// cache that will automatically get garbage-collected at the end of
// an API request when the context itself is garbage-collected.
//
// WithCache must be called near the start of an API request handling
// before being any memoized functions get executed in child goroutines.
//
// The given context will be used as the root context of this cache. If
// it gets cancelled, all pending memoized executions will be abandoned.
// On the other hand, the context given to Execute won't affect pending
// executions. Child goroutines can cancel the context given to Execute
// to stop waiting for the result from the memoized function, which will
// still proceed till completion.
//
// Note: the return DestroyFn must be deferred to minimize memory leaks.
func WithCache(ctx context.Context) (context.Context, DestroyFn) {
	c := newCache(ctx)
	return context.WithValue(ctx, memoizeStoreKey, c), c.destroy
}

// WithConcurrentCache returns a new context.Context that holds a reference
// to a cache for memoized functions. This is meant to be a request-level
// cache that will automatically get garbage-collected at the end of an API
// request when the context itself is garbage-collected.
//
// WithConcurrentCache must be called at the start of an API request handling
// before being any memoized functions get executed in child goroutines.
//
// The given context will be used as the root context of this cache. If it is
// cancelled, all pending memoized executions will be abandoned. In contrast,
// the context given to Execute won't affect pending executions. Child routines
// can cancel the context given to Execute to stop waiting for the result from
// the memoized function, which will still proceed till completion.
//
// Note: the return DestroyFn must be deferred to minimize memory leaks.
func WithConcurrentCache(ctx context.Context, concurrencyLevel int) (context.Context, DestroyFn) {
	c := func() iCache {
		if concurrencyLevel == 1 {
			return newCache(ctx)
		}

		return newConcurrentCache(ctx, concurrencyLevel)
	}()

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

	return &noMemoizeCache{}
}

// PopulateCache will put the given entries into this cache. The key
// of such entries should be the executionKey that would be used to
// call execute. The value should be the Outcome that you want to map
// to this executionKey.
//
// Note: the given entries can only be populated in the cache if the
// input context has been initialized using WithCache.
func PopulateCache(ctx context.Context, entries map[interface{}]Outcome) {
	c := extractCache(ctx)
	c.take(entries)
}

// Execute guarantees that the given memoizedFn will be invoked only
// once regardless of how many times Execute gets called with the same
// executionKey. All callers will receive the same result and error as
// the result of this call.
//
// Note 1: this promise can only be kept if the given context has been
// initialized using WithCache before calling Execute.
//
// Note 2: the provided key must be comparable and should not be of type
// string or any other built-in type to avoid collisions between packages
// using this context. Callers of Execute should define their own types
// for keys similar to the best practices for using context.WithValue.
//
// Note 3: cancelling the given context allows caller to stop waiting
// for the result from the memoizedFn. However, the memoizedFn will
// still proceed till completion unless the root context given to
// WithCache was cancelled.
func Execute[K comparable, V any](
	ctx context.Context,
	executionKey K,
	memoizedFn func(context.Context) (V, error),
) (TypedOutcome[V], Extra) {
	var convertedFn func(context.Context) (interface{}, error)
	if memoizedFn != nil {
		convertedFn = func(ctx context.Context) (interface{}, error) {
			return memoizedFn(ctx)
		}
	}

	c := extractCache(ctx)

	outcome, extra := c.execute(ctx, executionKey, convertedFn)
	return newTypedOutcome[V](outcome), extra
}

// FindOutcomes returns all Outcome that were memoized under the given
// executionKey type at the time findOutcomes was called. If a promise
// related to this executionKey type is still pending, the function
// will block and wait for it to complete to get its Outcome.
//
// Note: this function can only return all memoized Outcome if the given
// context has been initialized using WithCache.
func FindOutcomes[K comparable, V any](ctx context.Context, executionKey K) map[K]TypedOutcome[V] {
	c := extractCache(ctx)

	promises := c.findPromises(executionKey)
	if promises == nil {
		return nil
	}

	m := make(map[K]TypedOutcome[V], len(promises))
	for key, p := range promises {
		// Check if context was cancelled while we were waiting
		// for the previous promise.
		if ctx.Err() != nil {
			return nil
		}

		// Wait for the result
		m[key.(K)] = newTypedOutcome[V](p.get(ctx))
	}

	return m
}

// TypedOutcome ...
type TypedOutcome[V any] struct {
	Value V
	Err   error
}

func newTypedOutcome[V any](o Outcome) TypedOutcome[V] {
	if casted, ok := o.Value.(V); ok {
		return TypedOutcome[V]{
			Value: casted,
			Err:   o.Err,
		}
	}

	var temp V
	return TypedOutcome[V]{
		Value: temp,
		Err:   o.Err,
	}
}

// ResultOrDefault returns the final result if there's no error or the default
// result if there's an error.
func (o TypedOutcome[V]) ResultOrDefault(defaultResult V) V {
	if o.Err != nil {
		return defaultResult
	}

	return o.Value
}
