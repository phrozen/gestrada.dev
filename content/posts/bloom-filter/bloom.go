package bloom

import (
	"hash"
	"hash/fnv"
	"math"
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
// We use a uint64 slice (`[]uint64`) because it provides maximum performance
// on modern 64-bit architectures due to memory alignment and CPU word size.
// Fetching a 64-bit word and masking a bit is incredibly fast.
type Filter struct {
	bitset []uint64
	m      uint // Total number of bits in the filter
	k      uint // Number of hash functions to apply per element
	hasher HashFunc
}

// #endregion filter

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

// #region math
// optimalM calculates the ideal total number of bits (m) needed to store
// `n` elements while maintaining a target false positive rate `p`.
// The formula is: m = - (n * ln(p)) / (ln(2)^2)
func optimalM(n uint, p float64) uint {
	return uint(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

// optimalK calculates the ideal number of hash functions (k) to use
// for a given number of bits (`m`) and expected elements (`n`).
// The formula is: k = (m / n) * ln(2)
func optimalK(m uint, n uint) uint {
	return uint(math.Ceil((float64(m) / float64(n)) * math.Log(2)))
}

// #endregion math

// #region newfilter
// NewFilter creates a Bloom filter directly with a specific bit size (m)
// and number of hash iteration (k).
func NewFilter(m uint, k uint, opts ...Option) *Filter {
	// Since each uint64 holds 64 bits, we divide `m` by 64.
	// We add 63 before dividing to ensure we round up.
	size := (m + 63) / 64

	f := &Filter{
		bitset: make([]uint64, size),
		m:      m,
		k:      k,
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
func NewFilterFromProbability(n uint, p float64, opts ...Option) *Filter {
	m := optimalM(n, p)
	k := optimalK(m, n)
	return NewFilter(m, k, opts...)
}

// #endregion newfilter

// #region add
// Add inserts data into the Bloom filter.
// It applies `k` hash functions, calculating `k` bit positions,
// and sets all those bits to 1.
func (f *Filter) Add(data []byte) {
	// Allocate the hasher once per Add operation
	h := f.hasher()

	for i := uint(0); i < f.k; i++ {
		// Calculate the raw hash values
		hash1, hash2 := f.hash(h, data)

		// Combine them using the standard Kirsch-Mitzenmacher optimization:
		// hash_i = hash1 + i * hash2
		hashVal := hash1 + uint64(i)*hash2

		// Map the hash to a specific bit index inside our `m` bounds
		bitIdx := hashVal % uint64(f.m)

		// Locate the exact uint64 word that contains this bit
		wordIdx := bitIdx / 64
		// Locate the specific bit within that 64-bit word (0 through 63)
		bitOffset := bitIdx % 64

		// Use bitwise OR (|) to set the specific bit to 1, leaving others unchanged
		f.bitset[wordIdx] |= 1 << bitOffset
	}
}

// #endregion add

// #region contains
// Contains checks if data might be in the set.
// If *any* of the `k` hash positions are 0, the element was definitively never added.
// If *all* of the positions are 1, it *might* have been added (or we have a collision).
func (f *Filter) Contains(data []byte) bool {
	// Allocate the hasher once per Contains operation
	h := f.hasher()

	for i := uint(0); i < f.k; i++ {
		hash1, hash2 := f.hash(h, data)
		hashVal := hash1 + uint64(i)*hash2

		bitIdx := hashVal % uint64(f.m)

		wordIdx := bitIdx / 64
		bitOffset := bitIdx % 64

		// Use bitwise AND (&) to isolate the specific bit.
		// If the result is 0, the bit was not set.
		if (f.bitset[wordIdx] & (1 << bitOffset)) == 0 {
			return false // Definitively not in the set
		}
	}
	return true // Probably in the set
}

// #endregion contains

// #region hash
// hash returns two 64-bit hashes for the given data.
// A Bloom filter requires `k` distinct hash functions. Instead of actually
// implementing `k` different algorithms, we use the Kirsch-Mitzenmacher
// optimization. We hash the data once with a 64-bit hasher, and split it
// into two 32-bit halves (h1, h2). We can then simulate `k` independent
// hash functions using the formula: hash_i = h1 + i * h2.
func (f *Filter) hash(h hash.Hash64, data []byte) (uint64, uint64) {
	h.Reset()
	h.Write(data)
	sum := h.Sum64()

	// Split the 64-bit hash into two 32-bit hashes
	h1 := sum & 0xffffffff
	h2 := sum >> 32

	return h1, h2
}

// #endregion hash
