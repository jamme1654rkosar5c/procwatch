package monitor

import (
	"sync"
	"time"
)

// ProbeResult holds the outcome of a single liveness probe.
type ProbeResult struct {
	Process   string
	Success   bool
	CheckedAt time.Time
	Message   string
}

// ProbeStore tracks the latest liveness probe result per process.
type ProbeStore struct {
	mu      sync.RWMutex
	results map[string]ProbeResult
}

// NewProbeStore creates an empty ProbeStore.
func NewProbeStore() *ProbeStore {
	return &ProbeStore{
		results: make(map[string]ProbeResult),
	}
}

// Record stores a probe result for the given process.
func (p *ProbeStore) Record(process string, success bool, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.results[process] = ProbeResult{
		Process:   process,
		Success:   success,
		CheckedAt: time.Now(),
		Message:   message,
	}
}

// Get returns the latest probe result for a process and whether it exists.
func (p *ProbeStore) Get(process string) (ProbeResult, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	r, ok := p.results[process]
	return r, ok
}

// All returns a copy of all probe results.
func (p *ProbeStore) All() map[string]ProbeResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	copy := make(map[string]ProbeResult, len(p.results))
	for k, v := range p.results {
		copy[k] = v
	}
	return copy
}

// Delete removes the probe result for a process.
func (p *ProbeStore) Delete(process string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.results, process)
}
