package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/madalinpopa/procwatch/internal/api"
	"github.com/madalinpopa/procwatch/internal/monitor"
)

func buildRunbookServer(t *testing.T) (*httptest.Server, *monitor.RunbookStore) {
	t.Helper()
	store := monitor.NewRunbookStore()
	srv := buildTestServer(t)
	api.WithRunbookStoreExported(srv, store)
	return httptest.NewServer(srv.Handler()), store
}

func TestHandleRunbook_Get_Empty(t *testing.T) {
	ts, _ := buildRunbookServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/runbooks")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandleRunbook_Post_And_Get(t *testing.T) {
	ts, _ := buildRunbookServer(t)
	defer ts.Close()

	body, _ := json.Marshal(map[string]string{
		"process": "nginx",
		"url":     "https://wiki.example.com/nginx",
		"note":    "restart procedure",
	})
	resp, err := http.Post(ts.URL+"/api/runbooks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	resp2, _ := http.Get(ts.URL + "/api/runbooks?process=nginx")
	defer resp2.Body.Close()
	var entry map[string]interface{}
	_ = json.NewDecoder(resp2.Body).Decode(&entry)
	if entry["url"] != "https://wiki.example.com/nginx" {
		t.Errorf("unexpected url: %v", entry["url"])
	}
}

func TestHandleRunbook_Get_NotFound(t *testing.T) {
	ts, _ := buildRunbookServer(t)
	defer ts.Close()
	resp, _ := http.Get(ts.URL + "/api/runbooks?process=missing")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleRunbook_Delete(t *testing.T) {
	ts, store := buildRunbookServer(t)
	defer ts.Close()
	_ = store.Set("nginx", "https://example.com", "")

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/runbooks?process=nginx", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	_, ok := store.Get("nginx")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestHandleRunbook_MethodNotAllowed(t *testing.T) {
	ts, _ := buildRunbookServer(t)
	defer ts.Close()
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/runbooks", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
