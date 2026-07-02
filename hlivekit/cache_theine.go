package hlivekit

import (
	"fmt"

	"github.com/Yiling-J/theine-go"

	l "github.com/SamHennessy/hlive"
)

// NewCacheTheine wraps a Theine cache for use as an HLive Cache.
//
// HLive always uses string keys (a page's content hash), so the cache should
// be built with a string key type, e.g. theine.NewBuilder[string, any](maxSize).
func NewCacheTheine(cache *theine.Cache[string, any]) *CacheTheine {
	return &CacheTheine{cache: cache}
}

// CacheTheine is a Cache adapter for Theine (github.com/Yiling-J/theine-go).
type CacheTheine struct {
	cache *theine.Cache[string, any]
}

func (c *CacheTheine) Get(key any) (any, bool) {
	k, ok := key.(string)
	if !ok {
		l.LoggerDev.Error("CacheTheine: key is not a string", "callers", l.CallerStackStr(), "key", fmt.Sprintf("%#v", key))

		return nil, false
	}

	return c.cache.Get(k)
}

func (c *CacheTheine) Set(key any, value any) {
	k, ok := key.(string)
	if !ok {
		l.LoggerDev.Error("CacheTheine: key is not a string", "callers", l.CallerStackStr(), "key", fmt.Sprintf("%#v", key))

		return
	}

	c.cache.Set(k, value, 0)
}
