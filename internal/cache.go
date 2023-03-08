package internal

import (
	"container/heap"
	"fmt"
	"time"
)

// Op describes a set of cache operations.
type Op uint8

// These are the generalized cache operations that can trigger a event.
const (
	Read Op = iota + 1
	Write
	Remove
	maxOp
)

func (op Op) String() string {
	switch op {
	case Read:
		return "READ"
	case Write:
		return "WRITE"
	case Remove:
		return "REMOVE"
	default:
		return "UNKNOWN"
	}
}

type handler struct {
	mask [((maxOp - 1) + 7) / 8]uint8
}

func (h *handler) want(op Op) bool {
	return (h.mask[op/8]>>uint8(op&7))&1 != 0
}

func (h *handler) set(op Op) {
	h.mask[op/8] |= 1 << uint8(op&7)
}

func (h *handler) clear(op Op) {
	h.mask[op/8] &^= 1 << uint8(op&7)
}

// Collection represents the cache underlying data structure,
// and defines the functions or operations that can be applied to the data elements.
type Collection interface {
	Move(*Entry)
	Add(*Entry)
	Remove(*Entry)
	Discard() *Entry
	Front() *Entry
	Back() *Entry
	Len() int
	Init()
}

// Event represents a single cache entry change.
type Event struct {
	// Op represents cache operation that triggered the event.
	Op Op
	// Key represents cache entry key.
	Key interface{}
	// Value represents cache key value.
	Value interface{}
	// Expiry represents cache key value expiry time.
	Expiry time.Time
	// Ok report whether the read operation succeed.
	Ok bool
}

// String returns a string representation of the event in the form
// "file: REMOVE|WRITE|..."
func (e Event) String() string {
	return fmt.Sprintf("%v: %s", e.Key, e.Op.String())
}

// Entry is used to hold a value in the cache.
type Entry struct {
	Key     interface{}
	Value   interface{}
	Element interface{}
	Exp     time.Time
	index   int
}

// Cache is an abstracted cache that provides a skeletal implementation,
// of the Cache interface to minimize the effort required to implement interface.
type Cache struct {
	coll     Collection
	heap     expiringHeap
	entries  map[interface{}]*Entry
	handlers map[chan<- Event]*handler
	ttl      time.Duration
	capacity int
}

// Front returns the first key of cache or nil if the cache is empty.
func (c *Cache) Front() interface{} {
	// Run GC inline before get the front entry.
	c.GC()

	if e := c.coll.Front(); e != nil {
		return e.Key
	}

	return nil
}

// Back returns the last key of cache or nil if the cache is empty.
func (c *Cache) Back() interface{} {
	// Run GC inline before get the back entry.
	c.GC()

	if e := c.coll.Back(); e != nil {
		return e.Key
	}

	return nil
}

// Load returns key value.
func (c *Cache) Load(key interface{}) (interface{}, bool) {
	return c.get(key, false)
}

// Peek returns key value without updating the underlying "rank".
func (c *Cache) Peek(key interface{}) (interface{}, bool) {
	return c.get(key, true)
}

func (c *Cache) get(key interface{}, peek bool) (interface{}, bool) {
	// Run GC inline before return the entry.
	c.GC()

	e, ok := c.entries[key]
	if !ok {
		c.emit(Read, key, nil, time.Time{}, ok)
		return nil, ok
	}

	if !peek {
		c.coll.Move(e)
	}

	c.emit(Read, key, e.Value, e.Exp, ok)
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
	// Run GC inline before pushing the new entry.
	c.GC()

	if e, ok := c.entries[key]; ok {
		c.removeEntry(e)
	}

	e := &Entry{Key: key, Value: value}

	if ttl > 0 {
		e.Exp = time.Now().UTC().Add(ttl)
		heap.Push(&c.heap, e)
	}

	c.entries[key] = e
	if c.capacity != 0 && c.Len() >= c.capacity {
		c.Discard()
	}

	c.coll.Add(e)
	c.emit(Write, e.Key, e.Value, e.Exp, false)
}

// Update the key value without updating the underlying "rank".
func (c *Cache) Update(key, value interface{}) {
	// Run GC inline before update the entry.
	c.GC()

	if c.Contains(key) {
		e := c.entries[key]
		e.Value = value
		c.emit(Write, e.Key, e.Value, e.Exp, false)
	}
}

