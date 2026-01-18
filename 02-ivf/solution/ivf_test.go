package solution

import (
	"fmt"
	"testing"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/testdata"
	"github.com/tmdgusya/database-class/pkg/vector"
)

func TestNewIVFIndex(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		idx, err := NewIVFIndex(Config{
			Metric:      distance.L2Distance,
			NumClusters: 10,
			NumProbes:   3,
		})
		if err != nil {
			t.Fatalf("NewIVFIndex() failed: %v", err)
		}
		if idx == nil {
			t.Fatal("NewIVFIndex() returned nil")
		}
	})

	t.Run("nil metric", func(t *testing.T) {
		_, err := NewIVFIndex(Config{
			Metric:      nil,
			NumClusters: 10,
			NumProbes:   3,
		})
		if err == nil {
			t.Error("NewIVFIndex() should fail with nil metric")
		}
	})

	t.Run("nprobe > nlist", func(t *testing.T) {
		_, err := NewIVFIndex(Config{
			Metric:      distance.L2Distance,
			NumClusters: 10,
			NumProbes:   20, // More than clusters!
		})
		if err == nil {
			t.Error("NewIVFIndex() should fail when nprobe > nlist")
		}
	})
}

func TestIVFNotTrained(t *testing.T) {
	// TRAP: Using IVF before training should error
	idx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: 10,
		NumProbes:   3,
	})

	v := vector.Vector{1.0, 2.0, 3.0}

	// Try to add without training
	err := idx.Add(v)
	if err == nil {
		t.Error("Add() should fail when index not trained")
	}

	// Try to search without training
	_, err = idx.Search(v, 5)
	if err == nil {
		t.Error("Search() should fail when index not trained")
	}
}

func TestIVFBasicTrainAndSearch(t *testing.T) {
	// Generate clustered data (easier to get good results)
	vectors := testdata.GenerateClusteredVectors(100, 10, 5, 42)

	idx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: 5,
		NumProbes:   3,
	})

	// Train
	if err := idx.Train(vectors); err != nil {
		t.Fatalf("Train() failed: %v", err)
	}

	// Add vectors
	for _, v := range vectors {
		if err := idx.Add(v); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	// Search
	query := vectors[0]
	results, err := idx.Search(query, 5)
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Search() returned no results")
	}

	// First result should be the query itself (distance ~0)
	if results[0].Distance > 1e-6 {
		t.Errorf("First result distance = %f, want ~0", results[0].Distance)
	}
}

func TestIVFInsufficientTrainingData(t *testing.T) {
	// TRAP: Training with fewer vectors than clusters
	vectors := testdata.GenerateRandomVectors(5, 10, 42)

	idx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: 10, // More clusters than vectors!
		NumProbes:   3,
	})

	err := idx.Train(vectors)
	if err == nil {
		t.Error("Train() should fail or warn with insufficient data")
	}
}

// 🔥 핵심 함정 테스트!
// nprobe=1일 때 recall이 매우 낮음을 경험
func TestIVFRecallWithSmallNprobe(t *testing.T) {
	// Generate clustered data (realistic scenario)
	numVectors := 500
	dim := 64
	numClusters := 10

	vectors := testdata.GenerateClusteredVectors(numVectors, dim, numClusters, 42)
	queries := testdata.GenerateRandomVectors(20, dim, 123)

	// Build IVF index with nprobe=1 (TRAP!)
	ivfIdx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: numClusters,
		NumProbes:   1, // ⚠️ Too small!
	})

	if err := ivfIdx.Train(vectors); err != nil {
		t.Fatalf("Train() failed: %v", err)
	}

	for _, v := range vectors {
		if err := ivfIdx.Add(v); err != nil {
			t.Fatalf("Add() failed: %v", err)
		}
	}

	// Build Flat index for ground truth
	flatIdx := buildFlatIndex(vectors)

	// Measure recall
	k := 10
	recall := calculateRecall(ivfIdx, flatIdx, queries, k)

	fmt.Printf("\n📊 Recall with nprobe=1: %.1f%%\n", recall*100)

	// With nprobe=1, recall should be poor (typically 20-40%)
	if recall < 0.5 {
		t.Logf("\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n" +
			"❌ EXPECTED FAILURE: Recall is too low (%.1f%%)\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
			"💡 This is THE KEY LEARNING POINT of IVF!\n\n" +
			"What's happening?\n" +
			"  - nprobe=1 means we only search 1 cluster\n" +
			"  - But nearest neighbors might be in other clusters!\n" +
			"  - So we miss many correct results\n\n" +
			"How to fix?\n" +
			"  1. Increase nprobe parameter\n" +
			"  2. Try nprobe=3, 5, or 10\n" +
			"  3. Run TestIVFRecallVsNprobe to see the curve\n\n" +
			"Trade-off:\n" +
			"  nprobe ↑ → recall ↑ (good) but latency ↑ (slower)\n" +
			"  nprobe ↓ → latency ↓ (fast) but recall ↓ (bad)\n\n" +
			"Typical settings:\n" +
			"  - nprobe = 1:       30%% recall (too low!)\n" +
			"  - nprobe = 5:       75%% recall (acceptable)\n" +
			"  - nprobe = 10:      90%% recall (good) ✅\n" +
			"  - nprobe = nlist:  100%% recall (= Flat index)\n\n" +
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n",
			recall*100,
		)

		// Mark test as failed to ensure user sees this
		t.Errorf("Recall too low with nprobe=1: %.1f%%. This teaches the importance of nprobe parameter!",
			recall*100)
	}
}

