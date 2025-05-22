package pokeCache

import (
	"sync"
	"time"
)

type Cache struct {
	Cache	map[string]cacheEntry
	mu		sync.Mutex
}

type cacheEntry struct {
	CreatedAt	time.Time
	val			[]byte
}

func NewCache(interval time.Duration) *Cache {
	var cache = Cache{
		Cache: make(map[string]cacheEntry),
	}
	go cache.reapLoop(interval)
	return &cache
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.Cache {
			if entry.CreatedAt.Add(interval).Before(time.Now()) {
				delete(c.Cache, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	val, ok := c.Cache[key]
	c.mu.Unlock()
	if ok {
		return val.val, true
	}
	return nil, false
}

func (c *Cache) Add(key string, value []byte) {
	entry := cacheEntry{
		CreatedAt: time.Now(),
		val:       value,
	}
	c.mu.Lock()
	c.Cache[key] = entry
	c.mu.Unlock()
}
