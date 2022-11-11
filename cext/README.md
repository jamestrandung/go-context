# Context Extension

This package offers extra functionality we need from contexts that is not available in the standard context package.

## Functions

### func Detach

```go
// Detach returns a context that keeps all values of the parent context
// but detaches from its cancellation and error handling.
func Detach(ctx context.Context) context.Context
```

This function comes in handy when some code needs to follow through to completion instead of getting cancelled halfway
by the parent context (e.g. cancelled by)

### func Delegate

```go
// Delegate returns a context that keeps all values of the valueCtx while
// taking its cancellation signal and error from the cancelCtx.
func Delegate(cancelCtx context.Context, valueCtx context.Context) context.Context
```

The standard Context in Go is meant for 2 distinct purposes: carrying request-level data and looking out for cancelling
signals. In cases where we want to delegate these responsibilities to 2 different Contexts, this function will get the
job done.