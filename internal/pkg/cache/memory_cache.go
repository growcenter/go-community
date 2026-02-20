package cache

import (
	"sync"
	"time"
)

// MemoryCache is a thread-safe in-memory cache with TTL support
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	ttl   time.Duration
}

// cacheItem represents a cached value with expiration
type cacheItem struct {
	value      interface{}
	expiration int64
}

// New creates a new MemoryCache with the specified TTL
func New(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*cacheItem),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.startCleanup()

	return cache
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(c.ttl).UnixNano()
	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Get retrieves a value from the cache
// Returns (value, true) if found and not expired, (nil, false) otherwise
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if expired
	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}

	return item.value, true
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}

// Count returns the number of items in the cache (including expired)
func (c *MemoryCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// startCleanup runs a background goroutine to remove expired items
func (c *MemoryCache) startCleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired items from the cache
func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for key, item := range c.items {
		if now > item.expiration {
			delete(c.items, key)
		}
	}
}

// GetOrSet retrieves a value from cache, or sets it if not found
// The setter function is only called if the key is not in cache or expired
func (c *MemoryCache) GetOrSet(key string, setter func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := c.Get(key); found {
		return value, nil
	}

	// Not in cache, call setter
	// Setter function is only called if the key is not in cache or expired
	value, err := setter()
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(key, value)

	return value, nil
}
