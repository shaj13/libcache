// Package libcache provides in-memory caches based on different caches replacement algorithms.
package libcache

import (
	"sync"
	"time"
)

// Cache stores data so that future requests for that data can be served faster.
type Cache interface {
	// Load returns key value.
	Load(key interface{}) (interface{}, bool)
	// Peek returns key value without updating the underlying "recent-ness".
	Peek(key interface{}) (interface{}, bool)
	// Update the key value without updating the underlying "recent-ness".
	Update(key interface{}, value interface{})
	// Store sets the key value.
	Store(key interface{}, value interface{})
	// Set sets the key value with TTL overrides the default.
	Set(key interface{}, value interface{}, ttl time.Duration)
	// Delete deletes the key value.
	Delete(key interface{})
	// DeleteOldest Removes the oldest entry from cache.
	DeleteOldest() (key, value interface{})
	// Expiry returns key value expiry time.
	Expiry(key interface{}) (time.Time, bool)
	// Keys return cache records keys.
	Keys() []interface{}
	// Contains Checks if a key exists in cache.
	Contains(key interface{}) bool
	// Purge Clears all cache entries.
	Purge()
	// Resize cache, returning number evicted
	Resize(int) int
	// Len Returns the number of items in the cache.
	Len() int
	// Cap Returns the cache capacity.
	Cap() int
	// TTL returns entries default TTL.
	TTL() time.Duration
	// SetTTL sets entries default TTL.
	SetTTL(time.Duration)
	// RegisterOnEvicted registers a function,
	// to call in its own goroutine when an entry is purged from the cache.
	RegisterOnEvicted(f func(key, value interface{}))
	// RegisterOnExpired registers a function,
	// to call in its own goroutine when an entry TTL elapsed.
	// invocation of f, does not mean the entry is purged from the cache,
	// if need be, it must coordinate with the cache explicitly.
	//
	// 	var cache cache.Cache
	// 	onExpired := func(key interface{}) {
	//	 	_, _, _ = cache.Peek(key)
	// 	}
	//
	// This should not be done unless the cache thread-safe.
	RegisterOnExpired(f func(key interface{}))
}

type cache struct {
	mu        sync.RWMutex
	container Cache
}

func (c *cache) Load(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container.Load(key)
}

func (c *cache) Peek(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container.Peek(key)
}

func (c *cache) Update(key interface{}, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.Update(key, value)
}

func (c *cache) Store(key interface{}, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.Store(key, value)
}

func (c *cache) Set(key interface{}, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.Set(key, value, ttl)
}

func (c *cache) Delete(key interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.Delete(key)
}

func (c *cache) Keys() []interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.container.Keys()
}

func (c *cache) Contains(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container.Contains(key)
}

func (c *cache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.Purge()
}

func (c *cache) Resize(s int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container.Resize(s)
}

func (c *cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.container.Len()
}

func (c *cache) Cap() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.container.Cap()
}

func (c *cache) TTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.container.TTL()
}

func (c *cache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.SetTTL(ttl)
}

func (c *cache) RegisterOnEvicted(f func(key, value interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.RegisterOnEvicted(f)
}

func (c *cache) RegisterOnExpired(f func(key interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.container.RegisterOnExpired(f)
}

func (c *cache) DeleteOldest() (key, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container.DeleteOldest()
}

func (c *cache) Expiry(key interface{}) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.container.Expiry(key)
}
