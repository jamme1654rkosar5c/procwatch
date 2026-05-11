package monitor

import (
	"fmt"
	"sync"
	"time"
)

// OnCallEntry represents a single on-call rotation entry.
type OnCallEntry struct {
	Process   string    `json:"process"`
	Owner     string    `json:"owner"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	StartsAt  time.Time `json:"starts_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OnCallStore tracks on-call ownership per process.
type OnCallStore struct {
	mu      sync.RWMutex
	entries map[string]OnCallEntry
}

// NewOnCallStore returns an initialised OnCallStore.
func NewOnCallStore() *OnCallStore {
	return &OnCallStore{entries: make(map[string]OnCallEntry)}
}

// Set registers or replaces the on-call entry for a process.
func (s *OnCallStore) Set(e OnCallEntry) error {
	if e.Process == "" {
		return fmt.Errorf("oncall: process name required")
	}
	if e.Owner == "" {
		return fmt.Errorf("oncall: owner required")
	}
	if e.ExpiresAt.IsZero() || !e.ExpiresAt.After(e.StartsAt()) {
		return fmt.Errorf("oncall: expires_at must be after starts_at")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[e.Process] = e
	return nil
}

// Get returns the current on-call entry for a process, if any.
func (s *OnCallStore) Get(process string) (OnCallEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[process]
	return e, ok
}

// Active returns the entry only when the current time falls within the window.
func (s *OnCallStore) Active(process string, now time.Time) (OnCallEntry, bool) {
	e, ok := s.Get(process)
	if !ok {
		return OnCallEntry{}, false
	}
	if now.Before(e.StartsAt()) || now.After(e.ExpiresAt) {
		return OnCallEntry{}, false
	}
	return e, true
}

// Delete removes the on-call entry for a process.
func (s *OnCallStore) Delete(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, process)
}

// All returns a snapshot of all entries.
func (s *OnCallStore) All() []OnCallEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OnCallEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// StartsAt returns StartsAt, defaulting to zero time when unset.
func (e OnCallEntry) StartsAt() time.Time { return e.StartsAt }
