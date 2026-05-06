package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/briandowns/procwatch/internal/api"
	"github.com/briandowns/procwatch/internal/monitor"
)

func TestHandleRateLimit_Empty(t *testing.T) {
	srv := buildTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	api.HandleRateLimit(srv, w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	entries := body["entries"].([]interface{})
	if len(entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(entries))
	}
}

func TestHandleRateLimit_MethodNotAllowed(t *testing.T) {
	srv := buildTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ratelimit", nil)
	api.HandleRateLimit(srv, w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleRateLimit_WithCounts(t *testing.T) {
	rl := monitor.NewRateLimiter(time.Minute, 10)
	rl.Allow("nginx:down")
	rl.Allow("nginx:down")

	_ = rl // counts tracked internally; integration covered via Count()
	if c := rl.Count("nginx:down"); c != 2 {
		t.Fatalf("expected count 2, got %d", c)
	}
}
