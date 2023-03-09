// Package fifo implements an FIFO cache.
package fifo

import (
	"container/list"

	"github.com/shaj13/libcache"
	"github.com/shaj13/libcache/internal"
)

func init() {
	libcache.FIFO.Register(New)
}

// New returns a new non-thread safe cache.
func New(cap int) libcache.Cache {
	col := &collection{list.New()}
	return internal.New(col, cap)
}

type collection struct {
	ll *list.List
}

func (c *collection) Move(e *internal.Entry) {}

func (c *collection) Add(e *internal.Entry) {
	le := c.ll.PushBack(e)
	e.Element = le
}

func (c *collection) Remove(e *internal.Entry) {
	le := e.Element.(*list.Element)
	c.ll.Remove(le)
}

func (c *collection) Discard() (e *internal.Entry) {
	if le := c.ll.Front(); le != nil {
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
