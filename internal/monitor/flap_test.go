package monitor

import (
	"testing"
	"time"
)

func newTestFlapDetector(threshold int, window time.Duration) *FlapDetector {
	fd := NewFlapDetector(threshold, window)
	return fd
}

func TestFlapDetector_BelowThreshold(t *testing.T) {
	fd := newTestFlapDetector(3, 10*time.Second)
	for i := 0; i < 3; i++ {
		if fd.Record("nginx") {
			t.Fatalf("expected not flapping on transition %d", i+1)
		}
	}
}

func TestFlapDetector_AtThreshold(t *testing.T) {
	fd := newTestFlapDetector(3, 10*time.Second)
	for i := 0; i < 4; i++ {
		flapping := fd.Record("nginx")
		if i < 3 && flapping {
			t.Fatalf("expected not flapping at transition %d", i+1)
		}
		if i == 3 && !flapping {
			t.Fatal("expected flapping after exceeding threshold")
		}
	}
}

func TestFlapDetector_WindowExpiry(t *testing.T) {
	now := time.Now()
	fd := newTestFlapDetector(2, 5*time.Second)
	// Inject old transitions outside the window.
	fd.transitions["redis"] = []time.Time{
		now.Add(-10 * time.Second),
		now.Add(-8 * time.Second),
		now.Add(-6 * time.Second),
	}
	// One new transition should not push it over threshold.
	if fd.Record("redis") {
		t.Fatal("expected old transitions to be evicted")
	}
}

func TestFlapDetector_DifferentProcesses_Independent(t *testing.T) {
	fd := newTestFlapDetector(2, 10*time.Second)
	fd.Record("nginx")
	fd.Record("nginx")
	fd.Record("nginx")

	// redis has no transitions, should not be flapping.
	if fd.IsFlapping("redis") {
		t.Fatal("redis should not be flapping")
	}
}

func TestFlapDetector_Reset_ClearsState(t *testing.T) {
	fd := newTestFlapDetector(2, 10*time.Second)
	fd.Record("nginx")
	fd.Record("nginx")
	fd.Record("nginx")
	if !fd.IsFlapping("nginx") {
		t.Fatal("expected flapping before reset")
	}
	fd.Reset("nginx")
	if fd.IsFlapping("nginx") {
		t.Fatal("expected not flapping after reset")
	}
}

func TestFlapDetector_IsFlapping_NoRecord(t *testing.T) {
	fd := newTestFlapDetector(1, 10*time.Second)
	if fd.IsFlapping("unknown") {
		t.Fatal("unknown process should not be flapping")
	}
}
