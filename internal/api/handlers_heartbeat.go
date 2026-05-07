package api

import (
	"net/http"
	"time"

	"github.com/slimloans/procwatch/internal/monitor"
)

type heartbeatResponse struct {
	Process  string    `json:"process"`
	LastSeen time.Time `json:"last_seen"`
	Stale    bool      `json:"stale"`
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.heartbeat == nil {
		writeJSON(w, http.StatusOK, []heartbeatResponse{})
		return
	}

	all := s.heartbeat.All()
	results := make([]heartbeatResponse, 0, len(all))
	for name, ts := range all {
		results = append(results, heartbeatResponse{
			Process:  name,
			LastSeen: ts,
			Stale:    s.heartbeat.IsStale(name),
		})
	}

	writeJSON(w, http.StatusOK, results)
}

// WithHeartbeatTracker attaches a HeartbeatTracker to the server.
func WithHeartbeatTracker(ht *monitor.HeartbeatTracker) func(*Server) {
	return func(s *Server) {
		s.heartbeat = ht
	}
}
