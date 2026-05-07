package monitor

import (
	"sync"
	"time"
)

// CheckpointStore records the last successful "seen alive" timestamp for each
// watched process. It is used to detect processes that have not been observed
// within a configurable window, complementing the heartbeat tracker with a
// persistent, poll-driven perspective.
type CheckpointStore struct {
	mu          sync.RWMutex
	checkpoints map[string]time.Time
	now         func() time.Time
}

// NewCheckpointStore returns an initialised CheckpointStore.
func NewCheckpointStore() *CheckpointStore {
	return &CheckpointStore{
		checkpoints: make(map[string]time.Time),
		now:         time.Now,
	}
}

// Touch records the current time as the last checkpoint for the named process.
func (c *CheckpointStore) Touch(process string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checkpoints[process] = c.now()
}

// LastSeen returns the last checkpoint time for the named process and whether
// an entry exists.
func (c *CheckpointStore) LastSeen(process string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.checkpoints[process]
	return t, ok
}

// IsStale returns true when the process has a checkpoint that is older than
// maxAge, or when no checkpoint exists at all.
func (c *CheckpointStore) IsStale(process string, maxAge time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.checkpoints[process]
	if !ok {
		return true
	}
	return c.now().Sub(t) > maxAge
}

// Delete removes the checkpoint entry for the named process.
func (c *CheckpointStore) Delete(process string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checkpoints, process)
}

// All returns a snapshot of all current checkpoints keyed by process name.
func (c *CheckpointStore) All() map[string]time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]time.Time, len(c.checkpoints))
	for k, v := range c.checkpoints {
		out[k] = v
	}
	return out
}
