package monitor

import (
	"testing"
)

func TestPolicyStore_SetAndGet(t *testing.T) {
	s := NewPolicyStore()
	p := AlertPolicy{Process: "nginx", MinSeverity: "warn", Channels: []string{"slack"}}
	if err := s.Set(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("nginx")
	if !ok {
		t.Fatal("expected policy to be found")
	}
	if got.MinSeverity != "warn" {
		t.Errorf("expected warn, got %s", got.MinSeverity)
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPolicyStore_Get_Missing(t *testing.T) {
	s := NewPolicyStore()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected no policy")
	}
}

func TestPolicyStore_Set_EmptyProcess_ReturnsError(t *testing.T) {
	s := NewPolicyStore()
	err := s.Set(AlertPolicy{MinSeverity: "info"})
	if err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestPolicyStore_Set_InvalidSeverity_ReturnsError(t *testing.T) {
	s := NewPolicyStore()
	err := s.Set(AlertPolicy{Process: "nginx", MinSeverity: "urgent"})
	if err == nil {
		t.Fatal("expected error for invalid severity")
	}
}

func TestPolicyStore_Delete(t *testing.T) {
	s := NewPolicyStore()
	_ = s.Set(AlertPolicy{Process: "nginx", MinSeverity: "info"})
	s.Delete("nginx")
	_, ok := s.Get("nginx")
	if ok {
		t.Fatal("expected policy to be deleted")
	}
}

func TestPolicyStore_All_ReturnsCopy(t *testing.T) {
	s := NewPolicyStore()
	_ = s.Set(AlertPolicy{Process: "nginx", MinSeverity: "info"})
	_ = s.Set(AlertPolicy{Process: "redis", MinSeverity: "critical"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(all))
	}
	// mutating the returned slice should not affect the store
	all[0].Process = "modified"
	if _, ok := s.Get("modified"); ok {
		t.Error("store should not be affected by external mutation")
	}
}
