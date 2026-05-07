package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/procwatch/internal/monitor"
)

// WithLabelStore wires the LabelStore into the Server and registers routes.
func WithLabelStore(ls *monitor.LabelStore) Option {
	return func(s *Server) {
		s.mux.HandleFunc("/api/labels", handleLabels(ls))
	}
}

// handleLabels serves GET/POST/DELETE on /api/labels?process=<name>[&key=<k>].
//
//	GET    /api/labels?process=nginx          → all labels for process
//	POST   /api/labels                        → body {"process":"nginx","key":"env","value":"prod"}
//	DELETE /api/labels?process=nginx&key=env  → remove single label
func handleLabels(ls *monitor.LabelStore) http.HandlerFunc {
	type setRequest struct {
		Process string `json:"process"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			process := strings.TrimSpace(r.URL.Query().Get("process"))
			if process == "" {
				http.Error(w, "process query param required", http.StatusBadRequest)
				return
			}
			labels := ls.All(process)
			if labels == nil {
				labels = map[string]string{}
			}
			writeJSON(w, labels)

		case http.MethodPost:
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if err := ls.Set(req.Process, req.Key, req.Value); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			process := strings.TrimSpace(r.URL.Query().Get("process"))
			key := strings.TrimSpace(r.URL.Query().Get("key"))
			if process == "" || key == "" {
				http.Error(w, "process and key query params required", http.StatusBadRequest)
				return
			}
			ls.Delete(process, key)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
