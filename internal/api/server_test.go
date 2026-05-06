package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/procwatch/internal/api"
	"github.com/user/procwatch/internal/monitor"
)

func buildTestServer(t *testing.T) *api.Server {
	t.Helper()
	status := monitor.NewStatusRegistry()
	history := monitor.NewHistory(100)
	summary := monitor.NewSummaryBuilder(status, history)
	return api.NewServer(":0", status, history, summary)
}

func TestServer_Shutdown(t *testing.T) {
	srv := buildTestServer(t)
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
}

func TestHandleHealthz(t *testing.T) {
	srv := buildTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHandleHealthz_MethodNotAllowed(t *testing.T) {
	srv := buildTestServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/healthz", nil)
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
