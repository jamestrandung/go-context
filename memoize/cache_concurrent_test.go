package memoize

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentCache_Destroy(t *testing.T) {
	c := newConcurrentCache(context.Background(), 10)

	for _, shard := range c {
		assert.False(t, shard.isDestroyed)
		assert.NotNil(t, shard.promises)
	}

	c.destroy()

	for _, shard := range c {
		assert.True(t, shard.isDestroyed)
		assert.Nil(t, shard.promises)
	}
}

func TestConcurrentCache_PopulateCache(t *testing.T) {
	c := newConcurrentCache(context.Background(), 10)

	for _, shard := range c {
		assert.Empty(t, shard.promises)
	}

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

	promiseCount := 0
	for _, shard := range c {
		promiseCount += len(shard.promises)
	}

	assert.Equal(t, 2, promiseCount)

	outcome, extra := c.execute(
		context.Background(), "key1", func(ctx context.Context) (interface{}, error) {
			return 3, assert.AnError
		},
	)

	// Should get back result from populated entries
	assert.Equal(t, 1, outcome.Value)
	assert.Equal(t, assert.AnError, outcome.Err)
	assert.True(t, extra.IsMemoized)
	assert.False(t, extra.IsExecuted)

	outcome, extra = c.execute(
		context.Background(), "key2", func(ctx context.Context) (interface{}, error) {
			return 3, assert.AnError
		},
	)

	// Should get back result from populated entries
	assert.Equal(t, 2, outcome.Value)
	assert.Equal(t, assert.AnError, outcome.Err)
	assert.True(t, extra.IsMemoized)
	assert.False(t, extra.IsExecuted)

	c.destroy()

	for _, shard := range c {
		assert.True(t, shard.isDestroyed)
		assert.Nil(t, shard.promises)
	}

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

	for _, shard := range c {
		assert.Empty(t, shard.promises, "populating a destroyed cache must be a no-op")
	}
}

func TestConcurrentCache_Execute(t *testing.T) {
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

				c := newConcurrentCache(context.Background(), 10)

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

				c := newConcurrentCache(context.Background(), 10)

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

				c := newConcurrentCache(context.Background(), 10)
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

				c := newConcurrentCache(context.Background(), 10)

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

func TestConcurrentCache_FindPromises(t *testing.T) {
	c := newConcurrentCache(context.Background(), 10)

	for i := 0; i < 100; i++ {
		i := i
		c.execute(
			context.Background(), fmt.Sprintf("key%v", i), func(ctx context.Context) (interface{}, error) {
				return i, assert.AnError
			},
		)
	}

	c.execute(
		context.Background(), 101, func(ctx context.Context) (interface{}, error) {
			return 101, assert.AnError
		},
	)

	promises := c.findPromises("key")
	assert.Equal(t, 100, len(promises))

	for i := 0; i < 100; i++ {
		p, ok := promises[fmt.Sprintf("key%v", i)]
		assert.True(t, ok)
		assert.Equal(t, "string", p.executionKeyType)
	}

	// should get ALL promises when key is `nil`
	promises = c.findPromises(nil)
	assert.Equal(t, 101, len(promises))

	for i := 0; i < 100; i++ {
		p, ok := promises[fmt.Sprintf("key%v", i)]
		assert.True(t, ok)
		assert.Equal(t, "string", p.executionKeyType)
	}

	p, ok := promises[101]
	assert.True(t, ok)
	assert.Equal(t, "int", p.executionKeyType)

	c.destroy()

	promises = c.findPromises("key")
	assert.Equal(t, 0, len(promises), "no promises should come from a destroyed cache")
}
