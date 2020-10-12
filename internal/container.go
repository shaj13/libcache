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
	GetOldest() *Entry
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
	Collection Collection
	Entries    map[interface{}]*Entry
	OnEvicted  func(key interface{}, value interface{})
	OnExpired  func(key interface{})
	ttl        time.Duration
	Capacity   int
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
	e, ok := c.Entries[key]
	if !ok {
		return
	}

	if !e.Exp.IsZero() && time.Now().UTC().After(e.Exp) {
		c.Evict(e)
		return
	}

	if !peek {
		c.Collection.Move(e)
	}

	return e.Value, ok
}

// Store sets the value for a key.
func (c *Container) Store(key, value interface{}) {
	c.Set(key, value, c.ttl)
}

// Set sets the key value with TTL overrides the default.
func (c *Container) Set(key, value interface{}, ttl time.Duration) {
	if e, ok := c.Entries[key]; ok {
		c.RemoveEntry(e)
	}

	e := &Entry{Key: key, Value: value}

	if ttl > 0 {
		if c.OnExpired != nil {
			e.Timer = time.AfterFunc(ttl, func() {
				c.OnExpired(e.Key)
			})
		}
		e.Exp = time.Now().UTC().Add(ttl)
	}

	c.Entries[key] = e
	c.RemoveOldest()
	c.Collection.Add(e)
}

// Update the key value without updating the underlying "rank".
func (c *Container) Update(key, value interface{}) {
	if e, ok := c.Entries[key]; ok {
		e.Value = value
	}
}

// Purge Clears all cache entries.
func (c *Container) Purge() {
	defer c.Collection.Init()

	if c.OnEvicted == nil {
		c.Entries = make(map[interface{}]*Entry)
		return
	}

	for _, e := range c.Entries {
		c.Evict(e)
	}
}

// Resize cache, returning number evicted
func (c *Container) Resize(size int) int {
	c.Capacity = size
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
	if e, ok := c.Entries[key]; ok {
		c.Evict(e)
	}
}

// Contains Checks if a key exists in cache.
func (c *Container) Contains(key interface{}) (ok bool) {
	_, ok = c.Peek(key)
	return
}

// Keys return cache records keys.
func (c *Container) Keys() (keys []interface{}) {
	for k := range c.Entries {
		keys = append(keys, k)
	}
	return
}

// Len Returns the number of items in the cache.
func (c *Container) Len() int {
	return c.Collection.Len()
}

// RemoveOldest Removes the oldest entry from cache.
func (c *Container) RemoveOldest() {
	if c.Capacity != 0 && c.Len() >= c.Capacity {
		if e := c.Collection.GetOldest(); e != nil {
			c.Evict(e)
		}
	}
}

// RemoveEntry remove entry silently.
func (c *Container) RemoveEntry(e *Entry) {
	c.Collection.Remove(e)
	if e.Timer != nil {
		e.Timer.Stop()
	}
	delete(c.Entries, e.Key)
}

// Evict remove entry and fire on evicted callback.
func (c *Container) Evict(e *Entry) {
	c.RemoveEntry(e)
	if c.OnEvicted != nil {
		go c.OnEvicted(e.Key, e.Value)
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
	return c.Capacity
}

// RegisterOnEvicted registers a function,
// to call in its own goroutine when an entry is purged from the cache.
func (c *Container) RegisterOnEvicted(f func(key, value interface{})) {
	c.OnEvicted = f
}

// RegisterOnExpired registers a function,
// to call in its own goroutine when an entry TTL elapsed.
func (c *Container) RegisterOnExpired(f func(key interface{})) {
	c.OnExpired = f
}

// New return new container.
func New(c Collection, cap int) *Container {
	return &Container{
		Collection: c,
		Capacity:   cap,
		Entries:    make(map[interface{}]*Entry),
	}
}
