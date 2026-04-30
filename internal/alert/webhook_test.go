package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/procwatch/internal/alert"
)

func TestSender_Send_Success(t *testing.T) {
	var received alert.Payload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := alert.NewSender(srv.URL)
	p := alert.Payload{
		Process: "myapp",
		Event:   "crash",
		PID:     1234,
		Message: "process exited unexpectedly",
	}

	if err := s.Send(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Process != "myapp" {
		t.Errorf("expected process 'myapp', got %q", received.Process)
	}
	if received.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestSender_Send_NonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := alert.NewSender(srv.URL)
	err := s.Send(alert.Payload{Process: "svc", Event: "crash", Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestSender_Send_UnreachableURL(t *testing.T) {
	s := alert.NewSender("http://127.0.0.1:0/webhook")
	err := s.Send(alert.Payload{Process: "svc", Event: "crash"})
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}
