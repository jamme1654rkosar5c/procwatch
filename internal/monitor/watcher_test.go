package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/alert"
	"github.com/user/procwatch/internal/config"
)

func newTestWatcher(t *testing.T, webhookURL string, processes []config.Process) *Watcher {
	t.Helper()
	cfg := &config.Config{
		WebhookURL:   webhookURL,
		PollInterval: 1,
		Processes:    processes,
	}
	sender := alert.NewSender(cfg.WebhookURL, 5*time.Second)
	w := NewWatcher(cfg, sender)
	return w
}

func TestWatcher_tick_ProcessDown(t *testing.T) {
	var received []AlertPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p AlertPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		received = append(received, p)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	w := newTestWatcher(t, ts.URL, []config.Process{
		{Name: "nonexistent-proc-xyz", ProcessName: "nonexistent-proc-xyz"},
	})
	w.tick()

	if len(received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(received))
	}
	if received[0].Event != EventDown {
		t.Errorf("expected event %q, got %q", EventDown, received[0].Event)
	}
}

func TestWatcher_tick_NoRepeatDownAlert(t *testing.T) {
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	w := newTestWatcher(t, ts.URL, []config.Process{
		{Name: "nonexistent-proc-xyz", ProcessName: "nonexistent-proc-xyz"},
	})

	w.tick()
	w.tick() // second tick should NOT re-alert since state hasn't changed

	if count != 1 {
		t.Errorf("expected 1 alert total, got %d", count)
	}
}

func TestBuildPayload_Down(t *testing.T) {
	p := buildPayload("myapp", EventDown, &ProcessState{Running: false})
	if p.Event != EventDown {
		t.Errorf("unexpected event: %s", p.Event)
	}
	if p.ProcessName != "myapp" {
		t.Errorf("unexpected process name: %s", p.ProcessName)
	}
	if p.Details == "" {
		t.Error("details should not be empty")
	}
}
