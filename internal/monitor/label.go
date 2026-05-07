package monitor

import (
	"fmt"
	"sync"
)

// LabelStore holds arbitrary key-value labels attached to watched processes.
// Labels are user-defined metadata (e.g. env=production, team=platform) that
// are surfaced through the API but do not affect alerting logic.
type LabelStore struct {
	mu     sync.RWMutex
	labels map[string]map[string]string // process name → label key → value
}

// NewLabelStore returns an initialised, empty LabelStore.
func NewLabelStore() *LabelStore {
	return &LabelStore{
		labels: make(map[string]map[string]string),
	}
}

// Set adds or updates a label on the named process.
func (s *LabelStore) Set(process, key, value string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if key == "" {
		return fmt.Errorf("label key must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.labels[process]; !ok {
		s.labels[process] = make(map[string]string)
	}
	s.labels[process][key] = value
	return nil
}

// Get returns the value for a single label key on a process.
// The second return value is false when the process or key is unknown.
func (s *LabelStore) Get(process, key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.labels[process]; ok {
		v, found := m[key]
		return v, found
	}
	return "", false
}

// Delete removes a single label key from a process. It is a no-op if the
// process or key does not exist.
func (s *LabelStore) Delete(process, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.labels[process]; ok {
		delete(m, key)
	}
}

// All returns a copy of all labels for the named process, or nil if none exist.
func (s *LabelStore) All(process string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.labels[process]
	if !ok {
		return nil
	}
	copy := make(map[string]string, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// Processes returns the names of all processes that have at least one label.
func (s *LabelStore) Processes() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.labels))
	for p := range s.labels {
		out = append(out, p)
	}
	return out
}
