package monitor

import (
	"sync"
	"time"
)

// HeartbeatTracker records the last-seen time for each watched process
// and exposes whether a process has missed its expected check-in window.
type HeartbeatTracker struct {
	mu      sync.RWMutex
	hearts  map[string]time.Time
	maxAge  time.Duration
	nowFunc func() time.Time
}

// NewHeartbeatTracker creates a tracker that considers a process stale
// after maxAge without a recorded beat.
func NewHeartbeatTracker(maxAge time.Duration) *HeartbeatTracker {
	return &HeartbeatTracker{
		hearts:  make(map[string]time.Time),
		maxAge:  maxAge,
		nowFunc: time.Now,
	}
}

// Beat records that process name was observed alive right now.
func (h *HeartbeatTracker) Beat(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hearts[name] = h.nowFunc()
}

// IsStale returns true when the process has never beaten or its last beat
// is older than maxAge.
func (h *HeartbeatTracker) IsStale(name string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t, ok := h.hearts[name]
	if !ok {
		return true
	}
	return h.nowFunc().Sub(t) > h.maxAge
}

// LastSeen returns the last beat time for a process and whether it exists.
func (h *HeartbeatTracker) LastSeen(name string) (time.Time, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t, ok := h.hearts[name]
	return t, ok
}

// All returns a snapshot of all tracked process names and their last-seen times.
func (h *HeartbeatTracker) All() map[string]time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make(map[string]time.Time, len(h.hearts))
	for k, v := range h.hearts {
		out[k] = v
	}
	return out
}
