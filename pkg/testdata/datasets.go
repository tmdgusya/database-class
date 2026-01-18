package testdata

import "github.com/tmdgusya/database-class/pkg/vector"

// Predefined datasets for quick testing
// These are initialized once and can be reused across tests

var (
	// TinyDataset2D - 10 vectors in 2D for visualization
	// Useful for understanding algorithms visually
	TinyDataset2D []vector.Vector

	// SmallDataset - 100 vectors in 128D
	// Quick tests that don't need much data
	SmallDataset []vector.Vector

	// MediumDataset - 1000 vectors in 128D
	// Standard test size
	MediumDataset []vector.Vector

	// ClusteredData - 1000 vectors in 10 clusters, 128D
	// Critical for testing IVF and understanding clustering
	ClusteredData []vector.Vector

	// NormalizedData - 1000 normalized vectors in 128D
	// For cosine distance testing
	NormalizedData []vector.Vector
)

func init() {
	// Initialize predefined datasets with fixed seeds for reproducibility
	const (
		seed = 42 // The answer to everything
		dim  = 128
	)

	// 2D dataset for visualization
	TinyDataset2D = []vector.Vector{
		{0.1, 0.1}, {0.2, 0.1}, {0.1, 0.2},
		{0.8, 0.8}, {0.9, 0.8}, {0.8, 0.9},
		{0.5, 0.1}, {0.1, 0.5},
		{0.5, 0.9}, {0.9, 0.5},
	}

	// Small dataset for quick tests
	SmallDataset = GenerateRandomVectors(100, dim, seed)

	// Medium dataset for standard tests
	MediumDataset = GenerateRandomVectors(1000, dim, seed+1)

	// Clustered dataset - critical for IVF testing
	ClusteredData = GenerateClusteredVectors(1000, dim, 10, seed+2)

	// Normalized dataset for cosine distance
	NormalizedData = GenerateNormalizedVectors(1000, dim, seed+3)
}

// DatasetInfo provides metadata about predefined datasets
type DatasetInfo struct {
	Name        string
	Count       int
	Dimension   int
	Description string
}

// AvailableDatasets returns information about all predefined datasets
func AvailableDatasets() []DatasetInfo {
	return []DatasetInfo{
		{
			Name:        "TinyDataset2D",
			Count:       len(TinyDataset2D),
			Dimension:   2,
			Description: "10 vectors in 2D for visualization and debugging",
		},
		{
			Name:        "SmallDataset",
			Count:       len(SmallDataset),
			Dimension:   128,
			Description: "100 random vectors in 128D for quick tests",
		},
		{
			Name:        "MediumDataset",
			Count:       len(MediumDataset),
			Dimension:   128,
			Description: "1000 random vectors in 128D for standard tests",
		},
		{
			Name:        "ClusteredData",
			Count:       len(ClusteredData),
			Dimension:   128,
			Description: "1000 vectors in 10 clusters for IVF testing",
		},
		{
			Name:        "NormalizedData",
			Count:       len(NormalizedData),
			Dimension:   128,
			Description: "1000 normalized vectors for cosine distance",
		},
	}
}
