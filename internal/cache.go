package internal

import (
	"time"
)

// Collection represents the cache underlying data structure,
// and defines the functions or operations that can be applied to the data elements.
type Collection interface {
	Move(*Entry)
	Add(*Entry)
	Remove(*Entry)
	Discard() *Entry
	Len() int
	Init()
}

// Entry is used to hold a value in the cache.
type Entry struct {
	Key     interface{}
	Value   interface{}
	Element interface{}
	Exp     time.Time
	timer   *time.Timer
	cancel  chan struct{}
}

// start/stop timer added for safety to prevent fire on expired callback,
// when entry re-stored at the expiry time.
func (e *Entry) startTimer(d time.Duration, f func(key, value interface{})) {
	e.cancel = make(chan struct{})
	e.timer = time.AfterFunc(d, func() {
		select {
		case <-e.cancel:
		default:
			f(e.Key, e.Value)
		}
	})
}

func (e *Entry) stopTimer() {
	if e.timer == nil {
		return
	}
	e.timer.Stop()
	close(e.cancel)
}

// Cache is an abstracted cache that provides a skeletal implementation,
// of the Cache interface to minimize the effort required to implement interface.
type Cache struct {
	coll      Collection
	entries   map[interface{}]*Entry
	onEvicted func(key, value interface{})
	onExpired func(key, value interface{})
	ttl       time.Duration
	capacity  int
}

// Load returns key value.
func (c *Cache) Load(key interface{}) (interface{}, bool) {
	return c.get(key, false)
}

// Peek returns key value without updating the underlying "rank".
func (c *Cache) Peek(key interface{}) (interface{}, bool) {
	return c.get(key, true)
}

func (c *Cache) get(key interface{}, peek bool) (v interface{}, found bool) {
	e, ok := c.entries[key]
	if !ok {
		return
	}

	if !e.Exp.IsZero() && time.Now().UTC().After(e.Exp) {
		c.evict(e)
		return
	}

	if !peek {
		c.coll.Move(e)
	}

	return e.Value, ok
}

// Expiry returns key value expiry time.
func (c *Cache) Expiry(key interface{}) (t time.Time, ok bool) {
	ok = c.Contains(key)
	if ok {
		t = c.entries[key].Exp
	}
	return t, ok
}

// Store sets the value for a key.
func (c *Cache) Store(key, value interface{}) {
	c.StoreWithTTL(key, value, c.ttl)
}

// StoreWithTTL sets the key value with TTL overrides the default.
func (c *Cache) StoreWithTTL(key, value interface{}, ttl time.Duration) {
	if e, ok := c.entries[key]; ok {
		c.removeEntry(e)
	}

	e := &Entry{Key: key, Value: value}

	if ttl > 0 {
		if c.onExpired != nil {
			e.startTimer(ttl, c.onExpired)
		}
		e.Exp = time.Now().UTC().Add(ttl)
	}

	c.entries[key] = e
	if c.capacity != 0 && c.Len() >= c.capacity {
		c.Discard()
	}
	c.coll.Add(e)
}

// Update the key value without updating the underlying "rank".
func (c *Cache) Update(key, value interface{}) {
	if c.Contains(key) {
		c.entries[key].Value = value
	}
}

// Purge Clears all cache entries.
func (c *Cache) Purge() {
	defer c.coll.Init()

	if c.onEvicted == nil {
		c.entries = make(map[interface{}]*Entry)
		return
	}

	for _, e := range c.entries {
		c.evict(e)
	}
}

// Resize cache, returning number evicted
func (c *Cache) Resize(size int) int {
	c.capacity = size
	diff := c.Len() - size

	if diff < 0 {
		diff = 0
	}

	for i := 0; i < diff; i++ {
		c.Discard()
	}

	return diff
}

// DelSilently the key value silently without call onEvicted.
func (c *Cache) DelSilently(key interface{}) {
	if e, ok := c.entries[key]; ok {
		c.removeEntry(e)
	}
}

// Delete deletes the key value.
func (c *Cache) Delete(key interface{}) {
	if e, ok := c.entries[key]; ok {
		c.evict(e)
	}
}

// Contains Checks if a key exists in cache.
func (c *Cache) Contains(key interface{}) (ok bool) {
	_, ok = c.Peek(key)
	return
}

// Keys return cache records keys.
func (c *Cache) Keys() (keys []interface{}) {
	for k := range c.entries {
		keys = append(keys, k)
	}
	return
}

// Len Returns the number of items in the cache.
func (c *Cache) Len() int {
	return c.coll.Len()
}

// Discard oldest entry from cache to make room for the new ones.
func (c *Cache) Discard() (key, value interface{}) {
	if e := c.coll.Discard(); e != nil {
		c.evict(e)
		return e.Key, e.Value
	}

	return
}

func (c *Cache) removeEntry(e *Entry) {
	c.coll.Remove(e)
	e.stopTimer()
	delete(c.entries, e.Key)
}

// evict remove entry and fire on evicted callback.
func (c *Cache) evict(e *Entry) {
	c.removeEntry(e)
	if c.onEvicted != nil {
		go c.onEvicted(e.Key, e.Value)
	}
}

// TTL returns entries default TTL.
func (c *Cache) TTL() time.Duration {
	return c.ttl
}

// SetTTL sets entries default TTL.
func (c *Cache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Cap Returns the cache capacity.
func (c *Cache) Cap() int {
	return c.capacity
}

// RegisterOnEvicted registers a function,
// to call in its own goroutine when an entry is purged from the cache.
func (c *Cache) RegisterOnEvicted(f func(key, value interface{})) {
	c.onEvicted = f
}

// RegisterOnExpired registers a function,
// to call in its own goroutine when an entry TTL elapsed.
func (c *Cache) RegisterOnExpired(f func(key, value interface{})) {
	c.onExpired = f
}

// New return new abstracted cache.
func New(c Collection, cap int) *Cache {
	return &Cache{
		coll:     c,
		capacity: cap,
		entries:  make(map[interface{}]*Entry),
	}
}
