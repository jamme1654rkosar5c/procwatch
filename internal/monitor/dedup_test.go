package monitor

import (
	"testing"
	"time"
)

func TestDeduplicator_FirstCallNotDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	if d.IsDuplicate("nginx", "down", "") {
		t.Fatal("first call should not be a duplicate")
	}
}

func TestDeduplicator_SecondCallWithinWindowIsDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("nginx", "down", "")
	if !d.IsDuplicate("nginx", "down", "") {
		t.Fatal("second call within window should be a duplicate")
	}
}

func TestDeduplicator_AfterWindowExpiry_NotDuplicate(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(10 * time.Second)
	d.nowFunc = func() time.Time { return now }

	d.IsDuplicate("nginx", "cpu", "cpu_percent")

	// Advance clock past the window.
	d.nowFunc = func() time.Time { return now.Add(11 * time.Second) }

	if d.IsDuplicate("nginx", "cpu", "cpu_percent") {
		t.Fatal("event after window expiry should not be a duplicate")
	}
}

func TestDeduplicator_DifferentProcesses_NotDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("nginx", "down", "")
	if d.IsDuplicate("redis", "down", "") {
		t.Fatal("different process should not be a duplicate")
	}
}

func TestDeduplicator_DifferentEventTypes_NotDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("nginx", "down", "")
	if d.IsDuplicate("nginx", "cpu", "cpu_percent") {
		t.Fatal("different event type should not be a duplicate")
	}
}

func TestDeduplicator_DifferentFields_NotDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	d.IsDuplicate("nginx", "threshold", "cpu_percent")
	if d.IsDuplicate("nginx", "threshold", "mem_rss_bytes") {
		t.Fatal("different threshold field should not be a duplicate")
	}
}

func TestDeduplicator_Flush_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(5 * time.Second)
	d.nowFunc = func() time.Time { return now }

	d.IsDuplicate("nginx", "down", "")
	d.IsDuplicate("redis", "down", "")

	if d.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", d.Len())
	}

	d.nowFunc = func() time.Time { return now.Add(10 * time.Second) }
	d.Flush()

	if d.Len() != 0 {
		t.Fatalf("expected 0 entries after flush, got %d", d.Len())
	}
}

func TestDeduplicator_Flush_KeepsActiveEntries(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(30 * time.Second)
	d.nowFunc = func() time.Time { return now }

	d.IsDuplicate("nginx", "down", "")
	d.Flush()

	if d.Len() != 1 {
		t.Fatalf("expected 1 active entry after flush, got %d", d.Len())
	}
}
