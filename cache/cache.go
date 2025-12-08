package cache

import (
	"time"

	"github.com/maypok86/otter/v2"
)

type Cache struct {
	cache *otter.Cache[string, valueTTL]
}

type valueTTL struct {
	value string
	ttl   time.Duration
}

func NewCache() *Cache {
	cache := otter.Must(&otter.Options[string, valueTTL]{
		ExpiryCalculator: otter.ExpiryAccessingFunc(func(e otter.Entry[string, valueTTL]) time.Duration {
			if e.Value.ttl > 0 {
				return e.Value.ttl
			}
			// fallback default
			return 5 * time.Minute
		}),
	})
	return &Cache{
		cache: cache,
	}
}

func (c *Cache) Set(key string, value string, ttl time.Duration) {
	c.cache.Set(key, valueTTL{
		value: value,
		ttl:   ttl,
	})
}

func (c *Cache) Get(key string) (string, bool) {
	value, ok := c.cache.GetIfPresent(key)
	if !ok {
		return "", false
	}
	return value.value, true
}

func (c *Cache) Delete(key string) {
	c.cache.Invalidate(key)
}

func (c *Cache) Close() {
	c.cache.CleanUp()
}
