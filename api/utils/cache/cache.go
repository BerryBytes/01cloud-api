package cache

import (
	"os"
	"reflect"
	"sync"
	"time"
)

type ICache interface {
	Set(key string, value interface{}, duration ...time.Duration)
	Get(key string, result interface{}) (ok bool)
	Delete(key string)
	DeleteMulti(keys []string)
}

type Cache struct {
	data     map[string]*CacheEntry
	lock     sync.RWMutex
	duration time.Duration
}
type DisabledCache struct {
}

// Delete implements ICache.
func (*DisabledCache) Delete(key string) {

}

// DeleteMulti implements ICache.
func (*DisabledCache) DeleteMulti(keys []string) {

}

// Get implements ICache.
func (*DisabledCache) Get(key string, result interface{}) (ok bool) {
	return false
}

// Set implements ICache.
func (*DisabledCache) Set(key string, value interface{}, duration ...time.Duration) {
}

type CacheEntry struct {
	value      interface{}
	expiration time.Time
}

func NewCache(duration time.Duration) ICache {
	if os.Getenv("CACHE_TYPE") == "redis" {
		return NewRedisCache(duration)
	}
	if os.Getenv("CACHE_TYPE") == "local" {
		return NewLocalCache(duration)
	}
	return NewDisabledCache()
}

func NewDisabledCache() *DisabledCache {
	return &DisabledCache{}
}

func NewLocalCache(duration time.Duration) *Cache {
	return &Cache{
		data:     make(map[string]*CacheEntry),
		duration: duration,
	}
}

func (c *Cache) Set(key string, value interface{}, duration ...time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	expiration := time.Now().Add(c.duration)
	if len(duration) > 0 {
		expiration = time.Now().Add(duration[0])
	}

	entry := &CacheEntry{
		value:      value,
		expiration: expiration,
	}
	c.data[key] = entry
}

func (c *Cache) Get(key string, result interface{}) (ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	entry, found := c.data[key]
	if found && time.Now().Before(entry.expiration) {
		resultValue := reflect.ValueOf(result).Elem()
		entryValue := reflect.ValueOf(entry.value)

		// Check if the types are compatible
		if entryValue.Type().AssignableTo(resultValue.Type()) {
			resultValue.Set(entryValue)
			return true
		}
	}
	return false
}

func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.data, key)
}

func (c *Cache) DeleteMulti(keys []string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, key := range keys {
		delete(c.data, key)
	}
}
