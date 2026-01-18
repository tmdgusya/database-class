package exercise

import "github.com/tmdgusya/database-class/pkg/vector"

// Node represents a vector in the HNSW graph
type Node struct {
	ID          int              // Unique ID
	Vector      vector.Vector    // The actual vector
	Connections [][]int          // connections[layer] = list of neighbor IDs at that layer
	Level       int              // Maximum layer this node exists in (0 to Level)
}

// NewNode creates a new node
func NewNode(id int, v vector.Vector, level int) *Node {
	// TODO: Implement
	//
	// Tasks:
	// 1. Create node with given ID, vector, level
	// 2. Initialize Connections as slice of slices
	//    - Length: level+1 (layers 0 to level)
	//    - Each layer starts empty

	panic("not implemented")
}

// AddConnection adds a bidirectional connection at specified layer
// This is a helper - not required but useful
func (n *Node) AddConnection(neighborID int, layer int) {
	// TODO: Implement (optional)
	//
	// Add neighborID to this node's connections at layer
	// Note: This is one-directional - caller should also update neighbor

	panic("not implemented")
}
