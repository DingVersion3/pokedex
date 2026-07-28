package pokecache

import (
	"sync"
	"time"
)

// cacheEntry holds a cached value and the time it was created.
type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache is a thread-safe, in-memory key/value cache that automatically
// evicts entries older than its configured interval.
type Cache struct {
	entries  map[string]cacheEntry
	mu       sync.Mutex
	interval time.Duration
}

// NewCache creates a new Cache and starts a background reap loop that
// removes entries older than interval every time interval elapses.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}
	go c.reapLoop()
	return c
}

// Add inserts or overwrites a cache entry for the given key.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

// Get returns the cached value for key, and whether it was found.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

// reapLoop periodically removes entries older than the cache's interval.
func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.reap()
	}
}

// reap deletes all entries whose age exceeds the cache's interval.
func (c *Cache) reap() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.createdAt) > c.interval {
			delete(c.entries, key)
		}
	}
}
