package monitor

import (
	"testing"
	"time"
)

func TestThrottler_Allow_FirstCall(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	if !th.Allow("nginx", "down") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestThrottler_Allow_WithinCooldown(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	th.Allow("nginx", "down") // first call records time
	if th.Allow("nginx", "down") {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestThrottler_Allow_AfterCooldown(t *testing.T) {
	th := NewThrottler(10 * time.Millisecond)
	th.Allow("nginx", "down")
	time.Sleep(20 * time.Millisecond)
	if !th.Allow("nginx", "down") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestThrottler_Allow_DifferentEventTypes(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	th.Allow("nginx", "down")
	if !th.Allow("nginx", "cpu") {
		t.Fatal("expected different event type to be allowed independently")
	}
}

func TestThrottler_Allow_DifferentProcesses(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	th.Allow("nginx", "down")
	if !th.Allow("redis", "down") {
		t.Fatal("expected different process to be allowed independently")
	}
}

func TestThrottler_Reset_AllowsImmediately(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	th.Allow("nginx", "down")
	th.Reset("nginx", "down")
	if !th.Allow("nginx", "down") {
		t.Fatal("expected call after reset to be allowed")
	}
}

func TestThrottler_Reset_UnknownKey(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	// resetting a key that was never set should not panic
	th.Reset("unknown", "down")
}
