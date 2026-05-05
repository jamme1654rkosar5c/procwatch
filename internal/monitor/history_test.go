package monitor

import (
	"fmt"
	"testing"
	"time"
)

func TestHistory_Record_And_Len(t *testing.T) {
	h := NewHistory(10)
	if h.Len() != 0 {
		t.Fatalf("expected 0, got %d", h.Len())
	}
	h.Record("nginx", "down", "exit code 1")
	if h.Len() != 1 {
		t.Fatalf("expected 1, got %d", h.Len())
	}
}

func TestHistory_All_ReturnsCopy(t *testing.T) {
	h := NewHistory(10)
	h.Record("nginx", "down", "")
	h.Record("redis", "cpu", "90%")

	all := h.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 records, got %d", len(all))
	}
	// Mutating the returned slice must not affect internal state.
	all[0].ProcessName = "mutated"
	if h.All()[0].ProcessName == "mutated" {
		t.Fatal("All() returned a reference to internal slice")
	}
}

func TestHistory_ForProcess_Filters(t *testing.T) {
	h := NewHistory(20)
	h.Record("nginx", "down", "")
	h.Record("redis", "down", "")
	h.Record("nginx", "recovered", "")

	results := h.ForProcess("nginx")
	if len(results) != 2 {
		t.Fatalf("expected 2 nginx records, got %d", len(results))
	}
	for _, r := range results {
		if r.ProcessName != "nginx" {
			t.Errorf("unexpected process name %q", r.ProcessName)
		}
	}
}

func TestHistory_Eviction_WhenFull(t *testing.T) {
	const max = 5
	h := NewHistory(max)
	for i := 0; i < max+3; i++ {
		h.Record("svc", "cpu", fmt.Sprintf("iter %d", i))
	}
	if h.Len() != max {
		t.Fatalf("expected %d records after eviction, got %d", max, h.Len())
	}
	// The oldest entries should have been dropped; last record should be iter 7.
	all := h.All()
	if all[max-1].Details != fmt.Sprintf("iter %d", max+2) {
		t.Errorf("unexpected last record details: %q", all[max-1].Details)
	}
}

func TestHistory_Record_TimestampIsRecent(t *testing.T) {
	before := time.Now()
	h := NewHistory(10)
	h.Record("svc", "down", "")
	after := time.Now()

	r := h.All()[0]
	if r.OccurredAt.Before(before) || r.OccurredAt.After(after) {
		t.Errorf("timestamp %v out of expected range [%v, %v]", r.OccurredAt, before, after)
	}
}
