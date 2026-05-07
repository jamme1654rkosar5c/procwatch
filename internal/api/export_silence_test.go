package api

import "github.com/andream16/procwatch/internal/monitor"

// WithSilencer is an option exposed for testing that injects a Silencer into the Server.
func WithSilencer(s *monitor.Silencer) func(*Server) {
	return func(srv *Server) {
		srv.silencer = s
	}
}
