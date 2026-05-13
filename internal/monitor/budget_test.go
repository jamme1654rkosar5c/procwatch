package monitor

import (
	"testing"
	"time"
)

func newTestBudget(limit int, window time.Duration) (*AlertBudget, *time.Time) {
	now := time.Now()
	b := NewAlertBudget(limit, window)
	b.nowFunc = func() time.Time { return now }
	return b, &now
}

func TestBudget_ConsumeUnderLimit(t *testing.T) {
	b, _ := newTestBudget(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !b.Consume("nginx") {
			t.Fatalf("expected consume %d to succeed", i+1)
		}
	}
}

func TestBudget_ConsumeExceedsLimit(t *testing.T) {
	b, _ := newTestBudget(2, time.Minute)
	b.Consume("nginx")
	b.Consume("nginx")
	if b.Consume("nginx") {
		t.Fatal("expected third consume to be denied")
	}
}

func TestBudget_WindowExpiry_ResetsCount(t *testing.T) {
	now := time.Now()
	b := NewAlertBudget(2, time.Minute)
	b.nowFunc = func() time.Time { return now }

	b.Consume("nginx")
	b.Consume("nginx")

	// advance past the window
	now = now.Add(2 * time.Minute)
	if !b.Consume("nginx") {
		t.Fatal("expected consume to succeed after window expiry")
	}
}

func TestBudget_Remaining(t *testing.T) {
	b, _ := newTestBudget(5, time.Minute)
	if b.Remaining("nginx") != 5 {
		t.Fatalf("expected 5 remaining, got %d", b.Remaining("nginx"))
	}
	b.Consume("nginx")
	b.Consume("nginx")
	if b.Remaining("nginx") != 3 {
		t.Fatalf("expected 3 remaining, got %d", b.Remaining("nginx"))
	}
}

func TestBudget_DifferentProcesses_Independent(t *testing.T) {
	b, _ := newTestBudget(1, time.Minute)
	b.Consume("nginx")
	if !b.Consume("redis") {
		t.Fatal("redis budget should be independent of nginx")
	}
}

func TestBudget_Reset_ClearsState(t *testing.T) {
	b, _ := newTestBudget(1, time.Minute)
	b.Consume("nginx")
	if b.Consume("nginx") {
		t.Fatal("expected second consume to fail")
	}
	_ = b.Reset("nginx")
	if !b.Consume("nginx") {
		t.Fatal("expected consume to succeed after reset")
	}
}

func TestBudget_Reset_EmptyProcess_ReturnsError(t *testing.T) {
	b, _ := newTestBudget(3, time.Minute)
	if err := b.Reset(""); err == nil {
		t.Fatal("expected error for empty process name")
	}
}

func TestBudget_Consume_EmptyProcess_ReturnsFalse(t *testing.T) {
	b, _ := newTestBudget(3, time.Minute)
	if b.Consume("") {
		t.Fatal("expected false for empty process name")
	}
}
