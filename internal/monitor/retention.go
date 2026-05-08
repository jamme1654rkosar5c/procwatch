package monitor

import (
	"sync"
	"time"
)

// RetentionPolicy defines how long events/records are kept.
type RetentionPolicy struct {
	MaxAge time.Duration
}

// RetentionManager prunes stale entries from the History store
// according to a configured policy.
type RetentionManager struct {
	mu      sync.Mutex
	policy  RetentionPolicy
	history *History
	stopCh  chan struct{}
}

// NewRetentionManager creates a RetentionManager that will prune
// history entries older than policy.MaxAge.
func NewRetentionManager(h *History, policy RetentionPolicy) *RetentionManager {
	return &RetentionManager{
		policy:  policy,
		history: h,
		stopCh:  make(chan struct{}),
	}
}

// Start launches the background pruning loop with the given interval.
func (r *RetentionManager) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.Prune(time.Now())
			case <-r.stopCh:
				return
			}
		}
	}()
}

// Stop halts the background pruning loop.
func (r *RetentionManager) Stop() {
	close(r.stopCh)
}

// Prune removes all history entries whose timestamp is older than
// policy.MaxAge relative to now.
func (r *RetentionManager) Prune(now time.Time) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := now.Add(-r.policy.MaxAge)
	all := r.history.All()
	pruned := 0

	var keep []AlertEvent
	for _, e := range all {
		if e.Timestamp.After(cutoff) {
			keep = append(keep, e)
		} else {
			pruned++
		}
	}

	r.history.replace(keep)
	return pruned
}
