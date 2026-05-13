package monitor

import (
	"testing"
	"time"
)

func newTestCorrelation(window time.Duration) *CorrelationTracker {
	return NewCorrelationTracker(window)
}

func TestCorrelation_NoPeers_FirstProcess(t *testing.T) {
	ct := newTestCorrelation(5 * time.Second)
	now := time.Now()
	peers := ct.Record("svcA", "down", now)
	if len(peers) != 0 {
		t.Fatalf("expected no peers, got %v", peers)
	}
}

func TestCorrelation_PeerDetected_SameEventType(t *testing.T) {
	ct := newTestCorrelation(5 * time.Second)
	now := time.Now()
	ct.Record("svcA", "down", now)
	peers := ct.Record("svcB", "down", now.Add(time.Second))
	if len(peers) != 1 || peers[0] != "svcA" {
		t.Fatalf("expected peer svcA, got %v", peers)
	}
}

func TestCorrelation_NoPeer_DifferentEventType(t *testing.T) {
	ct := newTestCorrelation(5 * time.Second)
	now := time.Now()
	ct.Record("svcA", "down", now)
	peers := ct.Record("svcB", "cpu", now.Add(time.Second))
	if len(peers) != 0 {
		t.Fatalf("expected no peers for different event type, got %v", peers)
	}
}

func TestCorrelation_WindowExpiry_NoPeer(t *testing.T) {
	ct := newTestCorrelation(2 * time.Second)
	now := time.Now()
	ct.Record("svcA", "down", now)
	// svcB records well outside the window.
	peers := ct.Record("svcB", "down", now.Add(10*time.Second))
	if len(peers) != 0 {
		t.Fatalf("expected no peers after window expiry, got %v", peers)
	}
}

func TestCorrelation_All_RecordsEntry(t *testing.T) {
	ct := newTestCorrelation(5 * time.Second)
	now := time.Now()
	ct.Record("svcA", "down", now)
	ct.Record("svcB", "down", now.Add(time.Second))

	all := ct.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 correlation entry, got %d", len(all))
	}
	e := all[0]
	if e.EventType != "down" {
		t.Errorf("expected event type down, got %s", e.EventType)
	}
	if e.Count != 1 {
		t.Errorf("expected count 1, got %d", e.Count)
	}
}

func TestCorrelation_All_CountIncrementsOnRepeat(t *testing.T) {
	ct := newTestCorrelation(30 * time.Second)
	now := time.Now()
	ct.Record("svcA", "down", now)
	ct.Record("svcB", "down", now.Add(time.Second))
	ct.Record("svcA", "down", now.Add(2*time.Second))
	ct.Record("svcB", "down", now.Add(3*time.Second))

	all := ct.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 correlation entry, got %d", len(all))
	}
	if all[0].Count < 2 {
		t.Errorf("expected count >= 2, got %d", all[0].Count)
	}
}

func TestCorrelation_MultiplePairs(t *testing.T) {
	ct := newTestCorrelation(5 * time.Second)
	now := time.Now()
	ct.Record("svcA", "down", now)
	ct.Record("svcB", "down", now)
	peers := ct.Record("svcC", "down", now)
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %v", peers)
	}
}
