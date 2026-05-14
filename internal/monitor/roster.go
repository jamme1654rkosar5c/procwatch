package monitor

import (
	"fmt"
	"sync"
)

// RosterEntry holds metadata about a watched process.
type RosterEntry struct {
	Name    string
	Enabled bool
	Owner   string
}

// RosterStore tracks which processes are actively enrolled for monitoring.
type RosterStore struct {
	mu      sync.RWMutex
	entries map[string]RosterEntry
}

// NewRosterStore returns an empty RosterStore.
func NewRosterStore() *RosterStore {
	return &RosterStore{
		entries: make(map[string]RosterEntry),
	}
}

// Enroll adds or updates a process in the roster.
func (r *RosterStore) Enroll(process, owner string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[process] = RosterEntry{
		Name:    process,
		Enabled: true,
		Owner:   owner,
	}
	return nil
}

// Disable marks a process as disabled without removing it.
func (r *RosterStore) Disable(process string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[process]
	if !ok {
		return fmt.Errorf("process %q not found in roster", process)
	}
	e.Enabled = false
	r.entries[process] = e
	return nil
}

// Remove deletes a process from the roster entirely.
func (r *RosterStore) Remove(process string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, process)
}

// Get returns the RosterEntry for a process and whether it exists.
func (r *RosterStore) Get(process string) (RosterEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[process]
	return e, ok
}

// IsEnabled reports whether a process is enrolled and enabled.
func (r *RosterStore) IsEnabled(process string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[process]
	return ok && e.Enabled
}

// All returns a snapshot of all roster entries.
func (r *RosterStore) All() []RosterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RosterEntry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}
