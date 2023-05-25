package memoize

import (
    "context"
    "fmt"
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
    assert.Equal(t, &noMemoizeCache{}, c)

    ctxWithCache, destroyFn := WithCache(ctx)
    defer destroyFn()

    c = extractCache(ctxWithCache)
    assert.Equal(t, reflect.TypeOf((*cache)(nil)), reflect.TypeOf(c))
}

func TestPopulateCache(t *testing.T) {
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

                PopulateCache(
                    ctxWithCache, map[interface{}]Outcome{
                        "executionKey": {
                            Value: 2,
                            Err:   assert.AnError,
                        },
                    },
                )

                var wg sync.WaitGroup
                for i := 0; i < 100; i++ {
                    wg.Add(1)

                    go func() {
                        defer wg.Done()

                        outcome, extra := Execute(ctxWithCache, "executionKey", memoizedFn)
                        assert.Equal(t, 2, outcome.Value)
                        assert.Equal(t, assert.AnError, outcome.Err)
                        assert.True(t, extra.IsMemoized)
                        assert.False(t, extra.IsExecuted)
                    }()
                }

                wg.Wait()

                assert.Equal(t, (int32)(0), evaled, "got %v calls to function, wanted 0", evaled)
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

                PopulateCache(
                    ctxWithoutCache, map[interface{}]Outcome{
                        "executionKey": {
                            Value: 2,
                            Err:   assert.AnError,
                        },
                    },
                )

                var wg sync.WaitGroup
                for i := 0; i < 100; i++ {
                    wg.Add(1)

                    go func() {
                        defer wg.Done()

                        outcome, extra := Execute(ctxWithoutCache, "executionKey", memoizedFn)
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
    }

    for _, scenario := range scenarios {
        sc := scenario

        t.Run(sc.desc, sc.test)
    }
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

                        outcome, extra := Execute(ctxWithCache, "executionKey", memoizedFn)
                        assert.Equal(t, 1, outcome.Value)
                        assert.Equal(t, assert.AnError, outcome.Err)
                        assert.True(t, extra.IsMemoized)
                        assert.True(t, extra.IsExecuted)
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

                        outcome, extra := Execute(ctxWithoutCache, "executionKey", memoizedFn)
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
    }

    for _, scenario := range scenarios {
        sc := scenario

        t.Run(sc.desc, sc.test)
    }
}

func TestFindOutcomes(t *testing.T) {
    scenarios := []struct {
        desc string
        test func(t *testing.T)
    }{
        {
            desc: "context was initialized using WithCache",
            test: func(t *testing.T) {
                var evaled int32 = 0

                ctxWithCache, destroyFn := WithCache(context.Background())
                defer destroyFn()

                var wg sync.WaitGroup
                for i := 0; i < 100; i++ {
                    wg.Add(1)

                    i := i
                    go func() {
                        defer wg.Done()

                        outcome, extra := Execute(
                            ctxWithCache, fmt.Sprintf("key%v", i), func(ctx context.Context) (interface{}, error) {
                                atomic.AddInt32(&evaled, 1)
                                return i, assert.AnError
                            },
                        )

                        assert.Equal(t, i, outcome.Value)
                        assert.Equal(t, assert.AnError, outcome.Err)
                        assert.True(t, extra.IsMemoized)
                        assert.True(t, extra.IsExecuted)
                    }()
                }

                wg.Wait()

                assert.Equal(t, (int32)(100), evaled, "got %v calls to function, wanted 100", evaled)

                outcomes := FindOutcomes[string, int](ctxWithCache, "key")
                assert.Equal(t, 100, len(outcomes))

                for i := 0; i < 100; i++ {
                    expected := TypedOutcome[int]{
                        Value: i,
                        Err:   assert.AnError,
                    }

                    outcome, ok := outcomes[fmt.Sprintf("key%v", i)]
                    assert.True(t, ok)
                    assert.Equal(t, expected, outcome)
                }
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

                PopulateCache(
                    ctxWithoutCache, map[interface{}]Outcome{
                        "executionKey": {
                            Value: 2,
                            Err:   assert.AnError,
                        },
                    },
                )

                var wg sync.WaitGroup
                for i := 0; i < 100; i++ {
                    wg.Add(1)

                    go func() {
                        defer wg.Done()

                        outcome, extra := Execute(ctxWithoutCache, "executionKey", memoizedFn)
                        assert.Equal(t, 1, outcome.Value)
                        assert.Equal(t, assert.AnError, outcome.Err)
                        assert.False(t, extra.IsMemoized)
                        assert.True(t, extra.IsExecuted)
                    }()
                }

                wg.Wait()

                assert.Equal(t, (int32)(100), evaled, "got %v calls to function, wanted 100", evaled)

                outcomes := FindOutcomes[string, int](ctxWithoutCache, "key")
                assert.Equal(t, 0, len(outcomes))
            },
        },
    }

    for _, scenario := range scenarios {
        sc := scenario

        t.Run(sc.desc, sc.test)
    }
}

func TestFindAllOutcomes(t *testing.T) {
    scenarios := []struct {
        desc string
        test func(t *testing.T)
    }{
        {
            desc: "context was initialized using WithCache",
            test: func(t *testing.T) {
                var evaled int32 = 0

                ctxWithCache, destroyFn := WithCache(context.Background())
                defer destroyFn()

                var wg sync.WaitGroup
                for i := 0; i < 100; i++ {
                    wg.Add(1)

                    i := i
                    go func() {
                        defer wg.Done()

                        outcome, extra := Execute(
                            ctxWithCache, fmt.Sprintf("key%v", i), func(ctx context.Context) (interface{}, error) {
                                atomic.AddInt32(&evaled, 1)
                                return i, assert.AnError
                            },
                        )

                        assert.Equal(t, i, outcome.Value)
                        assert.Equal(t, assert.AnError, outcome.Err)
                        assert.True(t, extra.IsMemoized)
                        assert.True(t, extra.IsExecuted)
                    }()
                }

                wg.Wait()

                assert.Equal(t, (int32)(100), evaled, "got %v calls to function, wanted 100", evaled)

                outcomes := FindAllOutcomes(ctxWithCache)
                assert.Equal(t, 100, len(outcomes))

                for i := 0; i < 100; i++ {
                    expected := Outcome{
                        Value: i,
                        Err:   assert.AnError,
                    }

                    outcome, ok := outcomes[fmt.Sprintf("key%v", i)]
                    assert.True(t, ok)
                    assert.Equal(t, expected, outcome)
                }
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

                PopulateCache(
                    ctxWithoutCache, map[interface{}]Outcome{
                        "executionKey": {
                            Value: 2,
                            Err:   assert.AnError,
                        },
                    },
                )

                var wg sync.WaitGroup
                for i := 0; i < 100; i++ {
                    wg.Add(1)

                    go func() {
                        defer wg.Done()

                        outcome, extra := Execute(ctxWithoutCache, "executionKey", memoizedFn)
                        assert.Equal(t, 1, outcome.Value)
                        assert.Equal(t, assert.AnError, outcome.Err)
                        assert.False(t, extra.IsMemoized)
                        assert.True(t, extra.IsExecuted)
                    }()
                }

                wg.Wait()

                assert.Equal(t, (int32)(100), evaled, "got %v calls to function, wanted 100", evaled)

                outcomes := FindAllOutcomes(ctxWithoutCache)
                assert.Equal(t, 0, len(outcomes))
            },
        },
    }

    for _, scenario := range scenarios {
        sc := scenario

        t.Run(sc.desc, sc.test)
    }
}
