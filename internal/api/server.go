// Package api provides an HTTP server exposing procwatch status endpoints.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/user/procwatch/internal/monitor"
)

// Server is a lightweight HTTP server that exposes process status and history.
type Server struct {
	httpServer *http.Server
	registry   *monitor.StatusRegistry
	history    *monitor.History
	summary    *monitor.SummaryBuilder
}

// NewServer creates a new API server bound to addr.
func NewServer(addr string, registry *monitor.StatusRegistry, history *monitor.History, summary *monitor.SummaryBuilder) *Server {
	s := &Server{
		registry: registry,
		history:  history,
		summary:  summary,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/history", s.handleHistory)
	mux.HandleFunc("/summary", s.handleSummary)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s
}

// Start begins listening and serving. It blocks until the server stops.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
