package monitor

import (
	"testing"
)

func TestLabelStore_SetAndGet(t *testing.T) {
	s := NewLabelStore()
	if err := s.Set("nginx", "env", "production"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := s.Get("nginx", "env")
	if !ok {
		t.Fatal("expected label to exist")
	}
	if v != "production" {
		t.Fatalf("expected 'production', got %q", v)
	}
}

func TestLabelStore_Get_Missing(t *testing.T) {
	s := NewLabelStore()
	_, ok := s.Get("unknown", "env")
	if ok {
		t.Fatal("expected label to be missing")
	}
}

func TestLabelStore_Set_EmptyProcess_ReturnsError(t *testing.T) {
	s := NewLabelStore()
	if err := s.Set("", "env", "prod"); err == nil {
		t.Fatal("expected error for empty process name")
	}
}

func TestLabelStore_Set_EmptyKey_ReturnsError(t *testing.T) {
	s := NewLabelStore()
	if err := s.Set("nginx", "", "prod"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestLabelStore_Delete_RemovesKey(t *testing.T) {
	s := NewLabelStore()
	_ = s.Set("nginx", "env", "prod")
	s.Delete("nginx", "env")
	_, ok := s.Get("nginx", "env")
	if ok {
		t.Fatal("expected label to be deleted")
	}
}

func TestLabelStore_Delete_NoOp_UnknownProcess(t *testing.T) {
	s := NewLabelStore()
	s.Delete("ghost", "env") // must not panic
}

func TestLabelStore_All_ReturnsCopy(t *testing.T) {
	s := NewLabelStore()
	_ = s.Set("nginx", "env", "prod")
	_ = s.Set("nginx", "team", "platform")
	m := s.All("nginx")
	if len(m) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(m))
	}
	// Mutating the copy must not affect the store.
	m["env"] = "staging"
	v, _ := s.Get("nginx", "env")
	if v != "prod" {
		t.Fatal("copy mutation affected store")
	}
}

func TestLabelStore_All_NilForUnknownProcess(t *testing.T) {
	s := NewLabelStore()
	if s.All("ghost") != nil {
		t.Fatal("expected nil for unknown process")
	}
}

func TestLabelStore_Processes(t *testing.T) {
	s := NewLabelStore()
	_ = s.Set("nginx", "env", "prod")
	_ = s.Set("redis", "env", "prod")
	procs := s.Processes()
	if len(procs) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(procs))
	}
}
