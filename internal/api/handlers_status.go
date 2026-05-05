package api

import (
	"net/http"
	"time"
)

// ProcessStatusResponse is the JSON shape returned by GET /status.
type ProcessStatusResponse struct {
	Name      string    `json:"name"`
	Up        bool      `json:"up"`
	PID       int       `json:"pid,omitempty"`
	CPUPct    float64   `json:"cpu_pct,omitempty"`
	MemBytes  uint64    `json:"mem_bytes,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// SummaryResponse wraps the per-process statuses returned by GET /status.
type SummaryResponse struct {
	Processes []ProcessStatusResponse `json:"processes"`
	Total     int                     `json:"total"`
	Up        int                     `json:"up"`
	Down      int                     `json:"down"`
}

// handleStatus serves a snapshot of every watched process.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	all := s.registry.All()
	resp := SummaryResponse{
		Processes: make([]ProcessStatusResponse, 0, len(all)),
	}

	for name, st := range all {
		psr := ProcessStatusResponse{
			Name:      name,
			Up:        st.Up,
			PID:       st.PID,
			CPUPct:    st.CPUPct,
			MemBytes:  st.MemBytes,
			CheckedAt: st.CheckedAt,
		}
		resp.Processes = append(resp.Processes, psr)
		resp.Total++
		if st.Up {
			resp.Up++
		} else {
			resp.Down++
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
