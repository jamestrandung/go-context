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

				c := noMemoizeCache{}

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

						outcome, extra := c.execute(context.Background(), "executionKey", memoizedFn)
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
			desc: "has panic",
			test: func(t *testing.T) {
				c := noMemoizeCache{}

				memoizedFn := func(context.Context) (interface{}, error) {
					panic("some error")
				}

				assert.NotPanics(
					t, func() {
						outcome, extra := c.execute(context.Background(), "executionKey", memoizedFn)
						assert.Equal(t, nil, outcome.Value)
						assert.True(t, errors.Is(outcome.Err, ErrPanicExecutingMemoizedFn))
						assert.False(t, extra.IsMemoized)
						assert.True(t, extra.IsExecuted)
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
