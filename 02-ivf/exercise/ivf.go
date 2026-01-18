package exercise

import (
	"sync"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// IVFIndex implements Inverted File Index
// Vectors are clustered, and search only examines nearby clusters
type IVFIndex struct {
	// TODO: Add necessary fields
	// Hints:
	// - Store cluster centroids ([]vector.Vector)
	// - Store vectors in each cluster ([][]vector.Vector or map)
	// - Store the distance metric
	// - Track parameters: nlist (num clusters), nprobe (num to search)
	// - Track if trained (bool)
	// - Track dimension (int)

	mu sync.RWMutex // Provided for thread safety
}

// Config holds IVF configuration
type Config struct {
	Metric      distance.Metric // Distance function
	NumClusters int             // nlist: number of clusters
	NumProbes   int             // nprobe: number of clusters to search
}

// SearchResult represents a single search result
type SearchResult struct {
	Vector   vector.Vector
	Distance float64
	Index    int // Global index
}

// NewIVFIndex creates a new IVF index
func NewIVFIndex(cfg Config) (*IVFIndex, error) {
	// TODO: Implement
	//
	// Tasks:
	// 1. Validate config:
	//    - Metric not nil
	//    - NumClusters > 0
	//    - NumProbes > 0
	//    - NumProbes <= NumClusters
	// 2. Initialize index structure
	// 3. Set trained = false (no centroids yet)

	panic("not implemented")
}

// Train trains the index by clustering the provided vectors
// This must be called before Add/Search!
func (idx *IVFIndex) Train(vectors []vector.Vector) error {
	// TODO: Implement
	//
	// Tasks:
	// 1. Validate training data:
	//    - Not empty
	//    - Enough vectors (recommended: >= nlist * 30)
	//    - All same dimension
	// 2. Run k-means clustering:
	//    - Use KMeans function
	//    - Store resulting centroids
	// 3. Initialize empty clusters (one per centroid)
	// 4. Set trained = true
	// 5. Store dimension
	//
	// TRAP: What if len(vectors) < nlist?
	// - k-means can't create more clusters than vectors
	// - Return error or adjust nlist
	//
	// Hints:
	// - maxIter = 100 is typical
	// - Check if already trained?

	panic("not implemented")
}

// Add adds a vector to the index
func (idx *IVFIndex) Add(v vector.Vector) error {
	// TODO: Implement
	//
	// Tasks:
	// 1. Check if trained (return error if not)
	// 2. Validate vector
	// 3. Check dimension match
	// 4. Find nearest centroid
	// 5. Add vector to that cluster
	//
	// TRAP: Not checking trained = panic!
	//
	// Hints:
	// - Use FindNearestCentroid
	// - Thread safety: lock for write

	panic("not implemented")
}

// Search performs approximate k-NN search
func (idx *IVFIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	// TODO: Implement
	//
	// Tasks:
	// 1. Validate query and k
	// 2. Check trained
	// 3. Check dimension
	// 4. Find nprobe nearest centroids
	// 5. Search only those clusters:
	//    - Collect all vectors from nprobe clusters
	//    - Calculate distances
	//    - Sort by distance
	//    - Return top k
	//
	// TRAP: nprobe too small → poor recall!
	//       nprobe = 1 → recall ~30%
	//       nprobe = nlist → same as Flat
	//
	// Hints:
	// - Use FindNearestCentroid but find nprobe of them
	// - Collect candidates from multiple clusters
	// - Handle k > total candidates
	// - Thread safety: lock for read

	panic("not implemented")
}

// SetNumProbes adjusts nprobe parameter at runtime
func (idx *IVFIndex) SetNumProbes(nprobe int) error {
	// TODO: Implement
	//
	// Tasks:
	// 1. Validate: 1 <= nprobe <= nlist
	// 2. Update idx.nprobe
	//
	// Hints:
	// - This allows tuning without rebuilding index
	// - Useful for experiments

	panic("not implemented")
}

// Size returns the total number of vectors in the index
func (idx *IVFIndex) Size() int {
	// TODO: Implement
	//
	// Hints:
	// - Sum vectors across all clusters
	// - Thread safety

	panic("not implemented")
}

// Helper function: find nprobe nearest centroids to query
// You'll need this for Search
func (idx *IVFIndex) findNearestCentroids(query vector.Vector, nprobe int) ([]int, error) {
	// TODO: Implement
	//
	// Tasks:
	// 1. Calculate distance from query to all centroids
	// 2. Sort by distance
	// 3. Return indices of nprobe nearest
	//
	// Hints:
	// - Similar to Search in Flat index
	// - Create slice of {centroidIdx, distance}
	// - Sort and take first nprobe

	panic("not implemented")
}
