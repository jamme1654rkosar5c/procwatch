package api

import (
	"net/http"

	"github.com/robcowart/procwatch/internal/monitor"
)

// WithCorrelationTracker registers the /api/correlations endpoint.
func WithCorrelationTracker(ct *monitor.CorrelationTracker) Option {
	return func(s *Server) {
		s.mux.HandleFunc("/api/correlations", makeHandleCorrelations(ct))
	}
}

func makeHandleCorrelations(ct *monitor.CorrelationTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		entries := ct.All()

		type responseEntry struct {
			ProcessA  string `json:"process_a"`
			ProcessB  string `json:"process_b"`
			EventType string `json:"event_type"`
			Count     int    `json:"count"`
			LastSeen  string `json:"last_seen"`
		}

		out := make([]responseEntry, 0, len(entries))
		for _, e := range entries {
			out = append(out, responseEntry{
				ProcessA:  e.ProcessA,
				ProcessB:  e.ProcessB,
				EventType: e.EventType,
				Count:     e.Count,
				LastSeen:  e.LastSeen.UTC().Format("2006-01-02T15:04:05Z"),
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"correlations": out,
		})
	}
}
