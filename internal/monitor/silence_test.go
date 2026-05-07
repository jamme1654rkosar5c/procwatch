package monitor

import (
	"testing"
	"time"
)

func TestSilencer_IsSilenced_NotTracked(t *testing.T) {
	s := NewSilencer()
	if s.IsSilenced("nginx") {
		t.Fatal("expected false for unknown process")
	}
}

func TestSilencer_IsSilenced_Active(t *testing.T) {
	s := NewSilencer()
	s.Silence("nginx", time.Now().Add(10*time.Minute))
	if !s.IsSilenced("nginx") {
		t.Fatal("expected process to be silenced")
	}
}

func TestSilencer_IsSilenced_Expired(t *testing.T) {
	s := NewSilencer()
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Silence("nginx", fixed.Add(-1*time.Second))
	if s.IsSilenced("nginx") {
		t.Fatal("expected expired silence to return false")
	}
}

func TestSilencer_Lift(t *testing.T) {
	s := NewSilencer()
	s.Silence("nginx", time.Now().Add(10*time.Minute))
	s.Lift("nginx")
	if s.IsSilenced("nginx") {
		t.Fatal("expected silence to be lifted")
	}
}

func TestSilencer_All_ReturnsActiveOnly(t *testing.T) {
	s := NewSilencer()
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Silence("nginx", fixed.Add(5*time.Minute))
	s.Silence("redis", fixed.Add(-1*time.Second)) // expired

	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 active rule, got %d", len(all))
	}
	if all[0].ProcessName != "nginx" {
		t.Errorf("unexpected process name %q", all[0].ProcessName)
	}
}

func TestSilencer_All_ReturnsCopy(t *testing.T) {
	s := NewSilencer()
	s.Silence("nginx", time.Now().Add(5*time.Minute))
	a := s.All()
	a[0].ProcessName = "mutated"
	b := s.All()
	if b[0].ProcessName == "mutated" {
		t.Fatal("All() should return a copy, not a reference")
	}
}
