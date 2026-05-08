package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/procwatch/internal/monitor"
)

// WithRetentionManager registers the retention management endpoint.
func WithRetentionManager(rm *monitor.RetentionManager) Option {
	return func(s *Server) {
		s.mux.HandleFunc("/api/v1/retention/prune", handleRetentionPrune(rm))
	}
}

type pruneResponse struct {
	Pruned int    `json:"pruned"`
	At     string `json:"at"`
}

type pruneRequest struct {
	MaxAgeSecs int `json:"max_age_seconds"`
}

func handleRetentionPrune(rm *monitor.RetentionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req pruneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.MaxAgeSecs <= 0 {
			http.Error(w, "invalid request: max_age_seconds must be > 0", http.StatusBadRequest)
			return
		}

		now := time.Now()
		origPolicy := rm.Policy()
		rm.SetPolicy(monitor.RetentionPolicy{MaxAge: time.Duration(req.MaxAgeSecs) * time.Second})
		pruned := rm.Prune(now)
		rm.SetPolicy(origPolicy)

		writeJSON(w, pruneResponse{
			Pruned: pruned,
			At:     now.UTC().Format(time.RFC3339),
		})
	}
}
