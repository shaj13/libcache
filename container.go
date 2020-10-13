package libcache

import (
	"strconv"
	"sync"
)

const (
	// IDLE cache container.
	IDLE Container = iota + 1
	// FIFO cache container.
	FIFO
	// LRU cache container.
	LRU
	// LFU cache container.
	LFU
	// MRU cache container.
	MRU
	max
)

var containers = make([]func(cap int) Cache, max)

// Container identifies a cache container that implemented in another package.
type Container uint

// Register registers a function that returns a new instance,
// of the given cache container function.
// This is intended to be called from the init function in packages that implement container functions.
func (c Container) Register(function func(cap int) Cache) {
	if c <= 0 && c >= max { //nolint:staticcheck
		panic("cache: Register of unknown cache container function")
	}

	containers[c] = function
}

// Available reports whether the given cache container is linked into the binary.
func (c Container) Available() bool {
	return c > 0 && c < max && containers[c] != nil
}

// New returns a new thread safe cache.
// New panics if the cache container function is not linked into the binary.
func (c Container) New(cap int) Cache {
	cache := new(cache)
	cache.mu = sync.RWMutex{}
	cache.container = c.NewUnsafe(cap)
	return cache
}

// NewUnsafe returns a new non-thread safe cache.
// NewUnsafe panics if the cache container function is not linked into the binary.
func (c Container) NewUnsafe(cap int) Cache {
	if !c.Available() {
		panic("cache: Requested cache container function #" + strconv.Itoa(int(c)) + " is unavailable")
	}

	return containers[c](cap)
}

// String returns string describes the container function.
func (c Container) String() string {
	switch c {
	case IDLE:
		return "IDLE"
	case FIFO:
		return "FIFO"
	case LRU:
		return "LRU"
	case LFU:
		return "LFU"
	case MRU:
		return "MRU"
	default:
		return "unknown container value " + strconv.Itoa(int(c))
	}

}
