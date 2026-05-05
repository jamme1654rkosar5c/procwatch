package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arnavsurve/procwatch/internal/alert"
	"github.com/arnavsurve/procwatch/internal/config"
)

// TestWatcher_tick_RecordsDownEvent verifies that a down event is written to
// the History attached to a Watcher when a process cannot be found.
func TestWatcher_tick_RecordsDownEvent(t *testing.T) {
	received := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := config.Process{
		Name:       "ghost-proc",
		WebhookURL: ts.URL,
	}

	sender := alert.NewSender(5)
	hist := NewHistory(50)
	w := newTestWatcher(t, cfg, sender)
	w.history = hist

	w.tick()

	<-received // wait for webhook delivery

	records := hist.ForProcess("ghost-proc")
	if len(records) == 0 {
		t.Fatal("expected at least one history record for ghost-proc, got none")
	}
	if records[0].EventType != "down" {
		t.Errorf("expected event type 'down', got %q", records[0].EventType)
	}
}
