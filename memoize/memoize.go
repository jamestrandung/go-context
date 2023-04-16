package memoize

import (
    "context"
    "fmt"
    "github.com/jamestrandung/go-context/cext"
    "github.com/jamestrandung/go-context/helper"
    "github.com/pkg/errors"
    "reflect"
    "runtime/debug"
    "runtime/trace"
    "sync"
    "sync/atomic"
)

// Function is the type of function that can be memoized.
//
// The argument must not materially affect the result of the function in ways
// that are not captured by the promise's key, since if promise.get is called
// twice concurrently, with the same (implicit) key but different arguments,
// the Function is called once and only once but its result must be suitable
// for both callers.
type Function func(ctx context.Context) (interface{}, error)

// Outcome is the outcome of executing a memoized function.
type Outcome struct {
    Value interface{}
    Err   error
}

// Extra includes additional details about the returned outcome.
type Extra struct {
    // IsMemoized indicates if the outcome was memoized.
    IsMemoized bool
    // IsExecuted indicates if the outcome came from actual execution or
    // was pre-populated in the cache.
    IsExecuted bool
}

// A promise represents the future result of a call to a function.
type promise struct {
    executionKeyType string

    // the rootCtx that was used to initialize a cache and would provide
    // the cancelling signal for our execution
    rootCtx context.Context
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
    outcome Outcome
}

// newPromise returns a promise for the future result of calling the
// specified function.
//
// The executionKeyType string is used to classify promises in logs
// and metrics. It should be drawn from a small set.
func newPromise(executionKeyType string, rootCtx context.Context, function Function) *promise {
    if function == nil {
        panic("nil function")
    }

    return &promise{
        executionKeyType: executionKeyType,
        rootCtx:          rootCtx,
        done:             make(chan struct{}),
        function:         function,
    }
}

// completedPromise returns a promise that has already completed with
// the given Outcome.
func completedPromise(debug string, outcome Outcome) *promise {
    return &promise{
        executionKeyType: debug,
        finished:         true,
        outcome:          outcome,
    }
}

// isExecuted returns whether this promise has resolved and if it was
// actually executed or the result was pre-populated.
func (p *promise) isExecuted() bool {
    return p.finished && p.executed == 1
}

// get returns the value associated with a promise.
//
// All calls to promise.get on a given promise return the same result
// but the function is called (to completion) at most once.
//
// - If the underlying function has not been invoked, it will be.
// - If ctx is cancelled, get returns (nil, context.Canceled).
func (p *promise) get(ctx context.Context) Outcome {
    if ctx.Err() != nil {
        return Outcome{
            Value: nil,
            Err:   ctx.Err(),
        }
    }

    if !p.finished && atomic.CompareAndSwapInt32(&p.executed, 0, 1) {
        return p.run(ctx)
    }

    if p.finished {
        return p.outcome
    }

    return p.wait(ctx)
}

