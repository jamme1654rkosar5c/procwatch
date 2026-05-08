package monitor

import (
	"testing"
	"time"
)

func newTestRetentionHistory(events []AlertEvent) *History {
	h := NewHistory(100)
	for _, e := range events {
		h.Record(e)
	}
	return h
}

func TestRetentionManager_Prune_RemovesOldEntries(t *testing.T) {
	now := time.Now()
	h := newTestRetentionHistory([]AlertEvent{
		{Process: "a", Kind: "down", Timestamp: now.Add(-2 * time.Hour)},
		{Process: "b", Kind: "down", Timestamp: now.Add(-30 * time.Minute)},
		{Process: "c", Kind: "down", Timestamp: now.Add(-5 * time.Minute)},
	})

	rm := NewRetentionManager(h, RetentionPolicy{MaxAge: time.Hour})
	pruned := rm.Prune(now)

	if pruned != 1 {
		t.Fatalf("expected 1 pruned, got %d", pruned)
	}
	if h.Len() != 2 {
		t.Fatalf("expected 2 remaining, got %d", h.Len())
	}
}

func TestRetentionManager_Prune_NothingToRemove(t *testing.T) {
	now := time.Now()
	h := newTestRetentionHistory([]AlertEvent{
		{Process: "a", Kind: "down", Timestamp: now.Add(-10 * time.Minute)},
	})

	rm := NewRetentionManager(h, RetentionPolicy{MaxAge: time.Hour})
	pruned := rm.Prune(now)

	if pruned != 0 {
		t.Fatalf("expected 0 pruned, got %d", pruned)
	}
	if h.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", h.Len())
	}
}

func TestRetentionManager_Prune_AllExpired(t *testing.T) {
	now := time.Now()
	h := newTestRetentionHistory([]AlertEvent{
		{Process: "a", Kind: "down", Timestamp: now.Add(-3 * time.Hour)},
		{Process: "b", Kind: "down", Timestamp: now.Add(-2 * time.Hour)},
	})

	rm := NewRetentionManager(h, RetentionPolicy{MaxAge: time.Hour})
	pruned := rm.Prune(now)

	if pruned != 2 {
		t.Fatalf("expected 2 pruned, got %d", pruned)
	}
	if h.Len() != 0 {
		t.Fatalf("expected 0 remaining, got %d", h.Len())
	}
}

func TestRetentionManager_Prune_EmptyHistory(t *testing.T) {
	h := NewHistory(100)
	rm := NewRetentionManager(h, RetentionPolicy{MaxAge: time.Hour})
	pruned := rm.Prune(time.Now())
	if pruned != 0 {
		t.Fatalf("expected 0 pruned, got %d", pruned)
	}
}
