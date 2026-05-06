package api

import (
	"net/http"

	"github.com/briandowns/procwatch/internal/monitor"
)

// RateLimitStater is satisfied by monitor.RateLimiter.
type RateLimitStater interface {
	Count(key string) int
}

type rateLimitEntry struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type rateLimitSummary struct {
	Entries []rateLimitEntry `json:"entries"`
}

// handleRateLimit returns current in-window alert counts for all tracked keys.
func (s *Server) handleRateLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses := s.statusRegistry.All()
	var entries []rateLimitEntry

	for _, st := range statuses {
		for _, evType := range []string{"down", "cpu", "mem"} {
			key := st.Name + ":" + evType
			if rl, ok := s.rateLimiter.(*monitor.RateLimiter); ok {
				c := rl.Count(key)
				if c > 0 {
					entries = append(entries, rateLimitEntry{Key: key, Count: c})
				}
			}
		}
	}

	if entries == nil {
		entries = []rateLimitEntry{}
	}

	writeJSON(w, http.StatusOK, rateLimitSummary{Entries: entries})
}
