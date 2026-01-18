package testdata

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/tmdgusya/database-class/pkg/vector"
)

// GenerateRandomVectors creates uniformly random vectors
// All values are in [0, 1) range
// seed ensures reproducibility
func GenerateRandomVectors(count, dim int, seed int64) []vector.Vector {
	if count <= 0 || dim <= 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	vectors := make([]vector.Vector, count)

	for i := 0; i < count; i++ {
		v := make(vector.Vector, dim)
		for j := 0; j < dim; j++ {
			v[j] = rng.Float64()
		}
		vectors[i] = v
	}

	return vectors
}

// GenerateClusteredVectors creates vectors in clusters
// This is CRITICAL for IVF testing - creates realistic clustered data
// Each cluster has a random center, and vectors are distributed around it with Gaussian noise
func GenerateClusteredVectors(count, dim, numClusters int, seed int64) []vector.Vector {
	if count <= 0 || dim <= 0 || numClusters <= 0 {
		return nil
	}
	if numClusters > count {
		numClusters = count
	}

	rng := rand.New(rand.NewSource(seed))

	// Generate cluster centers
	centers := make([]vector.Vector, numClusters)
	for i := 0; i < numClusters; i++ {
		center := make(vector.Vector, dim)
		for j := 0; j < dim; j++ {
			center[j] = rng.Float64()
		}
		centers[i] = center
	}

	// Generate vectors around cluster centers
	vectors := make([]vector.Vector, count)
	clusterVariance := 0.1 // Standard deviation of Gaussian noise

	for i := 0; i < count; i++ {
		// Assign to random cluster
		clusterIdx := rng.Intn(numClusters)
		center := centers[clusterIdx]

		// Add Gaussian noise around center
		v := make(vector.Vector, dim)
		for j := 0; j < dim; j++ {
			v[j] = center[j] + rng.NormFloat64()*clusterVariance
		}
		vectors[i] = v
	}

	return vectors
}

// GenerateNormalizedVectors creates random vectors and normalizes them to unit length
// This is useful for cosine distance testing (normalized vectors have ||v|| = 1)
func GenerateNormalizedVectors(count, dim int, seed int64) []vector.Vector {
	vectors := GenerateRandomVectors(count, dim, seed)
	if vectors == nil {
		return nil
	}

	for i := range vectors {
		vectors[i] = normalize(vectors[i])
	}

	return vectors
}

// GenerateGridVectors creates vectors on a regular grid
// Useful for visualization and understanding in low dimensions (2D or 3D)
func GenerateGridVectors(gridSize, dim int) []vector.Vector {
	if gridSize <= 0 || dim <= 0 || dim > 3 {
		return nil
	}

	var vectors []vector.Vector

	switch dim {
	case 1:
		for i := 0; i < gridSize; i++ {
			v := vector.Vector{float64(i) / float64(gridSize)}
			vectors = append(vectors, v)
		}
	case 2:
		for i := 0; i < gridSize; i++ {
			for j := 0; j < gridSize; j++ {
				v := vector.Vector{
					float64(i) / float64(gridSize),
					float64(j) / float64(gridSize),
				}
				vectors = append(vectors, v)
			}
		}
	case 3:
		for i := 0; i < gridSize; i++ {
			for j := 0; j < gridSize; j++ {
				for k := 0; k < gridSize; k++ {
					v := vector.Vector{
						float64(i) / float64(gridSize),
						float64(j) / float64(gridSize),
						float64(k) / float64(gridSize),
					}
					vectors = append(vectors, v)
				}
			}
		}
	}

	return vectors
}

// GenerateVectorsInBall creates vectors uniformly distributed in a ball
// This creates more realistic high-dimensional data than uniform cube sampling
func GenerateVectorsInBall(count, dim int, radius float64, seed int64) []vector.Vector {
	if count <= 0 || dim <= 0 || radius <= 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	vectors := make([]vector.Vector, count)

	for i := 0; i < count; i++ {
		// Generate random direction (unit vector)
		v := make(vector.Vector, dim)
		for j := 0; j < dim; j++ {
			v[j] = rng.NormFloat64()
		}
		v = normalize(v)

		// Generate random radius with correct distribution
		// For uniform distribution in ball, we need r^(1/dim)
		r := math.Pow(rng.Float64(), 1.0/float64(dim)) * radius

		// Scale to random radius
		for j := 0; j < dim; j++ {
			v[j] *= r
		}

		vectors[i] = v
	}

	return vectors
}

// AddNoise adds Gaussian noise to existing vectors
// variance controls the amount of noise (standard deviation)
func AddNoise(vectors []vector.Vector, variance float64, seed int64) []vector.Vector {
	if len(vectors) == 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	noisy := make([]vector.Vector, len(vectors))

	for i, v := range vectors {
		noisyVec := make(vector.Vector, len(v))
		for j := range v {
			noisyVec[j] = v[j] + rng.NormFloat64()*variance
		}
		noisy[i] = noisyVec
	}

	return noisy
}

// ShuffleVectors randomly shuffles the order of vectors
// Useful for ensuring test independence from insertion order
func ShuffleVectors(vectors []vector.Vector, seed int64) []vector.Vector {
	if len(vectors) == 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	shuffled := make([]vector.Vector, len(vectors))
	copy(shuffled, vectors)

	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}

// normalize returns a unit vector (||v|| = 1)
func normalize(v vector.Vector) vector.Vector {
	var sumSquares float64
	for _, val := range v {
		sumSquares += val * val
	}

	norm := math.Sqrt(sumSquares)
	if norm == 0 {
		// If zero vector, return as-is (or could return error)
		return v
	}

	normalized := make(vector.Vector, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}

	return normalized
}

// ComputeGroundTruth computes exact k-NN for queries against a database
// This is the "ground truth" for evaluating approximate algorithms
// Uses brute force - slow but exact
func ComputeGroundTruth(
	queries []vector.Vector,
	database []vector.Vector,
	k int,
	metric func(a, b vector.Vector) (float64, error),
) ([][]int, error) {
	if len(queries) == 0 || len(database) == 0 || k <= 0 {
		return nil, fmt.Errorf("invalid parameters: queries=%d, database=%d, k=%d",
			len(queries), len(database), k)
	}

	if k > len(database) {
		k = len(database)
	}

	results := make([][]int, len(queries))

	for i, query := range queries {
		// Calculate distances to all database vectors
		type distIdx struct {
			dist  float64
			index int
		}

		distances := make([]distIdx, len(database))
		for j, dbVec := range database {
			dist, err := metric(query, dbVec)
			if err != nil {
				return nil, fmt.Errorf("metric error at query %d, db %d: %w", i, j, err)
			}
			distances[j] = distIdx{dist: dist, index: j}
		}

		// Sort by distance
		// Simple selection sort for k smallest
		for j := 0; j < k; j++ {
			minIdx := j
			for l := j + 1; l < len(distances); l++ {
				if distances[l].dist < distances[minIdx].dist {
					minIdx = l
				}
			}
			distances[j], distances[minIdx] = distances[minIdx], distances[j]
		}

		// Extract indices
		topK := make([]int, k)
		for j := 0; j < k; j++ {
			topK[j] = distances[j].index
		}

		results[i] = topK
	}

	return results, nil
}
