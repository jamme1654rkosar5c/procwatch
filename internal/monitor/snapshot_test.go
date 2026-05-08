package monitor

import (
	"testing"
	"time"
)

func TestSnapshotStore_Record_And_Get(t *testing.T) {
	s := NewSnapshotStore()
	snap := Snapshot{Process: "nginx", CPUPct: 12.5, MemBytes: 1024, PID: 42, Up: true}
	s.Record(snap)

	got, ok := s.Get("nginx")
	if !ok {
		t.Fatal("expected snapshot to exist")
	}
	if got.CPUPct != 12.5 {
		t.Errorf("CPUPct: got %v, want 12.5", got.CPUPct)
	}
	if got.MemBytes != 1024 {
		t.Errorf("MemBytes: got %v, want 1024", got.MemBytes)
	}
	if got.PID != 42 {
		t.Errorf("PID: got %v, want 42", got.PID)
	}
}

func TestSnapshotStore_Get_Missing(t *testing.T) {
	s := NewSnapshotStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no snapshot for unknown process")
	}
}

func TestSnapshotStore_Record_SetsTimestamp(t *testing.T) {
	s := NewSnapshotStore()
	before := time.Now()
	s.Record(Snapshot{Process: "redis"})
	after := time.Now()

	got, _ := s.Get("redis")
	if got.RecordedAt.Before(before) || got.RecordedAt.After(after) {
		t.Errorf("RecordedAt %v not between %v and %v", got.RecordedAt, before, after)
	}
}

func TestSnapshotStore_Record_Overwrites(t *testing.T) {
	s := NewSnapshotStore()
	s.Record(Snapshot{Process: "app", CPUPct: 5.0})
	s.Record(Snapshot{Process: "app", CPUPct: 99.9})

	got, _ := s.Get("app")
	if got.CPUPct != 99.9 {
		t.Errorf("expected overwritten CPUPct 99.9, got %v", got.CPUPct)
	}
}

func TestSnapshotStore_All_ReturnsCopy(t *testing.T) {
	s := NewSnapshotStore()
	s.Record(Snapshot{Process: "a", Up: true})
	s.Record(Snapshot{Process: "b", Up: false})

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(all))
	}
	// Mutating the copy must not affect the store.
	delete(all, "a")
	if _, ok := s.Get("a"); !ok {
		t.Error("deleting from All() copy should not affect the store")
	}
}

func TestSnapshotStore_Delete(t *testing.T) {
	s := NewSnapshotStore()
	s.Record(Snapshot{Process: "svc", Up: true})
	s.Delete("svc")

	if _, ok := s.Get("svc"); ok {
		t.Error("expected snapshot to be deleted")
	}
}
