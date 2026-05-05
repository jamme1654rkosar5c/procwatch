package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/api"
	"github.com/user/procwatch/internal/monitor"
)

func buildTestServer(t *testing.T) *api.Server {
	t.Helper()
	reg := monitor.NewStatusRegistry()
	hist := monitor.NewHistory(50)
	sum := monitor.NewSummaryBuilder()
	return api.NewServer(":0", reg, hist, sum)
}

func TestServer_Shutdown(t *testing.T) {
	s := buildTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
}

func TestHandleHealthz(t *testing.T) {
	s := buildTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTPForTest(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestHandleHealthz_MethodNotAllowed(t *testing.T) {
	s := buildTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTPForTest(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
