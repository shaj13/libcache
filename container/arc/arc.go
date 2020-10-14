package arc

import (
	"time"

	"github.com/shaj13/libcache"
	"github.com/shaj13/libcache/container/lru"
	"github.com/shaj13/libcache/internal"
)

func init() {
	libcache.ARC.Register(New)
}

// New creates an ARC of the given size
func New(cap int) libcache.Cache {
	return &arc{
		p:  0,
		t1: lru.New(cap).(*internal.Container),
		b1: lru.New(cap).(*internal.Container),
		t2: lru.New(cap).(*internal.Container),
		b2: lru.New(cap).(*internal.Container),
	}
}

type arc struct {
	p int // P is the dynamic preference towards T1 or T2

	t1 *internal.Container // T1 is the LRU for recently accessed items
	b1 *internal.Container // B1 is the LRU for evictions from t1
	t2 *internal.Container // T2 is the LRU for frequently accessed items
	b2 *internal.Container // B2 is the LRU for evictions from t2
}

func (a *arc) Load(key interface{}) (value interface{}, ok bool) {
	// If the value is contained in T1 (recent), then
	// promote it to T2 (frequent)
	if val, ok := a.t1.Peek(key); ok {
		// Remove silently.
		a.t1.Remove(key)
		a.t2.Store(key, val)
		return val, ok
	}

	// Check if the value is contained in T2 (frequent)
	if val, ok := a.t2.Load(key); ok {
		return val, ok
	}

	// No hit
	return nil, false
}

func (a *arc) Store(key, val interface{}) {
	a.Set(key, val, a.TTL())
}

func (a *arc) Set(key, val interface{}, ttl time.Duration) {
	// Check if the value is contained in T1 (recent), and potentially
	// promote it to frequent T2
	if a.t1.Contains(key) {
		// Remove silently.
		a.t1.Remove(key)
		a.t2.Set(key, val, ttl)
		return
	}

	// Check if the value is already in T2 (frequent) and update it
	if a.t2.Contains(key) {
		a.t2.Set(key, val, ttl)
		return
	}

	// Check if this value was recently evicted as part of the
	// recently used list
	if a.b1.Contains(key) {
		// T1 set is too small, increase P appropriately
		delta := 1
		b1Len := a.b1.Len()
		b2Len := a.b2.Len()
		if b2Len > b1Len {
			delta = b2Len / b1Len
		}
		if a.p+delta >= a.Cap() {
			a.p = a.Cap()
		} else {
			a.p += delta
		}

		// Potentially need to make room in the cache
		if a.t1.Len()+a.t2.Len() >= a.Cap() {
			a.replace(false)
		}

		a.b1.Delete(key)

		// Add the key to the frequently used list
		a.t2.Set(key, val, ttl)
		return
	}

	// Check if this value was recently evicted as part of the
	// frequently used list
	if a.b2.Contains(key) {
		// T2 set is too small, decrease P appropriately
		delta := 1
		b1Len := a.b1.Len()
		b2Len := a.b2.Len()
		if b1Len > b2Len {
			delta = b1Len / b2Len
		}
		if delta >= a.p {
			a.p = 0
		} else {
			a.p -= delta
		}

		// Potentially need to make room in the cache
		if a.t1.Len()+a.t2.Len() >= a.Cap() {
			a.replace(true)
		}

		a.b2.Delete(key)

		// Add the key to the frequently used list
		a.t2.Set(key, val, ttl)
		return
	}

	// Potentially need to make room in the cache
	if a.Cap() != 0 && a.t1.Len()+a.t2.Len() >= a.Cap() {
		a.replace(false)
	}

	// Keep the size of the ghost buffers trim
	if a.b1.Len() > a.Cap()-a.p {
		a.b1.DeleteOldest()
	}
	if a.b2.Len() > a.p {
		a.b2.DeleteOldest()
	}

	// Add to the recently seen list
	a.t1.Set(key, val, ttl)
}

// replace is used to adaptively evict from either T1 or T2
// based on the current learned value of P
func (a *arc) replace(b2ContainsKey bool) {
	t1Len := a.t1.Len()
	if t1Len > 0 && (t1Len > a.p || (t1Len == a.p && b2ContainsKey)) {
		k, _ := a.t1.DeleteOldest()
		if k != nil {
			a.b1.Store(k, nil)
		}
	} else {
		k, _ := a.t2.DeleteOldest()
		if k != nil {
			a.b2.Store(k, nil)
		}
	}
}

func (a *arc) Delete(key interface{}) {
	if a.t1.Contains(key) {
		a.t1.Delete(key)
		return
	}
	if a.t2.Contains(key) {
		a.t2.Delete(key)
		return
	}
	if a.b1.Contains(key) {
		a.b1.Delete(key)
		return
	}
	if a.b2.Contains(key) {
		a.b2.Delete(key)
		return
	}
}

func (a *arc) Update(key, value interface{}) {
	if a.t1.Contains(key) {
		a.t1.Update(key, value)
	}
	a.t2.Update(key, value)
}

func (a *arc) Peek(key interface{}) (value interface{}, ok bool) {
	if val, ok := a.t1.Peek(key); ok {
		return val, ok
	}
	return a.t2.Peek(key)
}

func (a *arc) DeleteOldest() (k, v interface{}) {
	return
}

func (a *arc) Expiry(key interface{}) (time.Time, bool) {
	if a.t1.Contains(key) {
		return a.t1.Expiry(key)
	}
	return a.t2.Expiry(key)
}

func (a *arc) Purge() {
	a.t1.Purge()
	a.t2.Purge()
	a.b1.Purge()
	a.b2.Purge()
}

func (a *arc) Resize(size int) int {
	a.b1.Resize(size)
	a.b2.Resize(size)
	return a.t1.Resize(size) + a.t2.Resize(size)
}

func (a *arc) SetTTL(ttl time.Duration) {
	a.t1.SetTTL(ttl)
	a.t2.SetTTL(ttl)
}

func (a *arc) TTL() time.Duration {
	// Both T1 and T2 LRU have the same ttl.
	return a.t1.TTL()
}

func (a *arc) Len() int {
	return a.t1.Len() + a.t2.Len()
}

func (a *arc) Keys() []interface{} {
	return append(a.t1.Keys(), a.t2.Keys()...)
}

func (a *arc) Cap() int {
	// ALL sub LRU have the same capacity.
	return a.t1.Cap()
}

func (a *arc) Contains(key interface{}) bool {
	return a.t1.Contains(key) || a.t2.Contains(key)
}

func (a *arc) RegisterOnEvicted(f func(key interface{}, value interface{})) {
	a.t1.RegisterOnEvicted(f)
	a.t2.RegisterOnEvicted(f)
}

func (a *arc) RegisterOnExpired(f func(key interface{})) {
	a.t1.RegisterOnExpired(f)
	a.t2.RegisterOnExpired(f)
}
