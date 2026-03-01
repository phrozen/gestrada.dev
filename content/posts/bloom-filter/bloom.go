package bloom

import (
	"hash"
	"hash/fnv"
	"math"
	"sync/atomic"
)

// #region hashfunc
// HashFunc is a function that returns a new hash.Hash64.
// By defining an interface for our hasher, we avoid hardcoding a
// specific algorithm. This gives users the flexibility to swap
// the default hasher (FNV-1a) with a faster one (like Murmur3 or xxHash)
// if their use case requires maximum throughput.
type HashFunc func() hash.Hash64

// #endregion hashfunc

// #region filter
// Filter represents a Bloom filter.
// We use atomic.Uint64 so that Add and Contains are fully concurrent-safe
// without requiring an external mutex. On modern 64-bit architectures,
// atomic operations on aligned 64-bit words are essentially free.
type Filter struct {
	bitset []atomic.Uint64
	m      uint64 // Total number of bits in the filter
	k      uint64 // Number of hash functions to apply per element
	hasher HashFunc
}

// #endregion filter

// #region options
// Option configures the Filter using the functional options pattern.
// This pattern allows us to provide sensible defaults while
// remaining open for extension later, without breaking the API.
type Option func(*Filter)

// WithHashFunc allows providing a custom hash function constructor.
func WithHashFunc(h HashFunc) Option {
	return func(f *Filter) {
		f.hasher = h
	}
}

// #endregion options

// #region math
// optimalM calculates the ideal total number of bits (m) needed to store
// `n` elements while maintaining a target false positive rate `p`.
// The formula is: m = - (n * ln(p)) / (ln(2)^2)
func optimalM(n int, p float64) uint64 {
	return uint64(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

// optimalK calculates the ideal number of hash functions (k) to use
// for a given number of bits (`m`) and expected elements (`n`).
// The formula is: k = (m / n) * ln(2)
func optimalK(m uint64, n int) uint64 {
	return uint64(math.Ceil((float64(m) / float64(n)) * math.Log(2)))
}

// #endregion math

// #region newfilter
// NewFilter creates a Bloom filter directly with a specific bit size (m)
// and number of hash iterations (k).
func NewFilter(m int, k int, opts ...Option) *Filter {
	// Since each uint64 holds 64 bits, we divide `m` by 64.
	// We add 63 before dividing to ensure we round up.
	bits := uint64(m)
	size := (bits + 63) / 64
	f := &Filter{
		bitset: make([]atomic.Uint64, size),
		m:      bits,
		k:      uint64(k),
		hasher: fnv.New64a, // FNV-1a is in the standard library and computationally cheap.
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// NewFilterFromProbability creates a Bloom filter tailored for an expected
// number of elements (`n`) and a desired false-positive probability (`p`).
// Example: NewFilterFromProbability(10000, 0.01) creates a filter for 10k items
// with a 1% chance of a false positive.
func NewFilterFromProbability(n int, p float64, opts ...Option) *Filter {
	m := optimalM(n, p)
	k := optimalK(m, n)
	return NewFilter(int(m), int(k), opts...)
}

// #endregion newfilter

// #region add
// Add inserts data into the Bloom filter.
// It hashes the data once, splits the 64-bit result into two 32-bit halves,
// and simulates `k` independent hashes via Kirsch-Mitzenmacher.
// This method is safe for concurrent use.
func (f *Filter) Add(data []byte) {
	h := f.hasher()
	h.Write(data)
	sum := h.Sum64()

	// Split the 64-bit hash into two 32-bit halves for Kirsch-Mitzenmacher
	h1 := sum & 0xffffffff
	h2 := sum >> 32

	for i := range f.k {
		// Simulate k independent hashes: hash_i = h1 + i * h2
		bitIdx := (h1 + i*h2) % f.m

		// Atomically OR the bit to 1, safe for concurrent writes
		f.bitset[bitIdx/64].Or(1 << (bitIdx % 64))
	}
}

// #endregion add

// #region contains
// Contains checks if data might be in the set.
// If *any* of the `k` hash positions are 0, the element was definitively never added.
// If *all* of the positions are 1, it *might* have been added (or we have a collision).
// This method is safe for concurrent use.
func (f *Filter) Contains(data []byte) bool {
	h := f.hasher()
	h.Write(data)
	sum := h.Sum64()

	// Split the 64-bit hash into two 32-bit halves for Kirsch-Mitzenmacher
	h1 := sum & 0xffffffff
	h2 := sum >> 32

	for i := range f.k {
		bitIdx := (h1 + i*h2) % f.m

		// Atomically load the word and check the specific bit
		if (f.bitset[bitIdx/64].Load() & (1 << (bitIdx % 64))) == 0 {
			return false // Definitively not in the set
		}
	}
	return true // Probably in the set
}

// #endregion contains
