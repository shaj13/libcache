package fifo

import (
	"github.com/shaj13/memc"
)

// Capacity set cache container capacity.
func Capacity(capacity int) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*fifo); ok {
			l.c.Capacity = capacity
		}
	})
}

// RegisterOnEvicted register OnEvicted callback.
func RegisterOnEvicted(cb memc.OnEvicted) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*fifo); ok {
			l.c.OnEvicted = cb
		}
	})
}

// RegisterOnExpired register OnExpired callback.
func RegisterOnExpired(cb memc.OnExpired) memc.Option {
	return memc.OptionFunc(func(c memc.Cache) {
		if l, ok := c.(*fifo); ok {
			l.c.OnExpired = cb
		}
	})
}
