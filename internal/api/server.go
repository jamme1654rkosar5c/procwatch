package api

import (
	"context"
	"net/http"

	"github.com/user/procwatch/internal/monitor"
)

// Server is the HTTP API server for procwatch.
type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	status     *monitor.StatusRegistry
	history    *monitor.History
	summary    *monitor.SummaryBuilder
}

// NewServer constructs a Server bound to addr.
func NewServer(
	addr string,
	status *monitor.StatusRegistry,
	history *monitor.History,
	summary *monitor.SummaryBuilder,
) *Server {
	s := &Server{
		mux:     http.NewServeMux(),
		status:  status,
		history: history,
		summary: summary,
	}
	s.httpServer = &http.Server{Addr: addr, Handler: s.mux}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/history", s.handleHistory)
}

// Handler returns the underlying http.Handler (for testing).
func (s *Server) Handler() http.Handler { return s.mux }

// Start begins listening in a goroutine.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
