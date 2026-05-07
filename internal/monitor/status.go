package monitor

import (
	"sync"
	"time"
)

// ProcessStatus holds the last known status of a watched process.
type ProcessStatus struct {
	Name      string
	Up        bool
	PID       int
	LastSeen  time.Time
	LastEvent string
}

// StatusRegistry tracks the current status of all watched processes.
type StatusRegistry struct {
	mu       sync.RWMutex
	statuses map[string]*ProcessStatus
}

// NewStatusRegistry creates an empty StatusRegistry.
func NewStatusRegistry() *StatusRegistry {
	return &StatusRegistry{
		statuses: make(map[string]*ProcessStatus),
	}
}

// Update sets or replaces the status for a process.
func (r *StatusRegistry) Update(name string, up bool, pid int, event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statuses[name] = &ProcessStatus{
		Name:      name,
		Up:        up,
		PID:       pid,
		LastSeen:  time.Now(),
		LastEvent: event,
	}
}

// Get returns the status for a named process and whether it was found.
func (r *StatusRegistry) Get(name string) (ProcessStatus, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.statuses[name]
	if !ok {
		return ProcessStatus{}, false
	}
	return *s, true
}

// All returns a snapshot of all current statuses.
func (r *StatusRegistry) All() []ProcessStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ProcessStatus, 0, len(r.statuses))
	for _, s := range r.statuses {
		out = append(out, *s)
	}
	return out
}

// Delete removes the status entry for a named process.
// It is a no-op if the process is not currently tracked.
func (r *StatusRegistry) Delete(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.statuses, name)
}
