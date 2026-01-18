package solution

import (
	"fmt"
	"math/rand"

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
	// Validate inputs
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no vectors provided")
	}
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive, got %d", k)
	}
	if k > len(vectors) {
		return nil, fmt.Errorf("k (%d) cannot exceed number of vectors (%d)", k, len(vectors))
	}
	if maxIter <= 0 {
		maxIter = 100 // Default
	}

	// Check dimension consistency
	dim := vectors[0].Dimension()
	for i, v := range vectors {
		if v.Dimension() != dim {
			return nil, fmt.Errorf("dimension mismatch at vector %d: expected %d, got %d",
				i, dim, v.Dimension())
		}
	}

	// Initialize centroids with k-means++
	centroids := initializeCentroidsKMeansPlusPlus(vectors, k, metric)

	// Main k-means loop
	for iter := 0; iter < maxIter; iter++ {
		// Assignment step: assign each vector to nearest centroid
		assignments := make([]int, len(vectors))
		for i, v := range vectors {
			nearest, err := FindNearestCentroid(v, centroids, metric)
			if err != nil {
				return nil, fmt.Errorf("assignment failed: %w", err)
			}
			assignments[i] = nearest
		}

		// Update step: compute new centroids
		newCentroids := make([]vector.Vector, k)
		clusterSizes := make([]int, k)

		for clusterIdx := 0; clusterIdx < k; clusterIdx++ {
			// Collect vectors in this cluster
			var clusterVectors []vector.Vector
			for vecIdx, assignment := range assignments {
				if assignment == clusterIdx {
					clusterVectors = append(clusterVectors, vectors[vecIdx])
				}
			}

			clusterSizes[clusterIdx] = len(clusterVectors)

			// Compute mean (new centroid)
			if len(clusterVectors) > 0 {
				newCentroids[clusterIdx] = computeMean(clusterVectors)
			} else {
				// Empty cluster: keep old centroid or reinitialize
				newCentroids[clusterIdx] = centroids[clusterIdx].Clone()
			}
		}

		// Check convergence
		if converged(centroids, newCentroids, 1e-6) {
			centroids = newCentroids
			break
		}

		centroids = newCentroids
	}

	return centroids, nil
}

// FindNearestCentroid finds the index of the nearest centroid to vector v
func FindNearestCentroid(
	v vector.Vector,
	centroids []vector.Vector,
	metric distance.Metric,
) (int, error) {
	if len(centroids) == 0 {
		return 0, fmt.Errorf("no centroids provided")
	}

	minIdx := 0
	minDist, err := metric(v, centroids[0])
	if err != nil {
		return 0, fmt.Errorf("distance calculation failed: %w", err)
	}

	for i := 1; i < len(centroids); i++ {
		dist, err := metric(v, centroids[i])
		if err != nil {
			return 0, fmt.Errorf("distance calculation failed: %w", err)
		}

		if dist < minDist {
			minDist = dist
			minIdx = i
		}
	}

	return minIdx, nil
}

// initializeCentroidsKMeansPlusPlus implements k-means++ initialization
// This gives better initial centroids than random selection
func initializeCentroidsKMeansPlusPlus(
	vectors []vector.Vector,
	k int,
	metric distance.Metric,
) []vector.Vector {
	centroids := make([]vector.Vector, 0, k)

	// First centroid: random
	firstIdx := rand.Intn(len(vectors))
	centroids = append(centroids, vectors[firstIdx].Clone())

	// Remaining k-1 centroids
	for len(centroids) < k {
		// Calculate distance from each vector to nearest centroid
		distances := make([]float64, len(vectors))
		totalDist := 0.0

		for i, v := range vectors {
			minDist := 1e10
			for _, centroid := range centroids {
				dist, _ := metric(v, centroid)
				if dist < minDist {
					minDist = dist
				}
			}
			distances[i] = minDist * minDist // Squared distance
			totalDist += distances[i]
		}

		// Select next centroid with probability proportional to distance^2
		target := rand.Float64() * totalDist
		cumsum := 0.0
		nextIdx := 0

		for i, dist := range distances {
			cumsum += dist
			if cumsum >= target {
				nextIdx = i
				break
			}
		}

		centroids = append(centroids, vectors[nextIdx].Clone())
	}

	return centroids
}

// computeMean computes the mean of vectors
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

// converged checks if centroids have converged
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
