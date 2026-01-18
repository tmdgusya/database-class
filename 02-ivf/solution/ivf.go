package solution

import (
	"fmt"
	"sort"
	"sync"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// IVFIndex implements Inverted File Index
type IVFIndex struct {
	centroids []vector.Vector   // Cluster centroids
	clusters  [][]vector.Vector // Vectors in each cluster
	metric    distance.Metric   // Distance function
	nlist     int               // Number of clusters
	nprobe    int               // Number of clusters to search
	trained   bool              // Whether index is trained
	dimension int               // Vector dimension
	mu        sync.RWMutex      // Thread safety
}

// Config holds IVF configuration
type Config struct {
	Metric      distance.Metric
	NumClusters int // nlist
	NumProbes   int // nprobe
}

// SearchResult represents a single search result
type SearchResult struct {
	Vector   vector.Vector
	Distance float64
	Index    int // Global index across all clusters
}

// NewIVFIndex creates a new IVF index
func NewIVFIndex(cfg Config) (*IVFIndex, error) {
	// Validate config
	if cfg.Metric == nil {
		return nil, fmt.Errorf("metric cannot be nil")
	}
	if cfg.NumClusters <= 0 {
		return nil, fmt.Errorf("NumClusters must be positive, got %d", cfg.NumClusters)
	}
	if cfg.NumProbes <= 0 {
		return nil, fmt.Errorf("NumProbes must be positive, got %d", cfg.NumProbes)
	}
	if cfg.NumProbes > cfg.NumClusters {
		return nil, fmt.Errorf("NumProbes (%d) cannot exceed NumClusters (%d)",
			cfg.NumProbes, cfg.NumClusters)
	}

	return &IVFIndex{
		metric:  cfg.Metric,
		nlist:   cfg.NumClusters,
		nprobe:  cfg.NumProbes,
		trained: false,
	}, nil
}

// Train trains the index by clustering the provided vectors
func (idx *IVFIndex) Train(vectors []vector.Vector) error {
	if len(vectors) == 0 {
		return fmt.Errorf("no training vectors provided")
	}

	// Check if enough training data
	// Recommended: at least 30 * nlist vectors
	minVectors := idx.nlist
	if len(vectors) < minVectors {
		return fmt.Errorf("insufficient training data: need at least %d vectors, got %d",
			minVectors, len(vectors))
	}

	// Validate all vectors
	dim := vectors[0].Dimension()
	for i, v := range vectors {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("invalid vector at index %d: %w", i, err)
		}
		if v.Dimension() != dim {
			return fmt.Errorf("dimension mismatch at vector %d: expected %d, got %d",
				i, dim, v.Dimension())
		}
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Run k-means clustering
	centroids, err := KMeans(vectors, idx.nlist, 100, idx.metric)
	if err != nil {
		return fmt.Errorf("k-means clustering failed: %w", err)
	}

	// Store centroids and initialize empty clusters
	idx.centroids = centroids
	idx.clusters = make([][]vector.Vector, idx.nlist)
	for i := range idx.clusters {
		idx.clusters[i] = make([]vector.Vector, 0)
	}

	idx.trained = true
	idx.dimension = dim

	return nil
}

// Add adds a vector to the index
func (idx *IVFIndex) Add(v vector.Vector) error {
	// Validate vector
	if err := v.Validate(); err != nil {
		return fmt.Errorf("invalid vector: %w", err)
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Check if trained
	if !idx.trained {
		return fmt.Errorf("index not trained: call Train() first")
	}

	// Check dimension match
	if v.Dimension() != idx.dimension {
		return fmt.Errorf("dimension mismatch: expected %d, got %d",
			idx.dimension, v.Dimension())
	}

	// Find nearest centroid
	centroidIdx, err := FindNearestCentroid(v, idx.centroids, idx.metric)
	if err != nil {
		return fmt.Errorf("failed to find nearest centroid: %w", err)
	}

	// Add to that cluster
	idx.clusters[centroidIdx] = append(idx.clusters[centroidIdx], v.Clone())

	return nil
}

// Search performs approximate k-NN search
func (idx *IVFIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	if k <= 0 {
		return nil, fmt.Errorf("k must be positive, got %d", k)
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Check if trained
	if !idx.trained {
		return nil, fmt.Errorf("index not trained: call Train() first")
	}

	// Check dimension match
	if query.Dimension() != idx.dimension {
		return nil, fmt.Errorf("query dimension mismatch: expected %d, got %d",
			idx.dimension, query.Dimension())
	}

	// Find nprobe nearest centroids
	nearestCentroids, err := idx.findNearestCentroids(query, idx.nprobe)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest centroids: %w", err)
	}

	// Collect candidates from selected clusters
	var candidates []SearchResult
	globalIdx := 0

	for clusterIdx := 0; clusterIdx < idx.nlist; clusterIdx++ {
		cluster := idx.clusters[clusterIdx]

		// Check if this cluster is in our search list
		searchThisCluster := false
		for _, nearestIdx := range nearestCentroids {
			if nearestIdx == clusterIdx {
				searchThisCluster = true
				break
			}
		}

		if searchThisCluster {
			// Search this cluster
			for _, v := range cluster {
				dist, err := idx.metric(query, v)
				if err != nil {
					return nil, fmt.Errorf("distance calculation failed: %w", err)
				}

				candidates = append(candidates, SearchResult{
					Vector:   v,
					Distance: dist,
					Index:    globalIdx,
				})
				globalIdx++
			}
		} else {
			// Skip this cluster but update global index
			globalIdx += len(cluster)
		}
	}

	// Handle case: no candidates found
	if len(candidates) == 0 {
		return []SearchResult{}, nil
	}

	// Sort by distance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Distance < candidates[j].Distance
	})

	// Return top k
	if k > len(candidates) {
		k = len(candidates)
	}

	return candidates[:k], nil
}

// SetNumProbes adjusts nprobe parameter at runtime
func (idx *IVFIndex) SetNumProbes(nprobe int) error {
	if nprobe <= 0 {
		return fmt.Errorf("nprobe must be positive, got %d", nprobe)
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if nprobe > idx.nlist {
		return fmt.Errorf("nprobe (%d) cannot exceed nlist (%d)", nprobe, idx.nlist)
	}

	idx.nprobe = nprobe
	return nil
}

// Size returns the total number of vectors in the index
func (idx *IVFIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if !idx.trained {
		return 0
	}

	total := 0
	for _, cluster := range idx.clusters {
		total += len(cluster)
	}
	return total
}

// findNearestCentroids finds nprobe nearest centroids to query
func (idx *IVFIndex) findNearestCentroids(query vector.Vector, nprobe int) ([]int, error) {
	if len(idx.centroids) == 0 {
		return nil, fmt.Errorf("no centroids available")
	}

	// Calculate distance to all centroids
	type centroidDist struct {
		index    int
		distance float64
	}

	distances := make([]centroidDist, len(idx.centroids))
	for i, centroid := range idx.centroids {
		dist, err := idx.metric(query, centroid)
		if err != nil {
			return nil, fmt.Errorf("distance calculation failed: %w", err)
		}
		distances[i] = centroidDist{
			index:    i,
			distance: dist,
		}
	}

	// Sort by distance
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].distance < distances[j].distance
	})

	// Return indices of nprobe nearest
	if nprobe > len(distances) {
		nprobe = len(distances)
	}

	result := make([]int, nprobe)
	for i := 0; i < nprobe; i++ {
		result[i] = distances[i].index
	}

	return result, nil
}
