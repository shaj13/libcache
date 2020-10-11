package memc_test

import (
	"fmt"
	"time"

	"github.com/shaj13/memc"
	_ "github.com/shaj13/memc/container/fifo"
	_ "github.com/shaj13/memc/container/idle"
	_ "github.com/shaj13/memc/container/lfu"
	_ "github.com/shaj13/memc/container/lru"
)

func Example_idle() {
	//  it can be unsafe, no any race conditions
	c := memc.IDLE.NewUnsafe()
	c.Store(1, 0)
	fmt.Println(c.Contains(1))
	// Output:
	// false
}

func Example_fifo() {
	c := memc.FIFO.New()
	c.Store(1, 0)
	c.Store(2, 0)
	c.Store(3, 0)
	fmt.Println(c.Contains(1))
	// Output:
	// false
}

func Example_lru() {
	c := memc.LRU.New()
	c.Store(1, 0)
	c.Store(2, 0)
	c.Store(3, 0)
	fmt.Println(c.Contains(1))
	// Output:
	// false
}

func Example_lfu() {
	c := memc.LFU.New()
	c.Store(1, 0)
	c.Store(2, 0)
	c.Load(1)
	c.Store(3, 0)
	fmt.Println(c.Contains(2))
	// Output:
	// false
}

func Example_onexpired() {
	// c must be thread safe
	var c memc.Cache

	exp := func(key interface{}) {
		// use Peek/Load over delete, perhaps a new entry added with the same key during expiration,
		// or entry refreshed from other thread.
		c.Peek(key)
	}

	c = memc.LRU.New()
	c.RegisterOnExpired(exp)
	c.SetTTL(time.Millisecond)
	c.Store(1, 0)

	time.Sleep(time.Millisecond * 5)
	fmt.Println(c.Contains(1))
	// Output:
	// false
}
