package hlivekit

import (
	"fmt"

	"github.com/maypok86/otter/v2"

	l "github.com/SamHennessy/hlive"
)

// NewCacheOtter wraps an Otter cache for use as an HLive Cache.
//
// HLive always uses string keys (a page's content hash), so the cache should
// be created with a string key type, e.g. otter.Must(&otter.Options[string, any]{...}).
func NewCacheOtter(cache *otter.Cache[string, any]) *CacheOtter {
	return &CacheOtter{cache: cache}
}

// CacheOtter is a Cache adapter for Otter (github.com/maypok86/otter/v2).
type CacheOtter struct {
	cache *otter.Cache[string, any]
}

func (c *CacheOtter) Get(key any) (any, bool) {
	k, ok := key.(string)
	if !ok {
		l.LoggerDev.Error("CacheOtter: key is not a string", "callers", l.CallerStackStr(), "key", fmt.Sprintf("%#v", key))

		return nil, false
	}

	return c.cache.GetIfPresent(k)
}

func (c *CacheOtter) Set(key any, value any) {
	k, ok := key.(string)
	if !ok {
		l.LoggerDev.Error("CacheOtter: key is not a string", "callers", l.CallerStackStr(), "key", fmt.Sprintf("%#v", key))

		return
	}

	c.cache.Set(k, value)
}
