package lfu

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shaj13/memc"
)

func TestStore(t *testing.T) {
	lfu := New()
	lfu.Store(1, 1)
	ok := lfu.Contains(1)
	assert.True(t, ok)
}

func TestSet(t *testing.T) {
	lfu := New()
	lfu.Set(1, 1, time.Nanosecond*10)
	time.Sleep(time.Nanosecond * 20)
	ok := lfu.Contains(1)
	assert.False(t, ok)
}

func TestLoad(t *testing.T) {
	lfu := New()
	lfu.Store("1", 1)
	v, ok := lfu.Load("1")
	assert.True(t, ok)
	assert.Equal(t, 1, v)
}

func TestDelete(t *testing.T) {
	lfu := New()
	lfu.Store(1, 1)
	lfu.Delete(1)
	ok := lfu.Contains(1)
	assert.False(t, ok)
}

func TestPeek(t *testing.T) {
	lfu := New().(*lfu)
	lfu.c.Capacity = 3

	lfu.Store(1, 0)
	lfu.Store(2, 0)
	lfu.Store(3, 0)
	v, ok := lfu.Peek(1)
	lfu.Store(4, 0)
	found := lfu.Contains(1)

	assert.Equal(t, 0, v)
	assert.True(t, ok)
	assert.False(t, found, "Peek should not move element")
}

func TestContains(t *testing.T) {
	lfu := New().(*lfu)
	lfu.c.Capacity = 3

	lfu.Store(1, 0)
	lfu.Store(2, 0)
	lfu.Store(3, 0)
	found := lfu.Contains(1)
	lfu.Store(4, 0)
	_, ok := lfu.Load(1)

	assert.True(t, found)
	assert.False(t, ok, "Contains should not move element")
}

func TestUpdate(t *testing.T) {
	lfu := New().(*lfu)
	lfu.c.Capacity = 3

	lfu.Store(1, 0)
	lfu.Store(2, 0)
	lfu.Store(3, 0)
	lfu.Update(1, 1)
	v, ok := lfu.Peek(1)
	lfu.Store(4, 0)
	found := lfu.Contains(1)

	assert.Equal(t, 1, v)
	assert.True(t, ok)
	assert.False(t, found, "Update should not move element")
}

func TestPurge(t *testing.T) {
	lfu := New().(*lfu)
	lfu.c.Capacity = 3

	lfu.Store(1, 0)
	lfu.Store(2, 0)
	lfu.Store(3, 0)
	lfu.Purge()

	assert.Equal(t, 0, lfu.Len())
}

func TestResize(t *testing.T) {
	lfu := New().(*lfu)
	lfu.c.Capacity = 3

	lfu.Store(1, 0)
	lfu.Store(2, 0)
	lfu.Store(3, 0)
	lfu.Resize(2)

	assert.Equal(t, 2, lfu.Len())
	assert.True(t, lfu.Contains(3))
	assert.True(t, lfu.Contains(2))
	assert.False(t, lfu.Contains(1))
}

func TestKeys(t *testing.T) {
	lfu := New()

	lfu.Store(1, 0)
	lfu.Store(2, 0)
	lfu.Store(3, 0)

	assert.ElementsMatch(t, []interface{}{1, 2, 3}, lfu.Keys())
}

func TestCap(t *testing.T) {
	lfu := New().(*lfu)
	lfu.c.Capacity = 3
	assert.Equal(t, 3, lfu.Cap())
}

func TestLFU(t *testing.T) {
	lfu := New().(*lfu)
	lfu.Store(1, 1)
	lfu.Store(2, 2)
	lfu.Store(3, 3)

	for _, k := range lfu.Keys() {
		v, _ := lfu.Peek(k)
		for i := 0; i < v.(int); i++ {
			lfu.Load(k)
		}
	}

	lfu.Resize(2)

	assert.False(t, lfu.Contains(1))
}

func TestOnEvicted(t *testing.T) {
	send := make(chan interface{})
	done := make(chan bool)

	evictedKeys := make([]interface{}, 0, 2)

	onEvictedFun := func(key, value interface{}) {
		send <- key
	}

	lfu := New().(*lfu)
	lfu.c.Capacity = 20
	lfu.c.OnEvicted = onEvictedFun

	go func() {
		for {
			key := <-send
			evictedKeys = append(evictedKeys, key)
			if len(evictedKeys) >= 2 {
				done <- true
				return
			}
		}
	}()

	for i := 0; i < 22; i++ {
		lfu.Store(fmt.Sprintf("myKey%d", i), i)
		for ii := 0; ii < i; ii++ {
			lfu.Load(fmt.Sprintf("myKey%d", i))
		}
	}

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatal("TestOnEvicted timeout exceeded, expected to receive evicted keys")
	}

	assert.ElementsMatch(t, []interface{}{"myKey0", "myKey1"}, evictedKeys)
}

func TestOnExpired(t *testing.T) {
	send := make(chan interface{})
	done := make(chan bool)

	expiredKeys := make([]interface{}, 0, 2)

	onExpiredFun := func(key interface{}) {
		send <- key
	}

	lfu := New().(*lfu)
	lfu.c.OnExpired = onExpiredFun
	lfu.SetTTL(time.Millisecond)

	go func() {
		for {
			key := <-send
			expiredKeys = append(expiredKeys, key)
			if len(expiredKeys) >= 2 {
				done <- true
				return
			}
		}
	}()

	lfu.Store(1, 1234)
	lfu.Store(2, 1234)

	select {
	case <-done:
	case <-time.After(time.Second * 2):
		t.Fatal("TestOnExpired timeout exceeded, expected to receive expired keys")
	}

	assert.ElementsMatch(t, []interface{}{1, 2}, expiredKeys)
}

func BenchmarkLFU(b *testing.B) {
	keys := []interface{}{}
	lfu := memc.LFU.New()

	for i := 0; i < 100; i++ {
		keys = append(keys, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := keys[rand.Intn(100)]
			_, ok := lfu.Load(key)
			if ok {
				lfu.Delete(key)
			} else {
				lfu.Store(key, struct{}{})
			}
		}
	})
}
