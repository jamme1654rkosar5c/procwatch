package api

import (
	"net/http"
)

// handleSummary returns a high-level summary of all watched processes,
// including current status and alert counts from history.
// Responds to GET requests only; returns 405 for other methods.
// The response is a JSON object produced by the configured summaryBuilder.
func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary, err := s.summaryBuilder.Build()
	if err != nil {
		http.Error(w, "failed to build summary", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
