package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nicholasgasior/procwatch/internal/api"
	"github.com/nicholasgasior/procwatch/internal/monitor"
)

func TestHandleStatus_Empty(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	api.ServeHTTP(srv, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp api.SummaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("expected total=0, got %d", resp.Total)
	}
}

func TestHandleStatus_WithProcesses(t *testing.T) {
	srv := buildTestServer(t)

	now := time.Now()
	api.RegistryUpdate(srv, "nginx", monitor.ProcessStatus{
		Up: true, PID: 42, CPUPct: 1.5, MemBytes: 1024, CheckedAt: now,
	})
	api.RegistryUpdate(srv, "redis", monitor.ProcessStatus{
		Up: false, CheckedAt: now,
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	api.ServeHTTP(srv, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp api.SummaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected total=2, got %d", resp.Total)
	}
	if resp.Up != 1 {
		t.Errorf("expected up=1, got %d", resp.Up)
	}
	if resp.Down != 1 {
		t.Errorf("expected down=1, got %d", resp.Down)
	}
}

func TestHandleStatus_MethodNotAllowed(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	api.ServeHTTP(srv, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHandleStatus_ContentTypeJSON(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	api.ServeHTTP(srv, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
