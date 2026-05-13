package monitor

import (
	"testing"
	"time"
)

func newTestCircuit() (*CircuitBreaker, *time.Time) {
	now := time.Now()
	cb := NewCircuitBreaker(3, 10*time.Second)
	cb.now = func() time.Time { return now }
	return cb, &now
}

func TestCircuit_ClosedUnderThreshold(t *testing.T) {
	cb, _ := newTestCircuit()
	for i := 0; i < 2; i++ {
		state := cb.Record("svc")
		if state != CircuitClosed {
			t.Fatalf("expected closed, got %s", state)
		}
	}
}

func TestCircuit_OpensAtThreshold(t *testing.T) {
	cb, _ := newTestCircuit()
	var state CircuitState
	for i := 0; i < 3; i++ {
		state = cb.Record("svc")
	}
	if state != CircuitOpen {
		t.Fatalf("expected open, got %s", state)
	}
}

func TestCircuit_StaysOpenWithinWindow(t *testing.T) {
	cb, _ := newTestCircuit()
	for i := 0; i < 3; i++ {
		cb.Record("svc")
	}
	if got := cb.State("svc"); got != CircuitOpen {
		t.Fatalf("expected open, got %s", got)
	}
}

func TestCircuit_HalfOpenAfterWindow(t *testing.T) {
	cb, nowPtr := newTestCircuit()
	for i := 0; i < 3; i++ {
		cb.Record("svc")
	}
	*nowPtr = nowPtr.Add(11 * time.Second)
	if got := cb.State("svc"); got != CircuitHalfOpen {
		t.Fatalf("expected half_open, got %s", got)
	}
}

func TestCircuit_Reset_ClosesBreakerManually(t *testing.T) {
	cb, _ := newTestCircuit()
	for i := 0; i < 3; i++ {
		cb.Record("svc")
	}
	cb.Reset("svc")
	if got := cb.State("svc"); got != CircuitClosed {
		t.Fatalf("expected closed after reset, got %s", got)
	}
}

func TestCircuit_DifferentProcesses_Independent(t *testing.T) {
	cb, _ := newTestCircuit()
	for i := 0; i < 3; i++ {
		cb.Record("svc-a")
	}
	if got := cb.State("svc-b"); got != CircuitClosed {
		t.Fatalf("svc-b should be closed, got %s", got)
	}
}

func TestCircuit_All_ReturnsSnapshot(t *testing.T) {
	cb, _ := newTestCircuit()
	for i := 0; i < 3; i++ {
		cb.Record("svc-a")
	}
	cb.Record("svc-b")
	all := cb.All()
	if all["svc-a"] != CircuitOpen {
		t.Errorf("svc-a should be open")
	}
	if all["svc-b"] != CircuitClosed {
		t.Errorf("svc-b should be closed")
	}
}
