package memoize

import (
    "context"
    "fmt"
    "github.com/jamestrandung/go-context/cext"
    "github.com/pkg/errors"
    "reflect"
    "runtime/debug"
    "runtime/trace"
    "sync"
    "sync/atomic"
)

// Function is the type of function that can be memoized.
//
// The argument must not materially affect the result of the function in
// ways that are not captured by the promise's key, since if promise.get
// is called twice concurrently, with the same (implicit) key but different
// arguments, the Function is called only once but its result must be
// suitable for both callers.
type Function func(ctx context.Context) (interface{}, error)

type outcome struct {
    value interface{}
    err   error
}

// A promise represents the future result of a call to a function.
type promise struct {
    debug string // for observability

    // executed is set when execution starts so that the memoized function
    // does not get executed more than once.
    executed int32
    // finished is set when execution completes so that future callers can
    // use outcome immediately instead of waiting on done.
    finished bool
    // done is closed when execution completes to unblock concurrent waiters.
    done chan struct{}
    // the function that will be used to populate the outcome.
    function Function
    // outcome is set when execution completes.
    outcome outcome
}

// newPromise returns a promise for the future result of calling the
// specified function.
//
// The debug string is used to classify promises in logs and metrics.
// It should be drawn from a small set.
func newPromise(debug string, function Function) *promise {
    if function == nil {
        panic("nil function")
    }

    return &promise{
        debug:    debug,
        done:     make(chan struct{}),
        function: function,
    }
}

// get returns the value associated with a promise.
//
// All calls to promise.get on a given promise return the same result
// but the function is called (to completion) at most once.
//
// - If the underlying function has not been invoked, it will be.
// - If ctx is cancelled, get returns (nil, context.Canceled).
func (p *promise) get(ctx context.Context) (interface{}, error) {
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }

    if atomic.CompareAndSwapInt32(&p.executed, 0, 1) {
        return p.run(ctx)
    }

    if p.finished {
        return p.outcome.value, p.outcome.err
    }

    return p.wait(ctx)
}

// run starts p.function and returns the result.
func (p *promise) run(ctx context.Context) (interface{}, error) {
    detachedCtx := cext.Detach(ctx)

    go func() {
        trace.WithRegion(
            detachedCtx, fmt.Sprintf("promise.run %s", p.debug), func() {
                v, err := execute(detachedCtx, p.function)

                p.outcome = outcome{
                    value: v,
                    err:   err,
                }
                p.function = nil // aid GC
                p.finished = true
                close(p.done)
            },
        )
    }()

    return p.wait(ctx)
}

// wait waits for the value to be computed, or ctx to be cancelled. p.mu must be locked.
func (p *promise) wait(ctx context.Context) (interface{}, error) {
    select {
    case <-p.done:
        return p.outcome.value, p.outcome.err

    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

// A store maps arbitrary keys to reference-counted promises.
type store struct {
    promisesMu sync.Mutex
    promises   map[interface{}]*promise
}

// newStore creates a new store with the given eviction policy.
func newStore() *store {
    return &store{
        promises: make(map[interface{}]*promise),
    }
}

// promise returns a promise for the future result of calling the given function.
// Calls to promise with the same key return the same promise.
func (s *store) promise(executionKey interface{}, function Function) *promise {
    s.promisesMu.Lock()
    defer s.promisesMu.Unlock()

    p, ok := s.promises[executionKey]
    if !ok {
        p = s.createPromise(executionKey, function)
    }

    return p
}

func (s *store) createPromise(executionKey interface{}, function Function) *promise {
    p := newPromise(reflect.TypeOf(executionKey).String(), function)
    if s.promises == nil {
        s.promises = make(map[interface{}]*promise)
    }

    s.promises[executionKey] = p

    return p
}

func execute(ctx context.Context, memoizedFn Function) (result interface{}, err error) {
    // Convert panics into standard errors for clients to handle gracefully
    defer func() {
        if r := recover(); r != nil {
            result = nil
            err = errors.Wrap(ErrPanicExecutingMemoizedFn, fmt.Sprintf("%v \n %v", r, debug.Stack()))
        }
    }()

    result, err = memoizedFn(ctx)
    return
}
