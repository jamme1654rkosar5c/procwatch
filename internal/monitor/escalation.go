package monitor

import (
	"sync"
	"time"
)

// EscalationPolicy defines how many times an alert must fire before
// escalating to a secondary webhook URL.
type EscalationPolicy struct {
	Threshold int           // number of alerts before escalation
	Window    time.Duration // rolling window for counting alerts
	URL       string        // secondary webhook URL to escalate to
}

// EscalationTracker tracks per-process alert counts and determines
// whether an alert should be escalated.
type EscalationTracker struct {
	mu      sync.Mutex
	policy  EscalationPolicy
	counts  map[string][]time.Time // process name -> alert timestamps
	clock   func() time.Time
}

// NewEscalationTracker creates a new EscalationTracker with the given policy.
func NewEscalationTracker(policy EscalationPolicy) *EscalationTracker {
	return &EscalationTracker{
		policy: policy,
		counts: make(map[string][]time.Time),
		clock:  time.Now,
	}
}

// Record registers an alert occurrence for the given process and event type key.
// Returns true if the alert count has reached or exceeded the escalation threshold
// within the configured window.
func (e *EscalationTracker) Record(processName, eventType string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := processName + ":" + eventType
	now := e.clock()
	cutoff := now.Add(-e.policy.Window)

	// Prune old entries outside the window.
	filtered := e.counts[key][:0]
	for _, t := range e.counts[key] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	e.counts[key] = filtered

	return len(filtered) >= e.policy.Threshold
}

// EscalationURL returns the secondary webhook URL defined in the policy.
func (e *EscalationTracker) EscalationURL() string {
	return e.policy.URL
}

// Reset clears the alert history for the given process and event type.
func (e *EscalationTracker) Reset(processName, eventType string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.counts, processName+":"+eventType)
}

// Count returns the number of alerts recorded within the window for the given key.
func (e *EscalationTracker) Count(processName, eventType string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := processName + ":" + eventType
	now := e.clock()
	cutoff := now.Add(-e.policy.Window)

	count := 0
	for _, t := range e.counts[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}
