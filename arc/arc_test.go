package arc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestARCc(t *testing.T) {
	a := New(2).(*arc)

	a.Store(1, 1)
	a.Store(2, 2)
	assert.Equal(t, 2, a.t1.Len())
	assert.Equal(t, 0, a.t2.Len())
	assert.Equal(t, 0, a.b1.Len())
	assert.Equal(t, 0, a.b2.Len())

	a.Load(1)
	assert.Equal(t, 1, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())
	assert.Equal(t, 0, a.b1.Len())
	assert.Equal(t, 0, a.b2.Len())

	a.Store(3, 3)
	assert.Equal(t, 1, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())
	assert.Equal(t, 1, a.b1.Len())
	assert.Equal(t, 0, a.b2.Len())

	a.Store(2, 2)
	assert.Equal(t, 1, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())
	assert.Equal(t, 0, a.b1.Len())
	assert.Equal(t, 1, a.b2.Len())

	a.Store(1, 1)
	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 2, a.t2.Len())
	assert.Equal(t, 1, a.b1.Len())
	assert.Equal(t, 0, a.b2.Len())

	a.Purge()
	a.Resize(1)

	a.Store(1, 1)
	assert.Equal(t, 1, a.t1.Len())
	assert.Equal(t, 0, a.t2.Len())

	a.Store(1, 1)
	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())

	a.Store(1, 1)
	a.Load(1)

	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())
	assert.Equal(t, 1, a.Front())
	assert.Equal(t, 1, a.Back())

	a.Delete(1)
}
