package exercise

import (
	"errors"
	"sort"
	"sync"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// FlatIndex implements a brute-force vector index
// This is the baseline - it searches through ALL vectors linearly
type FlatIndex struct {
	// - You need to store vectors (slice of vectors)
	vectors []vector.Vector
	// - You need to track the distance metric
	metric distance.Metric
	// - You need to track the dimension (for validation)
	dimension int
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
	if cfg.Metric == nil {
		return nil, errors.New("metric cannot be nil")
	}
	// 2. Initialize the index with empty vectors
	idx := &FlatIndex{
		vectors:   []vector.Vector{},
		metric:    cfg.Metric,
		dimension: 0,
		mu:        sync.RWMutex{},
	}
	// 3. Return the index
	return idx, nil
}

// Add adds a vector to the index
// Returns error if vector is invalid or dimension mismatch
func (idx *FlatIndex) Add(v vector.Vector) error {
	// TODO: Implement
	// Tasks:
	// 1. Validate vector (use v.Validate())
	idx.mu.Lock()
	defer idx.mu.Unlock()
	if err := v.Validate(); err != nil {
		return err
	}
	// 2. Check dimension consistency
	if idx.dimension == 0 {
		idx.dimension = v.Dimension()
	} else if idx.dimension != v.Dimension() {
		return errors.New("dimension mismatch")
	}
	// 3. Add vector to storage
	idx.vectors = append(idx.vectors, v)
	// 4. Handle thread safety (use idx.mu)
	return nil
}

// Search performs k-nearest neighbor search
// Returns the k closest vectors and their distances, sorted by distance (closest first)
func (idx *FlatIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	// TODO: Implement
	// Tasks:
	// 1. Validate query vector
	if err := query.Validate(); err != nil {
		return nil, err
	}
	// 2. Validate k (should be > 0)
	if k <= 0 {
		return nil, errors.New("invalid k")
	}

	if k > len(idx.vectors) {
		k = len(idx.vectors)
	}
	// 3. Check dimension match
	if query.Dimension() != idx.dimension {
		return nil, errors.New("dimension mismatch")
	}
	// 4. Calculate distance to ALL vectors (this is brute force!)
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	distances := make([]SearchResult, len(idx.vectors))
	for i, v := range idx.vectors {
		distance, err := idx.metric(query, v)
		if err != nil {
			return nil, err
		}
		distances[i] = SearchResult{
			Vector:   v,
			Distance: distance,
			Index:    i,
		}
	}
	// 5. Find k smallest distances
	//    - You can use sorting (simple)
	//    - Or use heap (more efficient for large n, small k)
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].Distance < distances[j].Distance
	})
	// 6. Return results sorted by distance
	return distances[:k], nil
}

// Size returns the number of vectors in the index
func (idx *FlatIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.vectors)
}
