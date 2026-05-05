package monitor

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/procwatch/internal/alert"
	"github.com/yourusername/procwatch/internal/config"
)

// TestWatcher_tick_ThrottlesRepeatedCPUAlerts verifies that the watcher does
// not fire repeated threshold alerts within the cooldown window.
func TestWatcher_tick_ThrottlesRepeatedCPUAlerts(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cooldown := 1 * time.Hour // large cooldown so second tick is blocked
	th := NewThrottler(cooldown)

	w := newTestWatcher(t, server.URL)
	// Inject throttler into watcher
	w.throttler = th

	// Configure a CPU threshold that will always fire (0%)
	w.cfg.Processes[0].MaxCPUPercent = 0.0
	w.cfg.Processes[0].MaxMemMB = 0

	// First tick — alert should fire
	w.tick()
	// Second tick — alert should be suppressed by throttler
	w.tick()

	time.Sleep(50 * time.Millisecond)
	if got := callCount.Load(); got > 1 {
		t.Errorf("expected at most 1 webhook call, got %d", got)
	}
}

// newTestWatcherWithSender is a helper used only in this file.
func newTestWatcherWithSender(t *testing.T, webhookURL string) *Watcher {
	t.Helper()
	return &Watcher{
		cfg: &config.Config{
			WebhookURL:   webhookURL,
			PollInterval: time.Second,
			Processes: []config.Process{
				{Name: "testproc", PIDFile: "/tmp/nonexistent.pid"},
			},
		},
		sender:    alert.NewSender(webhookURL, 3*time.Second),
		downSeen:  make(map[string]bool),
		throttler: NewThrottler(5 * time.Minute),
	}
}
