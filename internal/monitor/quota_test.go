package monitor

import (
	"testing"
	"time"
)

func newTestQuota(window time.Duration) *QuotaStore {
	qs := NewQuotaStore(window)
	qs.now = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }
	return qs
}

func TestQuota_AllowWithoutLimit(t *testing.T) {
	qs := newTestQuota(time.Minute)
	if !qs.Allow("nginx") {
		t.Fatal("expected Allow=true when no limit configured")
	}
}

func TestQuota_UnderLimit(t *testing.T) {
	qs := newTestQuota(time.Minute)
	_ = qs.SetLimit("nginx", 3)
	for i := 0; i < 3; i++ {
		if !qs.Allow("nginx") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestQuota_ExceedsLimit(t *testing.T) {
	qs := newTestQuota(time.Minute)
	_ = qs.SetLimit("nginx", 2)
	qs.Allow("nginx")
	qs.Allow("nginx")
	if qs.Allow("nginx") {
		t.Fatal("expected Allow=false after limit exceeded")
	}
}

func TestQuota_WindowExpiry_Resets(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	qs := NewQuotaStore(time.Minute)
	qs.now = func() time.Time { return now }
	_ = qs.SetLimit("nginx", 1)
	qs.Allow("nginx")
	if qs.Allow("nginx") {
		t.Fatal("should be blocked before window expires")
	}
	now = now.Add(2 * time.Minute)
	if !qs.Allow("nginx") {
		t.Fatal("expected Allow=true after window expiry")
	}
}

func TestQuota_SetLimit_EmptyProcess(t *testing.T) {
	qs := newTestQuota(time.Minute)
	if err := qs.SetLimit("", 5); err == nil {
		t.Fatal("expected error for empty process")
	}
}

func TestQuota_SetLimit_ZeroLimit(t *testing.T) {
	qs := newTestQuota(time.Minute)
	if err := qs.SetLimit("nginx", 0); err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestQuota_Get_Missing(t *testing.T) {
	qs := newTestQuota(time.Minute)
	if qs.Get("unknown") != nil {
		t.Fatal("expected nil for unknown process")
	}
}

func TestQuota_Get_ReturnsSnapshot(t *testing.T) {
	qs := newTestQuota(time.Minute)
	_ = qs.SetLimit("nginx", 5)
	qs.Allow("nginx")
	e := qs.Get("nginx")
	if e == nil || e.Used != 1 || e.Limit != 5 {
		t.Fatalf("unexpected entry: %+v", e)
	}
}

func TestQuota_Reset_ClearsUsage(t *testing.T) {
	qs := newTestQuota(time.Minute)
	_ = qs.SetLimit("nginx", 2)
	qs.Allow("nginx")
	qs.Allow("nginx")
	qs.Reset("nginx")
	if !qs.Allow("nginx") {
		t.Fatal("expected Allow=true after reset")
	}
}

func TestQuota_All_ReturnsCopy(t *testing.T) {
	qs := newTestQuota(time.Minute)
	_ = qs.SetLimit("a", 3)
	_ = qs.SetLimit("b", 5)
	all := qs.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
