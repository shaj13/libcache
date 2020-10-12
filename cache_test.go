package memc_test

import (
	"math/rand"
	"testing"

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
