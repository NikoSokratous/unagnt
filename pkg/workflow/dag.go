package workflow

import (
	"fmt"
	"sort"
)

// DAG represents a Directed Acyclic Graph for workflow execution.
type DAG struct {
	Nodes map[string]*Node
	Edges map[string][]string // node -> dependencies
}

// NodeType is the type of workflow node.
type NodeType string

const (
	NodeTypeAgent    NodeType = "agent"    // default: run agent
	NodeTypeApproval NodeType = "approval" // human-in-the-loop: pause for approval
)

// Node represents a node in the DAG.
type Node struct {
	ID           string
	Name         string
	Type         NodeType // agent (default) or approval
	Agent        string
	Goal         string
	Condition    string
	OutputKey    string
	Timeout      string
	Retry        int
	Dependencies []string
	Metadata     map[string]interface{}
	// Approval step fields (when Type == NodeTypeApproval)
	Approvers       []string
	ApprovalMessage string
}

// NewDAG creates a new DAG.
func NewDAG() *DAG {
	return &DAG{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
	}
}

// AddNode adds a node to the DAG.
func (d *DAG) AddNode(node *Node) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}
	if _, exists := d.Nodes[node.ID]; exists {
		return fmt.Errorf("node %s already exists", node.ID)
	}

	d.Nodes[node.ID] = node
	d.Edges[node.ID] = make([]string, 0)

	return nil
}

// AddEdge adds a dependency edge (from depends on to).
func (d *DAG) AddEdge(from, to string) error {
	if _, exists := d.Nodes[from]; !exists {
		return fmt.Errorf("node %s does not exist", from)
	}
	if _, exists := d.Nodes[to]; !exists {
		return fmt.Errorf("node %s does not exist", to)
	}

	// Add dependency
	d.Edges[from] = append(d.Edges[from], to)

	// Check for cycles
	if d.hasCycle() {
		// Remove the edge
		deps := d.Edges[from]
		for i, dep := range deps {
			if dep == to {
				d.Edges[from] = append(deps[:i], deps[i+1:]...)
				break
			}
		}
		return fmt.Errorf("adding edge %s -> %s would create a cycle", from, to)
	}

	return nil
}

// TopologicalSort returns nodes in execution order.
func (d *DAG) TopologicalSort() ([]string, error) {
	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for nodeID := range d.Nodes {
		inDegree[nodeID] = 0
	}

	for _, deps := range d.Edges {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// Find nodes with in-degree 0 (no dependencies)
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// Sort queue for deterministic order
	sort.Strings(queue)

	result := make([]string, 0, len(d.Nodes))

	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for neighbors
		neighbors := d.Edges[current]
		sort.Strings(neighbors) // Deterministic order

		for _, neighbor := range neighbors {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
				sort.Strings(queue) // Keep queue sorted
			}
		}
	}

	// Check if all nodes were processed
	if len(result) != len(d.Nodes) {
		return nil, fmt.Errorf("DAG contains a cycle")
	}

	return result, nil
}

