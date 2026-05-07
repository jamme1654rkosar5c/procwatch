package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brucechapman/procwatch/internal/api"
	"github.com/brucechapman/procwatch/internal/monitor"
)

func buildMaintenanceServer(t *testing.T) (*httptest.Server, *monitor.MaintenanceScheduler) {
	t.Helper()
	sched := monitor.NewMaintenanceScheduler()
	srv := buildTestServer(t, api.WithMaintenanceScheduler(sched))
	return srv, sched
}

func TestHandleMaintenance_Get_Empty(t *testing.T) {
	srv, _ := buildMaintenanceServer(t)
	resp, err := http.Get(srv.URL + "/maintenance")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var out []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&out)
	if len(out) != 0 {
		t.Fatalf("expected empty list, got %d items", len(out))
	}
}

func TestHandleMaintenance_Post_Success(t *testing.T) {
	srv, sched := buildMaintenanceServer(t)
	body := `{"process":"nginx","duration":"30m","reason":"deploy"}`
	resp, err := http.Post(srv.URL+"/maintenance", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !sched.IsUnderMaintenance("nginx") {
		t.Fatal("expected nginx to be under maintenance")
	}
}

func TestHandleMaintenance_Post_BadRequest(t *testing.T) {
	srv, _ := buildMaintenanceServer(t)
	body := `{"process":"nginx"}` // missing duration
	resp, err := http.Post(srv.URL+"/maintenance", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleMaintenance_Post_InvalidDuration(t *testing.T) {
	srv, _ := buildMaintenanceServer(t)
	body := `{"process":"nginx","duration":"notaduration"}`
	resp, err := http.Post(srv.URL+"/maintenance", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleMaintenance_Get_WithWindow(t *testing.T) {
	srv, sched := buildMaintenanceServer(t)
	now := time.Now()
	sched.Add(monitor.MaintenanceWindow{
		Process: "redis",
		Start:   now,
		End:     now.Add(time.Hour),
		Reason:  "upgrade",
	})
	resp, err := http.Get(srv.URL + "/maintenance")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&out)
	if len(out) != 1 {
		t.Fatalf("expected 1 window, got %d", len(out))
	}
	if out[0]["process"] != "redis" {
		t.Fatalf("expected process redis, got %v", out[0]["process"])
	}
}

func TestHandleMaintenance_MethodNotAllowed(t *testing.T) {
	srv, _ := buildMaintenanceServer(t)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/maintenance", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
