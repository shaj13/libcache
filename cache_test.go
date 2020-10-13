package memc_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shaj13/memc"
	_ "github.com/shaj13/memc/container/fifo"
	_ "github.com/shaj13/memc/container/lfu"
	_ "github.com/shaj13/memc/container/lru"
)

var cachetest = []memc.Container{
	memc.LFU,
	memc.LRU,
	memc.FIFO,
}

func TestCacheStore(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheStore", func(t *testing.T) {
			cache := c.New(0)
			cache.Store(1, 1)
			assert.True(t, cache.Contains(1))
		})
	}
}

func TestCacheSet(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheSet", func(t *testing.T) {
			cache := c.New(0)
			cache.Set(1, 1, time.Nanosecond*10)
			time.Sleep(time.Nanosecond * 20)
			assert.False(t, cache.Contains(1))
		})
	}
}

func TestCacheLoad(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheLoad", func(t *testing.T) {
			cache := c.New(0)
			cache.Store("1", 1)
			v, ok := cache.Load("1")
			assert.True(t, ok)
			assert.Equal(t, 1, v)
		})
	}
}

func TestCacheDelete(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheDelete", func(t *testing.T) {
			cache := c.New(0)
			cache.Store(1, 1)
			cache.Delete(1)
			assert.False(t, cache.Contains(1))
		})
	}
}

func TestCachePeek(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CachePeek", func(t *testing.T) {
			cache := c.New(3)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			v, ok := cache.Peek(1)
			cache.Store(4, 0)
			found := cache.Contains(1)
			assert.Equal(t, 0, v)
			assert.True(t, ok)
			assert.False(t, found, "Peek should not move element")
		})
	}
}

func TestCacheContains(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheContains", func(t *testing.T) {
			cache := c.New(3)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			found := cache.Contains(1)
			cache.Store(4, 0)
			_, ok := cache.Load(1)
			assert.True(t, found)
			assert.False(t, ok, "Contains should not move element")
		})
	}
}

func TestCacheUpdate(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheUpdate", func(t *testing.T) {
			cache := c.New(3)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			cache.Update(1, 1)
			v, ok := cache.Peek(1)
			cache.Store(4, 0)
			found := cache.Contains(1)
			assert.Equal(t, 1, v)
			assert.True(t, ok)
			assert.False(t, found, "Update should not move element")
		})
	}
}

func TestCachePurge(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CachePurge", func(t *testing.T) {
			cache := c.New(0)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			cache.Purge()

			assert.Equal(t, 0, cache.Len())
		})
	}
}

func TestCacheResize(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheResize", func(t *testing.T) {
			cache := c.New(0)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			cache.Resize(2)
			assert.Equal(t, 2, cache.Len())
			assert.True(t, cache.Contains(2))
			assert.True(t, cache.Contains(3))
			assert.False(t, cache.Contains(1))
		})
	}
}

func TestCacheKeys(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheKeys", func(t *testing.T) {
			cache := c.New(0)
			cache.Store(1, 0)
			cache.Store(2, 0)
			cache.Store(3, 0)
			assert.ElementsMatch(t, []interface{}{1, 2, 3}, cache.Keys())
		})
	}
}

func TestCacheCap(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheCap", func(t *testing.T) {
			cache := c.New(3)
			assert.Equal(t, 3, cache.Cap())
		})
	}
}

func TestCacheTTL(t *testing.T) {
	for _, c := range cachetest {
		t.Run("Test"+c.String()+"CacheTTL", func(t *testing.T) {
			cache := c.New(0)
			cache.SetTTL(time.Second)
			assert.Equal(t, time.Second, cache.TTL())
		})
	}
}

func BenchmarkCache(b *testing.B) {
	for _, c := range cachetest {
		b.Run("Benchmark"+c.String()+"Cache", func(b *testing.B) {
			keys := []interface{}{}
			cache := c.New(0)

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
