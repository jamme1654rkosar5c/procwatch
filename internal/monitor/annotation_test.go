package monitor

import (
	"testing"
	"time"
)

func TestAnnotation_SetAndGet(t *testing.T) {
	s := NewAnnotationStore()
	s.Set("nginx", "primary web server")

	a, ok := s.Get("nginx")
	if !ok {
		t.Fatal("expected annotation to exist")
	}
	if a.Text != "primary web server" {
		t.Errorf("got text %q, want %q", a.Text, "primary web server")
	}
}

func TestAnnotation_Get_Missing(t *testing.T) {
	s := NewAnnotationStore()
	_, ok := s.Get("unknown")
	if ok {
		t.Error("expected no annotation for unknown process")
	}
}

func TestAnnotation_Set_Overwrites_PreservesCreatedAt(t *testing.T) {
	s := NewAnnotationStore()
	s.Set("redis", "first")

	a1, _ := s.Get("redis")
	time.Sleep(2 * time.Millisecond)
	s.Set("redis", "second")

	a2, _ := s.Get("redis")
	if a2.Text != "second" {
		t.Errorf("got text %q, want %q", a2.Text, "second")
	}
	if !a2.CreatedAt.Equal(a1.CreatedAt) {
		t.Error("CreatedAt should not change on update")
	}
	if !a2.UpdatedAt.After(a1.UpdatedAt) {
		t.Error("UpdatedAt should advance on update")
	}
}

func TestAnnotation_Delete(t *testing.T) {
	s := NewAnnotationStore()
	s.Set("postgres", "main db")
	s.Delete("postgres")

	_, ok := s.Get("postgres")
	if ok {
		t.Error("expected annotation to be deleted")
	}
}

func TestAnnotation_All_ReturnsCopy(t *testing.T) {
	s := NewAnnotationStore()
	s.Set("nginx", "web")
	s.Set("redis", "cache")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("got %d entries, want 2", len(all))
	}

	// Mutating the returned map must not affect the store.
	delete(all, "nginx")
	if _, ok := s.Get("nginx"); !ok {
		t.Error("store should not be affected by mutation of returned map")
	}
}
