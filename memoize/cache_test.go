package memoize

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNoMemoizeCache_Execute(t *testing.T) {
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

				c := noMemoizeCache{}

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

				c := noMemoizeCache{}

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
			desc: "no panic",
			test: func(t *testing.T) {
				var evaled int32 = 0

				memoizedFn := func(context.Context) (interface{}, error) {
					atomic.AddInt32(&evaled, 1)
					return 1, assert.AnError
				}

				c := noMemoizeCache{}

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := c.execute(context.Background(), "executionKey", memoizedFn)
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
			desc: "has panic",
			test: func(t *testing.T) {
				c := noMemoizeCache{}

				memoizedFn := func(context.Context) (interface{}, error) {
					panic("some error")
				}

				assert.NotPanics(
					t, func() {
						result, err, isMemoized := c.execute(context.Background(), "executionKey", memoizedFn)
						assert.Equal(t, nil, result)
						assert.True(t, errors.Is(err, ErrPanicExecutingMemoizedFn))
						assert.False(t, isMemoized)
					},
				)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}

func TestCache_Destroy(t *testing.T) {
	c := newCache()

	assert.Equal(t, (int32)(0), c.isDestroyed)
	assert.NotNil(t, c.store)

	c.destroy()

	assert.Equal(t, (int32)(1), c.isDestroyed)
	assert.Nil(t, c.store)
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

				c := newCache()

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

				c := newCache()

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

				c := newCache()
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

				c := newCache()

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
