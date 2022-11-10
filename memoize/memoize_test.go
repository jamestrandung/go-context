package memoize

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGet(t *testing.T) {
	var s store

	evaled := 0

	p := s.promise(
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

func TestGetPanic(t *testing.T) {
	var s store

	p := s.promise(
		"key", func(context.Context) (interface{}, error) {
			panic("some error")
		},
	)

	assert.NotPanics(
		t, func() {
			result, err := p.get(context.Background())
			assert.Equal(t, nil, result)
			assert.True(t, errors.Is(err, ErrPanicExecutingMemoizedFn))
		},
	)
}

func TestNewPromise(t *testing.T) {
	calls := 0
	f := func(context.Context) (interface{}, error) {
		calls++
		return calls, assert.AnError
	}

	// All calls to Get on the same promise return the same result.
	p1 := newPromise("debug", f)
	expectGet(t, p1, 1, assert.AnError)
	expectGet(t, p1, 1, assert.AnError)

	// A new promise calls the function again.
	p2 := newPromise("debug", f)
	expectGet(t, p2, 2, assert.AnError)
	expectGet(t, p2, 2, assert.AnError)

	// The original promise is unchanged.
	expectGet(t, p1, 1, assert.AnError)
}

func expectGet(t *testing.T, h *promise, wantV interface{}, wantErr error) {
	t.Helper()

	gotV, gotErr := h.get(context.Background())
	if gotV != wantV || gotErr != wantErr {
		t.Fatalf("Get() = %v, %v, wanted %v, %v", gotV, gotErr, wantV, wantErr)
	}
}
