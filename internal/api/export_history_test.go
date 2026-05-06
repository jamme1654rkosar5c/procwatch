package api

import "github.com/user/procwatch/internal/monitor"

// History exposes the server's history store for testing.
func (s *Server) History() *monitor.History {
	return s.history
}
