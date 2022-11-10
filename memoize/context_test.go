package memoize

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

func TestWithCache(t *testing.T) {
	ctxWithCache, destroyFn := WithCache(context.Background())
	defer destroyFn()

	actual := ctxWithCache.Value(memoizeStoreKey)
	assert.Equal(t, reflect.TypeOf((*cache)(nil)), reflect.TypeOf(actual))
}

func TestExtractCache(t *testing.T) {
	ctx := context.Background()

	c := extractCache(ctx)
	assert.Equal(t, noMemoizeCache{}, c)

	ctxWithCache, destroyFn := WithCache(ctx)
	defer destroyFn()

	c = extractCache(ctxWithCache)
	assert.Equal(t, reflect.TypeOf((*cache)(nil)), reflect.TypeOf(c))
}

func TestExecute(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(t *testing.T)
	}{
		{
			desc: "context was initialized using WithCache",
			test: func(t *testing.T) {
				var evaled int32 = 0

				memoizedFn := func(context.Context) (interface{}, error) {
					atomic.AddInt32(&evaled, 1)
					return 1, assert.AnError
				}

				ctxWithCache, destroyFn := WithCache(context.Background())
				defer destroyFn()

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := Execute(ctxWithCache, "executionKey", memoizedFn)
						assert.Equal(t, 1, result)
						assert.Equal(t, assert.AnError, err)
						assert.True(t, isMemoized)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(1), evaled, "got %v calls to function, wanted 1", evaled)
			},
		},
		{
			desc: "context was NOT initialized using WithCache",
			test: func(t *testing.T) {
				var evaled int32 = 0

				memoizedFn := func(context.Context) (interface{}, error) {
					atomic.AddInt32(&evaled, 1)
					return 1, assert.AnError
				}

				ctxWithoutCache := context.Background()

				var wg sync.WaitGroup
				for i := 0; i < 100; i++ {
					wg.Add(1)

					go func() {
						defer wg.Done()

						result, err, isMemoized := Execute(ctxWithoutCache, "executionKey", memoizedFn)
						assert.Equal(t, 1, result)
						assert.Equal(t, assert.AnError, err)
						assert.False(t, isMemoized)
					}()
				}

				wg.Wait()

				assert.Equal(t, (int32)(100), evaled, "got %v calls to function, wanted 100", evaled)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}
