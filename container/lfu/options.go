package lfu

import (
	"time"

	"github.com/shaj13/memc"
)

// TTL set cache container entries TTL.
func TTL(ttl time.Duration) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*lfu); ok {
			l.c.TTL = ttl
		}
	})
}

// Capacity set cache container capacity.
func Capacity(capacity int) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*lfu); ok {
			l.c.Capacity = capacity
		}
	})
}

// RegisterOnEvicted register OnEvicted callback.
func RegisterOnEvicted(cb memc.OnEvicted) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*lfu); ok {
			l.c.OnEvicted = cb
		}
	})
}

// RegisterOnExpired register OnExpired callback.
func RegisterOnExpired(cb memc.OnExpired) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*lfu); ok {
			l.c.OnExpired = cb
		}
	})
}
