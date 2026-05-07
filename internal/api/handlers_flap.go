package api

import (
	"net/http"

	"github.com/user/procwatch/internal/monitor"
)

type flapHandler struct {
	detector *monitor.FlapDetector
}

// FlapEntry is the JSON representation of a process flap status.
type FlapEntry struct {
	Process    string `json:"process"`
	IsFlapping bool   `json:"is_flapping"`
}

func (h *flapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	process := r.URL.Query().Get("process")

	if r.Method == http.MethodDelete {
		if process == "" {
			http.Error(w, "process query param required", http.StatusBadRequest)
			return
		}
		h.detector.Reset(process)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET: return flap status for a specific process.
	if process == "" {
		http.Error(w, "process query param required", http.StatusBadRequest)
		return
	}

	entry := FlapEntry{
		Process:    process,
		IsFlapping: h.detector.IsFlapping(process),
	}
	writeJSON(w, http.StatusOK, entry)
}

// WithFlapDetector registers the /api/v1/flap endpoint on the server.
func WithFlapDetector(fd *monitor.FlapDetector) ServerOption {
	return func(s *Server) {
		s.mux.Handle("/api/v1/flap", &flapHandler{detector: fd})
	}
}