// GetExecutionLevels returns nodes grouped by execution level.
// Nodes in the same level can execute in parallel.
func (d *DAG) GetExecutionLevels() ([][]string, error) {
	// Calculate in-degree
	inDegree := make(map[string]int)
	for nodeID := range d.Nodes {
		inDegree[nodeID] = 0
	}

	for _, deps := range d.Edges {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	levels := make([][]string, 0)
	processed := make(map[string]bool)

	for len(processed) < len(d.Nodes) {
		// Find all nodes with in-degree 0 (not yet processed)
		currentLevel := make([]string, 0)

		for nodeID, degree := range inDegree {
			if !processed[nodeID] && degree == 0 {
				currentLevel = append(currentLevel, nodeID)
			}
		}

		if len(currentLevel) == 0 {
			return nil, fmt.Errorf("DAG contains a cycle or orphaned nodes")
		}

		// Sort level for deterministic order
		sort.Strings(currentLevel)
		levels = append(levels, currentLevel)

		// Mark as processed and reduce in-degree for neighbors
		for _, nodeID := range currentLevel {
			processed[nodeID] = true

			for _, neighbor := range d.Edges[nodeID] {
				inDegree[neighbor]--
			}
		}
	}

	return levels, nil
}

// hasCycle checks if the DAG contains a cycle using DFS.
func (d *DAG) hasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range d.Nodes {
		if !visited[nodeID] {
			if d.hasCycleDFS(nodeID, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// hasCycleDFS performs DFS to detect cycles.
func (d *DAG) hasCycleDFS(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, neighbor := range d.Edges[nodeID] {
		if !visited[neighbor] {
			if d.hasCycleDFS(neighbor, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// Validate checks if the DAG is valid.
func (d *DAG) Validate() error {
	if len(d.Nodes) == 0 {
		return fmt.Errorf("DAG has no nodes")
	}

	// Check for cycles
	if d.hasCycle() {
		return fmt.Errorf("DAG contains a cycle")
	}

	// Check that all edges reference existing nodes
	for nodeID, deps := range d.Edges {
		if _, exists := d.Nodes[nodeID]; !exists {
			return fmt.Errorf("edge references non-existent node: %s", nodeID)
		}
		for _, dep := range deps {
			if _, exists := d.Nodes[dep]; !exists {
				return fmt.Errorf("edge from %s references non-existent node: %s", nodeID, dep)
			}
		}
	}

	return nil
}

// GetRoots returns all root nodes (no dependencies).
func (d *DAG) GetRoots() []string {
	roots := make([]string, 0)

	for nodeID := range d.Nodes {
		hasIncoming := false
		for _, deps := range d.Edges {
			for _, dep := range deps {
				if dep == nodeID {
					hasIncoming = true
					break
				}
			}
			if hasIncoming {
				break
			}
		}
		if !hasIncoming {
			roots = append(roots, nodeID)
		}
	}

	sort.Strings(roots)
	return roots
}

// GetLeaves returns all leaf nodes (no dependents).
func (d *DAG) GetLeaves() []string {
	leaves := make([]string, 0)

	for nodeID := range d.Nodes {
		if len(d.Edges[nodeID]) == 0 {
			leaves = append(leaves, nodeID)
		}
	}

	sort.Strings(leaves)
	return leaves
}

// Clone creates a deep copy of the DAG.
func (d *DAG) Clone() *DAG {
	clone := NewDAG()

	// Copy nodes
	for id, node := range d.Nodes {
		nodeCopy := *node
		clone.Nodes[id] = &nodeCopy
	}

	// Copy edges
	for from, deps := range d.Edges {
		clone.Edges[from] = make([]string, len(deps))
		copy(clone.Edges[from], deps)
	}

	return clone
}

// ToDOT generates a Graphviz DOT representation.
func (d *DAG) ToDOT() string {
	dot := "digraph workflow {\n"
	dot += "  rankdir=TB;\n"
	dot += "  node [shape=box, style=rounded];\n\n"

	// Add nodes
	for id, node := range d.Nodes {
		label := node.Name
		if label == "" {
			label = id
		}
		dot += fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", id, label)
	}

	dot += "\n"

	// Add edges
	for from, deps := range d.Edges {
		for _, to := range deps {
			dot += fmt.Sprintf("  \"%s\" -> \"%s\";\n", from, to)
		}
	}

	dot += "}\n"
	return dot
}

// GetNodeDependencies returns all dependencies for a node.
func (d *DAG) GetNodeDependencies(nodeID string) []string {
	return d.Edges[nodeID]
}

// GetNodeDependents returns all nodes that depend on this node.
func (d *DAG) GetNodeDependents(nodeID string) []string {
	dependents := make([]string, 0)

	for id, deps := range d.Edges {
		for _, dep := range deps {
			if dep == nodeID {
				dependents = append(dependents, id)
				break
			}
		}
	}

	sort.Strings(dependents)
	return dependents
}
