package monitor

import (
	"sync"
	"time"
)

// CooldownEntry holds the expiry time for a cooldown period.
type CooldownEntry struct {
	ExpiresAt time.Time
}

// CooldownStore tracks per-process cooldown periods that suppress alerts
// after a recovery, preventing alert storms during unstable restarts.
type CooldownStore struct {
	mu      sync.Mutex
	entries map[string]CooldownEntry
	now     func() time.Time
}

// NewCooldownStore creates a new CooldownStore.
func NewCooldownStore() *CooldownStore {
	return &CooldownStore{
		entries: make(map[string]CooldownEntry),
		now:     time.Now,
	}
}

// Set registers a cooldown for the given process lasting the specified duration.
func (c *CooldownStore) Set(process string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[process] = CooldownEntry{
		ExpiresAt: c.now().Add(duration),
	}
}

// Active returns true if the given process is currently in a cooldown period.
func (c *CooldownStore) Active(process string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[process]
	if !ok {
		return false
	}
	if c.now().After(entry.ExpiresAt) {
		delete(c.entries, process)
		return false
	}
	return true
}

// Lift removes the cooldown for a process immediately.
func (c *CooldownStore) Lift(process string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, process)
}

// All returns a snapshot of all active cooldown entries keyed by process name.
func (c *CooldownStore) All() map[string]CooldownEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]CooldownEntry, len(c.entries))
	now := c.now()
	for k, v := range c.entries {
		if now.Before(v.ExpiresAt) {
			out[k] = v
		}
	}
	return out
}
