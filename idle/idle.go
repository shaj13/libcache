// Package idle implements an IDLE cache, that never finds/stores a key's value.
package idle

import (
	"time"

	"github.com/shaj13/libcache"
)

func init() {
	libcache.IDLE.Register(New)
}

// New return idle cache that never finds/stores a key's value.
func New(cap int) libcache.Cache {
	return idle{}
}

type idle struct{}

func (idle) Load(interface{}) (v interface{}, ok bool)            { return }
func (idle) Peek(interface{}) (v interface{}, ok bool)            { return }
func (idle) Keys() (keys []interface{})                           { return }
func (idle) Contains(interface{}) (ok bool)                       { return }
func (idle) Resize(int) (i int)                                   { return }
func (idle) Len() (len int)                                       { return }
func (idle) Cap() (cap int)                                       { return }
func (idle) TTL() (t time.Duration)                               { return }
func (idle) Expiry(interface{}) (t time.Time, ok bool)            { return }
func (idle) Update(interface{}, interface{})                      {}
func (idle) Store(interface{}, interface{})                       {}
func (idle) StoreWithTTL(interface{}, interface{}, time.Duration) {}
func (idle) Delete(interface{})                                   {}
func (idle) Purge()                                               {}
func (idle) SetTTL(ttl time.Duration)                             {}
func (idle) RegisterOnExpired(f func(key, value interface{}))     {}
func (idle) RegisterOnEvicted(f func(key, value interface{}))     {}
