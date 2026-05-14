package monitor

import (
	"testing"
)

func TestRosterStore_Enroll_And_Get(t *testing.T) {
	r := NewRosterStore()
	if err := r.Enroll("nginx", "team-ops"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := r.Get("nginx")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Name != "nginx" || e.Owner != "team-ops" || !e.Enabled {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestRosterStore_Enroll_EmptyProcess_ReturnsError(t *testing.T) {
	r := NewRosterStore()
	if err := r.Enroll("", "owner"); err == nil {
		t.Fatal("expected error for empty process name")
	}
}

func TestRosterStore_Get_Missing(t *testing.T) {
	r := NewRosterStore()
	_, ok := r.Get("unknown")
	if ok {
		t.Fatal("expected missing entry")
	}
}

func TestRosterStore_IsEnabled_True(t *testing.T) {
	r := NewRosterStore()
	_ = r.Enroll("redis", "")
	if !r.IsEnabled("redis") {
		t.Fatal("expected process to be enabled after enroll")
	}
}

func TestRosterStore_Disable(t *testing.T) {
	r := NewRosterStore()
	_ = r.Enroll("redis", "")
	if err := r.Disable("redis"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.IsEnabled("redis") {
		t.Fatal("expected process to be disabled")
	}
}

func TestRosterStore_Disable_Missing_ReturnsError(t *testing.T) {
	r := NewRosterStore()
	if err := r.Disable("ghost"); err == nil {
		t.Fatal("expected error for unknown process")
	}
}

func TestRosterStore_Remove(t *testing.T) {
	r := NewRosterStore()
	_ = r.Enroll("postgres", "dba")
	r.Remove("postgres")
	_, ok := r.Get("postgres")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestRosterStore_All_ReturnsCopy(t *testing.T) {
	r := NewRosterStore()
	_ = r.Enroll("svc-a", "team-a")
	_ = r.Enroll("svc-b", "team-b")
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the slice should not affect the store.
	all[0].Owner = "mutated"
	for _, e := range r.All() {
		if e.Owner == "mutated" {
			t.Error("store was mutated through returned slice")
		}
	}
}

func TestRosterStore_IsEnabled_NotEnrolled(t *testing.T) {
	r := NewRosterStore()
	if r.IsEnabled("nope") {
		t.Fatal("expected false for unenrolled process")
	}
}
