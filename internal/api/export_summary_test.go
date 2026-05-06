package api

import "net/http"

// HandleSummary exposes the summary handler for black-box testing.
func HandleSummary(s *Server, w http.ResponseWriter, r *http.Request) {
	s.handleSummary(w, r)
}
