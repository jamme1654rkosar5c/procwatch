package monitor

import (
	"testing"
	"time"
)

func newTestScheduler(now time.Time) *MaintenanceScheduler {
	s := NewMaintenanceScheduler()
	s.now = func() time.Time { return now }
	return s
}

func TestMaintenance_NotTracked(t *testing.T) {
	s := newTestScheduler(time.Now())
	if s.IsUnderMaintenance("nginx") {
		t.Fatal("expected false for untracked process")
	}
}

func TestMaintenance_ActiveWindow(t *testing.T) {
	now := time.Now()
	s := newTestScheduler(now)
	s.Add(MaintenanceWindow{
		Process: "nginx",
		Start:   now.Add(-1 * time.Minute),
		End:     now.Add(1 * time.Minute),
		Reason:  "deploy",
	})
	if !s.IsUnderMaintenance("nginx") {
		t.Fatal("expected process to be under maintenance")
	}
}

func TestMaintenance_ExpiredWindow(t *testing.T) {
	now := time.Now()
	s := newTestScheduler(now)
	s.Add(MaintenanceWindow{
		Process: "nginx",
		Start:   now.Add(-5 * time.Minute),
		End:     now.Add(-1 * time.Minute),
	})
	if s.IsUnderMaintenance("nginx") {
		t.Fatal("expected false for expired window")
	}
}

func TestMaintenance_FutureWindow(t *testing.T) {
	now := time.Now()
	s := newTestScheduler(now)
	s.Add(MaintenanceWindow{
		Process: "nginx",
		Start:   now.Add(5 * time.Minute),
		End:     now.Add(10 * time.Minute),
	})
	if s.IsUnderMaintenance("nginx") {
		t.Fatal("expected false for future window")
	}
}

func TestMaintenance_Prune_RemovesExpired(t *testing.T) {
	now := time.Now()
	s := newTestScheduler(now)
	s.Add(MaintenanceWindow{
		Process: "nginx",
		Start:   now.Add(-10 * time.Minute),
		End:     now.Add(-1 * time.Minute),
	})
	s.Prune()
	if len(s.All()) != 0 {
		t.Fatalf("expected 0 windows after prune, got %d", len(s.All()))
	}
}

func TestMaintenance_All_ReturnsCopy(t *testing.T) {
	now := time.Now()
	s := newTestScheduler(now)
	s.Add(MaintenanceWindow{Process: "redis", Start: now, End: now.Add(time.Hour)})
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 window, got %d", len(all))
	}
	all[0].Process = "mutated"
	if s.All()[0].Process != "redis" {
		t.Fatal("All() should return a copy, not a reference")
	}
}
