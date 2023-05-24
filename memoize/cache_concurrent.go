package memoize

import (
	"context"
	"github.com/mitchellh/hashstructure/v2"
	"sync"
)

const defaultConcurrencyLevel = 10

type concurrentCache []*cache

// newConcurrentCache creates a new concurrentCache.
func newConcurrentCache(rootCtx context.Context, concurrencyLevel int) concurrentCache {
	if concurrencyLevel == 0 {
		concurrencyLevel = defaultConcurrencyLevel
	}

	shards := make([]*cache, concurrencyLevel)

	for i := 0; i < concurrencyLevel; i++ {
		shards[i] = newCache(rootCtx)
	}

	return shards
}

func (c concurrentCache) getShard(executionKey interface{}) *cache {
	return c[c.hashIndex(executionKey)]
}

func (c concurrentCache) hashIndex(executionKey interface{}) uint64 {
	return hashAny(executionKey) % uint64(len(c))
}

func (c concurrentCache) destroy() {
	for _, shard := range c {
		shard.destroy()
	}
}

func (c concurrentCache) take(entries map[interface{}]Outcome) {
	shardEntries := make([]map[interface{}]Outcome, len(c))

	for k, v := range entries {
		hashIdx := c.hashIndex(k)

		m := func() map[interface{}]Outcome {
			if curEntries := shardEntries[hashIdx]; curEntries != nil {
				return curEntries
			}

			newEntries := make(map[interface{}]Outcome)
			shardEntries[hashIdx] = newEntries

			return newEntries
		}()

		m[k] = v
	}

	var wg sync.WaitGroup
	for idx, shard := range c {
		toTakeEntries := shardEntries[idx]
		if len(toTakeEntries) == 0 {
			continue
		}

		wg.Add(1)
		go func(shard *cache) {
			defer wg.Done()

			shard.take(toTakeEntries)
		}(shard)
	}

	wg.Wait()
}

func (c concurrentCache) execute(
	ctx context.Context,
	executionKey interface{},
	memoizedFn Function,
) (Outcome, Extra) {
	shard := c.getShard(executionKey)
	return shard.execute(ctx, executionKey, memoizedFn)
}

func (c concurrentCache) findPromises(executionKey interface{}) map[interface{}]*promise {
	m := make(map[interface{}]*promise)

	for _, shard := range c {
		promises := shard.findPromises(executionKey)

		for key, p := range promises {
			m[key] = p
		}
	}

	return m
}

var hashFn = hashstructure.Hash

func hashAny(key interface{}) uint64 {
	defer func() {
		// Fall back to 0 if panics
		recover()
	}()

	hash, err := hashFn(key, hashstructure.FormatV2, &hashstructure.HashOptions{UseStringer: true})
	if err != nil {
		// Use the 1st shard as fallback in case hashing fails
		return 0
	}

	return hash
}
