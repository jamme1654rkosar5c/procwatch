package monitor

import (
	"math"
	"sync"
	"time"
)

// BackoffEntry tracks the current backoff state for a single process.
type BackoffEntry struct {
	Attempts int
	NextAllowed time.Time
}

// BackoffStore manages exponential backoff per process, preventing alert
// floods during sustained outages by increasing the delay between alerts.
type BackoffStore struct {
	mu      sync.Mutex
	entries map[string]*BackoffEntry
	base    time.Duration
	max     time.Duration
}

// NewBackoffStore creates a BackoffStore with the given base and maximum delay.
func NewBackoffStore(base, max time.Duration) *BackoffStore {
	return &BackoffStore{
		entries: make(map[string]*BackoffEntry),
		base:    base,
		max:     max,
	}
}

// Allow returns true if an alert for the given process is permitted now.
// If allowed, it advances the backoff state for the next call.
func (b *BackoffStore) Allow(process string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	e, ok := b.entries[process]
	if !ok {
		delay := b.base
		b.entries[process] = &BackoffEntry{
			Attempts:    1,
			NextAllowed: now.Add(delay),
		}
		return true
	}

	if now.Before(e.NextAllowed) {
		return false
	}

	e.Attempts++
	delay := time.Duration(float64(b.base) * math.Pow(2, float64(e.Attempts-1)))
	if delay > b.max {
		delay = b.max
	}
	e.NextAllowed = now.Add(delay)
	return true
}

// Reset clears the backoff state for a process (e.g. after recovery).
func (b *BackoffStore) Reset(process string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, process)
}

// State returns a copy of the current backoff entry for a process, and
// a boolean indicating whether an entry exists.
func (b *BackoffStore) State(process string) (BackoffEntry, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e, ok := b.entries[process]
	if !ok {
		return BackoffEntry{}, false
	}
	return *e, true
}
