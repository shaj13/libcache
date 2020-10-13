// Package lru implements an LRU cache.
package lru

import (
	"container/list"

	"github.com/shaj13/memc"
	"github.com/shaj13/memc/internal"
)

func init() {
	memc.LRU.Register(New)
}

// New returns new thread unsafe cache container.
func New(cap int) memc.Cache {
	col := &collection{list.New()}
	return internal.New(col, cap)
}

type collection struct {
	ll *list.List
}

func (c *collection) Move(e *internal.Entry) {
	le := e.Element.(*list.Element)
	c.ll.MoveToFront(le)
}

func (c *collection) Add(e *internal.Entry) {
	le := c.ll.PushFront(e)
	e.Element = le
}

func (c *collection) Remove(e *internal.Entry) {
	le := e.Element.(*list.Element)
	c.ll.Remove(le)
}

func (c *collection) RemoveOldest() (e *internal.Entry) {
	if le := c.ll.Back(); le != nil {
		c.ll.Remove(le)
		e = le.Value.(*internal.Entry)
	}
	return
}

func (c *collection) Len() int {
	return c.ll.Len()
}

func (c *collection) Init() {
	c.ll.Init()
}
