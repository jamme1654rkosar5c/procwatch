package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brucechapman/procwatch/internal/monitor"
)

type maintenanceRequest struct {
	Process  string `json:"process"`
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
}

type maintenanceResponse struct {
	Process string    `json:"process"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
	Reason  string    `json:"reason"`
}

func (s *Server) handleMaintenance(w http.ResponseWriter, r *http.Request) {
	if s.scheduler == nil {
		http.Error(w, "maintenance not configured", http.StatusNotImplemented)
		return
	}
	switch r.Method {
	case http.MethodGet:
		windows := s.scheduler.All()
		out := make([]maintenanceResponse, 0, len(windows))
		for _, w2 := range windows {
			out = append(out, maintenanceResponse{
				Process: w2.Process,
				Start:   w2.Start,
				End:     w2.End,
				Reason:  w2.Reason,
			})
		}
		writeJSON(w, out)
	case http.MethodPost:
		var req maintenanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Process == "" || req.Duration == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		d, err := time.ParseDuration(req.Duration)
		if err != nil || d <= 0 {
			http.Error(w, "invalid duration", http.StatusBadRequest)
			return
		}
		now := time.Now()
		win := monitor.MaintenanceWindow{
			Process: req.Process,
			Start:   now,
			End:     now.Add(d),
			Reason:  req.Reason,
		}
		s.scheduler.Add(win)
		writeJSON(w, maintenanceResponse{
			Process: win.Process,
			Start:   win.Start,
			End:     win.End,
			Reason:  win.Reason,
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// WithMaintenanceScheduler attaches a MaintenanceScheduler to the server.
func WithMaintenanceScheduler(sched *monitor.MaintenanceScheduler) func(*Server) {
	return func(s *Server) {
		s.scheduler = sched
	}
}
