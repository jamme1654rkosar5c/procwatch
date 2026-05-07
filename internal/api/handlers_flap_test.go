package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/api"
	"github.com/user/procwatch/internal/monitor"
)

func buildFlapServer(t *testing.T, fd *monitor.FlapDetector) *httptest.Server {
	t.Helper()
	srv, err := api.NewServer(":0", api.WithFlapDetector(fd))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return httptest.NewServer(srv.Handler())
}

func TestHandleFlap_NotFlapping(t *testing.T) {
	fd := monitor.NewFlapDetector(3, 10*time.Second)
	ts := buildFlapServer(t, fd)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/flap?process=nginx")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entry api.FlapEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if entry.IsFlapping {
		t.Fatal("expected not flapping")
	}
	if entry.Process != "nginx" {
		t.Fatalf("expected process nginx, got %s", entry.Process)
	}
}

func TestHandleFlap_IsFlapping(t *testing.T) {
	fd := monitor.NewFlapDetector(2, 10*time.Second)
	fd.Record("redis")
	fd.Record("redis")
	fd.Record("redis")
	ts := buildFlapServer(t, fd)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/flap?process=redis")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	var entry api.FlapEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !entry.IsFlapping {
		t.Fatal("expected flapping")
	}
}

func TestHandleFlap_Reset(t *testing.T) {
	fd := monitor.NewFlapDetector(2, 10*time.Second)
	fd.Record("redis")
	fd.Record("redis")
	fd.Record("redis")
	ts := buildFlapServer(t, fd)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/flap?process=redis", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if fd.IsFlapping("redis") {
		t.Fatal("expected not flapping after reset")
	}
}

func TestHandleFlap_MissingProcess(t *testing.T) {
	fd := monitor.NewFlapDetector(3, 10*time.Second)
	ts := buildFlapServer(t, fd)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/flap")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleFlap_MethodNotAllowed(t *testing.T) {
	fd := monitor.NewFlapDetector(3, 10*time.Second)
	ts := buildFlapServer(t, fd)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/v1/flap", "application/json", nil)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
