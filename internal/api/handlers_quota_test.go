package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/weezel/procwatch/internal/api"
	"github.com/weezel/procwatch/internal/monitor"
)

func buildQuotaServer(t *testing.T) (*httptest.Server, *monitor.QuotaStore) {
	t.Helper()
	qs := monitor.NewQuotaStore(time.Minute)
	mux := http.NewServeMux()
	api.WithQuotaStore(mux, qs)
	return httptest.NewServer(mux), qs
}

func TestHandleQuota_Get_Empty(t *testing.T) {
	srv, _ := buildQuotaServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/quota")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var out []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(out))
	}
}

func TestHandleQuota_Post_And_Get(t *testing.T) {
	srv, _ := buildQuotaServer(t)
	defer srv.Close()
	body := `{"process":"nginx","limit":10,"window":"1m"}`
	resp, err := http.Post(srv.URL+"/api/quota", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp2, _ := http.Get(srv.URL + "/api/quota")
	defer resp2.Body.Close()
	var entries []map[string]interface{}
	_ = json.NewDecoder(resp2.Body).Decode(&entries)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestHandleQuota_Post_BadJSON(t *testing.T) {
	srv, _ := buildQuotaServer(t)
	defer srv.Close()
	resp, _ := http.Post(srv.URL+"/api/quota", "application/json", bytes.NewBufferString("not-json"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleQuota_Post_InvalidLimit(t *testing.T) {
	srv, _ := buildQuotaServer(t)
	defer srv.Close()
	body := `{"process":"nginx","limit":0}`
	resp, _ := http.Post(srv.URL+"/api/quota", "application/json", bytes.NewBufferString(body))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleQuota_Delete_Reset(t *testing.T) {
	srv, qs := buildQuotaServer(t)
	defer srv.Close()
	_ = qs.SetLimit("nginx", 1)
	qs.Allow("nginx") // exhaust
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/quota?process=nginx", nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if !qs.Allow("nginx") {
		t.Fatal("expected quota reset to allow delivery")
	}
}

func TestHandleQuota_MethodNotAllowed(t *testing.T) {
	srv, _ := buildQuotaServer(t)
	defer srv.Close()
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/api/quota", nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
