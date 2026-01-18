package exercise

import "math"

// RandomLevel generates a random level for a new node
// Uses exponential decay distribution
func RandomLevel(ml float64, maxLevel int) int {
	// TODO: Implement
	//
	// Algorithm:
	// level = 0
	// while random() < ml and level < maxLevel:
	//     level++
	// return level
	//
	// Parameters:
	// - ml: multiplier, typically 1/ln(2) ≈ 0.69
	// - maxLevel: maximum allowed level
	//
	// Distribution:
	// P(level=0) ≈ 50%
	// P(level=1) ≈ 25%
	// P(level=2) ≈ 12.5%
	// ...
	//
	// Hints:
	// - import "math/rand"
	// - Use rand.Float64() for random [0,1)

	panic("not implemented")
}

// DefaultMl returns the typical ml value
func DefaultMl() float64 {
	return 1.0 / math.Log(2.0) // ≈ 0.69
}
