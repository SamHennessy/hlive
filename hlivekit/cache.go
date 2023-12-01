package hlivekit

import "github.com/dgraph-io/ristretto"

func NewCacheRistretto(cache *ristretto.Cache) *CacheRistretto {
	return &CacheRistretto{cache: cache}
}

// CacheRistretto cache adapter for Ristretto
type CacheRistretto struct {
	cache *ristretto.Cache
}

func (c *CacheRistretto) Get(key any) (any, bool) {
	return c.cache.Get(key)
}

func (c *CacheRistretto) Set(key any, value any) {
	c.cache.Set(key, value, 0)
}
