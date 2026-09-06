package storage

import (
	"container/list"
	"sync"
	"time"
)

type cacheEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
}

// LRUCache provides a thread-safe, memory-bounded LRU cache with TTL support.
type LRUCache struct {
	mu         sync.RWMutex
	capacity   int
	defaultTTL time.Duration
	items      map[string]*list.Element
	evictList  *list.List
}

// NewLRUCache creates an LRUCache with given max items and default expiration TTL.
func NewLRUCache(capacity int, defaultTTL time.Duration) *LRUCache {
	if capacity <= 0 {
		capacity = 1000
	}
	if defaultTTL <= 0 {
		defaultTTL = 24 * time.Hour
	}
	return &LRUCache{
		capacity:   capacity,
		defaultTTL: defaultTTL,
		items:      make(map[string]*list.Element),
		evictList:  list.New(),
	}
}

// Get retrieves an item by key. Returns value and true if found and not expired.
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	element, exists := c.items[key]
	if !exists {
		return nil, false
	}

	entry := element.Value.(*cacheEntry)
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeElement(element)
		return nil, false
	}

	c.evictList.MoveToFront(element)
	return entry.value, true
}

// Set stores an item in the cache with the default TTL.
func (c *LRUCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores an item with a custom TTL duration.
func (c *LRUCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// Update existing element
	if element, exists := c.items[key]; exists {
		c.evictList.MoveToFront(element)
		entry := element.Value.(*cacheEntry)
		entry.value = value
		entry.expiresAt = expiresAt
		return
	}

	// Evict oldest if capacity exceeded
	if c.evictList.Len() >= c.capacity {
		c.evictOldest()
	}

	entry := &cacheEntry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}
	element := c.evictList.PushFront(entry)
	c.items[key] = element
}

// Delete removes an item by key.
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, exists := c.items[key]; exists {
		c.removeElement(element)
	}
}

// Len returns the current count of items in the cache.
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evictList.Len()
}

// Purge evicts all entries from the cache.
func (c *LRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.evictList.Init()
}

func (c *LRUCache) evictOldest() {
	element := c.evictList.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *LRUCache) removeElement(element *list.Element) {
	c.evictList.Remove(element)
	entry := element.Value.(*cacheEntry)
	delete(c.items, entry.key)
}
