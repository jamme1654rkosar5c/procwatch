package monitor

import (
	"testing"
)

func TestRunbookStore_SetAndGet(t *testing.T) {
	s := NewRunbookStore()
	if err := s.Set("nginx", "https://wiki.example.com/nginx", "primary runbook"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("nginx")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.URL != "https://wiki.example.com/nginx" {
		t.Errorf("got URL %q, want https://wiki.example.com/nginx", e.URL)
	}
	if e.Note != "primary runbook" {
		t.Errorf("got note %q, want 'primary runbook'", e.Note)
	}
}

func TestRunbookStore_Get_Missing(t *testing.T) {
	s := NewRunbookStore()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected no entry")
	}
}

func TestRunbookStore_Set_EmptyProcess_ReturnsError(t *testing.T) {
	s := NewRunbookStore()
	if err := s.Set("", "https://example.com", ""); err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestRunbookStore_Set_EmptyURL_ReturnsError(t *testing.T) {
	s := NewRunbookStore()
	if err := s.Set("nginx", "", ""); err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestRunbookStore_Delete(t *testing.T) {
	s := NewRunbookStore()
	_ = s.Set("nginx", "https://example.com", "")
	s.Delete("nginx")
	_, ok := s.Get("nginx")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestRunbookStore_All_ReturnsCopy(t *testing.T) {
	s := NewRunbookStore()
	_ = s.Set("nginx", "https://example.com/nginx", "")
	_ = s.Set("redis", "https://example.com/redis", "")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the slice should not affect the store.
	all[0] = RunbookEntry{}
	if len(s.All()) != 2 {
		t.Fatal("store was mutated via returned slice")
	}
}

func TestRunbookStore_Set_UpdatesTimestamp(t *testing.T) {
	s := NewRunbookStore()
	_ = s.Set("nginx", "https://example.com", "")
	e, _ := s.Get("nginx")
	if e.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}
