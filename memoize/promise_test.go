package memoize

import (
    "context"
    "errors"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestNewPromise(t *testing.T) {
    calls := 0
    f := func(context.Context) (interface{}, error) {
        calls++
        return calls, assert.AnError
    }

    // All calls to Get on the same promise return the same result.
    p1 := newPromise("executionKeyType", context.Background(), f)
    expectGet(t, p1, 1, assert.AnError)
    expectGet(t, p1, 1, assert.AnError)

    // A new promise calls the function again.
    p2 := newPromise("executionKeyType", context.Background(), f)
    expectGet(t, p2, 2, assert.AnError)
    expectGet(t, p2, 2, assert.AnError)

    // The original promise is unchanged.
    expectGet(t, p1, 1, assert.AnError)
}

func TestPromise_Get(t *testing.T) {
    var c cache

    evaled := 0

    p, _ := c.promise(
        "key", func(context.Context) (interface{}, error) {
            evaled++
            return "res", assert.AnError
        },
    )

    expectGet(t, p, "res", assert.AnError)
    expectGet(t, p, "res", assert.AnError)

    if evaled != 1 {
        t.Errorf("got %v calls to function, wanted 1", evaled)
    }
}

func TestPromise_Panic(t *testing.T) {
    var c cache

    p, _ := c.promise(
        "key", func(context.Context) (interface{}, error) {
            panic("some error")
        },
    )

    assert.NotPanics(
        t, func() {
            outcome := p.get(context.Background())
            assert.Equal(t, nil, outcome.Value)
            assert.True(t, errors.Is(outcome.Err, ErrPanicExecutingMemoizedFn))
        },
    )
}

func expectGet(t *testing.T, h *promise, wantV interface{}, wantErr error) {
    t.Helper()

    outcome := h.get(context.Background())
    if outcome.Value != wantV || outcome.Err != wantErr {
        t.Fatalf("Get() = %v, %v, wanted %v, %v", outcome.Value, outcome.Err, wantV, wantErr)
    }
}
