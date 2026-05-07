package monitor

import (
	"sort"
	"testing"
)

func TestDependencyGraph_AddAndDependencies(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("worker", "db")
	g.Add("worker", "cache")

	deps := g.Dependencies("worker")
	sort.Strings(deps)
	if len(deps) != 2 || deps[0] != "cache" || deps[1] != "db" {
		t.Fatalf("expected [cache db], got %v", deps)
	}
}

func TestDependencyGraph_Add_Idempotent(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("api", "db")
	g.Add("api", "db")

	if len(g.Dependencies("api")) != 1 {
		t.Fatal("duplicate edge should not be recorded")
	}
}

func TestDependencyGraph_Dependents(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("api", "db")
	g.Add("worker", "db")
	g.Add("scheduler", "cache")

	deps := g.Dependents("db")
	sort.Strings(deps)
	if len(deps) != 2 || deps[0] != "api" || deps[1] != "worker" {
		t.Fatalf("expected [api worker], got %v", deps)
	}
}

func TestDependencyGraph_Dependents_None(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("api", "db")

	if len(g.Dependents("api")) != 0 {
		t.Fatal("api has no dependents")
	}
}

func TestDependencyGraph_Remove(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("api", "db")
	g.Add("api", "cache")
	g.Remove("api", "db")

	deps := g.Dependencies("api")
	if len(deps) != 1 || deps[0] != "cache" {
		t.Fatalf("expected [cache] after remove, got %v", deps)
	}
}

func TestDependencyGraph_Remove_NonExistent(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("api", "db")
	g.Remove("api", "nonexistent") // should not panic

	if len(g.Dependencies("api")) != 1 {
		t.Fatal("remove of nonexistent edge should leave graph unchanged")
	}
}

func TestDependencyGraph_All(t *testing.T) {
	g := NewDependencyGraph()
	g.Add("a", "b")
	g.Add("a", "c")
	g.Add("d", "e")

	all := g.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(all))
	}
	for _, edge := range all {
		if edge.RecordedAt.IsZero() {
			t.Error("RecordedAt should not be zero")
		}
	}
}

func TestDependencyGraph_Dependencies_Unknown(t *testing.T) {
	g := NewDependencyGraph()
	if deps := g.Dependencies("ghost"); len(deps) != 0 {
		t.Fatalf("expected empty for unknown process, got %v", deps)
	}
}
