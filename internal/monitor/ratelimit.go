package monitor

import (
	"sync"
	"time"
)

// RateLimiter tracks per-process alert counts within a rolling time window.
// It is used to suppress excessive alerting when a process flaps repeatedly.
type RateLimiter struct {
	mu       sync.Mutex
	window   time.Duration
	maxCount int
	buckets  map[string][]time.Time
}

// NewRateLimiter creates a RateLimiter that allows at most maxCount alerts
// per process key within the given window duration.
func NewRateLimiter(window time.Duration, maxCount int) *RateLimiter {
	return &RateLimiter{
		window:   window,
		maxCount: maxCount,
		buckets:  make(map[string][]time.Time),
	}
}

// Allow returns true if the event identified by key is within the allowed
// rate, and records the occurrence. Returns false when the limit is exceeded.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	times := r.buckets[key]
	// evict timestamps outside the window
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= r.maxCount {
		r.buckets[key] = valid
		return false
	}

	r.buckets[key] = append(valid, now)
	return true
}

// Reset clears the rate-limit state for a specific key.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, key)
}

// Count returns the number of recorded events for key within the current window.
func (r *RateLimiter) Count(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-r.window)
	count := 0
	for _, t := range r.buckets[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}
