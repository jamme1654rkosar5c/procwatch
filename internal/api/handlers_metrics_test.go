package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/monitor"
)

func TestHandleMetrics_Empty(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["processes"]; !ok {
		t.Error("expected 'processes' key in response")
	}
}

func TestHandleMetrics_WithData(t *testing.T) {
	srv := buildTestServer(t)

	srv.StatusRegistry().Update("nginx", monitor.ProcessStatus{
		Name:    "nginx",
		Up:      true,
		PID:     1234,
		CPUPct:  12.5,
		MemRSSMB: 64.0,
		CheckedAt: time.Now(),
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body struct {
		Processes []struct {
			Name     string  `json:"name"`
			Up       bool    `json:"up"`
			PID      int     `json:"pid"`
			CPUPct   float64 `json:"cpu_pct"`
			MemRSSMB float64 `json:"mem_rss_mb"`
		} `json:"processes"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(body.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(body.Processes))
	}

	p := body.Processes[0]
	if p.Name != "nginx" {
		t.Errorf("expected name 'nginx', got %q", p.Name)
	}
	if !p.Up {
		t.Error("expected process to be up")
	}
	if p.CPUPct != 12.5 {
		t.Errorf("expected cpu_pct 12.5, got %f", p.CPUPct)
	}
	if p.MemRSSMB != 64.0 {
		t.Errorf("expected mem_rss_mb 64.0, got %f", p.MemRSSMB)
	}
}

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	srv := buildTestServer(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