// 파라미터 튜닝 학습을 위한 테스트
func TestIVFRecallVsNprobe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping parameter tuning test in short mode")
	}

	vectors := testdata.GenerateClusteredVectors(500, 64, 10, 42)
	queries := testdata.GenerateRandomVectors(20, 64, 123)

	idx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: 10,
		NumProbes:   1, // Will change dynamically
	})

	idx.Train(vectors)
	for _, v := range vectors {
		idx.Add(v)
	}

	flatIdx := buildFlatIndex(vectors)

	fmt.Println("\n📊 Recall vs nprobe:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━")

	for nprobe := 1; nprobe <= 10; nprobe++ {
		idx.SetNumProbes(nprobe)
		recall := calculateRecall(idx, flatIdx, queries, 10)
		fmt.Printf("nprobe=%2d: recall=%5.1f%%\n", nprobe, recall*100)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func TestIVFDimensionMismatch(t *testing.T) {
	vectors := testdata.GenerateRandomVectors(100, 10, 42)

	idx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: 5,
		NumProbes:   2,
	})

	idx.Train(vectors)
	for _, v := range vectors {
		idx.Add(v)
	}

	// Try to add different dimension
	wrongDimVec := vector.Vector{1.0, 2.0, 3.0} // 3D instead of 10D
	if err := idx.Add(wrongDimVec); err == nil {
		t.Error("Add() should fail with dimension mismatch")
	}

	// Try to search with different dimension
	wrongDimQuery := vector.Vector{1.0, 2.0}
	if _, err := idx.Search(wrongDimQuery, 5); err == nil {
		t.Error("Search() should fail with dimension mismatch")
	}
}

func TestSetNumProbes(t *testing.T) {
	idx, _ := NewIVFIndex(Config{
		Metric:      distance.L2Distance,
		NumClusters: 10,
		NumProbes:   5,
	})

	// Valid update
	if err := idx.SetNumProbes(7); err != nil {
		t.Errorf("SetNumProbes(7) failed: %v", err)
	}

	// Invalid: too large
	if err := idx.SetNumProbes(20); err == nil {
		t.Error("SetNumProbes(20) should fail when > nlist(10)")
	}

	// Invalid: too small
	if err := idx.SetNumProbes(0); err == nil {
		t.Error("SetNumProbes(0) should fail")
	}
}

// Helper functions

type flatIndex struct {
	vectors []vector.Vector
	metric  distance.Metric
}

func buildFlatIndex(vectors []vector.Vector) *flatIndex {
	return &flatIndex{
		vectors: vectors,
		metric:  distance.L2Distance,
	}
}

func (idx *flatIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	results := make([]SearchResult, len(idx.vectors))
	for i, v := range idx.vectors {
		dist, _ := idx.metric(query, v)
		results[i] = SearchResult{
			Vector:   v,
			Distance: dist,
			Index:    i,
		}
	}

	// Sort by distance
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Distance < results[i].Distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if k > len(results) {
		k = len(results)
	}

	return results[:k], nil
}

// Simple recall calculation
func calculateRecall(ivfIdx *IVFIndex, flatIdx *flatIndex, queries []vector.Vector, k int) float64 {
	var totalRecall float64

	for _, query := range queries {
		// Get IVF results
		ivfResults, err := ivfIdx.Search(query, k)
		if err != nil {
			continue
		}

		// Get ground truth
		flatResults, err := flatIdx.Search(query, k)
		if err != nil {
			continue
		}

		// Build set of ground truth indices
		truthSet := make(map[int]bool)
		for _, r := range flatResults {
			truthSet[r.Index] = true
		}

		// Count matches
		matches := 0
		for _, r := range ivfResults {
			if truthSet[r.Index] {
				matches++
			}
		}

		recall := float64(matches) / float64(k)
		totalRecall += recall
	}

	return totalRecall / float64(len(queries))
}
