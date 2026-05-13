package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robmorgan/procwatch/internal/monitor"
)

func buildCircuitServer(cb *monitor.CircuitBreaker) *httptest.Server {
	mux := http.NewServeMux()
	WithCircuitBreaker(mux, cb)
	return httptest.NewServer(mux)
}

func newCB() *monitor.CircuitBreaker {
	return monitor.NewCircuitBreaker(3, 10*time.Second)
}

func TestHandleCircuit_Get_Empty(t *testing.T) {
	srv := buildCircuitServer(newCB())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/circuit")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleCircuit_Get_ByProcess(t *testing.T) {
	cb := newCB()
	for i := 0; i < 3; i++ {
		cb.Record("nginx")
	}
	srv := buildCircuitServer(cb)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/circuit?process=nginx")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var out struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.State != "open" {
		t.Errorf("expected open, got %s", out.State)
	}
}

func TestHandleCircuit_Reset_Success(t *testing.T) {
	cb := newCB()
	for i := 0; i < 3; i++ {
		cb.Record("nginx")
	}
	srv := buildCircuitServer(cb)
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{"process": "nginx"})
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/circuit", bytes.NewReader(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if cb.State("nginx") != monitor.CircuitClosed {
		t.Error("expected circuit closed after reset")
	}
}

func TestHandleCircuit_Reset_BadRequest(t *testing.T) {
	srv := buildCircuitServer(newCB())
	defer srv.Close()

	body := bytes.NewBufferString(`{}`)
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/circuit", body)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleCircuit_MethodNotAllowed(t *testing.T) {
	srv := buildCircuitServer(newCB())
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/circuit", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
