package solution

import (
	"math"
	"sync"
	"testing"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/testdata"
	"github.com/tmdgusya/database-class/pkg/vector"
)

func TestNewFlatIndex(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		idx, err := NewFlatIndex(Config{
			Metric: distance.L2Distance,
		})
		if err != nil {
			t.Fatalf("NewFlatIndex() failed: %v", err)
		}
		if idx == nil {
			t.Fatal("NewFlatIndex() returned nil index")
		}
	})

	t.Run("nil metric", func(t *testing.T) {
		_, err := NewFlatIndex(Config{
			Metric: nil,
		})
		if err == nil {
			t.Error("NewFlatIndex() should fail with nil metric")
		}
	})
}

func TestBasicAdd(t *testing.T) {
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	// Add a single vector
	v1 := vector.Vector{1.0, 2.0, 3.0}
	if err := idx.Add(v1); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if idx.Size() != 1 {
		t.Errorf("Size() = %d, want 1", idx.Size())
	}

	// Add another vector with same dimension
	v2 := vector.Vector{4.0, 5.0, 6.0}
	if err := idx.Add(v2); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if idx.Size() != 2 {
		t.Errorf("Size() = %d, want 2", idx.Size())
	}
}

func TestEmptyIndex(t *testing.T) {
	// TRAP: Did you handle searching an empty index?
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	query := vector.Vector{1.0, 2.0, 3.0}
	results, err := idx.Search(query, 10)

	if err != nil {
		t.Fatalf("Search() on empty index should not error, got: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Search() on empty index should return empty results, got %d", len(results))
	}
}

func TestDimensionMismatch(t *testing.T) {
	// TRAP: Did you validate dimensions?
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	// Add 3D vector
	v1 := vector.Vector{1.0, 2.0, 3.0}
	if err := idx.Add(v1); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Try to add 2D vector - should fail!
	v2 := vector.Vector{1.0, 2.0}
	if err := idx.Add(v2); err == nil {
		t.Error("Add() should fail with dimension mismatch")
	}

	// Try to search with 2D query - should fail!
	query := vector.Vector{1.0, 2.0}
	if _, err := idx.Search(query, 5); err == nil {
		t.Error("Search() should fail with dimension mismatch")
	}
}

func TestDuplicateVectors(t *testing.T) {
	// TRAP: Can your index handle duplicate vectors?
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	v := vector.Vector{1.0, 2.0, 3.0}

	// Add same vector twice
	if err := idx.Add(v); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}
	if err := idx.Add(v); err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if idx.Size() != 2 {
		t.Errorf("Size() = %d, want 2 (duplicates should be allowed)", idx.Size())
	}

	// Search should find both
	results, err := idx.Search(v, 2)
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Search() returned %d results, want 2", len(results))
	}

	// Both should have distance 0
	for i, r := range results {
		if r.Distance > 1e-9 {
			t.Errorf("results[%d].Distance = %f, want ~0", i, r.Distance)
		}
	}
}

func TestSearchKLargerThanSize(t *testing.T) {
	// TRAP: What if k=100 but only 10 vectors exist?
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	// Add 10 vectors
	for i := 0; i < 10; i++ {
		v := vector.Vector{float64(i), float64(i), float64(i)}
		if err := idx.Add(v); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	// Search for k=100
	query := vector.Vector{5.0, 5.0, 5.0}
	results, err := idx.Search(query, 100)

	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	// Should return all available vectors (10)
	if len(results) != 10 {
		t.Errorf("Search() returned %d results, want 10 (all available)", len(results))
	}
}

func TestSearchResultsSorted(t *testing.T) {
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	// Add vectors at different distances from origin
	vectors := []vector.Vector{
		{5.0, 0.0, 0.0}, // distance 5
		{1.0, 0.0, 0.0}, // distance 1
		{3.0, 0.0, 0.0}, // distance 3
		{2.0, 0.0, 0.0}, // distance 2
		{4.0, 0.0, 0.0}, // distance 4
	}

	for _, v := range vectors {
		if err := idx.Add(v); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	// Search from origin
	query := vector.Vector{0.0, 0.0, 0.0}
	results, err := idx.Search(query, 5)

	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	// Check results are sorted by distance
	for i := 0; i < len(results)-1; i++ {
		if results[i].Distance > results[i+1].Distance {
			t.Errorf("Results not sorted: results[%d].Distance=%f > results[%d].Distance=%f",
				i, results[i].Distance, i+1, results[i+1].Distance)
		}
	}

	// Expected order: 1, 2, 3, 4, 5
	expectedDistances := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for i, expected := range expectedDistances {
		if math.Abs(results[i].Distance-expected) > 1e-9 {
			t.Errorf("results[%d].Distance = %f, want %f",
				i, results[i].Distance, expected)
		}
	}
}

func TestSearchAccuracy(t *testing.T) {
	// Test with predefined dataset
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	// Add small dataset
	for _, v := range testdata.SmallDataset {
		if err := idx.Add(v); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	// Search for a vector that exists in the index
	query := testdata.SmallDataset[0]
	results, err := idx.Search(query, 1)

	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Search() returned no results")
	}

	// First result should be the query itself (distance ~0)
	if results[0].Distance > 1e-9 {
		t.Errorf("First result distance = %f, want ~0 (exact match)", results[0].Distance)
	}
}

func TestConcurrency(t *testing.T) {
	// TRAP: Did you use the mutex correctly?
	// This test will catch race conditions
	// Run with: go test -race
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	// Add some initial vectors
	for i := 0; i < 100; i++ {
		v := vector.Vector{float64(i), float64(i), float64(i)}
		if err := idx.Add(v); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	var wg sync.WaitGroup

	// Concurrent reads (searches)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			query := vector.Vector{float64(i), float64(i), float64(i)}
			_, err := idx.Search(query, 10)
			if err != nil {
				t.Errorf("Concurrent Search() failed: %v", err)
			}
		}(i)
	}

	// Concurrent writes (adds)
	for i := 100; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := vector.Vector{float64(i), float64(i), float64(i)}
			if err := idx.Add(v); err != nil {
				t.Errorf("Concurrent Add() failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Final size should be 200
	if idx.Size() != 200 {
		t.Errorf("After concurrent operations, Size() = %d, want 200", idx.Size())
	}
}

func TestInvalidVector(t *testing.T) {
	idx, _ := NewFlatIndex(Config{Metric: distance.L2Distance})

	tests := []struct {
		name string
		vec  vector.Vector
	}{
		{"nil vector", nil},
		{"empty vector", vector.Vector{}},
		{"NaN vector", vector.Vector{1.0, math.NaN(), 3.0}},
		{"Inf vector", vector.Vector{1.0, math.Inf(1), 3.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := idx.Add(tt.vec); err == nil {
				t.Errorf("Add(%s) should fail but didn't", tt.name)
			}
		})
	}
}
