package metrics

import (
	"fmt"

	"github.com/tmdgusya/database-class/pkg/vector"
)

// SearchResult represents a single search result
type SearchResult struct {
	Vector   vector.Vector
	Distance float64
	Index    int // Position in the original index
}

// Index interface for any vector index
// All indexes (Flat, IVF, HNSW) must implement this
type Index interface {
	Add(v vector.Vector) error
	Search(query vector.Vector, k int) ([]SearchResult, error)
	Size() int
}

// CalculateRecall measures recall@k by comparing approximate results against ground truth
// Recall = (number of correct results) / k
// Returns average recall across all queries
func CalculateRecall(
	approxResults [][]int, // Approximate k-NN indices for each query
	groundTruth [][]int, // True k-NN indices for each query
	k int,
) (float64, error) {
	if len(approxResults) == 0 || len(groundTruth) == 0 {
		return 0, fmt.Errorf("empty results or ground truth")
	}

	if len(approxResults) != len(groundTruth) {
		return 0, fmt.Errorf("mismatch: %d approximate results vs %d ground truth",
			len(approxResults), len(groundTruth))
	}

	var totalRecall float64

	for i := range approxResults {
		approx := approxResults[i]
		truth := groundTruth[i]

		if len(approx) > k {
			approx = approx[:k]
		}
		if len(truth) > k {
			truth = truth[:k]
		}

		// Create set of ground truth indices for fast lookup
		truthSet := make(map[int]bool)
		for _, idx := range truth {
			truthSet[idx] = true
		}

		// Count matches
		matches := 0
		for _, idx := range approx {
			if truthSet[idx] {
				matches++
			}
		}

		// Recall for this query
		recall := float64(matches) / float64(k)
		totalRecall += recall
	}

	// Average recall across all queries
	return totalRecall / float64(len(approxResults)), nil
}

// CalculateRecallAtK measures recall at different k values
// Returns a map of k -> recall
func CalculateRecallAtK(
	approxResults [][]int,
	groundTruth [][]int,
	kValues []int,
) (map[int]float64, error) {
	results := make(map[int]float64)

	for _, k := range kValues {
		recall, err := CalculateRecall(approxResults, groundTruth, k)
		if err != nil {
			return nil, fmt.Errorf("recall@%d error: %w", k, err)
		}
		results[k] = recall
	}

	return results, nil
}

// RecallResult holds detailed recall metrics
type RecallResult struct {
	K              int     // Number of neighbors
	Recall         float64 // Average recall
	MinRecall      float64 // Worst recall across queries
	MaxRecall      float64 // Best recall across queries
	NumQueries     int     // Number of queries evaluated
	PerfectMatches int     // Queries with 100% recall
}

// CalculateDetailedRecall computes detailed recall statistics
func CalculateDetailedRecall(
	approxResults [][]int,
	groundTruth [][]int,
	k int,
) (*RecallResult, error) {
	if len(approxResults) == 0 || len(groundTruth) == 0 {
		return nil, fmt.Errorf("empty results or ground truth")
	}

	if len(approxResults) != len(groundTruth) {
		return nil, fmt.Errorf("mismatch: %d approximate results vs %d ground truth",
			len(approxResults), len(groundTruth))
	}

	result := &RecallResult{
		K:          k,
		MinRecall:  1.0,
		MaxRecall:  0.0,
		NumQueries: len(approxResults),
	}

	var totalRecall float64

	for i := range approxResults {
		approx := approxResults[i]
		truth := groundTruth[i]

		if len(approx) > k {
			approx = approx[:k]
		}
		if len(truth) > k {
			truth = truth[:k]
		}

		// Create set of ground truth indices
		truthSet := make(map[int]bool)
		for _, idx := range truth {
			truthSet[idx] = true
		}

		// Count matches
		matches := 0
		for _, idx := range approx {
			if truthSet[idx] {
				matches++
			}
		}

		// Query recall
		queryRecall := float64(matches) / float64(k)
		totalRecall += queryRecall

		// Update statistics
		if queryRecall < result.MinRecall {
			result.MinRecall = queryRecall
		}
		if queryRecall > result.MaxRecall {
			result.MaxRecall = queryRecall
		}
		if queryRecall == 1.0 {
			result.PerfectMatches++
		}
	}

	result.Recall = totalRecall / float64(len(approxResults))

	return result, nil
}

// ExtractIndices extracts index values from SearchResults
// Helper function to convert SearchResult slices to index slices
func ExtractIndices(results []SearchResult) []int {
	indices := make([]int, len(results))
	for i, r := range results {
		indices[i] = r.Index
	}
	return indices
}

// ExtractIndicesFromBatch extracts indices from a batch of search results
func ExtractIndicesFromBatch(batch [][]SearchResult) [][]int {
	indices := make([][]int, len(batch))
	for i, results := range batch {
		indices[i] = ExtractIndices(results)
	}
	return indices
}
