package monitor

import (
	"testing"
	"time"
)

func newTestCooldown() *CooldownStore {
	c := NewCooldownStore()
	return c
}

func TestCooldown_Active_NotTracked(t *testing.T) {
	c := newTestCooldown()
	if c.Active("nginx") {
		t.Fatal("expected no active cooldown for untracked process")
	}
}

func TestCooldown_Set_Active(t *testing.T) {
	c := newTestCooldown()
	c.Set("nginx", 10*time.Second)
	if !c.Active("nginx") {
		t.Fatal("expected cooldown to be active after Set")
	}
}

func TestCooldown_Active_Expired(t *testing.T) {
	now := time.Now()
	c := NewCooldownStore()
	c.now = func() time.Time { return now }
	c.Set("nginx", 1*time.Second)

	// Advance time past expiry
	c.now = func() time.Time { return now.Add(2 * time.Second) }
	if c.Active("nginx") {
		t.Fatal("expected cooldown to be expired")
	}
}

func TestCooldown_Lift_RemovesCooldown(t *testing.T) {
	c := newTestCooldown()
	c.Set("redis", 30*time.Second)
	c.Lift("redis")
	if c.Active("redis") {
		t.Fatal("expected cooldown to be removed after Lift")
	}
}

func TestCooldown_All_ReturnsActiveOnly(t *testing.T) {
	now := time.Now()
	c := NewCooldownStore()
	c.now = func() time.Time { return now }

	c.Set("nginx", 10*time.Second)
	c.Set("redis", 1*time.Second)

	// Advance past redis expiry only
	c.now = func() time.Time { return now.Add(2 * time.Second) }

	all := c.All()
	if _, ok := all["nginx"]; !ok {
		t.Error("expected nginx to appear in All()")
	}
	if _, ok := all["redis"]; ok {
		t.Error("expected expired redis to be excluded from All()")
	}
}

func TestCooldown_All_ReturnsCopy(t *testing.T) {
	c := newTestCooldown()
	c.Set("nginx", 10*time.Second)

	all := c.All()
	all["injected"] = CooldownEntry{ExpiresAt: time.Now().Add(time.Hour)}

	if _, ok := c.All()["injected"]; ok {
		t.Error("modifying returned map should not affect internal state")
	}
}
