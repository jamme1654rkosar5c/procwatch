package monitor

import (
	"testing"
	"time"
)

func newTestHeartbeat(maxAge time.Duration) *HeartbeatTracker {
	ht := NewHeartbeatTracker(maxAge)
	return ht
}

func TestHeartbeat_IsStale_NotTracked(t *testing.T) {
	ht := newTestHeartbeat(5 * time.Second)
	if !ht.IsStale("nginx") {
		t.Fatal("expected untracked process to be stale")
	}
}

func TestHeartbeat_Beat_NotStale(t *testing.T) {
	ht := newTestHeartbeat(5 * time.Second)
	ht.Beat("nginx")
	if ht.IsStale("nginx") {
		t.Fatal("expected fresh beat to not be stale")
	}
}

func TestHeartbeat_IsStale_AfterMaxAge(t *testing.T) {
	now := time.Now()
	ht := newTestHeartbeat(5 * time.Second)
	// Inject a past beat via nowFunc trick
	ht.hearts["nginx"] = now.Add(-10 * time.Second)
	if !ht.IsStale("nginx") {
		t.Fatal("expected stale after max age exceeded")
	}
}

func TestHeartbeat_LastSeen_Missing(t *testing.T) {
	ht := newTestHeartbeat(5 * time.Second)
	_, ok := ht.LastSeen("redis")
	if ok {
		t.Fatal("expected ok=false for missing process")
	}
}

func TestHeartbeat_LastSeen_Present(t *testing.T) {
	ht := newTestHeartbeat(5 * time.Second)
	before := time.Now()
	ht.Beat("redis")
	after := time.Now()

	ts, ok := ht.LastSeen("redis")
	if !ok {
		t.Fatal("expected ok=true after beat")
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

func TestHeartbeat_All_ReturnsCopy(t *testing.T) {
	ht := newTestHeartbeat(5 * time.Second)
	ht.Beat("nginx")
	ht.Beat("redis")

	all := ht.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating returned map must not affect tracker
	delete(all, "nginx")
	if _, ok := ht.LastSeen("nginx"); !ok {
		t.Fatal("deleting from copy should not affect tracker")
	}
}

func TestHeartbeat_MultipleBeats_UpdatesTime(t *testing.T) {
	ht := newTestHeartbeat(5 * time.Second)
	old := time.Now().Add(-3 * time.Second)
	ht.hearts["nginx"] = old

	ht.Beat("nginx")
	ts, _ := ht.LastSeen("nginx")
	if !ts.After(old) {
		t.Error("expected updated timestamp to be more recent")
	}
}
