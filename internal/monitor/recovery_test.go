package monitor

import (
	"testing"
	"time"
)

func TestCheckRecovery_NotTracked(t *testing.T) {
	downSince := map[string]time.Time{}
	ev, ok := checkRecovery("nginx", true, downSince)
	if ok || ev != nil {
		t.Fatal("expected no recovery event for a process that was never tracked as down")
	}
}

func TestCheckRecovery_StillDown(t *testing.T) {
	downSince := map[string]time.Time{
		"nginx": time.Now().Add(-30 * time.Second),
	}
	ev, ok := checkRecovery("nginx", false, downSince)
	if ok || ev != nil {
		t.Fatal("expected no recovery event while process is still down")
	}
	if _, stillTracked := downSince["nginx"]; !stillTracked {
		t.Fatal("downSince entry should remain while process is still down")
	}
}

func TestCheckRecovery_Recovered(t *testing.T) {
	downAt := time.Now().Add(-45 * time.Second)
	downSince := map[string]time.Time{"nginx": downAt}

	ev, ok := checkRecovery("nginx", true, downSince)
	if !ok {
		t.Fatal("expected recovery event")
	}
	if ev.ProcessName != "nginx" {
		t.Errorf("expected process name 'nginx', got %q", ev.ProcessName)
	}
	if ev.Downtime < 44*time.Second {
		t.Errorf("expected downtime >= 44s, got %v", ev.Downtime)
	}
	if _, stillTracked := downSince["nginx"]; stillTracked {
		t.Fatal("downSince entry should be removed after recovery")
	}
}

func TestBuildRecoveryPayload(t *testing.T) {
	now := time.Now()
	ev := &RecoveryEvent{
		ProcessName: "redis",
		PID:         1234,
		DownAt:      now.Add(-60 * time.Second),
		RecoveredAt: now,
		Downtime:    60 * time.Second,
	}

	payload := buildRecoveryPayload(ev)

	if payload["event"] != "recovered" {
		t.Errorf("expected event=recovered, got %v", payload["event"])
	}
	if payload["process"] != "redis" {
		t.Errorf("expected process=redis, got %v", payload["process"])
	}
	if payload["downtime_sec"] != "60.0" {
		t.Errorf("expected downtime_sec=60.0, got %v", payload["downtime_sec"])
	}
}
