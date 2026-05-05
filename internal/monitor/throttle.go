package monitor

import (
	"sync"
	"time"
)

// throttleKey uniquely identifies an alert type for a given process.
type throttleKey struct {
	process string
	eventType string
}

// Throttler prevents duplicate alerts from firing within a cooldown window.
type Throttler struct {
	mu       sync.Mutex
	lastSent map[throttleKey]time.Time
	cooldown time.Duration
}

// NewThrottler creates a Throttler with the given cooldown duration.
func NewThrottler(cooldown time.Duration) *Throttler {
	return &Throttler{
		lastSent: make(map[throttleKey]time.Time),
		cooldown: cooldown,
	}
}

// Allow returns true if the alert for the given process and event type
// should be sent, i.e. it has not been sent within the cooldown window.
// If allowed, it records the current time as the last sent time.
func (t *Throttler) Allow(process, eventType string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := throttleKey{process: process, eventType: eventType}
	last, exists := t.lastSent[key]
	if exists && time.Since(last) < t.cooldown {
		return false
	}
	t.lastSent[key] = time.Now()
	return true
}

// Reset clears the throttle record for the given process and event type,
// allowing the next alert to be sent immediately.
func (t *Throttler) Reset(process, eventType string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastSent, throttleKey{process: process, eventType: eventType})
}
