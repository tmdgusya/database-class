package vector

import (
	"math"
	"testing"
)

func TestDimension(t *testing.T) {
	tests := []struct {
		name     string
		vector   Vector
		expected int
	}{
		{"empty vector", Vector{}, 0},
		{"1D vector", Vector{1.0}, 1},
		{"3D vector", Vector{1.0, 2.0, 3.0}, 3},
		{"128D vector", make(Vector, 128), 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vector.Dimension(); got != tt.expected {
				t.Errorf("Dimension() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClone(t *testing.T) {
	tests := []struct {
		name   string
		vector Vector
	}{
		{"nil vector", nil},
		{"empty vector", Vector{}},
		{"simple vector", Vector{1.0, 2.0, 3.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.vector.Clone()

			// Check if nil handling works
			if tt.vector == nil {
				if clone != nil {
					t.Errorf("Clone() of nil should be nil, got %v", clone)
				}
				return
			}

			// Check deep copy
			if len(tt.vector) > 0 {
				if &tt.vector[0] == &clone[0] {
					t.Error("Clone() should create a deep copy, not share memory")
				}
			}

			// Check values are equal
			if !clone.Equal(tt.vector, 1e-9) {
				t.Errorf("Clone() values differ: got %v, want %v", clone, tt.vector)
			}

			// Modify clone and ensure original is unchanged
			if len(clone) > 0 {
				clone[0] = 999.0
				if tt.vector[0] == 999.0 {
					t.Error("Modifying clone affected original vector")
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		vector  Vector
		wantErr bool
	}{
		{"valid vector", Vector{1.0, 2.0, 3.0}, false},
		{"valid zero vector", Vector{0.0, 0.0, 0.0}, false},
		{"valid negative", Vector{-1.0, -2.0, -3.0}, false},
		{"nil vector", nil, true},
		{"empty vector", Vector{}, true},
		{"contains NaN", Vector{1.0, math.NaN(), 3.0}, true},
		{"contains +Inf", Vector{1.0, math.Inf(1), 3.0}, true},
		{"contains -Inf", Vector{1.0, math.Inf(-1), 3.0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vector.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		v1       Vector
		v2       Vector
		epsilon  float64
		expected bool
	}{
		{
			"identical vectors",
			Vector{1.0, 2.0, 3.0},
			Vector{1.0, 2.0, 3.0},
			1e-9,
			true,
		},
		{
			"different lengths",
			Vector{1.0, 2.0},
			Vector{1.0, 2.0, 3.0},
			1e-9,
			false,
		},
		{
			"different values",
			Vector{1.0, 2.0, 3.0},
			Vector{1.0, 2.0, 4.0},
			1e-9,
			false,
		},
		{
			"within epsilon",
			Vector{1.0, 2.0, 3.0},
			Vector{1.0, 2.0000001, 3.0},
			1e-6,
			true,
		},
		{
			"outside epsilon",
			Vector{1.0, 2.0, 3.0},
			Vector{1.0, 2.001, 3.0},
			1e-6,
			false,
		},
		{
			"empty vectors",
			Vector{},
			Vector{},
			1e-9,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Equal(tt.v2, tt.epsilon); got != tt.expected {
				t.Errorf("Equal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name   string
		vector Vector
	}{
		{"empty", Vector{}},
		{"short vector", Vector{1.0, 2.0, 3.0}},
		{"long vector", Vector{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.vector.String()
			if s == "" {
				t.Error("String() should not return empty string")
			}
		})
	}
}
