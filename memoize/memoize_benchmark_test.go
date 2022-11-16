package memoize

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func BenchmarkPromise_Get(b *testing.B) {
	p := newPromise(
		"executionKeyType", context.Background(), func(context.Context) (interface{}, error) {
			return "res", assert.AnError
		},
	)

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			p.get(context.Background())
		}()
	}

	wg.Wait()
}

func BenchmarkStore_Get(b *testing.B) {
	var c cache

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			p, _ := c.promise(
				"key", func(context.Context) (interface{}, error) {
					return "res", assert.AnError
				},
			)

			p.get(context.Background())
		}()
	}

	wg.Wait()
}

func BenchmarkStore_Promise(b *testing.B) {
	var c cache

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			c.promise(
				"key", func(ctx context.Context) (interface{}, error) {
					return 1, assert.AnError
				},
			)
		}()
	}

	wg.Wait()
}
