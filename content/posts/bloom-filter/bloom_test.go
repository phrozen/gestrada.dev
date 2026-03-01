package bloom

import (
	"crypto/rand"
	"fmt"
	"hash"
	"hash/fnv"
	"testing"

	"github.com/cespare/xxhash/v2"
	"github.com/twmb/murmur3"
)

func TestFilter_AddAndContains(t *testing.T) {
	// Create a small filter for testing
	f := NewFilter(100, 3)

	// Test adding and retrieving a single item
	item1 := []byte("apple")
	if f.Contains(item1) {
		t.Errorf("Empty filter should not contain 'apple'")
	}

	f.Add(item1)
	if !f.Contains(item1) {
		t.Errorf("Filter should contain 'apple' after it was added")
	}

	// Test adding multiple items
	items := [][]byte{
		[]byte("banana"),
		[]byte("cherry"),
		[]byte("date"),
	}

	for _, item := range items {
		f.Add(item)
	}

	// Verify all added items are present
	for _, item := range items {
		if !f.Contains(item) {
			t.Errorf("Filter should contain '%s'", item)
		}
	}

	// Verify an unadded item is not present (might fail due to false positive,
	// but highly unlikely with this capacity and item count)
	unadded := []byte("elderberry")
	if f.Contains(unadded) {
		t.Errorf("Filter should not contain '%s' (or we got a very unlucky false positive)", unadded)
	}
}

func TestFilter_EmptyData(t *testing.T) {
	f := NewFilter(100, 3)
	empty := []byte("")

	if f.Contains(empty) {
		t.Errorf("Empty filter should not contain empty string")
	}

	f.Add(empty)

	if !f.Contains(empty) {
		t.Errorf("Filter should contain empty string after it was added")
	}
}

func TestFilter_CustomHasher(t *testing.T) {
	f := NewFilter(100, 3, WithHashFunc(func() hash.Hash64 { return xxhash.New() }))
	item := []byte("xxhash-test")
	f.Add(item)
	if !f.Contains(item) {
		t.Errorf("Filter with custom hasher should contain item after it was added")
	}
}

func TestFilter_FalsePositiveRate(t *testing.T) {
	// Let's test the probability math
	// We want to store 10,000 items with a 1% false positive rate
	n := 10000
	p := 0.01

	f := NewFilterFromProbability(n, p)

	// 1. Add 10,000 random items
	addedItems := make([][]byte, n)
	for i := range n {
		b := make([]byte, 16)
		rand.Read(b)
		addedItems[i] = b
		f.Add(b)
	}

	// 2. Verify all added items return true (no false negatives!)
	for _, item := range addedItems {
		if !f.Contains(item) {
			t.Fatalf("False negative detected! Bloom filters MUST never have false negatives.")
		}
	}

	// 3. Check 10,000 items we NEVER added, and count the false positives
	falsePositives := 0
	tests := 10000
	for range tests {
		b := make([]byte, 16)
		rand.Read(b)

		// Very tiny chance we randomly generated a byte slice we already added,
		// but 16 bytes of crypto/rand makes a collision statistically impossible in this test.
		if f.Contains(b) {
			falsePositives++
		}
	}

	// Calculate actual rate
	actualRate := float64(falsePositives) / float64(tests)

	// We expected ~1% (0.01). We should allow some variance for randomness,
	// say up to 1.5% (0.015) before failing the test.
	if actualRate > 0.015 {
		t.Errorf("False positive rate too high! Expected ~%f, got %f (%d/%d)", p, actualRate, falsePositives, tests)
	}

	t.Logf("Configured P: %f, Actual P: %f (m: %d, k: %d)\n", p, actualRate, f.m, f.k)
}

func ExampleNewFilterFromProbability() {
	// Create a filter for 1000 items with a 1% false positive rate
	f := NewFilterFromProbability(1000, 0.01)

	// Add an item
	f.Add([]byte("my-database-key"))

	// Check if it exists
	if f.Contains([]byte("my-database-key")) {
		fmt.Println("Item is probably in the set.")
	}

	if !f.Contains([]byte("non-existent-key")) {
		fmt.Println("Item is definitively NOT in the set.")
	}

	// Output:
	// Item is probably in the set.
	// Item is definitively NOT in the set.
}

// keySize is the size of each key in bytes, matching a UUID (128 bits).
const keySize = 16

// generateRandomBuffer creates a single contiguous random byte slice of
// length n + keySize - 1. By sliding a window of keySize bytes across it,
// we get n unique keys with zero per-key allocations.
// array[0:16] and array[1:17] are entirely different because the underlying
// data is cryptographically random.
func generateRandomBuffer(n int) []byte {
	buf := make([]byte, n+keySize-1)
	rand.Read(buf)
	return buf
}

// key returns the i-th key from a sliding window buffer.
func key(buf []byte, i int) []byte {
	return buf[i : i+keySize]
}

func BenchmarkFilter_Add_FNV(b *testing.B) {
	f := NewFilter(1000000, 7)
	buf := generateRandomBuffer(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Add(key(buf, i))
	}
}

func BenchmarkFilter_Add_xxHash(b *testing.B) {
	f := NewFilter(1000000, 7, WithHashFunc(func() hash.Hash64 { return xxhash.New() }))
	buf := generateRandomBuffer(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Add(key(buf, i))
	}
}

func BenchmarkFilter_Add_Murmur3(b *testing.B) {
	f := NewFilter(1000000, 7, WithHashFunc(func() hash.Hash64 { return murmur3.New64() }))
	buf := generateRandomBuffer(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Add(key(buf, i))
	}
}

