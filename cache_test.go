package libcache_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/arc"
	_ "github.com/shaj13/libcache/fifo"
	_ "github.com/shaj13/libcache/lfu"
	_ "github.com/shaj13/libcache/lifo"
	_ "github.com/shaj13/libcache/lru"
	_ "github.com/shaj13/libcache/mru"
)

var cacheTests = []struct {
	cont          libcache.ReplacementPolicy
	evictedKey    interface{}
	onEvictedKeys interface{}
}{
	{
		cont:          libcache.LFU,
		evictedKey:    1,
		onEvictedKeys: []interface{}{0, 19},
	},
	{
		cont:          libcache.LRU,
		evictedKey:    1,
		onEvictedKeys: []interface{}{0, 1},
	},
	{
		cont:          libcache.FIFO,
		evictedKey:    1,
		onEvictedKeys: []interface{}{0, 1},
	},
	{
		cont:          libcache.LIFO,
		evictedKey:    3,
		onEvictedKeys: []interface{}{20, 19},
	},
	{
		cont:          libcache.MRU,
		evictedKey:    3,
		onEvictedKeys: []interface{}{20, 19},
	},
	{
		cont:          libcache.ARC,
		evictedKey:    1,
		onEvictedKeys: []interface{}{0, 1},
	},
}

func TestCacheStore(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheStore", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.Store(1, 1)
			assert.True(t, cache.Contains(1))
		})
	}
}

func TestCacheStoreWithTTL(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheSet", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.StoreWithTTL(1, 1, time.Hour)
			got, ok := cache.Expiry(1)
			expect := time.Now().UTC().Add(time.Hour)
			assert.True(t, ok)
			assert.WithinDuration(t, expect, got, time.Hour)
		})
	}
}

func TestCacheLoad(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheLoad", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.Store("1", 1)
			v, ok := cache.Load("1")
			assert.True(t, ok)
			assert.Equal(t, 1, v)
		})
	}
}

func TestCacheDelete(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheDelete", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.Store(1, 1)
			cache.Delete(1)
			assert.False(t, cache.Contains(1))
		})
	}
}

func TestCachePeek(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CachePeek", func(t *testing.T) {
			cache := tt.cont.New(3)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			v, ok := cache.Peek(1)
			cache.Store(4, 0)
			found := cache.Contains(tt.evictedKey)
			assert.Equal(t, 0, v)
			assert.True(t, ok)
			assert.False(t, found, "Peek should not update recent-ness")
		})
	}
}

func TestCacheContains(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheContains", func(t *testing.T) {
			cache := tt.cont.New(3)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			found := cache.Contains(1)
			cache.Store(4, 0)
			_, ok := cache.Load(tt.evictedKey)
			assert.True(t, found)
			assert.False(t, ok, "Contains should not update recent-ness")
		})
	}
}

func TestCacheUpdate(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheUpdate", func(t *testing.T) {
			cache := tt.cont.New(3)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			cache.Update(1, 1)
			v, ok := cache.Peek(1)
			cache.Store(4, 0)
			found := cache.Contains(tt.evictedKey)
			assert.Equal(t, 1, v)
			assert.True(t, ok)
			assert.False(t, found, "Update should not move element")
		})
	}
}

func TestCachePurge(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CachePurge", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			cache.Purge()

			assert.Equal(t, 0, cache.Len())
		})
	}
}

func TestCacheResize(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheResize", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			cache.Resize(2)
			assert.Equal(t, 2, cache.Len())
		})
	}
}

func TestCacheKeys(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheKeys", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			assert.ElementsMatch(t, []interface{}{1, 2, 3}, cache.Keys())
		})
	}
}

func TestCacheCap(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheCap", func(t *testing.T) {
			cache := tt.cont.New(3)
			assert.Equal(t, 3, cache.Cap())
		})
	}
}

func TestCacheTTL(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheTTL", func(t *testing.T) {
			cache := tt.cont.New(0)
			cache.SetTTL(time.Second)
			assert.Equal(t, time.Second, cache.TTL())
		})
	}
}

func TestOnEvicted(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheOnEvicted", func(t *testing.T) {
			cache := tt.cont.New(20)
			send := make(chan interface{})
			done := make(chan bool)
			evictedKeys := make([]interface{}, 0, 2)
			cache.RegisterOnEvicted(func(key, value interface{}) {
				send <- key
			})

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
				cache.Store(i, i)
			}

			select {
			case <-done:
			case <-time.After(time.Second * 2):
				t.Fatal("TestOnEvicted timeout exceeded, expected to receive evicted keys")
			}

			assert.ElementsMatch(t, tt.onEvictedKeys, evictedKeys)
		})
	}
}

func TestOnExpired(t *testing.T) {
	for _, tt := range cacheTests {
		t.Run("Test"+tt.cont.String()+"CacheOnExpired", func(t *testing.T) {
			send := make(chan interface{})
			done := make(chan bool)
			expiredKeys := make([]interface{}, 0, 2)
			cache := tt.cont.New(0)
			cache.RegisterOnExpired(func(key, _ interface{}) {
				send <- key
			})
			cache.SetTTL(time.Millisecond)

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

			cache.Store(1, 1234)
			cache.Store(2, 1234)

			select {
			case <-done:
			case <-time.After(time.Second * 2):
				t.Fatal("TestOnExpired timeout exceeded, expected to receive expired keys")
			}

			assert.ElementsMatch(t, []interface{}{1, 2}, expiredKeys)
		})
	}
}

func BenchmarkCache(b *testing.B) {
	for _, tt := range cacheTests {
		b.Run("Benchmark"+tt.cont.String()+"Cache", func(b *testing.B) {
			keys := []interface{}{}
			cache := tt.cont.New(0)

			for i := 0; i < 100; i++ {
				keys = append(keys, i)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					key := keys[rand.Intn(100)]
					_, ok := cache.Load(key)
					if ok {
						cache.Delete(key)
					} else {
						cache.Store(key, struct{}{})
					}
				}
			})
		})
	}
}
