package monitor

import (
	"testing"
	"time"
)

func newTestSuppression() *SuppressionStore {
	s := NewSuppressionStore()
	return s
}

func TestSuppression_IsSuppressed_NotTracked(t *testing.T) {
	s := newTestSuppression()
	if s.IsSuppressed("nginx", "down") {
		t.Fatal("expected not suppressed for unknown process")
	}
}

func TestSuppression_Add_IsSuppressed(t *testing.T) {
	s := newTestSuppression()
	s.Add("nginx", "down", "planned maintenance", time.Minute)
	if !s.IsSuppressed("nginx", "down") {
		t.Fatal("expected suppressed after Add")
	}
}

func TestSuppression_Expired_NotSuppressed(t *testing.T) {
	s := newTestSuppression()
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Add("nginx", "down", "test", time.Second)
	// advance clock past expiry
	s.now = func() time.Time { return fixed.Add(2 * time.Second) }
	if s.IsSuppressed("nginx", "down") {
		t.Fatal("expected not suppressed after expiry")
	}
}

func TestSuppression_Remove(t *testing.T) {
	s := newTestSuppression()
	s.Add("nginx", "cpu", "reason", time.Minute)
	s.Remove("nginx", "cpu")
	if s.IsSuppressed("nginx", "cpu") {
		t.Fatal("expected not suppressed after Remove")
	}
}

func TestSuppression_DifferentEventTypes_Independent(t *testing.T) {
	s := newTestSuppression()
	s.Add("nginx", "down", "r", time.Minute)
	if s.IsSuppressed("nginx", "cpu") {
		t.Fatal("cpu should not be suppressed when only down is suppressed")
	}
	if !s.IsSuppressed("nginx", "down") {
		t.Fatal("down should be suppressed")
	}
}

func TestSuppression_All_ReturnsActiveOnly(t *testing.T) {
	s := newTestSuppression()
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Add("nginx", "down", "r1", time.Minute)
	s.Add("redis", "cpu", "r2", time.Millisecond)
	// expire redis/cpu
	s.now = func() time.Time { return fixed.Add(time.Second) }
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 active rule, got %d", len(all))
	}
	if all[0].Process != "nginx" || all[0].EventType != "down" {
		t.Fatalf("unexpected rule: %+v", all[0])
	}
}

func TestSuppression_All_Empty(t *testing.T) {
	s := newTestSuppression()
	if got := s.All(); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}
