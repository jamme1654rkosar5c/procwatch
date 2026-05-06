package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/monitor"
)

func TestHandleHistory_Empty(t *testing.T) {
	srv := buildTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["count"].(float64) != 0 {
		t.Errorf("expected count 0, got %v", body["count"])
	}
}

func TestHandleHistory_WithEvents(t *testing.T) {
	srv := buildTestServer(t)

	srv.History().Record(monitor.HistoryEntry{
		Process:   "nginx",
		EventType: "down",
		Timestamp: time.Now(),
	})
	srv.History().Record(monitor.HistoryEntry{
		Process:   "redis",
		EventType: "cpu_threshold",
		Timestamp: time.Now(),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	srv.Handler().ServeHTTP(rec, req)

	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body)
	if body["count"].(float64) != 2 {
		t.Errorf("expected count 2, got %v", body["count"])
	}
}

func TestHandleHistory_FilterByProcess(t *testing.T) {
	srv := buildTestServer(t)

	srv.History().Record(monitor.HistoryEntry{Process: "nginx", EventType: "down", Timestamp: time.Now()})
	srv.History().Record(monitor.HistoryEntry{Process: "redis", EventType: "down", Timestamp: time.Now()})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history?process=nginx", nil)
	srv.Handler().ServeHTTP(rec, req)

	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body)
	if body["count"].(float64) != 1 {
		t.Errorf("expected count 1 for nginx filter, got %v", body["count"])
	}
}

func TestHandleHistory_MethodNotAllowed(t *testing.T) {
	srv := buildTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/history", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