func BenchmarkFilter_Contains_FNV(b *testing.B) {
	f := NewFilter(1000000, 7)
	benchmarkContains(b, f)
}

func BenchmarkFilter_Contains_xxHash(b *testing.B) {
	f := NewFilter(1000000, 7, WithHashFunc(func() hash.Hash64 { return xxhash.New() }))
	benchmarkContains(b, f)
}

func BenchmarkFilter_Contains_Murmur3(b *testing.B) {
	f := NewFilter(1000000, 7, WithHashFunc(func() hash.Hash64 { return murmur3.New64() }))
	benchmarkContains(b, f)
}

func benchmarkContains(b *testing.B, f *Filter) {
	// Pre-fill the filter with 1000 items
	fillBuf := generateRandomBuffer(1000)
	for i := range 1000 {
		f.Add(key(fillBuf, i))
	}

	testBuf := generateRandomBuffer(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Contains(key(testBuf, i))
	}
}

// --- Micro-benchmarks ---

// BenchmarkFilter_Insert1M measures the throughput of inserting 1M UUID-sized
// keys into a filter sized for 1M items at 1% FPR.
func BenchmarkFilter_Insert1M(b *testing.B) {
	const numItems = 1_000_000
	buf := generateRandomBuffer(numItems)

	b.ResetTimer()
	for range b.N {
		f := NewFilterFromProbability(numItems, 0.01)
		for i := range numItems {
			f.Add(key(buf, i))
		}
	}

	// Report memory footprint of the filter itself
	f := NewFilterFromProbability(numItems, 0.01)
	b.ReportMetric(float64(len(f.bitset)*8), "bytes/filter")
	b.ReportMetric(float64(f.m), "bits")
	b.ReportMetric(float64(f.k), "hash_fns")
}

// BenchmarkFilter_Contains1M populates a filter with 1M UUID-sized keys,
// then benchmarks lookup speed against different random keys.
func BenchmarkFilter_Contains1M(b *testing.B) {
	const numItems = 1_000_000
	f := NewFilterFromProbability(numItems, 0.01)

	// Fill the filter
	fillBuf := generateRandomBuffer(numItems)
	for i := range numItems {
		f.Add(key(fillBuf, i))
	}

	// Generate separate lookup buffer
	lookupBuf := generateRandomBuffer(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Contains(key(lookupBuf, i))
	}
}

// BenchmarkFilter_ConcurrentAddContains exercises the atomic.Uint64 design
// under real contention. GOMAXPROCS goroutines do 50/50 Add/Contains on a
// shared filter simultaneously using UUID-sized keys.
func BenchmarkFilter_ConcurrentAddContains(b *testing.B) {
	const numItems = 1_000_000
	f := NewFilterFromProbability(numItems, 0.01)

	// Pre-fill with 1M items so Contains hits are realistic
	fillBuf := generateRandomBuffer(1_000_000)
	for i := range 1_000_000 {
		f.Add(key(fillBuf, i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine gets its own random buffer
		buf := generateRandomBuffer(1_000_000)
		i := 0
		for pb.Next() {
			idx := i % 1_000_000
			if i%2 == 0 {
				f.Add(key(buf, idx))
			} else {
				f.Contains(key(buf, idx))
			}
			i++
		}
	})
}

// TestFilter_FalsePositiveRate1M validates the actual false positive rate
// at scale across all three hash functions: FNV, xxHash, and Murmur3.
// This proves whether hash "quality" actually affects FPR in practice.
func TestFilter_FalsePositiveRate1M(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 1M FPR test in short mode")
	}

	const numItems = 1_000_000
	const targetFPR = 0.01

	hashers := []struct {
		name   string
		hasher HashFunc
	}{
		{"FNV", fnv.New64a},
		{"xxHash", func() hash.Hash64 { return xxhash.New() }},
		{"Murmur3", func() hash.Hash64 { return murmur3.New64() }},
	}

	// Generate data once, reuse across all hashers for fair comparison
	insertBuf := generateRandomBuffer(numItems)
	queryBuf := generateRandomBuffer(numItems)

	for _, h := range hashers {
		t.Run(h.name, func(t *testing.T) {
			f := NewFilterFromProbability(numItems, targetFPR, WithHashFunc(h.hasher))

			// Insert 1M keys
			for i := range numItems {
				f.Add(key(insertBuf, i))
			}

			// Verify zero false negatives
			for i := range numItems {
				if !f.Contains(key(insertBuf, i)) {
					t.Fatalf("False negative at index %d! Bloom filters MUST never have false negatives.", i)
				}
			}

			// Query 1M keys that were never inserted
			falsePositives := 0
			for i := range numItems {
				if f.Contains(key(queryBuf, i)) {
					falsePositives++
				}
			}

			actualRate := float64(falsePositives) / float64(numItems)
			t.Logf("=== FPR Validation [%s] (1M items) ===", h.name)
			t.Logf("  Filter size (m): %d bits (%.2f MB)", f.m, float64(f.m)/(8*1024*1024))
			t.Logf("  Hash functions (k): %d", f.k)
			t.Logf("  Items inserted: %d", numItems)
			t.Logf("  Items queried:  %d", numItems)
			t.Logf("  False positives: %d", falsePositives)
			t.Logf("  Target FPR: %.4f%%", targetFPR*100)
			t.Logf("  Actual FPR: %.4f%%", actualRate*100)

			// Allow some variance — up to 1.5% for a 1% target
			if actualRate > 0.015 {
				t.Errorf("FPR too high! Target: %.4f, Actual: %.4f (%d/%d)", targetFPR, actualRate, falsePositives, numItems)
			}
		})
	}
}
