package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sethgrid/procwatch/internal/monitor"
)

func buildProbeServer(ps *monitor.ProbeStore) *httptest.Server {
	srv, _ := NewServer(":0", WithProbeStore(ps))
	return httptest.NewServer(srv.mux)
}

func TestHandleProbe_Empty(t *testing.T) {
	ps := monitor.NewProbeStore()
	ts := buildProbeServer(ps)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/probes")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var out map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&out)
	if len(out) != 0 {
		t.Errorf("expected empty map, got %v", out)
	}
}

func TestHandleProbe_WithData(t *testing.T) {
	ps := monitor.NewProbeStore()
	ps.Record("nginx", true, "ok")
	ts := buildProbeServer(ps)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/probes")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var out map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&out)
	if _, ok := out["nginx"]; !ok {
		t.Error("expected nginx in response")
	}
}

func TestHandleProbe_FilterByProcess(t *testing.T) {
	ps := monitor.NewProbeStore()
	ps.Record("nginx", true, "ok")
	ts := buildProbeServer(ps)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/probes?process=nginx")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleProbe_FilterByProcess_NotFound(t *testing.T) {
	ps := monitor.NewProbeStore()
	ts := buildProbeServer(ps)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/api/v1/probes?process=missing")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleProbe_MethodNotAllowed(t *testing.T) {
	ps := monitor.NewProbeStore()
	ts := buildProbeServer(ps)
	defer ts.Close()

	resp, _ := http.Post(ts.URL+"/api/v1/probes", "application/json", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}
