package monitor

import (
	"sync"
	"time"
)

// EventRecord stores a single alert event for a process.
type EventRecord struct {
	ProcessName string
	EventType   string
	OccurredAt  time.Time
	Details     string
}

// History maintains a bounded in-memory log of recent alert events.
type History struct {
	mu      sync.Mutex
	records []EventRecord
	maxSize int
}

// NewHistory creates a History that retains at most maxSize records.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &History{maxSize: maxSize}
}

// Record appends an event to the history, evicting the oldest entry when full.
func (h *History) Record(processName, eventType, details string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry := EventRecord{
		ProcessName: processName,
		EventType:   eventType,
		OccurredAt:  time.Now(),
		Details:     details,
	}

	if len(h.records) >= h.maxSize {
		h.records = h.records[1:]
	}
	h.records = append(h.records, entry)
}

// All returns a snapshot of all stored records, oldest first.
func (h *History) All() []EventRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	snap := make([]EventRecord, len(h.records))
	copy(snap, h.records)
	return snap
}

// ForProcess returns records that belong to the named process.
func (h *History) ForProcess(name string) []EventRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	var out []EventRecord
	for _, r := range h.records {
		if r.ProcessName == name {
			out = append(out, r)
		}
	}
	return out
}

// Len returns the current number of stored records.
func (h *History) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.records)
}
