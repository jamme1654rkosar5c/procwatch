package monitor

import (
	"testing"
	"time"
)

func newTestOncall() (*OnCallStore, time.Time) {
	now := time.Now()
	return NewOnCallStore(), now
}

func TestOnCallStore_Set_And_Get(t *testing.T) {
	s, now := newTestOncall()
	e := OnCallEntry{
		Process:   "api",
		Owner:     "alice",
		Email:     "alice@example.com",
		StartsAt:  now.Add(-time.Hour),
		ExpiresAt: now.Add(time.Hour),
	}
	if err := s.Set(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("api")
	if !ok {
		t.Fatal("expected entry")
	}
	if got.Owner != "alice" {
		t.Errorf("got owner %q, want alice", got.Owner)
	}
}

func TestOnCallStore_Get_Missing(t *testing.T) {
	s, _ := newTestOncall()
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected no entry")
	}
}

func TestOnCallStore_Active_InWindow(t *testing.T) {
	s, now := newTestOncall()
	e := OnCallEntry{Process: "api", Owner: "bob", Email: "b@x.com",
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)}
	_ = s.Set(e)
	_, ok := s.Active("api", now)
	if !ok {
		t.Fatal("expected active entry")
	}
}

func TestOnCallStore_Active_Expired(t *testing.T) {
	s, now := newTestOncall()
	e := OnCallEntry{Process: "api", Owner: "bob", Email: "b@x.com",
		StartsAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Hour)}
	_ = s.Set(e)
	_, ok := s.Active("api", now)
	if ok {
		t.Fatal("expected inactive entry")
	}
}

func TestOnCallStore_Set_EmptyProcess(t *testing.T) {
	s, now := newTestOncall()
	err := s.Set(OnCallEntry{Owner: "x", Email: "x@x.com",
		StartsAt: now, ExpiresAt: now.Add(time.Hour)})
	if err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestOnCallStore_Delete(t *testing.T) {
	s, now := newTestOncall()
	_ = s.Set(OnCallEntry{Process: "api", Owner: "c", Email: "c@x.com",
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)})
	s.Delete("api")
	_, ok := s.Get("api")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestOnCallStore_All_ReturnsCopy(t *testing.T) {
	s, now := newTestOncall()
	_ = s.Set(OnCallEntry{Process: "api", Owner: "d", Email: "d@x.com",
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)})
	_ = s.Set(OnCallEntry{Process: "worker", Owner: "e", Email: "e@x.com",
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("got %d entries, want 2", len(all))
	}
}
