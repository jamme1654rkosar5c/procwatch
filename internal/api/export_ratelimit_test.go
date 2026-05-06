package api

import "net/http"

// HandleRateLimit exposes the unexported handleRateLimit method for testing.
func HandleRateLimit(s *Server, w http.ResponseWriter, r *http.Request) {
	s.handleRateLimit(w, r)
}
