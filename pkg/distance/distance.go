package distance

import (
	"fmt"
	"math"

	"github.com/tmdgusya/database-class/pkg/vector"
)

// Metric represents a distance metric function
// It takes two vectors and returns their distance and any error
type Metric func(a, b vector.Vector) (float64, error)

// L2Distance calculates Euclidean distance (L2 norm)
// Formula: sqrt(sum((a[i] - b[i])^2))
func L2Distance(a, b vector.Vector) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("dimension mismatch: %d vs %d", len(a), len(b))
	}
	if len(a) == 0 {
		return 0, fmt.Errorf("cannot calculate distance for empty vectors")
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum), nil
}

// L2DistanceSquared calculates squared Euclidean distance
// This is faster than L2Distance since it avoids the square root
// Useful when only comparing distances (order is preserved)
// Formula: sum((a[i] - b[i])^2)
func L2DistanceSquared(a, b vector.Vector) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("dimension mismatch: %d vs %d", len(a), len(b))
	}
	if len(a) == 0 {
		return 0, fmt.Errorf("cannot calculate distance for empty vectors")
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return sum, nil
}

// CosineDistance calculates cosine distance (1 - cosine similarity)
// Formula: 1 - (dot(a, b) / (||a|| * ||b||))
// Range: [0, 2], where 0 = identical direction, 2 = opposite direction
func CosineDistance(a, b vector.Vector) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("dimension mismatch: %d vs %d", len(a), len(b))
	}
	if len(a) == 0 {
		return 0, fmt.Errorf("cannot calculate distance for empty vectors")
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	// Handle zero vectors
	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("cannot calculate cosine distance for zero vector")
	}

	// Cosine similarity is in [-1, 1], so distance is in [0, 2]
	similarity := dotProduct / (normA * normB)
	// Clamp to [-1, 1] to handle floating point errors
	similarity = math.Max(-1.0, math.Min(1.0, similarity))

	return 1.0 - similarity, nil
}

// DotProduct calculates negative dot product
// This is used as a "distance" metric where higher dot product = closer
// Formula: -sum(a[i] * b[i])
// Note: Negative because we want smaller values to mean "closer"
func DotProduct(a, b vector.Vector) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("dimension mismatch: %d vs %d", len(a), len(b))
	}
	if len(a) == 0 {
		return 0, fmt.Errorf("cannot calculate distance for empty vectors")
	}

	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return -sum, nil
}

// InnerProduct is an alias for DotProduct (negative dot product)
// Some systems use this terminology
var InnerProduct = DotProduct

// ValidatePair checks if two vectors can be used together for distance calculation
func ValidatePair(a, b vector.Vector) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("first vector invalid: %w", err)
	}
	if err := b.Validate(); err != nil {
		return fmt.Errorf("second vector invalid: %w", err)
	}
	if len(a) != len(b) {
		return fmt.Errorf("dimension mismatch: %d vs %d", len(a), len(b))
	}
	return nil
}
