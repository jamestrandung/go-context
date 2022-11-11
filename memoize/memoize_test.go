package memoize

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewPromise(t *testing.T) {
	calls := 0
	f := func(context.Context) (interface{}, error) {
		calls++
		return calls, assert.AnError
	}

	// All calls to Get on the same promise return the same result.
	p1 := newPromise("debug", context.Background(), f)
	expectGet(t, p1, 1, assert.AnError)
	expectGet(t, p1, 1, assert.AnError)

	// A new promise calls the function again.
	p2 := newPromise("debug", context.Background(), f)
	expectGet(t, p2, 2, assert.AnError)
	expectGet(t, p2, 2, assert.AnError)

	// The original promise is unchanged.
	expectGet(t, p1, 1, assert.AnError)
}

func TestPromise_Get(t *testing.T) {
	var c cache

	evaled := 0

	p := c.promise(
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

	p := c.promise(
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

func expectGet(t *testing.T, h *promise, wantV interface{}, wantErr error) {
	t.Helper()

	gotV, gotErr := h.get(context.Background())
	if gotV != wantV || gotErr != wantErr {
		t.Fatalf("Get() = %v, %v, wanted %v, %v", gotV, gotErr, wantV, wantErr)
	}
}

func TestCache_Destroy(t *testing.T) {
	c := newCache(context.Background())

	assert.Equal(t, (int32)(0), c.isDestroyed)
	assert.NotNil(t, c.promises)

	c.destroy()

	assert.Equal(t, (int32)(1), c.isDestroyed)
	assert.Nil(t, c.promises)
}

func TestCache_Execute(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(t *testing.T)
	}{
		{
			desc: "nil executionKey",
			test: func(t *testing.T) {
				var evaled int32 = 0

				memoizedFn := func(context.Context) (interface{}, error) {
					atomic.AddInt32(&evaled, 1)
					return 1, assert.AnError
				}

				c := newCache(context.Background())

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := c.execute(context.Background(), nil, memoizedFn)
						assert.Equal(t, 1, result)
						assert.Equal(t, assert.AnError, err)
						assert.False(t, isMemoized)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(100), evaled, "got %v calls to function, wanted 100", evaled)
			},
		},
		{
			desc: "nil memoizedFn",
			test: func(t *testing.T) {
				var evaled int32 = 0

				c := newCache(context.Background())

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := c.execute(context.Background(), "executionKey", nil)
						assert.Equal(t, nil, result)
						assert.Equal(t, ErrMemoizedFnCannotBeNil, err)
						assert.False(t, isMemoized)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(0), evaled, "got %v calls to function, wanted 0", evaled)
			},
		},
		{
			desc: "cache was destroyed",
			test: func(t *testing.T) {
				var evaled int32 = 0

				memoizedFn := func(context.Context) (interface{}, error) {
					atomic.AddInt32(&evaled, 1)
					return 1, assert.AnError
				}

				c := newCache(context.Background())
				c.destroy()

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := c.execute(context.Background(), "executionKey", memoizedFn)
						assert.Equal(t, nil, result)
						assert.Equal(t, ErrCacheAlreadyDestroyed, err)
						assert.False(t, isMemoized)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(0), evaled, "got %v calls to function, wanted 0", evaled)
			},
		},
		{
			desc: "happy path",
			test: func(t *testing.T) {
				var evaled int32 = 0

				memoizedFn := func(context.Context) (interface{}, error) {
					atomic.AddInt32(&evaled, 1)
					return 1, assert.AnError
				}

				c := newCache(context.Background())

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := c.execute(context.Background(), "executionKey", memoizedFn)
						assert.Equal(t, 1, result)
						assert.Equal(t, assert.AnError, err)
						assert.True(t, isMemoized)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(1), evaled, "got %v calls to function, wanted 1", evaled)

				c.destroy()

				result, err, isMemoized := c.execute(context.Background(), "executionKey", memoizedFn)
				assert.Equal(t, nil, result)
				assert.Equal(t, ErrCacheAlreadyDestroyed, err)
				assert.False(t, isMemoized)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}
