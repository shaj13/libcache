// Package arc implements an ARC cache.
package arc

import (
	"time"

	"github.com/shaj13/libcache"
	"github.com/shaj13/libcache/internal"
	"github.com/shaj13/libcache/lru"
)

func init() {
	libcache.ARC.Register(New)
}

// New returns a new non-thread safe cache.
func New(cap int) libcache.Cache {
	return &arc{
		p:  0,
		t1: lru.New(cap).(*internal.Cache),
		b1: lru.New(cap).(*internal.Cache),
		t2: lru.New(cap).(*internal.Cache),
		b2: lru.New(cap).(*internal.Cache),
	}
}

type arc struct {
	p  int
	t1 *internal.Cache
	t2 *internal.Cache
	b1 *internal.Cache
	b2 *internal.Cache
}

func (a *arc) Load(key interface{}) (value interface{}, ok bool) {
	if val, ok := a.t1.Peek(key); ok {
		exp, _ := a.t1.Expiry(key)
		a.t1.DelSilently(key)
		a.t2.StoreWithTTL(key, val, time.Until(exp))
		return val, ok
	}

	return a.t2.Load(key)
}

func (a *arc) Store(key, val interface{}) {
	a.StoreWithTTL(key, val, a.TTL())
}

func (a *arc) StoreWithTTL(key, val interface{}, ttl time.Duration) {
	defer func() {
		if a.Cap() != 0 && a.t1.Len()+a.t2.Len() > a.Cap() {
			a.replace(key)
		}
	}()

	if a.t1.Contains(key) {
		a.t1.DelSilently(key)
		a.t2.StoreWithTTL(key, val, ttl)
		return
	}

	if a.t2.Contains(key) {
		a.t2.StoreWithTTL(key, val, ttl)
		return
	}

	if a.b1.Contains(key) {
		a.p = min(a.Cap(), a.p+max(a.b2.Len()/a.b1.Len(), 1))
		a.b1.Delete(key)
		a.t2.StoreWithTTL(key, val, ttl)
		return
	}

	if a.b2.Contains(key) {
		a.p = max(0, a.p-max(a.b1.Len()/a.b2.Len(), 1))
		a.b2.Delete(key)
		a.t2.StoreWithTTL(key, val, ttl)
		return
	}

	if a.b1.Len() > a.Cap()-a.p {
		a.b1.Discard()
	}

	if a.b2.Len() > a.p {
		a.b2.Discard()
	}

	a.t1.StoreWithTTL(key, val, ttl)
}

func (a *arc) replace(key interface{}) {
	if (a.t1.Len() > 0 && a.b2.Contains(key) && a.t1.Len() == a.p) || (a.t1.Len() > a.p) {
		k, _ := a.t1.Discard()
		a.b1.Store(k, nil)
		return
	}

	k, _ := a.t2.Discard()
	a.b2.Store(k, nil)
}

func (a *arc) Delete(key interface{}) {
	a.t1.Delete(key)
	a.t2.Delete(key)
	a.b1.Delete(key)
	a.b2.Delete(key)
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

func (a *arc) RegisterOnEvicted(f func(key, value interface{})) {
	a.t1.RegisterOnEvicted(f)
	a.t2.RegisterOnEvicted(f)
}

func (a *arc) RegisterOnExpired(f func(key, value interface{})) {
	a.t1.RegisterOnExpired(f)
	a.t2.RegisterOnExpired(f)
}

func (a *arc) Notify(ch chan<- libcache.Event, ops ...libcache.Op) {
	a.t1.Notify(ch, ops...)
	a.t2.Notify(ch, ops...)
}

func (a *arc) Ignore(ch chan<- libcache.Event, ops ...libcache.Op) {
	a.t1.Ignore(ch, ops...)
	a.t2.Ignore(ch, ops...)
}

func (a *arc) GC() time.Duration {
	x := a.t1.GC()
	y := a.t2.GC()

	// return the next nearer gc cycle.
	if y == 0 {
		return x
	} else if x == 0 {
		return y
	} else if x < y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
