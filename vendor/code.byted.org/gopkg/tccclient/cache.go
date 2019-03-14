package tccclient

import (
	"sync"
	"time"
)

var now = func() time.Time {
	return time.Now()
}

// Item represents a record in the cache map
type Item struct {
	Value   string
	Expires time.Time
}

func (e *Item) Expired() bool {
	return now().After(e.Expires)
}

// Cache is a synchronised map of items
type Cache struct {
	mu    sync.RWMutex
	items map[string]*Item
}

// NewCache creates instance of Cache
func NewCache() *Cache {
	return &Cache{items: map[string]*Item{}}
}

// Set adds key => value to cache
func (c *Cache) Set(key string, item Item) {
	c.mu.Lock()
	c.items[key] = &item
	c.stepClean()
	c.mu.Unlock()
}

// Get returns value of key
func (c *Cache) Get(key string) *Item {
	c.mu.RLock()
	item := c.items[key]
	c.mu.RUnlock()
	if item == nil {
		return nil
	}
	return item
}

// Len returns items number of cache
func (c *Cache) Len() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

func (c *Cache) stepClean() {
	const steps = 10
	n := 0
	for k, e := range c.items {
		if n >= steps {
			break
		}
		if now().After(e.Expires.Add(10 * time.Minute)) {
			// key has expired for a long time
			delete(c.items, k)
		}
		n++
	}
}