// run starts p.function and returns the result.
func (p *promise) run(ctx context.Context) Outcome {
    // To prevent one child goroutines from cancelling the execution of the memoized
    // function that is still meaningful to other goroutines, we will delegate the
    // value retrieving responsibility to the input context while letting the root
    // context handle cancelling signals.
    //
    // This makes sense because the root context that was used to initialize a cache
    // should be the parent of all child contexts, including the input context. If
    // the root context get cancelled, all child contexts must be cancelled as well.
    delegatingCtx := cext.Delegate(p.rootCtx, ctx)

    go func() {
        trace.WithRegion(
            delegatingCtx, fmt.Sprintf("promise.run %s", p.executionKeyType), func() {
                v, err := doExecute(delegatingCtx, p.function)

                p.outcome = Outcome{
                    Value: v,
                    Err:   err,
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
func (p *promise) wait(ctx context.Context) Outcome {
    select {
    case <-p.done:
        return p.outcome

    case <-ctx.Done():
        return Outcome{
            Value: nil,
            Err:   ctx.Err(),
        }
    }
}

// A cache maps arbitrary keys to promises.
type cache struct {
    rootCtx     context.Context
    isDestroyed bool
    promisesMu  sync.Mutex
    promises    map[interface{}]*promise
}

// newCache creates a new cache.
func newCache(rootCtx context.Context) *cache {
    return &cache{
        rootCtx:  rootCtx,
        promises: make(map[interface{}]*promise),
    }
}

// destroy clears existing items in this cache and mark it as destroyed.
// Subsequent calls to Execute will return ErrCacheAlreadyDestroyed.
func (c *cache) destroy() {
    c.promisesMu.Lock()
    defer c.promisesMu.Unlock()

    c.isDestroyed = true
    c.promises = nil
}

func (c *cache) take(entries map[interface{}]Outcome) {
    c.promisesMu.Lock()
    defer c.promisesMu.Unlock()

    if c.isDestroyed {
        return
    }

    if c.promises == nil {
        c.promises = make(map[interface{}]*promise)
    }

    for executionKey, outcome := range entries {
        if executionKey == nil {
            continue
        }

        p := completedPromise(c.extractExecutionKeyType(executionKey), outcome)
        c.promises[executionKey] = p
    }
}

func (c *cache) execute(
    ctx context.Context,
    executionKey interface{},
    memoizedFn Function,
) (Outcome, Extra) {
    if memoizedFn == nil {
        return Outcome{
                Value: nil,
                Err:   ErrMemoizedFnCannotBeNil,
            }, Extra{
                IsMemoized: false,
                IsExecuted: false,
            }
    }

    if !helper.IsComparable(executionKey) {
        result, err := doExecute(ctx, memoizedFn)
        return Outcome{
                Value: result,
                Err:   err,
            }, Extra{
                IsMemoized: false,
                IsExecuted: true,
            }
    }

    p, err := c.promise(executionKey, memoizedFn)
    if err != nil {
        return Outcome{
                Value: nil,
                Err:   err,
            }, Extra{
                IsMemoized: false,
                IsExecuted: false,
            }
    }

    return p.get(ctx), Extra{
        IsMemoized: true,
        IsExecuted: p.isExecuted(),
    }
}

// promise returns a promise for the future result of calling the given function.
// Calls to promise with the same key return the same promise.
func (c *cache) promise(executionKey interface{}, function Function) (*promise, error) {
    c.promisesMu.Lock()
    defer c.promisesMu.Unlock()

    if c.isDestroyed {
        return nil, ErrCacheAlreadyDestroyed
    }

    p, ok := c.promises[executionKey]
    if !ok {
        return c.createPromise(executionKey, function), nil
    }

    return p, nil
}

func (c *cache) createPromise(executionKey interface{}, function Function) *promise {
    p := newPromise(c.extractExecutionKeyType(executionKey), c.rootCtx, function)
    if c.promises == nil {
        c.promises = make(map[interface{}]*promise)
    }

    c.promises[executionKey] = p

    return p
}

func (c *cache) findPromises(executionKey interface{}) map[interface{}]*promise {
    returnAll := false
    if executionKey == nil {
        returnAll = true
    }

    c.promisesMu.Lock()
    defer c.promisesMu.Unlock()

    if c.isDestroyed {
        return nil
    }

    executionKeyType := func() string {
        if returnAll {
            return ""
        }

        return c.extractExecutionKeyType(executionKey)
    }()

    m := make(map[interface{}]*promise)
    for key, p := range c.promises {
        if !returnAll && p.executionKeyType != executionKeyType {
            continue
        }

        m[key] = p
    }

    return m
}

func (c *cache) extractExecutionKeyType(executionKey interface{}) string {
    return reflect.TypeOf(executionKey).String()
}

func doExecute(ctx context.Context, memoizedFn Function) (result interface{}, err error) {
    // Convert panics into standard errors for clients to handle gracefully
    defer func() {
        if r := recover(); r != nil {
            result = nil

            stackTrace := func() string {
                stack := debug.Stack()
                if len(stack) == 0 {
                    return ""
                }

                return string(stack)
            }()

            err = errors.Wrap(ErrPanicExecutingMemoizedFn, fmt.Sprintf("%v \n %v", r, stackTrace))
        }
    }()

    result, err = memoizedFn(ctx)
    return
}
