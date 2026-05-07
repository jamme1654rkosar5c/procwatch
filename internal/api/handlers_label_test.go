package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/procwatch/internal/monitor"
)

func buildLabelServer(ls *monitor.LabelStore) *Server {
	return buildTestServer(WithLabelStore(ls))
}

func TestHandleLabel_Get_Empty(t *testing.T) {
	s := buildLabelServer(monitor.NewLabelStore())
	req := httptest.NewRequest(http.MethodGet, "/api/labels?process=nginx", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var m map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(m) != 0 {
		t.Fatalf("expected empty map, got %v", m)
	}
}

func TestHandleLabel_Get_MissingProcess(t *testing.T) {
	s := buildLabelServer(monitor.NewLabelStore())
	req := httptest.NewRequest(http.MethodGet, "/api/labels", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleLabel_Post_And_Get(t *testing.T) {
	ls := monitor.NewLabelStore()
	s := buildLabelServer(ls)

	body := `{"process":"nginx","key":"env","value":"production"}`
	req := httptest.NewRequest(http.MethodPost, "/api/labels", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/labels?process=nginx", nil)
	rr2 := httptest.NewRecorder()
	s.ServeHTTP(rr2, req2)
	var m map[string]string
	_ = json.Unmarshal(rr2.Body.Bytes(), &m)
	if m["env"] != "production" {
		t.Fatalf("expected 'production', got %q", m["env"])
	}
}

func TestHandleLabel_Post_BadJSON(t *testing.T) {
	s := buildLabelServer(monitor.NewLabelStore())
	req := httptest.NewRequest(http.MethodPost, "/api/labels", bytes.NewBufferString("not-json"))
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleLabel_Delete(t *testing.T) {
	ls := monitor.NewLabelStore()
	_ = ls.Set("nginx", "env", "prod")
	s := buildLabelServer(ls)

	req := httptest.NewRequest(http.MethodDelete, "/api/labels?process=nginx&key=env", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	_, ok := ls.Get("nginx", "env")
	if ok {
		t.Fatal("expected label to be deleted")
	}
}

func TestHandleLabel_Delete_MissingParams(t *testing.T) {
	s := buildLabelServer(monitor.NewLabelStore())
	req := httptest.NewRequest(http.MethodDelete, "/api/labels?process=nginx", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleLabel_MethodNotAllowed(t *testing.T) {
	s := buildLabelServer(monitor.NewLabelStore())
	req := httptest.NewRequest(http.MethodPatch, "/api/labels", nil)
	rr := httptest.NewRecorder()
	s.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
