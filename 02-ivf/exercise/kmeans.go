package exercise

import (
	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// KMeans performs k-means clustering on vectors
// Returns k centroids (cluster centers)
func KMeans(
	vectors []vector.Vector,
	k int,
	maxIter int,
	metric distance.Metric,
) ([]vector.Vector, error) {
	// TODO: Implement k-means clustering
	//
	// Algorithm:
	// 1. Initialize centroids
	//    - Simple: pick first k vectors
	//    - Better: k-means++ (random with distance-based probability)
	//
	// 2. Iterate until convergence or maxIter:
	//    a. Assignment step:
	//       - Assign each vector to nearest centroid
	//    b. Update step:
	//       - Compute new centroids as mean of assigned vectors
	//    c. Check convergence:
	//       - If centroids don't change, break
	//
	// 3. Return final centroids
	//
	// Hints:
	// - Use FindNearestCentroid for assignment
	// - For update, compute mean: sum all vectors in cluster, divide by count
	// - Empty cluster handling: keep old centroid or reinitialize
	//
	// Validation:
	// - Check k <= len(vectors)
	// - Check all vectors have same dimension

	panic("not implemented")
}

// FindNearestCentroid finds the index of the nearest centroid to vector v
func FindNearestCentroid(
	v vector.Vector,
	centroids []vector.Vector,
	metric distance.Metric,
) (int, error) {
	// TODO: Implement
	//
	// Tasks:
	// 1. Calculate distance from v to each centroid
	// 2. Return index of centroid with minimum distance
	//
	// Hints:
	// - Similar to finding minimum in an array
	// - Track both minDistance and minIndex

	panic("not implemented")
}

// Helper function: compute mean of vectors
// You can use this in KMeans
func computeMean(vectors []vector.Vector) vector.Vector {
	if len(vectors) == 0 {
		return nil
	}

	dim := vectors[0].Dimension()
	mean := make(vector.Vector, dim)

	// Sum all vectors
	for _, v := range vectors {
		for i := 0; i < dim; i++ {
			mean[i] += v[i]
		}
	}

	// Divide by count
	count := float64(len(vectors))
	for i := 0; i < dim; i++ {
		mean[i] /= count
	}

	return mean
}

// Helper function: check if centroids converged
// You can use this in KMeans
func converged(old, new []vector.Vector, epsilon float64) bool {
	if len(old) != len(new) {
		return false
	}

	for i := range old {
		if !old[i].Equal(new[i], epsilon) {
			return false
		}
	}

	return true
}
