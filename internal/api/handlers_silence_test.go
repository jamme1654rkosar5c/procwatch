package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andream16/procwatch/internal/api"
	"github.com/andream16/procwatch/internal/monitor"
)

func buildSilenceServer(t *testing.T) (*httptest.Server, *monitor.Silencer) {
	t.Helper()
	silencer := monitor.NewSilencer()
	srv := buildTestServer(t, api.WithSilencer(silencer))
	return srv, silencer
}

func TestHandleSilence_Post_Success(t *testing.T) {
	srv, silencer := buildSilenceServer(t)
	body, _ := json.Marshal(map[string]interface{}{"process_name": "nginx", "duration_seconds": 300})
	resp, err := http.Post(srv.URL+"/silence", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if !silencer.IsSilenced("nginx") {
		t.Fatal("expected nginx to be silenced")
	}
}

func TestHandleSilence_Post_BadRequest(t *testing.T) {
	srv, _ := buildSilenceServer(t)
	body, _ := json.Marshal(map[string]interface{}{"process_name": "", "duration_seconds": 0})
	resp, err := http.Post(srv.URL+"/silence", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleSilence_Delete(t *testing.T) {
	srv, silencer := buildSilenceServer(t)
	silencer.Silence("redis", time.Now().Add(10*time.Minute))
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/silence?process=redis", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if silencer.IsSilenced("redis") {
		t.Fatal("expected redis to no longer be silenced")
	}
}

func TestHandleSilence_Get(t *testing.T) {
	srv, silencer := buildSilenceServer(t)
	silencer.Silence("postgres", time.Now().Add(10*time.Minute))
	resp, err := http.Get(srv.URL + "/silence")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var rules []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}

func TestHandleSilence_MethodNotAllowed(t *testing.T) {
	srv, _ := buildSilenceServer(t)
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/silence", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
