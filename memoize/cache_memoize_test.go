package memoize

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

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

    intPromise, _ := c.promise(
        101, func(ctx context.Context) (interface{}, error) {
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
    assert.Equal(t, intPromise, p)

    c.destroy()

    promises = c.findPromises("key")
    assert.Equal(t, 0, len(promises), "no promises should come from a destroyed cache")
}
