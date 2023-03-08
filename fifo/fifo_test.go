package fifo

import (
	"container/list"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shaj13/libcache/internal"
)

func TestCollection(t *testing.T) {
	entries := []*internal.Entry{}
	entries = append(entries, &internal.Entry{Key: 1})
	entries = append(entries, &internal.Entry{Key: 2})
	entries = append(entries, &internal.Entry{Key: 3})

	c := &collection{ll: list.New()}
	c.Init()

	for _, e := range entries {
		c.Add(e)
	}

	for _, e := range entries {
		for i := 0; i < e.Key.(int); i++ {
			c.Move(e)
		}
	}

	front := c.Front()
	back := c.Back()
	oldest := c.Discard()
	c.Remove(entries[2])

	assert.Equal(t, 1, front.Key)
	assert.Equal(t, 3, back.Key)
	assert.Equal(t, 1, oldest.Key)
	assert.Equal(t, 1, c.Len())
	assert.Equal(t, 2, c.Back().Key)
}
