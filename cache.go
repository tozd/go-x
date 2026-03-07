package x

import (
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
	"gitlab.com/tozd/go/errors"
)

// Cache is a LRU cache which counts cache misses.
type Cache[K comparable, V any] struct {
	*lru.Cache[K, V]

	missCount uint64
}

// NewCache creates a new LRU cache with the specified size.
func NewCache[K comparable, V any](size int) (*Cache[K, V], errors.E) {
	cache, err := lru.New[K, V](size)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Cache[K, V]{
		Cache:     cache,
		missCount: 0,
	}, nil
}

// Get retrieves a document from the cache and tracks cache misses.
func (c *Cache[K, V]) Get(key K) (V, bool) { //nolint:ireturn
	value, ok := c.Cache.Get(key)
	if !ok {
		atomic.AddUint64(&c.missCount, 1)
	}
	return value, ok
}

// MissCount returns the number of cache misses since the last call
// of MissCount (or since the initialization of the cache).
func (c *Cache[K, V]) MissCount() uint64 {
	return atomic.SwapUint64(&c.missCount, 0)
}
