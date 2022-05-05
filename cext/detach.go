package cext

import (
    "context"
    "fmt"
    "time"
)

// Detach returns a context that keeps all values of the parent context
// but detaches from its cancellation and error handling.
func Detach(ctx context.Context) context.Context {
    return &detachedContext{
        ctx,
    }
}

type detachedContext struct {
    parent context.Context
}

// Deadline ...
func (c *detachedContext) Deadline() (deadline time.Time, ok bool) {
    return
}

// Done ...
func (c *detachedContext) Done() <-chan struct{} {
    return nil
}

// Err ...
func (c *detachedContext) Err() error {
    return nil
}

// Value ...
func (c *detachedContext) Value(key interface{}) interface{} {
    return c.parent.Value(key)
}

// String ...
func (c *detachedContext) String() string {
    return fmt.Sprintf("detached context from %v", c.parent)
}
