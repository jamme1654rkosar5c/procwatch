package monitor

import (
	"fmt"
	"sync"
	"time"
)

// RunbookEntry associates a process with a URL pointing to its operational runbook.
type RunbookEntry struct {
	Process   string    `json:"process"`
	URL       string    `json:"url"`
	Note      string    `json:"note,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RunbookStore persists runbook URLs keyed by process name.
type RunbookStore struct {
	mu      sync.RWMutex
	entries map[string]RunbookEntry
}

// NewRunbookStore returns an initialised RunbookStore.
func NewRunbookStore() *RunbookStore {
	return &RunbookStore{entries: make(map[string]RunbookEntry)}
}

// Set creates or replaces the runbook entry for a process.
func (s *RunbookStore) Set(process, url, note string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if url == "" {
		return fmt.Errorf("runbook URL must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[process] = RunbookEntry{
		Process:   process,
		URL:       url,
		Note:      note,
		UpdatedAt: time.Now(),
	}
	return nil
}

// Get returns the runbook entry for a process and whether it was found.
func (s *RunbookStore) Get(process string) (RunbookEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[process]
	return e, ok
}

// Delete removes the runbook entry for a process.
func (s *RunbookStore) Delete(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, process)
}

// All returns a snapshot of all runbook entries.
func (s *RunbookStore) All() []RunbookEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]RunbookEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
