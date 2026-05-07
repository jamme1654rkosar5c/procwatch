package monitor

import (
	"testing"
	"time"
)

func newTestTracker(threshold int, window time.Duration, url string) *EscalationTracker {
	t := NewEscalationTracker(EscalationPolicy{
		Threshold: threshold,
		Window:    window,
		URL:       url,
	})
	return t
}

func TestEscalationTracker_BelowThreshold(t *testing.T) {
	tracker := newTestTracker(3, time.Minute, "https://escalation.example.com")

	if tracker.Record("nginx", "down") {
		t.Error("expected no escalation on first alert")
	}
	if tracker.Record("nginx", "down") {
		t.Error("expected no escalation on second alert")
	}
}

func TestEscalationTracker_AtThreshold(t *testing.T) {
	tracker := newTestTracker(3, time.Minute, "https://escalation.example.com")

	tracker.Record("nginx", "down")
	tracker.Record("nginx", "down")
	escalated := tracker.Record("nginx", "down")

	if !escalated {
		t.Error("expected escalation at threshold")
	}
}

func TestEscalationTracker_WindowExpiry(t *testing.T) {
	tracker := newTestTracker(2, 100*time.Millisecond, "https://escalation.example.com")

	// Fake clock starting in the past.
	now := time.Now()
	call := 0
	tracker.clock = func() time.Time {
		call++
		if call <= 1 {
			return now.Add(-200 * time.Millisecond) // outside window
		}
		return now
	}

	tracker.Record("nginx", "down") // recorded at -200ms, outside window
	escalated := tracker.Record("nginx", "down") // only 1 within window

	if escalated {
		t.Error("expected no escalation after window expiry pruned old entry")
	}
}

func TestEscalationTracker_DifferentProcesses_Independent(t *testing.T) {
	tracker := newTestTracker(2, time.Minute, "https://escalation.example.com")

	tracker.Record("nginx", "down")
	if tracker.Record("redis", "down") {
		t.Error("redis should not escalate based on nginx alerts")
	}
}

func TestEscalationTracker_Reset(t *testing.T) {
	tracker := newTestTracker(2, time.Minute, "https://escalation.example.com")

	tracker.Record("nginx", "down")
	tracker.Record("nginx", "down")
	tracker.Reset("nginx", "down")

	if tracker.Count("nginx", "down") != 0 {
		t.Error("expected count to be 0 after reset")
	}
}

func TestEscalationTracker_EscalationURL(t *testing.T) {
	expected := "https://pagerduty.example.com/webhook"
	tracker := newTestTracker(1, time.Minute, expected)

	if tracker.EscalationURL() != expected {
		t.Errorf("expected URL %q, got %q", expected, tracker.EscalationURL())
	}
}

func TestEscalationTracker_Count(t *testing.T) {
	tracker := newTestTracker(5, time.Minute, "")

	tracker.Record("nginx", "cpu")
	tracker.Record("nginx", "cpu")
	tracker.Record("nginx", "cpu")

	if got := tracker.Count("nginx", "cpu"); got != 3 {
		t.Errorf("expected count 3, got %d", got)
	}
}
