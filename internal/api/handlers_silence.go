package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andream16/procwatch/internal/monitor"
)

type silenceRequest struct {
	ProcessName string `json:"process_name"`
	DurationSec int    `json:"duration_seconds"`
}

type silenceEntry struct {
	ProcessName string    `json:"process_name"`
	Until       time.Time `json:"until"`
}

func (s *Server) handleSilence(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req silenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.ProcessName == "" || req.DurationSec <= 0 {
			http.Error(w, "process_name and duration_seconds > 0 required", http.StatusBadRequest)
			return
		}
		until := time.Now().Add(time.Duration(req.DurationSec) * time.Second)
		s.silencer.Silence(req.ProcessName, until)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		process := r.URL.Query().Get("process")
		if process == "" {
			http.Error(w, "process query param required", http.StatusBadRequest)
			return
		}
		s.silencer.Lift(process)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		rules := s.silencer.All()
		out := make([]silenceEntry, 0, len(rules))
		for _, r := range rules {
			out = append(out, silenceEntry{ProcessName: r.ProcessName, Until: r.Until})
		}
		writeJSON(w, http.StatusOK, out)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ensure Silencer interface used by Server
var _ silencerIface = (*monitor.Silencer)(nil)

type silencerIface interface {
	Silence(processName string, until time.Time)
	Lift(processName string)
	All() []monitor.SilenceRule
	IsSilenced(processName string) bool
}
