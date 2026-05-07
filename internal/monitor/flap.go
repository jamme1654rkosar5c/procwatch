package monitor

import (
	"sync"
	"time"
)

// FlapDetector tracks rapid state transitions (up→down→up) for a process
// and flags it as "flapping" when transitions exceed a threshold within a window.
type FlapDetector struct {
	mu         sync.Mutex
	transitions map[string][]time.Time
	window     time.Duration
	threshold  int
	now        func() time.Time
}

// NewFlapDetector creates a FlapDetector that considers a process flapping
// when it changes state more than threshold times within window.
func NewFlapDetector(threshold int, window time.Duration) *FlapDetector {
	return &FlapDetector{
		transitions: make(map[string][]time.Time),
		window:     window,
		threshold:  threshold,
		now:        time.Now,
	}
}

// Record notes a state transition for the named process and returns true
// if the process is currently considered to be flapping.
func (f *FlapDetector) Record(processName string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	cutoff := f.now().Add(-f.window)
	times := f.transitions[processName]

	// Evict entries outside the window.
	active := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			active = append(active, t)
		}
	}
	active = append(active, f.now())
	f.transitions[processName] = active

	return len(active) > f.threshold
}

// IsFlapping returns whether the process is currently considered flapping
// without recording a new transition.
func (f *FlapDetector) IsFlapping(processName string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	cutoff := f.now().Add(-f.window)
	count := 0
	for _, t := range f.transitions[processName] {
		if t.After(cutoff) {
			count++
		}
	}
	return count > f.threshold
}

// Reset clears the transition history for a process.
func (f *FlapDetector) Reset(processName string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.transitions, processName)
}
