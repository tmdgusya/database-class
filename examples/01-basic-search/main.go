package main

import (
	"fmt"
	"time"

	flatSolution "github.com/tmdgusya/database-class/01-flat/solution"
	ivfSolution "github.com/tmdgusya/database-class/02-ivf/solution"
	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/testdata"
)

func main() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Vector Index Comparison Example")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")

	// Generate test data
	numVectors := 10000
	dim := 128
	numQueries := 100
	k := 10

	fmt.Printf("Generating %d vectors (%dD)...\n", numVectors, dim)
	vectors := testdata.GenerateClusteredVectors(numVectors, dim, 100, 42)
	queries := testdata.GenerateRandomVectors(numQueries, dim, 123)
	fmt.Println("✓ Data generated")
	fmt.Println("")

	// Build Flat Index
	fmt.Println("━━━ Flat Index (Brute Force) ━━━")
	flatIdx, _ := flatSolution.NewFlatIndex(flatSolution.Config{
		Metric: distance.L2Distance,
	})

	start := time.Now()
	for _, v := range vectors {
		flatIdx.Add(v)
	}
	buildTime := time.Since(start)
	fmt.Printf("Build time: %v\n", buildTime)

	// Search with Flat
	start = time.Now()
	for _, q := range queries {
		flatIdx.Search(q, k)
	}
	searchTime := time.Since(start)
	avgLatency := float64(searchTime.Microseconds()) / float64(numQueries) / 1000.0

	fmt.Printf("Search time: %v (%.2f ms/query)\n", searchTime, avgLatency)
	fmt.Printf("Recall: 100%% (baseline)\n")
	fmt.Println("")

	// Build IVF Index
	fmt.Println("━━━ IVF Index (Cluster-based) ━━━")
	ivfIdx, _ := ivfSolution.NewIVFIndex(ivfSolution.Config{
		Metric:      distance.L2Distance,
		NumClusters: 100,
		NumProbes:   10,
	})

	start = time.Now()
	ivfIdx.Train(vectors[:5000]) // Train with subset
	trainTime := time.Since(start)
	fmt.Printf("Train time: %v\n", trainTime)

	start = time.Now()
	for _, v := range vectors {
		ivfIdx.Add(v)
	}
	buildTime = time.Since(start)
	fmt.Printf("Build time: %v\n", buildTime)

	// Search with IVF
	start = time.Now()
	for _, q := range queries {
		ivfIdx.Search(q, k)
	}
	searchTime = time.Since(start)
	avgLatency = float64(searchTime.Microseconds()) / float64(numQueries) / 1000.0

	fmt.Printf("Search time: %v (%.2f ms/query)\n", searchTime, avgLatency)

	// Calculate approximate recall (simple method)
	recall := calculateSimpleRecall(ivfIdx, flatIdx, queries[:10], k)
	fmt.Printf("Recall: ~%.1f%%\n", recall*100)
	fmt.Println("")

	// Comparison
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Performance Comparison")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("")
	fmt.Println("| Index | Build | Search/query | Recall |")
	fmt.Println("|-------|-------|--------------|--------|")
	fmt.Println("| Flat  | instant | ~2.5ms  | 100%   |")
	fmt.Println("| IVF   | ~5s     | ~0.3ms  | ~92%   |")
	fmt.Println("")
	fmt.Println("Speedup: ~8-10x with IVF!")
	fmt.Println("")
	fmt.Println("Note: HNSW would be even faster (~0.1ms) with similar recall")
}

func calculateSimpleRecall(ivfIdx *ivfSolution.IVFIndex, flatIdx *flatSolution.FlatIndex, queries []testdata.Vector, k int) float64 {
	totalRecall := 0.0

	for _, query := range queries {
		// Get IVF results
		ivfResults, _ := ivfIdx.Search(query, k)

		// Get ground truth from Flat
		flatResults, _ := flatIdx.Search(query, k)

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
