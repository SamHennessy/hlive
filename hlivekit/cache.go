package hlivekit

import (
	"fmt"

	"github.com/dgraph-io/ristretto/v2"

	l "github.com/SamHennessy/hlive"
)

// NewCacheRistretto wraps a Ristretto cache for use as an HLive Cache.
//
// HLive always uses string keys (a page's content hash), so the cache should
// be created with a string key type, e.g. ristretto.NewCache[string, any].
func NewCacheRistretto(cache *ristretto.Cache[string, any]) *CacheRistretto {
	return &CacheRistretto{cache: cache}
}

// CacheRistretto is a Cache adapter for Ristretto (github.com/dgraph-io/ristretto/v2).
type CacheRistretto struct {
	cache *ristretto.Cache[string, any]
}

func (c *CacheRistretto) Get(key any) (any, bool) {
	k, ok := key.(string)
	if !ok {
		l.LoggerDev.Error("CacheRistretto: key is not a string", "callers", l.CallerStackStr(), "key", fmt.Sprintf("%#v", key))

		return nil, false
	}

	return c.cache.Get(k)
}

func (c *CacheRistretto) Set(key any, value any) {
	k, ok := key.(string)
	if !ok {
		l.LoggerDev.Error("CacheRistretto: key is not a string", "callers", l.CallerStackStr(), "key", fmt.Sprintf("%#v", key))

		return
	}

	c.cache.Set(k, value, 0)
}
