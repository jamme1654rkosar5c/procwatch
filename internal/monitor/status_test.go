package monitor

import (
	"testing"
	"time"
)

func TestStatusRegistry_UpdateAndGet(t *testing.T) {
	r := NewStatusRegistry()
	before := time.Now()
	r.Update("nginx", true, 1234, "running")

	s, ok := r.Get("nginx")
	if !ok {
		t.Fatal("expected to find nginx status")
	}
	if s.Name != "nginx" {
		t.Errorf("expected name nginx, got %s", s.Name)
	}
	if !s.Up {
		t.Error("expected Up=true")
	}
	if s.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", s.PID)
	}
	if s.LastEvent != "running" {
		t.Errorf("expected event 'running', got %s", s.LastEvent)
	}
	if s.LastSeen.Before(before) {
		t.Error("LastSeen should be >= time before update")
	}
}

func TestStatusRegistry_Get_Missing(t *testing.T) {
	r := NewStatusRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected not found for unknown process")
	}
}

func TestStatusRegistry_Update_Overwrites(t *testing.T) {
	r := NewStatusRegistry()
	r.Update("redis", true, 100, "running")
	r.Update("redis", false, 0, "down")

	s, ok := r.Get("redis")
	if !ok {
		t.Fatal("expected redis status to exist")
	}
	if s.Up {
		t.Error("expected Up=false after overwrite")
	}
	if s.LastEvent != "down" {
		t.Errorf("expected event 'down', got %s", s.LastEvent)
	}
}

func TestStatusRegistry_All(t *testing.T) {
	r := NewStatusRegistry()
	r.Update("nginx", true, 10, "running")
	r.Update("redis", false, 0, "down")

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(all))
	}
}

func TestStatusRegistry_All_ReturnsCopy(t *testing.T) {
	r := NewStatusRegistry()
	r.Update("nginx", true, 10, "running")

	all := r.All()
	all[0].Name = "mutated"

	s, _ := r.Get("nginx")
	if s.Name != "nginx" {
		t.Error("All() should return a copy, not a reference")
	}
}
