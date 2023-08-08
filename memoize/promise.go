package memoize

import (
	"context"
	"fmt"
	"runtime/trace"
	"sync/atomic"

	"github.com/jamestrandung/go-context/cext"
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

// State represents the state enumeration for a promise.
type State byte

// Various states.
const (
	IsCreated   State = iota // IsCreated represents a newly created promise
	IsExecuted               // IsExecuted represents a promise which was executed
	IsPopulated              // IsPopulated represents a completed promise carrying populated outcome
)

// A promise represents the future result of a call to a function.
type promise struct {
	executionKeyType string

	// the rootCtx that was used to initialize a cache and would provide
	// the cancelling signal for our execution.
	rootCtx context.Context
	// state is the current memoize.State of this promise.
	state int32
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
	done := make(chan struct{})
	close(done)

	return &promise{
		executionKeyType: debug,
		state:            int32(IsPopulated),
		done:             done,
		outcome:          outcome,
	}
}

// isExecuted returns whether this promise was actually
// executed or the result was pre-populated.
func (p *promise) isExecuted() bool {
	return p.state == int32(IsExecuted)
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

	if p.changeState(IsCreated, IsExecuted) {
		return p.run(ctx)
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
				close(p.done)
			},
		)
	}()

	return p.wait(ctx)
}

// wait waits for the value to be computed, or ctx to be cancelled.
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

func (p *promise) changeState(from, to State) bool {
	return atomic.CompareAndSwapInt32(&p.state, int32(from), int32(to))
}
