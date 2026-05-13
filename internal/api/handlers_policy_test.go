package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shawnflorida/procwatch/internal/api"
	"github.com/shawnflorida/procwatch/internal/monitor"
)

func buildPolicyServer(t *testing.T) (*httptest.Server, *monitor.PolicyStore) {
	t.Helper()
	store := api.NewPolicyStoreExported()
	mux := http.NewServeMux()
	api.WithPolicyStoreExported(mux, store)
	return httptest.NewServer(mux), store
}

func TestHandlePolicy_Get_Empty(t *testing.T) {
	srv, _ := buildPolicyServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/policies")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlePolicy_Post_And_Get(t *testing.T) {
	srv, _ := buildPolicyServer(t)
	defer srv.Close()

	p := monitor.AlertPolicy{Process: "nginx", MinSeverity: "warn", Channels: []string{"slack"}}
	body, _ := json.Marshal(p)
	resp, err := http.Post(srv.URL+"/api/policies", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/api/policies?process=nginx")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	var got monitor.AlertPolicy
	if err := json.NewDecoder(resp2.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.MinSeverity != "warn" {
		t.Errorf("expected warn, got %s", got.MinSeverity)
	}
}

func TestHandlePolicy_Post_BadJSON(t *testing.T) {
	srv, _ := buildPolicyServer(t)
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/api/policies", "application/json", bytes.NewBufferString("notjson"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePolicy_Get_NotFound(t *testing.T) {
	srv, _ := buildPolicyServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/policies?process=unknown")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandlePolicy_Delete(t *testing.T) {
	srv, store := buildPolicyServer(t)
	defer srv.Close()
	_ = store.Set(monitor.AlertPolicy{Process: "redis", MinSeverity: "info"})

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/policies?process=redis", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	_, ok := store.Get("redis")
	if ok {
		t.Error("expected policy to be deleted")
	}
}

func TestHandlePolicy_MethodNotAllowed(t *testing.T) {
	srv, _ := buildPolicyServer(t)
	defer srv.Close()
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/policies", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
