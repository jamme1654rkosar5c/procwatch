package monitor

import (
	"fmt"
	"sync"
)

// TagRegistry stores arbitrary key-value tags associated with a process name.
// Tags are useful for grouping, filtering, or enriching alert payloads.
type TagRegistry struct {
	mu   sync.RWMutex
	tags map[string]map[string]string // process -> key -> value
}

// NewTagRegistry creates an empty TagRegistry.
func NewTagRegistry() *TagRegistry {
	return &TagRegistry{
		tags: make(map[string]map[string]string),
	}
}

// Set associates a key-value tag with the given process.
func (r *TagRegistry) Set(process, key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tags[process]; !ok {
		r.tags[process] = make(map[string]string)
	}
	r.tags[process][key] = value
}

// Get returns the value for a tag key on a process, and whether it was found.
func (r *TagRegistry) Get(process, key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.tags[process]; ok {
		v, found := m[key]
		return v, found
	}
	return "", false
}

// All returns a copy of all tags for the given process.
// Returns nil if the process has no tags.
func (r *TagRegistry) All(process string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.tags[process]
	if !ok {
		return nil
	}
	copy := make(map[string]string, len(m))
	for k, v := range m {
		copy[k] = v
	}
	return copy
}

// Delete removes a single tag key from a process.
func (r *TagRegistry) Delete(process, key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.tags[process]; ok {
		delete(m, key)
		if len(m) == 0 {
			delete(r.tags, process)
		}
	}
}

// Processes returns a list of all process names that have at least one tag.
func (r *TagRegistry) Processes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.tags))
	for p := range r.tags {
		out = append(out, p)
	}
	return out
}

// String returns a human-readable representation of all tags for a process.
func (r *TagRegistry) String(process string) string {
	tags := r.All(process)
	if len(tags) == 0 {
		return fmt.Sprintf("%s: (no tags)", process)
	}
	return fmt.Sprintf("%s: %v", process, tags)
}
