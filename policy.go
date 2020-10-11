package libcache

import (
	"strconv"
	"sync"
)

const (
	// IDLE cache replacement policy.
	IDLE ReplacementPolicy = iota + 1
	// FIFO cache replacement policy.
	FIFO
	// LIFO cache replacement policy.
	LIFO
	// LRU cache replacement policy.
	LRU
	// LFU cache replacement policy.
	LFU
	// MRU cache replacement policy.
	MRU
	// ARC cache replacement policy.
	ARC
	max
)

var policies = make([]func(cap int) Cache, max)

// ReplacementPolicy identifies a cache replacement policy function that implemented in another package.
type ReplacementPolicy uint

// Register registers a function that returns a new cache instance,
// of the given cache replacement policy function.
// This is intended to be called from the init function,
// in packages that implement cache replacement policy function.
func (c ReplacementPolicy) Register(function func(cap int) Cache) {
	if c <= 0 && c >= max { //nolint:staticcheck
		panic("libcache: Register of unknown cache replacement policy function")
	}

	policies[c] = function
}

// Available reports whether the given cache replacement policy is linked into the binary.
func (c ReplacementPolicy) Available() bool {
	return c > 0 && c < max && policies[c] != nil
}

// New returns a new thread safe cache.
// New panics if the cache replacement policy function is not linked into the binary.
func (c ReplacementPolicy) New(cap int) Cache {
	cache := new(cache)
	cache.mu = sync.RWMutex{}
	cache.unsafe = c.NewUnsafe(cap)
	return cache
}

// NewUnsafe returns a new non-thread safe cache.
// NewUnsafe panics if the cache replacement policy function is not linked into the binary.
func (c ReplacementPolicy) NewUnsafe(cap int) Cache {
	if !c.Available() {
		panic("libcache: Requested cache replacement policy function #" + strconv.Itoa(int(c)) + " is unavailable")
	}

	return policies[c](cap)
}

// String returns string describes the cache replacement policy function.
func (c ReplacementPolicy) String() string {
	switch c {
	case IDLE:
		return "IDLE"
	case FIFO:
		return "FIFO"
	case LIFO:
		return "LIFO"
	case LRU:
		return "LRU"
	case LFU:
		return "LFU"
	case MRU:
		return "MRU"
	case ARC:
		return "ARC"
	default:
		return "unknown cache replacement policy value " + strconv.Itoa(int(c))
	}

}
