package monitor

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 3)
	for i := 0; i < 3; i++ {
		if !rl.Allow("proc:down") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 2)
	rl.Allow("proc:down")
	rl.Allow("proc:down")
	if rl.Allow("proc:down") {
		t.Fatal("expected Allow=false when limit exceeded")
	}
}

func TestRateLimiter_Allow_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 1)
	if !rl.Allow("proc:down") {
		t.Fatal("first call should be allowed")
	}
	if rl.Allow("proc:down") {
		t.Fatal("second call within window should be denied")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow("proc:down") {
		t.Fatal("call after window expiry should be allowed")
	}
}

func TestRateLimiter_Reset_ClearsState(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 1)
	rl.Allow("proc:down")
	if rl.Allow("proc:down") {
		t.Fatal("should be denied before reset")
	}
	rl.Reset("proc:down")
	if !rl.Allow("proc:down") {
		t.Fatal("should be allowed after reset")
	}
}

func TestRateLimiter_Count_ReflectsWindow(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 10)
	rl.Allow("proc:cpu")
	rl.Allow("proc:cpu")
	rl.Allow("proc:cpu")
	if c := rl.Count("proc:cpu"); c != 3 {
		t.Fatalf("expected count 3, got %d", c)
	}
}

func TestRateLimiter_DifferentKeys_Independent(t *testing.T) {
	rl := NewRateLimiter(time.Minute, 1)
	rl.Allow("nginx:down")
	if !rl.Allow("redis:down") {
		t.Fatal("different key should have independent limit")
	}
}
