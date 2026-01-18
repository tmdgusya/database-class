package solution

import (
	"fmt"
	"sort"
	"sync"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// FlatIndex implements a brute-force vector index
type FlatIndex struct {
	vectors   []vector.Vector  // All stored vectors
	metric    distance.Metric  // Distance function
	dimension int              // Vector dimension (for validation)
	mu        sync.RWMutex     // Thread safety
}

// Config holds configuration for FlatIndex
type Config struct {
	Metric distance.Metric
}

// SearchResult represents a single search result
type SearchResult struct {
	Vector   vector.Vector
	Distance float64
	Index    int
}

// NewFlatIndex creates a new flat index
func NewFlatIndex(cfg Config) (*FlatIndex, error) {
	// Validate config
	if cfg.Metric == nil {
		return nil, fmt.Errorf("metric cannot be nil")
	}

	return &FlatIndex{
		vectors:   make([]vector.Vector, 0),
		metric:    cfg.Metric,
		dimension: -1, // -1 means not set yet
	}, nil
}

// Add adds a vector to the index
func (idx *FlatIndex) Add(v vector.Vector) error {
	// Validate vector
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid vector: %w", err)
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Check dimension consistency
	if idx.dimension == -1 {
		// First vector - set dimension
		idx.dimension = v.Dimension()
	} else {
		// Check dimension matches
		if v.Dimension() != idx.dimension {
			return fmt.Errorf("dimension mismatch: expected %d, got %d",
				idx.dimension, v.Dimension())
		}
	}

	// Store a clone to avoid external modifications
	idx.vectors = append(idx.vectors, v.Clone())

	return nil
}

// Search performs k-nearest neighbor search
func (idx *FlatIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	if k <= 0 {
		return nil, fmt.Errorf("k must be positive, got %d", k)
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Handle empty index
	if len(idx.vectors) == 0 {
		return []SearchResult{}, nil
	}

	// Check dimension match
	if query.Dimension() != idx.dimension {
		return nil, fmt.Errorf("query dimension mismatch: expected %d, got %d",
			idx.dimension, query.Dimension())
	}

	// Calculate distances to all vectors
	results := make([]SearchResult, len(idx.vectors))
	for i, v := range idx.vectors {
		dist, err := idx.metric(query, v)
		if err != nil {
			return nil, fmt.Errorf("distance calculation failed at index %d: %w", i, err)
		}

		results[i] = SearchResult{
			Vector:   v,
			Distance: dist,
			Index:    i,
		}
	}

	// Sort by distance (ascending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	// Return top k (or all if k > size)
	if k > len(results) {
		k = len(results)
	}

	return results[:k], nil
}

// Size returns the number of vectors in the index
func (idx *FlatIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.vectors)
}
