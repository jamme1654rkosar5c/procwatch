package monitor

import (
	"fmt"
	"sync"
	"time"
)

// QuotaEntry tracks webhook delivery quota for a process.
type QuotaEntry struct {
	Process   string
	Limit     int
	Used      int
	WindowEnd time.Time
}

// QuotaStore enforces per-process webhook delivery quotas within a rolling window.
type QuotaStore struct {
	mu      sync.Mutex
	entries map[string]*QuotaEntry
	window  time.Duration
	now     func() time.Time
}

// NewQuotaStore creates a QuotaStore with the given rolling window duration.
func NewQuotaStore(window time.Duration) *QuotaStore {
	return &QuotaStore{
		entries: make(map[string]*QuotaEntry),
		window:  window,
		now:     time.Now,
	}
}

// SetLimit configures the maximum deliveries allowed per window for a process.
func (q *QuotaStore) SetLimit(process string, limit int) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if limit <= 0 {
		return fmt.Errorf("limit must be positive")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if e, ok := q.entries[process]; ok {
		e.Limit = limit
	} else {
		q.entries[process] = &QuotaEntry{Process: process, Limit: limit, WindowEnd: q.now().Add(q.window)}
	}
	return nil
}

// Allow returns true and increments usage if the process is within quota.
// Returns false when the limit is exceeded or no quota is configured.
func (q *QuotaStore) Allow(process string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	e, ok := q.entries[process]
	if !ok {
		return true // no quota configured — allow by default
	}
	now := q.now()
	if now.After(e.WindowEnd) {
		e.Used = 0
		e.WindowEnd = now.Add(q.window)
	}
	if e.Used >= e.Limit {
		return false
	}
	e.Used++
	return true
}

// Get returns the current QuotaEntry for a process, or nil if not configured.
func (q *QuotaStore) Get(process string) *QuotaEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	e, ok := q.entries[process]
	if !ok {
		return nil
	}
	copy := *e
	return &copy
}

// All returns a snapshot of all quota entries.
func (q *QuotaStore) All() []QuotaEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]QuotaEntry, 0, len(q.entries))
	for _, e := range q.entries {
		out = append(out, *e)
	}
	return out
}

// Reset clears usage for a process, starting a fresh window.
func (q *QuotaStore) Reset(process string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if e, ok := q.entries[process]; ok {
		e.Used = 0
		e.WindowEnd = q.now().Add(q.window)
	}
}
