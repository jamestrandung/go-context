package memoize

import (
	"context"
	"fmt"
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

func TestCache_Destroy(t *testing.T) {
	c := newCache(context.Background())

	assert.False(t, c.isDestroyed)
	assert.NotNil(t, c.promises)

	c.destroy()

	assert.True(t, c.isDestroyed)
	assert.Nil(t, c.promises)
}

func TestCache_PopulateCache(t *testing.T) {
	var c cache

	assert.Empty(t, c.promises)

	c.take(
		map[interface{}]Outcome{
			"key1": {
				Value: 1,
				Err:   assert.AnError,
			},
			"key2": {
				Value: 2,
				Err:   assert.AnError,
			},
		},
	)

	assert.Equal(t, 2, len(c.promises))

	p1, _ := c.promise(
		"key1", func(ctx context.Context) (interface{}, error) {
			return 3, assert.AnError
		},
	)

	// Should get back result from populated entries
	outcome := p1.get(context.Background())
	assert.Equal(t, 1, outcome.Value)
	assert.Equal(t, assert.AnError, outcome.Err)

	p2, _ := c.promise(
		"key2", func(ctx context.Context) (interface{}, error) {
			return 3, assert.AnError
		},
	)

	// Should get back result from populated entries
	outcome = p2.get(context.Background())
	assert.Equal(t, 2, outcome.Value)
	assert.Equal(t, assert.AnError, outcome.Err)

	c.destroy()

	assert.Empty(t, c.promises)

	c.take(
		map[interface{}]Outcome{
			"key1": {
				Value: 1,
				Err:   assert.AnError,
			},
			"key2": {
				Value: 2,
				Err:   assert.AnError,
			},
		},
	)

	assert.Empty(t, c.promises, "populating a destroyed cache must be a no-op")
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

						outcome, extra := c.execute(context.Background(), nil, memoizedFn)
						assert.Equal(t, 1, outcome.Value)
						assert.Equal(t, assert.AnError, outcome.Err)
						assert.False(t, extra.IsMemoized)
						assert.True(t, extra.IsExecuted)
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

						outcome, extra := c.execute(context.Background(), "executionKey", nil)
						assert.Equal(t, nil, outcome.Value)
						assert.Equal(t, ErrMemoizedFnCannotBeNil, outcome.Err)
						assert.False(t, extra.IsMemoized)
						assert.False(t, extra.IsExecuted)
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

						outcome, extra := c.execute(context.Background(), "executionKey", memoizedFn)
						assert.Equal(t, nil, outcome.Value)
						assert.Equal(t, ErrCacheAlreadyDestroyed, outcome.Err)
						assert.False(t, extra.IsMemoized)
						assert.False(t, extra.IsExecuted)
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

						outcome, extra := c.execute(context.Background(), "executionKey", memoizedFn)
						assert.Equal(t, 1, outcome.Value)
						assert.Equal(t, assert.AnError, outcome.Err)
						assert.True(t, extra.IsMemoized)
						assert.True(t, extra.IsExecuted)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(1), evaled, "got %v calls to function, wanted 1", evaled)

				c.destroy()

				outcome, extra := c.execute(context.Background(), "executionKey", memoizedFn)
				assert.Equal(t, nil, outcome.Value)
				assert.Equal(t, ErrCacheAlreadyDestroyed, outcome.Err)
				assert.False(t, extra.IsMemoized)
				assert.False(t, extra.IsExecuted)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}

func TestCache_FindPromises(t *testing.T) {
	var c cache

	for i := 0; i < 100; i++ {
		i := i
		c.promise(
			fmt.Sprintf("key%v", i), func(ctx context.Context) (interface{}, error) {
				return i, assert.AnError
			},
		)
	}

	promises := c.findPromises("key")
	assert.Equal(t, 100, len(promises))

	for i := 0; i < 100; i++ {
		p, ok := promises[fmt.Sprintf("key%v", i)]
		assert.True(t, ok)
		assert.Equal(t, "string", p.executionKeyType)
	}

	c.destroy()

	promises = c.findPromises("key")
	assert.Equal(t, 0, len(promises), "no promises should come from a destroyed cache")
}
