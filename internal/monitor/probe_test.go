package monitor

import (
	"testing"
	"time"
)

func TestProbeStore_Record_And_Get(t *testing.T) {
	ps := NewProbeStore()
	ps.Record("nginx", true, "ok")

	r, ok := ps.Get("nginx")
	if !ok {
		t.Fatal("expected result to exist")
	}
	if r.Process != "nginx" {
		t.Errorf("expected process nginx, got %s", r.Process)
	}
	if !r.Success {
		t.Error("expected success=true")
	}
	if r.Message != "ok" {
		t.Errorf("expected message 'ok', got %s", r.Message)
	}
}

func TestProbeStore_Get_Missing(t *testing.T) {
	ps := NewProbeStore()
	_, ok := ps.Get("unknown")
	if ok {
		t.Error("expected missing result to return false")
	}
}

func TestProbeStore_Record_SetsTimestamp(t *testing.T) {
	ps := NewProbeStore()
	before := time.Now()
	ps.Record("svc", false, "timeout")
	after := time.Now()

	r, _ := ps.Get("svc")
	if r.CheckedAt.Before(before) || r.CheckedAt.After(after) {
		t.Error("CheckedAt timestamp out of expected range")
	}
}

func TestProbeStore_Record_Overwrites(t *testing.T) {
	ps := NewProbeStore()
	ps.Record("svc", false, "down")
	ps.Record("svc", true, "recovered")

	r, _ := ps.Get("svc")
	if !r.Success {
		t.Error("expected overwritten result to be success")
	}
	if r.Message != "recovered" {
		t.Errorf("expected message 'recovered', got %s", r.Message)
	}
}

func TestProbeStore_All_ReturnsCopy(t *testing.T) {
	ps := NewProbeStore()
	ps.Record("a", true, "")
	ps.Record("b", false, "err")

	all := ps.All()
	if len(all) != 2 {
		t.Errorf("expected 2 results, got %d", len(all))
	}
	// Mutating the copy should not affect the store
	delete(all, "a")
	if _, ok := ps.Get("a"); !ok {
		t.Error("deleting from copy should not affect store")
	}
}

func TestProbeStore_Delete(t *testing.T) {
	ps := NewProbeStore()
	ps.Record("svc", true, "ok")
	ps.Delete("svc")

	_, ok := ps.Get("svc")
	if ok {
		t.Error("expected result to be deleted")
	}
}
