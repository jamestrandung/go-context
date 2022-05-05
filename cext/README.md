# Context Extension

This package offers extra functionality we need from contexts that is not available in the standard context package.

## Functions

### func Detach

```go
func Detach(ctx context.Context) context.Context
```

Detach returns a context that keeps all values of the parent context but detaches from its cancellation and error handling.