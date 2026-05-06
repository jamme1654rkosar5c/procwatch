package api

import (
	"net/http"
	"time"
)

// metricsProcess is the per-process payload returned by GET /metrics.
type metricsProcess struct {
	Name      string    `json:"name"`
	Up        bool      `json:"up"`
	PID       int       `json:"pid"`
	CPUPct    float64   `json:"cpu_pct"`
	MemRSSMB  float64   `json:"mem_rss_mb"`
	CheckedAt time.Time `json:"checked_at"`
}

// metricsResponse is the top-level envelope for GET /metrics.
type metricsResponse struct {
	Processes []metricsProcess `json:"processes"`
}

// handleMetrics returns live resource metrics for every tracked process.
//
// GET /metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	all := s.status.All()

	procs := make([]metricsProcess, 0, len(all))
	for _, st := range all {
		procs = append(procs, metricsProcess{
			Name:      st.Name,
			Up:        st.Up,
			PID:       st.PID,
			CPUPct:    st.CPUPct,
			MemRSSMB:  st.MemRSSMB,
			CheckedAt: st.CheckedAt,
		})
	}

	writeJSON(w, http.StatusOK, metricsResponse{Processes: procs})
}
