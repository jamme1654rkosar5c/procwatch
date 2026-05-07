package api

import (
	"encoding/json"
	"net/http"

	"github.com/andrebq/procwatch/internal/monitor"
)

// WithAnnotationStore wires the annotation endpoints into the server mux.
func WithAnnotationStore(store *monitor.AnnotationStore) Option {
	return func(s *Server) {
		s.mux.HandleFunc("/api/annotations", func(w http.ResponseWriter, r *http.Request) {
			handleAnnotations(w, r, store)
		})
	}
}

type annotationRequest struct {
	Text string `json:"text"`
}

func handleAnnotations(w http.ResponseWriter, r *http.Request, store *monitor.AnnotationStore) {
	process := r.URL.Query().Get("process")

	switch r.Method {
	case http.MethodGet:
		if process == "" {
			writeJSON(w, http.StatusOK, store.All())
			return
		}
		a, ok := store.Get(process)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, a)

	case http.MethodPost:
		if process == "" {
			http.Error(w, "process query param required", http.StatusBadRequest)
			return
		}
		var req annotationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		store.Set(process, req.Text)
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	case http.MethodDelete:
		if process == "" {
			http.Error(w, "process query param required", http.StatusBadRequest)
			return
		}
		store.Delete(process)
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
