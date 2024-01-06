package cache

import (
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

const (
	defaultExpiration = 48 * time.Hour
	cleanupPeriod     = 72 * time.Hour
)

var (
	//lint:ignore GLOBAL this is the only cache variable
	DataCache *Cache
)

func init() {
	DataCache = NewCache()
}

// Cache is a wrapper around go-cache with a mutex for concurrent access.
type Cache struct {
	cache *cache.Cache
	mutex sync.Mutex
}

// NewCache creates a new instance of the Cache.
func NewCache() *Cache {
	return &Cache{
		cache: cache.New(defaultExpiration, cleanupPeriod),
		// no need to initialize mutex
	}
}

// Get retrieves a value from the cache.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.cache.Get(key)
}

// Set sets a value in the cache.
func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache.Set(key, value, cache.DefaultExpiration)
}

// Delete deletes a value from the cache.
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache.Delete(key)
}

// DeleteAll deletes all cache entries matching the specified prefix.
func (c *Cache) DeleteAll(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key := range c.cache.Items() {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			c.cache.Delete(key)
		}
	}
}
