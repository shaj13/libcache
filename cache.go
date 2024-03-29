// Package libcache provides in-memory caches based on different caches replacement algorithms.
package libcache

import (
	"context"
	"sync"
	"time"

	"github.com/shaj13/libcache/internal"
)

// These are the generalized cache operations that can trigger a event.
const (
	Read   = internal.Read
	Write  = internal.Write
	Remove = internal.Remove
)

// Op describes a set of cache operations.
type Op = internal.Op

// Event represents a single cache entry change.
type Event = internal.Event

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
	// StoreWithTTL sets the key value with TTL overrides the default.
	StoreWithTTL(key interface{}, value interface{}, ttl time.Duration)
	// Delete deletes the key value.
	Delete(key interface{})
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
	// to call it when an entry is purged from the cache.
	//
	// Deprecated: use Notify instead.
	RegisterOnEvicted(f func(key, value interface{}))
	// RegisterOnExpired registers a function,
	// to call it when an entry TTL elapsed.
	//
	// Deprecated: use Notify instead.
	RegisterOnExpired(f func(key, value interface{}))
	// Notify causes cache to relay events to ch.
	// If no operations are provided, all incoming operations will be relayed to ch.
	// Otherwise, just the provided operations will.
	Notify(ch chan<- Event, ops ...Op)
	// Ignore causes the provided operations to be ignored. Ignore undoes the effect
	// of any prior calls to Notify for the provided operations.
	// If no operations are provided, ch removed.
	Ignore(ch chan<- Event, ops ...Op)
	// GC runs a garbage collection and blocks the caller until the
	// all expired items from the cache evicted.
	//
	// GC returns the remaining time duration for the next gc cycle
	// if there any, Otherwise, it return 0.
	//
	// Calling GC without waits for the duration to elapsed considered a no-op.
	GC() time.Duration
}

// GC runs a garbage collection to evict expired items from the cache on time.
//
// GC trace expired items based on read-write barrier, therefore it listen to
// cache write events and capture the result of calling the GC method on cache
// to trigger the garbage collection loop at the right point in time.
//
// GC is a long running function, it returns when ctx done, therefore the
// caller must start it in its own goroutine.
//
// Experimental
//
// Notice: This func is EXPERIMENTAL and may be changed or removed in a
// later release.
func GC(ctx context.Context, cache Cache) {
	remaining := time.Duration(0)

	t := time.NewTimer(remaining)
	defer t.Stop()

	c := make(chan Event, 1)
	cache.Notify(c, Write)
	defer func() {
		cache.Ignore(c)
		close(c)
	}()

	gc := func() {
		remaining = cache.GC()
		t.Stop()
		if remaining > 0 {
			t.Reset(remaining)
		}
	}

	for {
		select {
		case e := <-c:
			if e.Expiry.IsZero() {
				continue
			}

			if remaining == 0 || time.Until(e.Expiry) < remaining {
				gc()
			}
		case <-t.C:
			gc()
		case <-ctx.Done():
			return
		}
	}
}

type cache struct {
	// mu guards unsafe cache.
	// Calls to mu.Unlock are currently not deferred,
	// because defer adds ~200 ns (as of go1.)
	mu     sync.Mutex
	unsafe Cache
}

func (c *cache) Load(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	v, ok := c.unsafe.Load(key)
	c.mu.Unlock()
	return v, ok
}

func (c *cache) Peek(key interface{}) (interface{}, bool) {
	c.mu.Lock()
	v, ok := c.unsafe.Peek(key)
	c.mu.Unlock()
	return v, ok
}

func (c *cache) Update(key interface{}, value interface{}) {
	c.mu.Lock()
	c.unsafe.Update(key, value)
	c.mu.Unlock()
}

func (c *cache) Store(key interface{}, value interface{}) {
	c.mu.Lock()
	c.unsafe.Store(key, value)
	c.mu.Unlock()
}

func (c *cache) StoreWithTTL(key interface{}, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	c.unsafe.StoreWithTTL(key, value, ttl)
	c.mu.Unlock()
}

func (c *cache) Delete(key interface{}) {
	c.mu.Lock()
	c.unsafe.Delete(key)
	c.mu.Unlock()
}

func (c *cache) Keys() []interface{} {
	c.mu.Lock()
	keys := c.unsafe.Keys()
	c.mu.Unlock()
	return keys
}

func (c *cache) Contains(key interface{}) bool {
	c.mu.Lock()
	ok := c.unsafe.Contains(key)
	c.mu.Unlock()
	return ok
}

func (c *cache) Purge() {
	c.mu.Lock()
	c.unsafe.Purge()
	c.mu.Unlock()
}

func (c *cache) Resize(s int) int {
	c.mu.Lock()
	n := c.unsafe.Resize(s)
	c.mu.Unlock()
	return n
}

func (c *cache) Len() int {
	c.mu.Lock()
	n := c.unsafe.Len()
	c.mu.Unlock()
	return n
}

func (c *cache) Cap() int {
	c.mu.Lock()
	n := c.unsafe.Cap()
	c.mu.Unlock()
	return n
}

func (c *cache) TTL() time.Duration {
	c.mu.Lock()
	ttl := c.unsafe.TTL()
	c.mu.Unlock()
	return ttl
}

func (c *cache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	c.unsafe.SetTTL(ttl)
	c.mu.Unlock()
}

func (c *cache) RegisterOnEvicted(f func(key, value interface{})) {
	c.mu.Lock()
	c.unsafe.RegisterOnEvicted(f)
	c.mu.Unlock()
}

func (c *cache) RegisterOnExpired(f func(key, value interface{})) {
	c.mu.Lock()
	c.unsafe.RegisterOnExpired(f)
	c.mu.Unlock()
}

func (c *cache) Notify(ch chan<- Event, ops ...Op) {
	c.mu.Lock()
	c.unsafe.Notify(ch, ops...)
	c.mu.Unlock()
}

func (c *cache) Ignore(ch chan<- Event, ops ...Op) {
	c.mu.Lock()
	c.unsafe.Ignore(ch, ops...)
	c.mu.Unlock()
}

func (c *cache) Expiry(key interface{}) (time.Time, bool) {
	c.mu.Lock()
	exp, ok := c.unsafe.Expiry(key)
	c.mu.Unlock()
	return exp, ok
}

func (c *cache) GC() time.Duration {
	c.mu.Lock()
	dur := c.unsafe.GC()
	c.mu.Unlock()
	return dur
}
