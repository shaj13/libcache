package lfu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapacity(t *testing.T) {
	opt := Capacity(100)
	lfu := New(opt).(*lfu)

	assert.Equal(t, lfu.c.Capacity, 100)
}

func TestRegisterOnEvicted(t *testing.T) {
	opt := RegisterOnEvicted(func(key, value interface{}) {})
	lfu := New(opt).(*lfu)

	assert.NotNil(t, lfu.c.OnEvicted)
}

func TestRegisterOnExpired(t *testing.T) {
	opt := RegisterOnExpired(func(key interface{}) {})
	lfu := New(opt).(*lfu)

	assert.NotNil(t, lfu.c.OnExpired)
}
