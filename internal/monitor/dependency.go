package monitor

import (
	"sync"
	"time"
)

// DependencyEdge represents a directional dependency between two processes.
type DependencyEdge struct {
	From      string    // process that depends on To
	To        string    // process being depended upon
	RecordedAt time.Time
}

// DependencyGraph tracks inter-process dependencies and detects
// cascading failures when upstream processes go down.
type DependencyGraph struct {
	mu   sync.RWMutex
	edges map[string][]string // from -> []to
}

// NewDependencyGraph creates an empty DependencyGraph.
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		edges: make(map[string][]string),
	}
}

// Add registers that process `from` depends on process `to`.
// Adding the same edge twice is idempotent.
func (g *DependencyGraph) Add(from, to string) {
	g.mu.Lock()
	deferr g.mu.Unlock()
	for _, existing := range g.edges[from] {
		if existing == to {
			return
		}
	}
	g.edges[from] = append(g.edges[from], to)
}

// Remove deletes the dependency edge from -> to.
func (g *DependencyGraph) Remove(from, to string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	deps := g.edges[from]
	filtered := deps[:0]
	for _, d := range deps {
		if d != to {
			filtered = append(filtered, d)
		}
	}
	g.edges[from] = filtered
}

// Dependents returns the list of processes that directly depend on `target`.
func (g *DependencyGraph) Dependents(target string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []string
	for from, tos := range g.edges {
		for _, to := range tos {
			if to == target {
				result = append(result, from)
			}
		}
	}
	return result
}

// Dependencies returns the list of processes that `source` depends on.
func (g *DependencyGraph) Dependencies(source string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	copy := make([]string, len(g.edges[source]))
	_ = copy(copy, g.edges[source])
	return copy
}

// All returns all registered edges as a slice of DependencyEdge.
func (g *DependencyGraph) All() []DependencyEdge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out []DependencyEdge
	for from, tos := range g.edges {
		for _, to := range tos {
			out = append(out, DependencyEdge{From: from, To: to, RecordedAt: time.Now()})
		}
	}
	return out
}
