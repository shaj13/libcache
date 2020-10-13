package internal

import (
	"time"
)

// Collection represents the container underlying data structure,
// and defines the functions or operations that can be applied to the data elements.
type Collection interface {
	Move(*Entry)
	Add(*Entry)
	Remove(*Entry)
	RemoveOldest() *Entry
	Len() int
	Init()
}

// Entry is used to hold a value in the cache.
type Entry struct {
	Key     interface{}
	Value   interface{}
	Element interface{}
	Exp     time.Time
	Timer   *time.Timer
}

// Container represent core cache container.
type Container struct {
	coll      Collection
	entries   map[interface{}]*Entry
	onEvicted func(key interface{}, value interface{})
	onExpired func(key interface{})
	ttl       time.Duration
	capacity  int
}

// Load returns key's value.
func (c *Container) Load(key interface{}) (interface{}, bool) {
	return c.get(key, false)
}

// Peek returns key's value without updating the underlying "rank".
func (c *Container) Peek(key interface{}) (interface{}, bool) {
	return c.get(key, true)
}

func (c *Container) get(key interface{}, peek bool) (v interface{}, found bool) {
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

// Store sets the value for a key.
func (c *Container) Store(key, value interface{}) {
	c.Set(key, value, c.ttl)
}

// Set sets the key value with TTL overrides the default.
func (c *Container) Set(key, value interface{}, ttl time.Duration) {
	if e, ok := c.entries[key]; ok {
		c.removeEntry(e)
	}

	e := &Entry{Key: key, Value: value}

	if ttl > 0 {
		if c.onExpired != nil {
			e.Timer = time.AfterFunc(ttl, func() {
				c.onExpired(e.Key)
			})
		}
		e.Exp = time.Now().UTC().Add(ttl)
	}

	c.entries[key] = e
	if c.capacity != 0 && c.Len() >= c.capacity {
		c.RemoveOldest()
	}
	c.coll.Add(e)
}

// Update the key value without updating the underlying "rank".
func (c *Container) Update(key, value interface{}) {
	if e, ok := c.entries[key]; ok {
		e.Value = value
	}
}

// Purge Clears all cache entries.
func (c *Container) Purge() {
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
func (c *Container) Resize(size int) int {
	c.capacity = size
	diff := c.Len() - size

	if diff < 0 {
		diff = 0
	}

	for i := 0; i < diff; i++ {
		c.RemoveOldest()
	}

	return diff
}

// Delete deletes the key value.
func (c *Container) Delete(key interface{}) {
	if e, ok := c.entries[key]; ok {
		c.evict(e)
	}
}

// Contains Checks if a key exists in cache.
func (c *Container) Contains(key interface{}) (ok bool) {
	_, ok = c.Peek(key)
	return
}

// Keys return cache records keys.
func (c *Container) Keys() (keys []interface{}) {
	for k := range c.entries {
		keys = append(keys, k)
	}
	return
}

// Len Returns the number of items in the cache.
func (c *Container) Len() int {
	return c.coll.Len()
}

// RemoveOldest Removes the oldest entry from cache.
func (c *Container) RemoveOldest() (key, value interface{}) {
	if e := c.coll.RemoveOldest(); e != nil {
		c.evict(e)
		return e.Key, e.Value
	}
	return
}

// removeEntry remove entry silently.
func (c *Container) removeEntry(e *Entry) {
	c.coll.Remove(e)
	if e.Timer != nil {
		e.Timer.Stop()
	}
	delete(c.entries, e.Key)
}

// evict remove entry and fire on evicted callback.
func (c *Container) evict(e *Entry) {
	c.removeEntry(e)
	if c.onEvicted != nil {
		go c.onEvicted(e.Key, e.Value)
	}
}

// TTL returns entries default TTL.
func (c *Container) TTL() time.Duration {
	return c.ttl
}

// SetTTL sets entries default TTL.
func (c *Container) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Cap Returns the cache capacity.
func (c *Container) Cap() int {
	return c.capacity
}

// RegisterOnEvicted registers a function,
// to call in its own goroutine when an entry is purged from the cache.
func (c *Container) RegisterOnEvicted(f func(key, value interface{})) {
	c.onEvicted = f
}

// RegisterOnExpired registers a function,
// to call in its own goroutine when an entry TTL elapsed.
func (c *Container) RegisterOnExpired(f func(key interface{})) {
	c.onExpired = f
}

// New return new container.
func New(c Collection, cap int) *Container {
	return &Container{
		coll:     c,
		capacity: cap,
		entries:  make(map[interface{}]*Entry),
	}
}
