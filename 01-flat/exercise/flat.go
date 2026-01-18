package exercise

import (
	"sync"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// FlatIndex implements a brute-force vector index
// This is the baseline - it searches through ALL vectors linearly
type FlatIndex struct {
	// TODO: Add necessary fields
	// Hints:
	// - You need to store vectors (slice of vectors)
	// - You need to track the distance metric
	// - You need to track the dimension (for validation)
	// - The mutex is provided below for thread safety

	mu sync.RWMutex // Provided for thread safety
}

// Config holds configuration for FlatIndex
type Config struct {
	Metric distance.Metric // Distance function to use
}

// SearchResult represents a single search result
type SearchResult struct {
	Vector   vector.Vector // The found vector
	Distance float64       // Distance to query
	Index    int           // Position in the index (0-based)
}

// NewFlatIndex creates a new flat index
func NewFlatIndex(cfg Config) (*FlatIndex, error) {
	// TODO: Implement
	// Tasks:
	// 1. Validate config (metric should not be nil)
	// 2. Initialize the index with empty vectors
	// 3. Return the index

	panic("not implemented")
}

// Add adds a vector to the index
// Returns error if vector is invalid or dimension mismatch
func (idx *FlatIndex) Add(v vector.Vector) error {
	// TODO: Implement
	// Tasks:
	// 1. Validate vector (use v.Validate())
	// 2. Check dimension consistency
	//    - If this is the first vector, store its dimension
	//    - If not first, check if dimension matches
	// 3. Add vector to storage
	// 4. Handle thread safety (use idx.mu)
	//
	// Thread safety pattern:
	//   idx.mu.Lock()
	//   defer idx.mu.Unlock()

	panic("not implemented")
}

// Search performs k-nearest neighbor search
// Returns the k closest vectors and their distances, sorted by distance (closest first)
func (idx *FlatIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	// TODO: Implement
	// Tasks:
	// 1. Validate query vector
	// 2. Validate k (should be > 0)
	// 3. Check dimension match
	// 4. Calculate distance to ALL vectors (this is brute force!)
	// 5. Find k smallest distances
	//    - You can use sorting (simple)
	//    - Or use heap (more efficient for large n, small k)
	// 6. Return results sorted by distance
	// 7. Handle thread safety (use idx.mu.RLock for read-only)
	//
	// Thread safety pattern for reading:
	//   idx.mu.RLock()
	//   defer idx.mu.RUnlock()
	//
	// Edge cases to handle:
	// - Empty index: return empty results
	// - k > index size: return all vectors
	//
	// Hint for k-smallest:
	//   1. Create slice of SearchResult with all distances
	//   2. Sort by distance
	//   3. Return first k elements (or all if k > size)

	panic("not implemented")
}

// Size returns the number of vectors in the index
func (idx *FlatIndex) Size() int {
	// TODO: Implement
	// Don't forget thread safety!

	panic("not implemented")
}
