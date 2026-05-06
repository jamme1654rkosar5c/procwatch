package monitor

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// Deduplicator suppresses identical alert payloads within a configurable window.
// Two payloads are considered identical when they share the same process name,
// event type, and — for threshold events — the same threshold field.
type Deduplicator struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	window  time.Duration
	nowFunc func() time.Time
}

// NewDeduplicator returns a Deduplicator that suppresses duplicate events
// within window duration.
func NewDeduplicator(window time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:    make(map[string]time.Time),
		window:  window,
		nowFunc: time.Now,
	}
}

// IsDuplicate returns true when an identical event was already seen within
// the deduplication window. If not a duplicate, the event is recorded.
func (d *Deduplicator) IsDuplicate(process, eventType, field string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := d.key(process, eventType, field)
	now := d.nowFunc()

	if t, ok := d.seen[key]; ok && now.Sub(t) < d.window {
		return true
	}

	d.seen[key] = now
	return false
}

// Flush removes all entries whose window has expired. Useful for long-running
// processes to reclaim memory.
func (d *Deduplicator) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.nowFunc()
	for k, t := range d.seen {
		if now.Sub(t) >= d.window {
			delete(d.seen, k)
		}
	}
}

// Len returns the number of tracked entries (including expired ones not yet
// flushed).
func (d *Deduplicator) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}

func (d *Deduplicator) key(process, eventType, field string) string {
	h := sha256.Sum256([]byte(process + "\x00" + eventType + "\x00" + field))
	return fmt.Sprintf("%x", h[:])
}
