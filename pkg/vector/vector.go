package vector

import (
	"fmt"
	"math"
)

// Vector represents a high-dimensional vector
type Vector []float64

// Dimension returns the dimensionality of the vector
func (v Vector) Dimension() int {
	return len(v)
}

// Clone creates a deep copy of the vector
func (v Vector) Clone() Vector {
	if v == nil {
		return nil
	}
	clone := make(Vector, len(v))
	copy(clone, v)
	return clone
}

// Validate checks if vector is valid (not nil, not empty, no NaN/Inf)
func (v Vector) Validate() error {
	if v == nil {
		return fmt.Errorf("vector is nil")
	}
	if len(v) == 0 {
		return fmt.Errorf("vector is empty")
	}
	for i, val := range v {
		if math.IsNaN(val) {
			return fmt.Errorf("invalid value at index %d: NaN", i)
		}
		if math.IsInf(val, 0) {
			return fmt.Errorf("invalid value at index %d: Inf", i)
		}
	}
	return nil
}

// Equal checks if two vectors are equal within epsilon tolerance
func (v Vector) Equal(other Vector, epsilon float64) bool {
	if len(v) != len(other) {
		return false
	}
	for i := range v {
		if math.Abs(v[i]-other[i]) > epsilon {
			return false
		}
	}
	return true
}

// String returns a string representation of the vector (truncated if too long)
func (v Vector) String() string {
	if len(v) == 0 {
		return "[]"
	}
	if len(v) <= 5 {
		return fmt.Sprintf("%v", []float64(v))
	}
	return fmt.Sprintf("[%v %v %v ... %v %v] (dim=%d)",
		v[0], v[1], v[2], v[len(v)-2], v[len(v)-1], len(v))
}
