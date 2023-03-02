[![PkgGoDev](https://pkg.go.dev/badge/github.com/shaj13/libcache@v1.0.0)](https://pkg.go.dev/github.com/shaj13/libcache@v1.0.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/shaj13/libcache)](https://goreportcard.com/report/github.com/shaj13/libcache)
[![Coverage Status](https://coveralls.io/repos/github/shaj13/libcache/badge.svg?branch=master)](https://coveralls.io/github/shaj13/libcache?branch=master)
[![CircleCI](https://circleci.com/gh/shaj13/libcache/tree/master.svg?style=svg)](https://circleci.com/gh/shaj13/libcache/tree/master)

# Libcache
A Lightweight in-memory key:value cache library for Go. 

## Introduction 
Caches are tremendously useful in a wide variety of use cases.<br>
you should consider using caches when a value is expensive to compute or retrieve,<br>
and you will need its value on a certain input more than once.<br>
libcache is here to help with that.

Libcache are local to a single run of your application.<br>
They do not store data in files, or on outside servers.

Libcache previously an [go-guardian](https://github.com/shaj13/go-guardian) package and designed to be a companion with it.<br>
While both can operate completely independently.<br>


## Features
- Rich [caching API](https://pkg.go.dev/github.com/shaj13/libcache@v1.0.0#Cache)
- Maximum cache size enforcement
- Default cache TTL (time-to-live) as well as custom TTLs per cache entry
- Thread safe as well as non-thread safe
- Event-Driven callbacks ([Notify](https://pkg.go.dev/github.com/shaj13/libcache@v1.0.0#Cache))
- Dynamic cache creation
- Multiple cache replacement policies:
  - FIFO (First In, First Out)
  - LIFO (Last In, First Out)
  - LRU (Least Recently Used)
  - MRU (Most Recently Used)
  - LFU (Least Frequently Used)
  - ARC (Adaptive Replacement Cache)

## Quickstart 
### Installing 
Using libcache is easy. First, use go get to install the latest version of the library.

```sh
go get github.com/shaj13/libcache
```
Next, include libcache in your application:
```go
import (
    _ "github.com/shaj13/libcache/<desired-replacement-policy>"
    "github.com/shaj13/libcache"
)
```

### Examples
**Note:** All examples use the LRU cache replacement policy for simplicity, any other cache replacement policy can be applied to them.
#### Basic 
```go
package main 
import (
    "fmt" 

    "github.com/shaj13/libcache"
    _ "github.com/shaj13/libcache/lru"
)

func main() {
    size := 10
    cache := libcache.LRU.NewUnsafe(size)
    for i:= 0 ; i < 10 ; i++ {
        cache.Store(i, i)
    }
    fmt.Println(cache.Load(0)) // nil, false  
    fmt.Println(cache.Load(1)) // 1, true
}
```

#### Thread Safe 
```go
package main

import (
	"fmt"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

func main() {
	done := make(chan struct{})

	f := func(c libcache.Cache) {
		for !c.Contains(5) {
		}
		fmt.Println(c.Load(5)) // 5, true
		done <- struct{}{}
	}

	size := 10
	cache := libcache.LRU.New(size)
	go f(cache)

	for i := 0; i < 10; i++ {
		cache.Store(i, i)
	}

	<-done
}
```
#### Unlimited Size
zero capacity means cache has no limit and replacement policy turned off.
```go
package main 
import (
    "fmt" 

    "github.com/shaj13/libcache"
    _ "github.com/shaj13/libcache/lru"
)

func main() {
	cache := libcache.LRU.New(0)
    for i:= 0 ; i < 100000 ; i++ {
        cache.Store(i, i)
    }
	fmt.Println(cache.Load(55555))
}
```
#### TTL
```go
package main 
import (
	"fmt"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

func main() {
	cache := libcache.LRU.New(10)
	cache.SetTTL(time.Second) // default TTL 
	
	for i:= 0 ; i < 10 ; i++ {
        cache.Store(i, i)
	}
	fmt.Println(cache.Expiry(1))

	cache.StoreWithTTL("mykey", "value", time.Hour) // TTL per cache entry 
	fmt.Println(cache.Expiry("mykey"))

}
```

#### Events 
```go
package main 
import (
	"fmt"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

func main() {
	cache := libcache.LRU.New(10)

	eventc := make(chan libcache.Event, 10)
	cache.Notify(eventc)
	defer cache.Ignore(eventc)

	go func() {
		for {
			e := <-eventc
			fmt.Printf("Operation %s on Key %v \n", e.Op, e.Key)
		}
	}()

	cache.Load(1)
	cache.Store(1, 1)
	cache.Peek(1)
	cache.Delete(1)
}
```
#### GC 
```go
package main 
import (
	"fmt"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

func main() {
	cache := libcache.LRU.New(10)

	eventc := make(chan libcache.Event, 10)
	cache.Notify(eventc)
	defer cache.Ignore(eventc)

	go func() {
		for {
			e := <-eventc
			fmt.Printf("Operation %s on Key %v \n", e.Op, e.Key)
		}
	}()

	ctx, cacnel := context.WithTimeout(context.Background(), time.Second*2)
	defer cacnel()

	cache.StoreWithTTL(1, 1, time.Second)

	// GC is a long running function, evict expired items from the cache on time.
	libcache.GC(ctx, cache)

	cache.StoreWithTTL(1, 1, time.Second)
	time.Sleep(time.Second)

	// Runs a garbage collection and blocks the caller until the garbage collection is complete
	cache.GC()
}
```


# Contributing
1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/libcache && cd libcache`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

# License
Libcache is released under the MIT license. See [LICENSE](https://github.com/shaj13/libcache/blob/master/LICENSE)
