package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapacity(t *testing.T) {
	opt := Capacity(100)
	lru := New(opt).(*lru)

	assert.Equal(t, lru.c.Capacity, 100)
}

func TestRegisterOnEvicted(t *testing.T) {
	opt := RegisterOnEvicted(func(key, value interface{}) {})
	lru := New(opt).(*lru)

	assert.NotNil(t, lru.c.OnEvicted)
}

func TestRegisterOnExpired(t *testing.T) {
	opt := RegisterOnExpired(func(key interface{}) {})
	lru := New(opt).(*lru)

	assert.NotNil(t, lru.c.OnExpired)
}
