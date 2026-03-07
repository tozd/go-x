package x_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

func TestNewLRUCache(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[string, int](10)
	require.NoError(t, errE, "% -+#.1v", errE)
	assert.NotNil(t, cache)
}

func TestNewLRUCacheInvalidSize(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[string, int](0)
	assert.Error(t, errE)
	assert.Nil(t, cache)
}

func TestLRUCacheGetMiss(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[string, int](10)
	require.NoError(t, errE, "% -+#.1v", errE)

	val, ok := cache.Get("missing")
	assert.False(t, ok)
	assert.Equal(t, 0, val)
	assert.Equal(t, uint64(1), cache.MissCount())
}

func TestLRUCacheGetHit(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[string, int](10)
	require.NoError(t, errE, "% -+#.1v", errE)

	cache.Add("key", 42)

	val, ok := cache.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 42, val)
	assert.Equal(t, uint64(0), cache.MissCount())
}

func TestLRUCacheMissCountResets(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[string, int](10)
	require.NoError(t, errE, "% -+#.1v", errE)

	cache.Get("a")
	cache.Get("b")
	cache.Get("c")

	assert.Equal(t, uint64(3), cache.MissCount())
	// MissCount resets after reading.
	assert.Equal(t, uint64(0), cache.MissCount())
}

func TestLRUCacheMissCountMixedHitsAndMisses(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[string, int](10)
	require.NoError(t, errE, "% -+#.1v", errE)

	cache.Add("x", 1)
	cache.Get("x") // Hit.
	cache.Get("y") // Miss.
	cache.Get("z") // Miss.
	cache.Get("x") // Hit.

	assert.Equal(t, uint64(2), cache.MissCount())
}

func TestLRUCacheEviction(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[int, int](3)
	require.NoError(t, errE, "% -+#.1v", errE)

	cache.Add(1, 1)
	cache.Add(2, 2)
	cache.Add(3, 3)
	// Adding a 4th entry evicts the least recently used (key 1).
	cache.Add(4, 4)

	_, ok := cache.Get(1)
	assert.False(t, ok)
	assert.Equal(t, uint64(1), cache.MissCount())

	_, ok = cache.Get(4)
	assert.True(t, ok)
	assert.Equal(t, uint64(0), cache.MissCount())
}

func TestLRUCacheConcurrentMissCount(t *testing.T) {
	t.Parallel()

	cache, errE := x.NewLRUCache[int, int](100)
	require.NoError(t, errE, "% -+#.1v", errE)

	const goroutines = 10
	const misses = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for i := range misses {
				cache.Get(1000 + i)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, uint64(goroutines*misses), cache.MissCount())
}
