package fifo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapacity(t *testing.T) {
	opt := Capacity(100)
	fifo := New(opt).(*fifo)

	assert.Equal(t, fifo.c.Capacity, 100)
}

func TestRegisterOnEvicted(t *testing.T) {
	opt := RegisterOnEvicted(func(key, value interface{}) {})
	fifo := New(opt).(*fifo)

	assert.NotNil(t, fifo.c.OnEvicted)
}

func TestRegisterOnExpired(t *testing.T) {
	opt := RegisterOnExpired(func(key interface{}) {})
	fifo := New(opt).(*fifo)

	assert.NotNil(t, fifo.c.OnExpired)
}
