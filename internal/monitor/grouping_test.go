package monitor

import (
	"sort"
	"testing"
)

func TestGroupRegistry_AssignAndGroupOf(t *testing.T) {
	r := NewGroupRegistry()
	if err := r.Assign("nginx", "web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.GroupOf("nginx"); got != "web" {
		t.Errorf("expected 'web', got %q", got)
	}
}

func TestGroupRegistry_GroupOf_Missing(t *testing.T) {
	r := NewGroupRegistry()
	if got := r.GroupOf("unknown"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGroupRegistry_Assign_EmptyProcess_ReturnsError(t *testing.T) {
	r := NewGroupRegistry()
	if err := r.Assign("", "web"); err == nil {
		t.Error("expected error for empty process")
	}
}

func TestGroupRegistry_Assign_EmptyGroup_ReturnsError(t *testing.T) {
	r := NewGroupRegistry()
	if err := r.Assign("nginx", ""); err == nil {
		t.Error("expected error for empty group")
	}
}

func TestGroupRegistry_Members(t *testing.T) {
	r := NewGroupRegistry()
	_ = r.Assign("nginx", "web")
	_ = r.Assign("haproxy", "web")
	_ = r.Assign("postgres", "db")

	members := r.Members("web")
	sort.Strings(members)
	if len(members) != 2 || members[0] != "haproxy" || members[1] != "nginx" {
		t.Errorf("unexpected members: %v", members)
	}
}

func TestGroupRegistry_Members_Missing(t *testing.T) {
	r := NewGroupRegistry()
	if m := r.Members("nonexistent"); m != nil {
		t.Errorf("expected nil, got %v", m)
	}
}

func TestGroupRegistry_Reassign_MovesProcess(t *testing.T) {
	r := NewGroupRegistry()
	_ = r.Assign("nginx", "web")
	_ = r.Assign("nginx", "edge")

	if got := r.GroupOf("nginx"); got != "edge" {
		t.Errorf("expected 'edge', got %q", got)
	}
	if m := r.Members("web"); len(m) != 0 {
		t.Errorf("old group should be empty, got %v", m)
	}
}

func TestGroupRegistry_Remove(t *testing.T) {
	r := NewGroupRegistry()
	_ = r.Assign("nginx", "web")
	r.Remove("nginx")

	if got := r.GroupOf("nginx"); got != "" {
		t.Errorf("expected empty after remove, got %q", got)
	}
	if m := r.Members("web"); len(m) != 0 {
		t.Errorf("group should be empty after remove, got %v", m)
	}
}

func TestGroupRegistry_All_ReturnsCopy(t *testing.T) {
	r := NewGroupRegistry()
	_ = r.Assign("nginx", "web")
	_ = r.Assign("postgres", "db")

	all := r.All()
	if len(all) != 2 {
		t.Errorf("expected 2 groups, got %d", len(all))
	}
	// Mutating the copy should not affect the registry.
	delete(all, "web")
	if r.GroupOf("nginx") != "web" {
		t.Error("registry was mutated by modifying All() result")
	}
}
