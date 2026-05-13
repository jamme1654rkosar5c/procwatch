package monitor

import (
	"fmt"
	"sync"
	"time"
)

// AlertBudget tracks how many alerts a process may fire within a rolling window.
// Once the budget is exhausted alerts are suppressed until the window resets.
type AlertBudget struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	counts  map[string][]time.Time // process -> timestamps of alerts
	nowFunc func() time.Time
}

// NewAlertBudget creates a new AlertBudget with the given per-window limit and
// window duration.
func NewAlertBudget(limit int, window time.Duration) *AlertBudget {
	return &AlertBudget{
		limit:   limit,
		window:  window,
		counts:  make(map[string][]time.Time),
		nowFunc: time.Now,
	}
}

// Consume attempts to consume one unit from the budget for the given process.
// It returns true if the alert is allowed, false if the budget is exhausted.
func (b *AlertBudget) Consume(process string) bool {
	if process == "" {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.nowFunc()
	cutoff := now.Add(-b.window)

	ts := b.counts[process]
	filtered := ts[:0]
	for _, t := range ts {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= b.limit {
		b.counts[process] = filtered
		return false
	}

	b.counts[process] = append(filtered, now)
	return true
}

// Remaining returns how many alerts are left in the current window for process.
func (b *AlertBudget) Remaining(process string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.nowFunc()
	cutoff := now.Add(-b.window)

	used := 0
	for _, t := range b.counts[process] {
		if t.After(cutoff) {
			used++
		}
	}

	rem := b.limit - used
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the alert budget for the given process.
func (b *AlertBudget) Reset(process string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.counts, process)
	return nil
}
