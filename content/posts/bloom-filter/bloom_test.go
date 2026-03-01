package bloom

import (
	"crypto/rand"
	"fmt"
	"hash"
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

// generateRandomData creates a slice of n words, each 8 bytes long
func generateRandomData(n int) [][]byte {
	data := make([][]byte, n)
	for i := range n {
		data[i] = make([]byte, 8)
		rand.Read(data[i])
	}
	return data
}

func BenchmarkFilter_Add_FNV(b *testing.B) {
	f := NewFilter(1000000, 7)
	data := generateRandomData(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Add(data[i])
	}
}

func BenchmarkFilter_Add_xxHash(b *testing.B) {
	f := NewFilter(1000000, 7, WithHashFunc(func() hash.Hash64 { return xxhash.New() }))
	data := generateRandomData(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Add(data[i])
	}
}

func BenchmarkFilter_Add_Murmur3(b *testing.B) {
	f := NewFilter(1000000, 7, WithHashFunc(func() hash.Hash64 { return murmur3.New64() }))
	data := generateRandomData(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Add(data[i])
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
	// Pre-fill the filter
	fillData := generateRandomData(1000)
	for _, d := range fillData {
		f.Add(d)
	}

	testData := generateRandomData(b.N)

	b.ResetTimer()
	for i := range b.N {
		f.Contains(testData[i])
	}
}
