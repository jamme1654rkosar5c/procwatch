package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ncolesummers/procwatch/internal/config"
)

// TestWatcher_tick_Recovery verifies that a process transitioning from down
// back to up triggers a recovery webhook call.
func TestWatcher_tick_Recovery(t *testing.T) {
	var mu sync.Mutex
	var received []map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
			mu.Lock()
			received = append(received, payload)
			mu.Unlock()
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	w := newTestWatcher(ts.URL)

	// Simulate process being down first.
	w.downSince["testproc"] = time.Now().Add(-10 * time.Second)
	w.alerted["testproc"] = true

	// Inject a live state so the next tick sees the process as up.
	w.lastState["testproc"] = &processState{Up: true, PID: 42}

	proc := config.Process{Name: "testproc"}
	w.handleRecovery(proc, w.lastState["testproc"])

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", len(received))
	}
	if received[0]["event"] != "recovered" {
		t.Errorf("expected event=recovered, got %v", received[0]["event"])
	}
	if _, exists := w.downSince["testproc"]; exists {
		t.Error("downSince should be cleared after recovery")
	}
	if w.alerted["testproc"] {
		t.Error("alerted flag should be cleared after recovery")
	}
}
