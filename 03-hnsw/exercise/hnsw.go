package exercise

import (
	"sync"

	"github.com/tmdgusya/database-class/pkg/distance"
	"github.com/tmdgusya/database-class/pkg/vector"
)

// HNSWIndex implements Hierarchical Navigable Small World graph
type HNSWIndex struct {
	// TODO: Add necessary fields
	// Hints:
	// - nodes []Node (all nodes in graph)
	// - entryPoint int (ID of entry node)
	// - maxLayer int (current max layer in graph)
	// - M int (max connections per layer)
	// - Mmax int (max connections at layer 0)
	// - efConstruction int (construction-time ef)
	// - efSearch int (search-time ef)
	// - ml float64 (level generation multiplier)
	// - metric distance.Metric
	// - dimension int
	// - mu sync.RWMutex

	mu sync.RWMutex // Provided
}

// Config holds HNSW parameters
type Config struct {
	Metric         distance.Metric
	M              int     // Max bidirectional connections per layer
	Mmax           int     // Max connections at layer 0 (typically M*2)
	EfConstruction int     // Construction-time candidate list size
	EfSearch       int     // Search-time candidate list size
	Ml             float64 // Level generation multiplier (default: 1/ln(2))
}

// SearchResult represents a search result
type SearchResult struct {
	Vector   vector.Vector
	Distance float64
	Index    int
}

// NewHNSWIndex creates a new HNSW index
func NewHNSWIndex(cfg Config) (*HNSWIndex, error) {
	// TODO: Implement
	//
	// Tasks:
	// 1. Validate parameters
	//    - M > 0
	//    - EfConstruction >= M
	//    - EfSearch > 0
	//    - If Mmax not set, use M*2
	//    - If Ml not set, use 1/ln(2)
	// 2. Initialize index
	//
	// Typical values:
	// M = 16, Mmax = 32, EfConstruction = 200, EfSearch = 50

	panic("not implemented")
}

// Add inserts a vector into the graph
func (idx *HNSWIndex) Add(v vector.Vector) error {
	// TODO: Implement - This is the MOST COMPLEX method!
	//
	// Algorithm:
	// 1. Validate vector
	// 2. Generate random level
	// 3. Create new node
	// 4. If first node: set as entry point
	// 5. Find insertion points layer by layer:
	//    - Start from entry point at top layer
	//    - Greedy search to current layer
	//    - Insert and connect at each layer from top down to 0
	// 6. Update entry point if new node is higher
	//
	// Detailed steps:
	// currNearest = [entryPoint]
	//
	// for layer from maxLayer down to newNode.Level+1:
	//     currNearest = searchLayer(v, currNearest, ef=1, layer)
	//
	// for layer from newNode.Level down to 0:
	//     candidates = searchLayer(v, currNearest, efConstruction, layer)
	//     neighbors = selectNeighbors(candidates, M, layer)
	//
	//     # Bidirectional connect
	//     for each neighbor:
	//         add bidirectional connection
	//
	//     # Prune if needed
	//     for each neighbor:
	//         if len(neighbor.connections[layer]) > Mmax:
	//             prune neighbor connections
	//
	//     currNearest = neighbors
	//
	// if newNode.Level > maxLayer:
	//     entryPoint = newNode
	//     maxLayer = newNode.Level

	panic("not implemented")
}

// Search performs k-NN search
func (idx *HNSWIndex) Search(query vector.Vector, k int) ([]SearchResult, error) {
	// TODO: Implement
	//
	// Algorithm:
	// 1. Validate query and k
	// 2. Start from entry point
	// 3. Greedy search through layers:
	//    for layer from maxLayer down to 1:
	//        currNearest = searchLayer(query, [currNearest], ef=1, layer)
	// 4. Precise search at layer 0:
	//    candidates = searchLayer(query, [currNearest], efSearch, 0)
	// 5. Return top k
	//
	// TRAP: efSearch < k → error or fewer results!

	panic("not implemented")
}

// searchLayer performs greedy search within a single layer
// This is the CORE algorithm of HNSW!
func (idx *HNSWIndex) searchLayer(
	query vector.Vector,
	entryPoints []int,
	ef int,
	layer int,
) []nodeWithDistance {
	// TODO: Implement - CRITICAL METHOD!
	//
	// Algorithm:
	// visited = set()
	// candidates = min-heap(entryPoints)  // To explore
	// best = max-heap(entryPoints)        // Keep top ef
	//
	// while candidates not empty:
	//     curr = pop closest from candidates
	//
	//     if curr.distance > best.worst.distance:
	//         break  // Can't improve
	//
	//     for neighbor in curr.connections[layer]:
	//         if neighbor in visited:
	//             continue
	//
	//         visited.add(neighbor)
	//         dist = distance(query, neighbor.vector)
	//
	//         if dist < best.worst.distance or len(best) < ef:
	//             push to candidates
	//             push to best
	//
	//             if len(best) > ef:
	//                 pop worst from best
	//
	// return best (sorted by distance)
	//
	// Hints:
	// - Use container/heap package
	// - Need min-heap and max-heap
	// - visited prevents infinite loops (CRITICAL!)

	panic("not implemented")
}

// selectNeighbors selects M best neighbors from candidates
func (idx *HNSWIndex) selectNeighbors(
	candidates []nodeWithDistance,
	M int,
	layer int,
) []int {
	// TODO: Implement
	//
	// Simple version (recommended for first implementation):
	//   1. Sort candidates by distance
	//   2. Return first M IDs
	//
	// Advanced version (better quality):
	//   Use heuristic that considers both distance and diversity
	//   Avoids too many clustered neighbors
	//
	// Start with simple version!

	panic("not implemented")
}

// SetEfSearch updates efSearch parameter at runtime
func (idx *HNSWIndex) SetEfSearch(ef int) error {
	// TODO: Implement
	//
	// Validate: ef > 0
	// Update idx.efSearch

	panic("not implemented")
}

// Size returns number of vectors
func (idx *HNSWIndex) Size() int {
	// TODO: Implement
	panic("not implemented")
}

// Helper type for search
type nodeWithDistance struct {
	nodeID   int
	distance float64
}

// You'll need to implement heap interface for nodeWithDistance
// Example for min-heap:
//
// type minHeap []nodeWithDistance
//
// func (h minHeap) Len() int { return len(h) }
// func (h minHeap) Less(i, j int) bool { return h[i].distance < h[j].distance }
// func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
// func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(nodeWithDistance)) }
// func (h *minHeap) Pop() interface{} {
//     old := *h
//     n := len(old)
//     x := old[n-1]
//     *h = old[0 : n-1]
//     return x
// }
//
// Similarly for max-heap
