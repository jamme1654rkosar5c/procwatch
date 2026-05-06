package api

import (
	"net/http"
)

// handleSummary returns a high-level summary of all watched processes,
// including current status and alert counts from history.
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary := s.summaryBuilder.Build()
	writeJSON(w, http.StatusOK, summary)
}
