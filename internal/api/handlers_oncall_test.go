package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/procwatch/internal/api"
	"github.com/yourorg/procwatch/internal/monitor"
)

func buildOncallServer(t *testing.T) (*httptest.Server, *monitor.OnCallStore) {
	t.Helper()
	store := monitor.NewOnCallStore()
	srv := api.NewServer(":0", nil, nil)
	api.WithOnCallStoreExported(srv, store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts, store
}

func TestHandleOncall_Get_Empty(t *testing.T) {
	ts, _ := buildOncallServer(t)
	resp, err := http.Get(ts.URL + "/api/oncall")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got %d, want 200", resp.StatusCode)
	}
}

func TestHandleOncall_Post_And_Get(t *testing.T) {
	ts, _ := buildOncallServer(t)
	now := time.Now()
	e := monitor.OnCallEntry{
		Process:   "api",
		Owner:     "alice",
		Email:     "alice@example.com",
		StartsAt:  now.Add(-time.Hour),
		ExpiresAt: now.Add(time.Hour),
	}
	body, _ := json.Marshal(e)
	resp, err := http.Post(ts.URL+"/api/oncall", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("got %d, want 201", resp.StatusCode)
	}
	resp2, _ := http.Get(ts.URL + "/api/oncall?process=api")
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("active lookup got %d, want 200", resp2.StatusCode)
	}
}

func TestHandleOncall_Get_NotFound(t *testing.T) {
	ts, _ := buildOncallServer(t)
	resp, _ := http.Get(ts.URL + "/api/oncall?process=ghost")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("got %d, want 404", resp.StatusCode)
	}
}

func TestHandleOncall_Delete(t *testing.T) {
	ts, store := buildOncallServer(t)
	now := time.Now()
	_ = store.Set(monitor.OnCallEntry{Process: "api", Owner: "bob", Email: "b@x.com",
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)})
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/oncall?process=api", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("got %d, want 204", resp.StatusCode)
	}
}

func TestHandleOncall_MethodNotAllowed(t *testing.T) {
	ts, _ := buildOncallServer(t)
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/oncall", nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("got %d, want 405", resp.StatusCode)
	}
}
