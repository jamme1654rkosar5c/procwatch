package monitor

import (
	"testing"
)

func TestTagRegistry_SetAndGet(t *testing.T) {
	r := NewTagRegistry()
	r.Set("nginx", "env", "production")

	v, ok := r.Get("nginx", "env")
	if !ok {
		t.Fatal("expected tag to be found")
	}
	if v != "production" {
		t.Errorf("expected 'production', got %q", v)
	}
}

func TestTagRegistry_Get_Missing(t *testing.T) {
	r := NewTagRegistry()
	_, ok := r.Get("unknown", "env")
	if ok {
		t.Error("expected tag to be missing")
	}
}

func TestTagRegistry_All_ReturnsCopy(t *testing.T) {
	r := NewTagRegistry()
	r.Set("nginx", "env", "staging")
	r.Set("nginx", "team", "platform")

	tags := r.All("nginx")
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	// Mutating the copy should not affect the registry
	tags["env"] = "mutated"
	v, _ := r.Get("nginx", "env")
	if v != "staging" {
		t.Errorf("registry was mutated, expected 'staging', got %q", v)
	}
}

func TestTagRegistry_All_NilForUnknownProcess(t *testing.T) {
	r := NewTagRegistry()
	if tags := r.All("ghost"); tags != nil {
		t.Errorf("expected nil for unknown process, got %v", tags)
	}
}

func TestTagRegistry_Delete_RemovesKey(t *testing.T) {
	r := NewTagRegistry()
	r.Set("redis", "env", "prod")
	r.Delete("redis", "env")

	_, ok := r.Get("redis", "env")
	if ok {
		t.Error("expected tag to be deleted")
	}
}

func TestTagRegistry_Delete_CleansUpEmptyProcess(t *testing.T) {
	r := NewTagRegistry()
	r.Set("redis", "env", "prod")
	r.Delete("redis", "env")

	procs := r.Processes()
	for _, p := range procs {
		if p == "redis" {
			t.Error("expected process to be removed after all tags deleted")
		}
	}
}

func TestTagRegistry_Processes(t *testing.T) {
	r := NewTagRegistry()
	r.Set("nginx", "env", "prod")
	r.Set("redis", "env", "prod")

	procs := r.Processes()
	if len(procs) != 2 {
		t.Errorf("expected 2 processes, got %d", len(procs))
	}
}

func TestTagRegistry_Set_Overwrites(t *testing.T) {
	r := NewTagRegistry()
	r.Set("nginx", "env", "staging")
	r.Set("nginx", "env", "production")

	v, _ := r.Get("nginx", "env")
	if v != "production" {
		t.Errorf("expected overwritten value 'production', got %q", v)
	}
}
