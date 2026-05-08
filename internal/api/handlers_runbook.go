package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/madalinpopa/procwatch/internal/monitor"
)

// WithRunbookStore registers the /api/runbooks routes on srv.
func WithRunbookStore(srv *Server, store *monitor.RunbookStore) {
	srv.mux.HandleFunc("/api/runbooks", makeHandleRunbooks(store))
}

func makeHandleRunbooks(store *monitor.RunbookStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := strings.TrimSpace(r.URL.Query().Get("process"))

		switch r.Method {
		case http.MethodGet:
			if process != "" {
				e, ok := store.Get(process)
				if !ok {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				writeJSON(w, e)
				return
			}
			writeJSON(w, store.All())

		case http.MethodPost:
			var body struct {
				Process string `json:"process"`
				URL     string `json:"url"`
				Note    string `json:"note"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if err := store.Set(body.Process, body.URL, body.Note); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if process == "" {
				http.Error(w, "process query param required", http.StatusBadRequest)
				return
			}
			store.Delete(process)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
