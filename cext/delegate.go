package cext

import (
    "context"
    "fmt"
    "time"
)

// Delegate returns a context that keeps all values of the valueCtx while
// taking its cancellation signal and error from the cancelCtx.
func Delegate(cancelCtx context.Context, valueCtx context.Context) context.Context {
    return &delegatingContext{
        cancelCtx: cancelCtx,
        valueCtx:  valueCtx,
    }
}

type delegatingContext struct {
    cancelCtx context.Context
    valueCtx  context.Context
}

// Deadline ...
func (c *delegatingContext) Deadline() (deadline time.Time, ok bool) {
    return c.cancelCtx.Deadline()
}

// Done ...
func (c *delegatingContext) Done() <-chan struct{} {
    return c.cancelCtx.Done()
}

// Err ...
func (c *delegatingContext) Err() error {
    return c.cancelCtx.Err()
}

// Value ...
func (c *delegatingContext) Value(key interface{}) interface{} {
    return c.valueCtx.Value(key)
}

// String ...
func (c *delegatingContext) String() string {
    return fmt.Sprintf("delegating context from cancelCtx %v and valueCtx %v", c.cancelCtx, c.valueCtx)
}
