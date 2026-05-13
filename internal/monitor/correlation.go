package monitor

import (
	"sync"
	"time"
)

// CorrelationEntry records a co-occurrence of events across processes.
type CorrelationEntry struct {
	ProcessA  string
	ProcessB  string
	EventType string
	Count     int
	LastSeen  time.Time
}

// CorrelationTracker detects when multiple processes experience the same
// event type within a short window, suggesting a common cause.
type CorrelationTracker struct {
	mu      sync.Mutex
	window  time.Duration
	events  map[string][]time.Time // key: "process:eventType"
	correls map[string]*CorrelationEntry // key: "procA|procB:eventType"
}

// NewCorrelationTracker creates a tracker with the given co-occurrence window.
func NewCorrelationTracker(window time.Duration) *CorrelationTracker {
	return &CorrelationTracker{
		window:  window,
		events:  make(map[string][]time.Time),
		correls: make(map[string]*CorrelationEntry),
	}
}

// Record notes an event for the given process and returns any processes that
// experienced the same event type within the correlation window.
func (c *CorrelationTracker) Record(process, eventType string, now time.Time) []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := now.Add(-c.window)

	// Prune and record for this process.
	key := process + ":" + eventType
	c.events[key] = pruneOld(c.events[key], cutoff)
	c.events[key] = append(c.events[key], now)

	// Find correlated peers.
	var peers []string
	for k, times := range c.events {
		if k == key {
			continue
		}
		// Check same event type suffix.
		if !hasSuffix(k, ":"+eventType) {
			continue
		}
		times = pruneOld(times, cutoff)
		c.events[k] = times
		if len(times) == 0 {
			continue
		}
		peer := k[:len(k)-len(":"+eventType)]
		peers = append(peers, peer)
		c.updateCorrel(process, peer, eventType, now)
	}
	return peers
}

func (c *CorrelationTracker) updateCorrel(a, b, eventType string, now time.Time) {
	key := corrKey(a, b, eventType)
	e, ok := c.correls[key]
	if !ok {
		e = &CorrelationEntry{ProcessA: a, ProcessB: b, EventType: eventType}
		c.correls[key] = e
	}
	e.Count++
	e.LastSeen = now
}

// All returns a snapshot of all correlation entries.
func (c *CorrelationTracker) All() []CorrelationEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]CorrelationEntry, 0, len(c.correls))
	for _, e := range c.correls {
		out = append(out, *e)
	}
	return out
}

func pruneOld(times []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for _, t := range times {
		if !t.Before(cutoff) {
			times[i] = t
			i++
		}
	}
	return times[:i]
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func corrKey(a, b, eventType string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b + ":" + eventType
}
