package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/monitor"
)

func buildRetentionServer(rm *monitor.RetentionManager) *Server {
	return buildTestServer(WithRetentionManager(rm))
}

func TestHandleRetentionPrune_Success(t *testing.T) {
	now := time.Now()
	h := monitor.NewHistory(100)
	h.Record(monitor.AlertEvent{Process: "svc", Kind: "down", Timestamp: now.Add(-2 * time.Hour)})
	h.Record(monitor.AlertEvent{Process: "svc", Kind: "down", Timestamp: now.Add(-10 * time.Minute)})

	rm := monitor.NewRetentionManager(h, monitor.RetentionPolicy{MaxAge: 24 * time.Hour})
	srv := buildRetentionServer(rm)

	body := bytes.NewBufferString(`{"max_age_seconds":3600}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/retention/prune", body)
	rw := httptest.NewRecorder()
	srv.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var resp struct {
		Pruned int    `json:"pruned"`
		At     string `json:"at"`
	}
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Pruned != 1 {
		t.Errorf("expected 1 pruned, got %d", resp.Pruned)
	}
	if resp.At == "" {
		t.Error("expected non-empty 'at' field")
	}
}

func TestHandleRetentionPrune_BadRequest_ZeroAge(t *testing.T) {
	rm := monitor.NewRetentionManager(monitor.NewHistory(10), monitor.RetentionPolicy{MaxAge: time.Hour})
	srv := buildRetentionServer(rm)

	body := bytes.NewBufferString(`{"max_age_seconds":0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/retention/prune", body)
	rw := httptest.NewRecorder()
	srv.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleRetentionPrune_MethodNotAllowed(t *testing.T) {
	rm := monitor.NewRetentionManager(monitor.NewHistory(10), monitor.RetentionPolicy{MaxAge: time.Hour})
	srv := buildRetentionServer(rm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/retention/prune", nil)
	rw := httptest.NewRecorder()
	srv.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
