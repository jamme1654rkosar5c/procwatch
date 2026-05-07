package monitor

import (
	"testing"
	"time"
)

func newTestCheckpoint() *CheckpointStore {
	cs := NewCheckpointStore()
	return cs
}

func TestCheckpoint_LastSeen_Missing(t *testing.T) {
	cs := newTestCheckpoint()
	_, ok := cs.LastSeen("nginx")
	if ok {
		t.Fatal("expected no entry for untracked process")
	}
}

func TestCheckpoint_Touch_And_LastSeen(t *testing.T) {
	cs := newTestCheckpoint()
	before := time.Now()
	cs.Touch("nginx")
	after := time.Now()

	t, ok := cs.LastSeen("nginx")
	if !ok {
		t.Fatal("expected entry after Touch")
	}
	if t.Before(before) || t.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", t, before, after)
	}
}

func TestCheckpoint_IsStale_NotTracked(t *testing.T) {
	cs := newTestCheckpoint()
	if !cs.IsStale("missing", time.Minute) {
		t.Fatal("untracked process should be considered stale")
	}
}

func TestCheckpoint_IsStale_Fresh(t *testing.T) {
	cs := newTestCheckpoint()
	cs.Touch("redis")
	if cs.IsStale("redis", time.Minute) {
		t.Fatal("recently touched process should not be stale")
	}
}

func TestCheckpoint_IsStale_Expired(t *testing.T) {
	cs := newTestCheckpoint()
	fixed := time.Now().Add(-2 * time.Minute)
	cs.mu.Lock()
	cs.checkpoints["redis"] = fixed
	cs.mu.Unlock()

	if !cs.IsStale("redis", time.Minute) {
		t.Fatal("old checkpoint should be stale")
	}
}

func TestCheckpoint_Delete(t *testing.T) {
	cs := newTestCheckpoint()
	cs.Touch("postgres")
	cs.Delete("postgres")
	_, ok := cs.LastSeen("postgres")
	if ok {
		t.Fatal("expected entry to be removed after Delete")
	}
}

func TestCheckpoint_All_ReturnsCopy(t *testing.T) {
	cs := newTestCheckpoint()
	cs.Touch("svc-a")
	cs.Touch("svc-b")

	all := cs.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	// Mutating the returned map must not affect the store.
	delete(all, "svc-a")
	if _, ok := cs.LastSeen("svc-a"); !ok {
		t.Fatal("deleting from snapshot should not affect store")
	}
}
