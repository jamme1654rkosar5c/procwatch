package api

import (
	"net/http"

	"github.com/sethgrid/procwatch/internal/monitor"
)

// WithProbeStore registers the liveness probe endpoint on the server.
func WithProbeStore(ps *monitor.ProbeStore) Option {
	return func(s *Server) {
		s.mux.HandleFunc("/api/v1/probes", makeHandleProbes(ps))
	}
}

func makeHandleProbes(ps *monitor.ProbeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		process := r.URL.Query().Get("process")
		if process != "" {
			result, ok := ps.Get(process)
			if !ok {
				http.Error(w, "process not found", http.StatusNotFound)
				return
			}
			writeJSON(w, result)
			return
		}

		writeJSON(w, ps.All())
	}
}
