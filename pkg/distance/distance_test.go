package distance

import (
	"math"
	"testing"

	"github.com/tmdgusya/database-class/pkg/vector"
)

func TestL2Distance(t *testing.T) {
	tests := []struct {
		name     string
		a        vector.Vector
		b        vector.Vector
		expected float64
		wantErr  bool
	}{
		{
			"identical vectors",
			vector.Vector{1.0, 2.0, 3.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0,
			false,
		},
		{
			"simple distance",
			vector.Vector{0.0, 0.0, 0.0},
			vector.Vector{3.0, 4.0, 0.0},
			5.0, // 3-4-5 triangle
			false,
		},
		{
			"1D distance",
			vector.Vector{0.0},
			vector.Vector{5.0},
			5.0,
			false,
		},
		{
			"dimension mismatch",
			vector.Vector{1.0, 2.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0,
			true,
		},
		{
			"empty vectors",
			vector.Vector{},
			vector.Vector{},
			0.0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := L2Distance(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("L2Distance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("L2Distance() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestL2DistanceSquared(t *testing.T) {
	tests := []struct {
		name     string
		a        vector.Vector
		b        vector.Vector
		expected float64
		wantErr  bool
	}{
		{
			"identical vectors",
			vector.Vector{1.0, 2.0, 3.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0,
			false,
		},
		{
			"simple distance",
			vector.Vector{0.0, 0.0, 0.0},
			vector.Vector{3.0, 4.0, 0.0},
			25.0, // 3^2 + 4^2 = 25
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := L2DistanceSquared(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("L2DistanceSquared() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("L2DistanceSquared() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCosineDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        vector.Vector
		b        vector.Vector
		expected float64
		wantErr  bool
	}{
		{
			"identical vectors",
			vector.Vector{1.0, 2.0, 3.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0, // distance = 1 - 1 = 0
			false,
		},
		{
			"orthogonal vectors",
			vector.Vector{1.0, 0.0, 0.0},
			vector.Vector{0.0, 1.0, 0.0},
			1.0, // distance = 1 - 0 = 1
			false,
		},
		{
			"opposite vectors",
			vector.Vector{1.0, 0.0, 0.0},
			vector.Vector{-1.0, 0.0, 0.0},
			2.0, // distance = 1 - (-1) = 2
			false,
		},
		{
			"parallel vectors different magnitude",
			vector.Vector{1.0, 2.0, 3.0},
			vector.Vector{2.0, 4.0, 6.0},
			0.0, // same direction
			false,
		},
		{
			"zero vector",
			vector.Vector{0.0, 0.0, 0.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0,
			true,
		},
		{
			"dimension mismatch",
			vector.Vector{1.0, 2.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CosineDistance(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("CosineDistance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("CosineDistance() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDotProduct(t *testing.T) {
	tests := []struct {
		name     string
		a        vector.Vector
		b        vector.Vector
		expected float64
		wantErr  bool
	}{
		{
			"simple dot product",
			vector.Vector{1.0, 2.0, 3.0},
			vector.Vector{4.0, 5.0, 6.0},
			-(1*4 + 2*5 + 3*6), // negative because we want smaller = closer
			false,
		},
		{
			"orthogonal vectors",
			vector.Vector{1.0, 0.0, 0.0},
			vector.Vector{0.0, 1.0, 0.0},
			0.0,
			false,
		},
		{
			"zero vectors",
			vector.Vector{0.0, 0.0, 0.0},
			vector.Vector{0.0, 0.0, 0.0},
			0.0,
			false,
		},
		{
			"dimension mismatch",
			vector.Vector{1.0, 2.0},
			vector.Vector{1.0, 2.0, 3.0},
			0.0,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DotProduct(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("DotProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && math.Abs(got-tt.expected) > 1e-9 {
				t.Errorf("DotProduct() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidatePair(t *testing.T) {
	tests := []struct {
		name    string
		a       vector.Vector
		b       vector.Vector
		wantErr bool
	}{
		{
			"valid pair",
			vector.Vector{1.0, 2.0, 3.0},
			vector.Vector{4.0, 5.0, 6.0},
			false,
		},
		{
			"dimension mismatch",
			vector.Vector{1.0, 2.0},
			vector.Vector{1.0, 2.0, 3.0},
			true,
		},
		{
			"first vector nil",
			nil,
			vector.Vector{1.0, 2.0, 3.0},
			true,
		},
		{
			"second vector nil",
			vector.Vector{1.0, 2.0, 3.0},
			nil,
			true,
		},
		{
			"first vector contains NaN",
			vector.Vector{1.0, math.NaN(), 3.0},
			vector.Vector{1.0, 2.0, 3.0},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePair(tt.a, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePair() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark tests
func BenchmarkL2Distance(b *testing.B) {
	dims := []int{128, 512, 1024}
	for _, dim := range dims {
		b.Run(fmt.Sprintf("dim=%d", dim), func(b *testing.B) {
			v1 := make(vector.Vector, dim)
			v2 := make(vector.Vector, dim)
			for i := 0; i < dim; i++ {
				v1[i] = float64(i)
				v2[i] = float64(i + 1)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = L2Distance(v1, v2)
			}
		})
	}
}

func BenchmarkCosineDistance(b *testing.B) {
	dims := []int{128, 512, 1024}
	for _, dim := range dims {
		b.Run(fmt.Sprintf("dim=%d", dim), func(b *testing.B) {
			v1 := make(vector.Vector, dim)
			v2 := make(vector.Vector, dim)
			for i := 0; i < dim; i++ {
				v1[i] = float64(i)
				v2[i] = float64(i + 1)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = CosineDistance(v1, v2)
			}
		})
	}
}
