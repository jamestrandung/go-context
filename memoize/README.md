# Memoize

## Why context?

As a matter of fact, it's not always possible to cache some data across API calls. At times, there's a need to memoize 
the result of a function call within the boundary of only one single API call. For example, let's say we call a Map 
service to get the distance to travel from point A to point B. Across different API calls at different points in time, 
the resulting distance might be different due to changing traffic conditions. Hence, it's only feasible to cache the 
result within a single API call so that all logic in this boundary uses the same distance value.

In this case, context is a perfect candidate for storing such request-level data.

## How to use

At the beginning of your API handling logic, you must initialize a cache using the following function.

```go
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
func WithCache(ctx context.Context) (context.Context, DestroyFn)
```

Subsequently, you can pass the context you got back from the above function down to lower-level code. Whenever there's
a need to memoize function calls, you just need to execute those functions using the provided function below.

```go
// Execute guarantees that the given memoizedFn will be invoked only
// once regardless of how many times Execute gets called with the same
// executionKey. All callers will receive the same result and error as
// the result of this call.
//
// Note 1: this promise can only be kept if the given context has been
// initialized using WithCache before calling Execute. The last return
// value indicates whether this function call was memoized or not.
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
func Execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn func(context.Context) (interface{}, error),
) (result interface{}, err error, isMemoized bool)
```