// Purge Clears all cache entries.
func (c *Cache) Purge() {
	defer c.coll.Init()

	if len(c.handlers) == 0 {
		c.entries = make(map[interface{}]*Entry)
		c.heap = nil
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
	delete(c.entries, e.Key)
	// Remove entry from the heap, the entry may does not exist because
	// it has zero ttl or already popped up by gc
	if len(c.heap) > 0 && e.index < len(c.heap) && e.Key == c.heap[e.index].Key {
		heap.Remove(&c.heap, e.index)
	}
}

// evict remove entry and fire on evicted callback.
func (c *Cache) evict(e *Entry) {
	c.removeEntry(e)
	c.emit(Remove, e.Key, e.Value, e.Exp, false)
}

func (c *Cache) emit(op Op, k, v interface{}, exp time.Time, ok bool) {
	e := Event{
		Op:     op,
		Key:    k,
		Value:  v,
		Expiry: exp,
		Ok:     ok,
	}

	for c, h := range c.handlers {
		if h.want(op) {
			// send but do not block for it
			select {
			case c <- e:
			default:
			}
		}
	}
}

// GC returns the remaining time duration for the next gc cycle if there any,
// Otherwise, it return 0.
//
// Calling GC without waits for the duration to elapsed considered a no-op.
func (c *Cache) GC() time.Duration {
	now := time.Now()
	for {

		// Return from gc if the heap is empty or the next element is not yet
		// expired.

		if len(c.heap) == 0 {
			return 0
		}

		if now.Before(c.heap[0].Exp) {
			return c.heap[0].Exp.Sub(now)
		}

		e := heap.Pop(&c.heap).(*Entry)
		c.evict(e)
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

// Notify causes cache to relay events to ch.
// If no operations are provided, all incoming operations will be relayed to ch.
// Otherwise, just the provided operations will.
func (c *Cache) Notify(ch chan<- Event, ops ...Op) {
	if ch == nil {
		panic("libcache: Notify using nil channel")
	}

	h := new(handler)
	c.handlers[ch] = h

	if len(ops) == 0 {
		for i := 1; i <= int(maxOp); i++ {
			h.set(Op(i))
		}
		return
	}

	for _, op := range ops {
		h.set(op)
	}
}

// Ignore causes the provided ops to be ignored. Ignore undoes the effect
// of any prior calls to Notify for the provided ops.
// If no ops are provided, ch removed.
func (c *Cache) Ignore(ch chan<- Event, ops ...Op) {
	if len(ops) == 0 {
		delete(c.handlers, ch)
		return
	}

	h, ok := c.handlers[ch]
	if !ok {
		return
	}

	for _, op := range ops {
		h.clear(op)
	}
}

// RegisterOnEvicted registers a function,
// to call it when an entry is purged from the cache.
func (c *Cache) RegisterOnEvicted(fn func(key, value interface{})) {
	panic("RegisterOnEvicted no longer available")
}

// RegisterOnExpired registers a function,
// to call it when an entry TTL elapsed.
func (c *Cache) RegisterOnExpired(fn func(key, value interface{})) {
	panic("RegisterOnExpired no longer available")
}

// New return new abstracted cache.
func New(c Collection, cap int) *Cache {
	return &Cache{
		coll:     c,
		capacity: cap,
		entries:  make(map[interface{}]*Entry),
		handlers: make(map[chan<- Event]*handler),
	}
}

// expiringHeap is a min-heap ordered by expiration time of its entries. The
// expiring cache uses this as a priority queue to efficiently organize entries
// which will be garbage collected once they expire.
type expiringHeap []*Entry

var _ heap.Interface = &expiringHeap{}

func (cq expiringHeap) Len() int {
	return len(cq)
}

func (cq expiringHeap) Less(i, j int) bool {
	return cq[i].Exp.Before(cq[j].Exp)
}

func (cq expiringHeap) Swap(i, j int) {
	cq[i].index, cq[j].index = cq[j].index, cq[i].index
	cq[i], cq[j] = cq[j], cq[i]
}

func (cq *expiringHeap) Push(c interface{}) {
	c.(*Entry).index = len(*cq)
	*cq = append(*cq, c.(*Entry))
}

func (cq *expiringHeap) Pop() interface{} {
	c := (*cq)[cq.Len()-1]
	*cq = (*cq)[:cq.Len()-1]
	return c
}
