package monitor

import (
	"sync"
	"time"
)

// TrendDirection indicates whether a metric is rising, falling, or stable.
type TrendDirection string

const (
	TrendRising  TrendDirection = "rising"
	TrendFalling TrendDirection = "falling"
	TrendStable  TrendDirection = "stable"
)

// TrendSample is a single data point recorded for trend analysis.
type TrendSample struct {
	Value     float64
	RecordedAt time.Time
}

// TrendResult summarises the direction and magnitude of change.
type TrendResult struct {
	Direction TrendDirection
	Delta     float64 // latest minus earliest in the window
	Samples   int
}

// TrendAnalyzer tracks recent metric samples per process and reports trends.
type TrendAnalyzer struct {
	mu         sync.Mutex
	samples    map[string][]TrendSample
	window     time.Duration
	minSamples int
}

// NewTrendAnalyzer creates a TrendAnalyzer that retains samples within window
// and requires at least minSamples before reporting a non-stable trend.
func NewTrendAnalyzer(window time.Duration, minSamples int) *TrendAnalyzer {
	if minSamples < 2 {
		minSamples = 2
	}
	return &TrendAnalyzer{
		samples:    make(map[string][]TrendSample),
		window:     window,
		minSamples: minSamples,
	}
}

// Record adds a new sample for the named process, pruning stale entries.
func (t *TrendAnalyzer) Record(process string, value float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.window)

	existing := t.samples[process]
	pruned := existing[:0]
	for _, s := range existing {
		if s.RecordedAt.After(cutoff) {
			pruned = append(pruned, s)
		}
	}
	pruned = append(pruned, TrendSample{Value: value, RecordedAt: now})
	t.samples[process] = pruned
}

// Analyze returns the TrendResult for the named process.
// If fewer than minSamples exist the direction is always TrendStable.
func (t *TrendAnalyzer) Analyze(process string) TrendResult {
	t.mu.Lock()
	defer t.mu.Unlock()

	ss := t.samples[process]
	if len(ss) < t.minSamples {
		return TrendResult{Direction: TrendStable, Samples: len(ss)}
	}

	delta := ss[len(ss)-1].Value - ss[0].Value
	dir := TrendStable
	switch {
	case delta > 0:
		dir = TrendRising
	case delta < 0:
		dir = TrendFalling
	}
	return TrendResult{Direction: dir, Delta: delta, Samples: len(ss)}
}

// All returns a snapshot of the latest TrendResult for every tracked process.
func (t *TrendAnalyzer) All() map[string]TrendResult {
	t.mu.Lock()
	defer t.mu.Unlock()

	out := make(map[string]TrendResult, len(t.samples))
	for proc, ss := range t.samples {
		if len(ss) < t.minSamples {
			out[proc] = TrendResult{Direction: TrendStable, Samples: len(ss)}
			continue
		}
		delta := ss[len(ss)-1].Value - ss[0].Value
		dir := TrendStable
		switch {
		case delta > 0:
			dir = TrendRising
		case delta < 0:
			dir = TrendFalling
		}
		out[proc] = TrendResult{Direction: dir, Delta: delta, Samples: len(ss)}
	}
	return out
}
