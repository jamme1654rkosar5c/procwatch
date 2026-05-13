package monitor

import (
	"testing"
	"time"
)

func newTestBackoff() *BackoffStore {
	return NewBackoffStore(100*time.Millisecond, 800*time.Millisecond)
}

func TestBackoff_FirstCall_Allowed(t *testing.T) {
	b := newTestBackoff()
	if !b.Allow("svc") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestBackoff_SecondCall_WithinDelay_Blocked(t *testing.T) {
	b := newTestBackoff()
	b.Allow("svc")
	if b.Allow("svc") {
		t.Fatal("expected second call within delay to be blocked")
	}
}

func TestBackoff_AfterDelay_Allowed(t *testing.T) {
	b := NewBackoffStore(50*time.Millisecond, 800*time.Millisecond)
	b.Allow("svc")
	time.Sleep(60 * time.Millisecond)
	if !b.Allow("svc") {
		t.Fatal("expected call after delay to be allowed")
	}
}

func TestBackoff_ExponentialIncrease(t *testing.T) {
	b := NewBackoffStore(50*time.Millisecond, 10*time.Second)

	// First allow seeds attempt=1, delay=50ms
	b.Allow("svc")
	time.Sleep(55 * time.Millisecond)

	// Second allow seeds attempt=2, delay=100ms
	b.Allow("svc")

	state, ok := b.State("svc")
	if !ok {
		t.Fatal("expected state to exist")
	}
	if state.Attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", state.Attempts)
	}
	// NextAllowed should be ~100ms from now, so blocked immediately
	if b.Allow("svc") {
		t.Fatal("expected call to be blocked during increased backoff")
	}
}

func TestBackoff_MaxDelayCapped(t *testing.T) {
	b := NewBackoffStore(100*time.Millisecond, 200*time.Millisecond)

	// Drive attempts high enough that uncapped delay would exceed max
	for i := 0; i < 5; i++ {
		b.Allow("svc")
		time.Sleep(5 * time.Millisecond)
	}

	state, _ := b.State("svc")
	remaining := time.Until(state.NextAllowed)
	if remaining > 210*time.Millisecond {
		t.Fatalf("expected delay capped at max, got %v", remaining)
	}
}

func TestBackoff_Reset_ClearsState(t *testing.T) {
	b := newTestBackoff()
	b.Allow("svc")
	b.Reset("svc")

	_, ok := b.State("svc")
	if ok {
		t.Fatal("expected state to be cleared after reset")
	}
	// Should be allowed again immediately after reset
	if !b.Allow("svc") {
		t.Fatal("expected allow after reset")
	}
}

func TestBackoff_DifferentProcesses_Independent(t *testing.T) {
	b := newTestBackoff()
	b.Allow("svc-a")

	if !b.Allow("svc-b") {
		t.Fatal("expected svc-b to be independent of svc-a")
	}
}

func TestBackoff_State_Missing(t *testing.T) {
	b := newTestBackoff()
	_, ok := b.State("unknown")
	if ok {
		t.Fatal("expected no state for unknown process")
	}
}
