package lfu

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shaj13/memc/internal"
)

func TestFrequently(t *testing.T) {
	entries := []*internal.Entry{}
	entries = append(entries, &internal.Entry{Key: 1})
	entries = append(entries, &internal.Entry{Key: 2})
	entries = append(entries, &internal.Entry{Key: 3})

	f := &frequently{}
	f.Init()

	for _, e := range entries {
		f.Add(e)
	}

	for _, e := range entries {
		for i := 0; i < e.Key.(int); i++ {
			f.Move(e)
		}
	}

	oldest := f.GetOldest()
	f.Remove(entries[2])

	assert.Equal(t, oldest.Key, 1)
	assert.Equal(t, f.Len(), 1)
	assert.Equal(t, (*f)[0].value.Key, 2)
}
