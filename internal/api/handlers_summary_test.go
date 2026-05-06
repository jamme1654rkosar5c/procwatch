package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robtec/procwatch/internal/api"
	"github.com/robtec/procwatch/internal/monitor"
)

func TestHandleSummary_Empty(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/summary", nil)
	api.HandleSummary(srv, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["processes"]; !ok {
		t.Error("expected 'processes' key in summary response")
	}
}

func TestHandleSummary_WithProcesses(t *testing.T) {
	srv := buildTestServer(t)

	api.StatusRegistry(srv).Update("nginx", monitor.ProcessStatus{
		Name:      "nginx",
		Up:        true,
		CheckedAt: time.Now(),
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/summary", nil)
	api.HandleSummary(srv, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	procs, ok := body["processes"].([]interface{})
	if !ok {
		t.Fatal("expected 'processes' to be an array")
	}
	if len(procs) != 1 {
		t.Errorf("expected 1 process, got %d", len(procs))
	}
}

func TestHandleSummary_MethodNotAllowed(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/summary", nil)
	api.HandleSummary(srv, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
