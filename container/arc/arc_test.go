package arc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRecentToFrequent(t *testing.T) {
	a := New(128).(*arc)

	// Touch all the entries, should be in t1
	for i := 0; i < 128; i++ {
		a.Store(i, i)
	}

	assert.Equal(t, 128, a.t1.Len())
	assert.Equal(t, 0, a.t2.Len())

	// Get should upgrade to t2
	for i := 0; i < 128; i++ {
		a.Load(i)
	}

	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 128, a.t2.Len())

	// Get be from t2
	for i := 0; i < 128; i++ {
		a.Load(i)
	}

	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 128, a.t2.Len())
}

func TestStoreRecentToFrequent(t *testing.T) {
	a := New(2).(*arc)

	// Add initially to t1
	a.Store(1, 1)
	assert.Equal(t, 1, a.t1.Len())
	assert.Equal(t, 0, a.t2.Len())

	// Add should upgrade to t2
	a.Store(1, 1)
	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())

	// Add should remain in t2
	a.Store(1, 1)
	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 1, a.t2.Len())
}

func TestARC_Adaptive(t *testing.T) {
	a := New(4).(*arc)

	// Fill t1
	for i := 0; i < 4; i++ {
		a.Store(i, i)
	}

	assert.Equal(t, 4, a.t1.Len())

	// Move to t2
	a.Load(0)
	a.Load(1)
	assert.Equal(t, 2, a.t2.Len())

	// Evict from t1
	a.Store(4, 4)
	assert.Equal(t, 1, a.b1.Len())

	// Current state
	// t1 : (MRU) [4, 3] (LRU)
	// t2 : (MRU) [1, 0] (LRU)
	// b1 : (MRU) [2] (LRU)
	// b2 : (MRU) [] (LRU)

	// Add 2, should cause hit on b1
	a.Store(2, 2)
	assert.Equal(t, 3, a.t2.Len())
	assert.Equal(t, 1, a.b1.Len())
	assert.Equal(t, 1, a.p)

	// Current state
	// t1 : (MRU) [4] (LRU)
	// t2 : (MRU) [2, 1, 0] (LRU)
	// b1 : (MRU) [3] (LRU)
	// b2 : (MRU) [] (LRU)

	// Add 4, should migrate to t2
	a.Store(4, 4)
	assert.Equal(t, 4, a.t2.Len())
	assert.Equal(t, 0, a.t1.Len())

	// Current state
	// t1 : (MRU) [] (LRU)
	// t2 : (MRU) [4, 2, 1, 0] (LRU)
	// b1 : (MRU) [3] (LRU)
	// b2 : (MRU) [] (LRU)

	// Add 4, should evict to b2
	a.Store(5, 5)
	assert.Equal(t, 3, a.t2.Len())
	assert.Equal(t, 1, a.t1.Len())
	assert.Equal(t, 1, a.b2.Len())

	// Current state
	// t1 : (MRU) [5] (LRU)
	// t2 : (MRU) [4, 2, 1] (LRU)
	// b1 : (MRU) [3] (LRU)
	// b2 : (MRU) [0] (LRU)

	// Add 0, should decrease p
	a.Store(0, 0)
	assert.Equal(t, 0, a.t1.Len())
	assert.Equal(t, 4, a.t2.Len())
	assert.Equal(t, 2, a.b1.Len())
	assert.Equal(t, 0, a.b2.Len())
	assert.Equal(t, 0, a.p)

	// Current state
	// t1 : (MRU) [] (LRU)
	// t2 : (MRU) [0, 4, 2, 1] (LRU)
	// b1 : (MRU) [5, 3] (LRU)
	// b2 : (MRU) [0] (LRU)
}
