package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/yourorg/procwatch/internal/monitor"
)

// WithOnCallStore registers the /api/oncall routes on srv.
func WithOnCallStore(srv *Server, store *monitor.OnCallStore) {
	srv.mux.HandleFunc("/api/oncall", makeHandleOncall(store))
}

func makeHandleOncall(store *monitor.OnCallStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		process := strings.TrimSpace(r.URL.Query().Get("process"))
		switch r.Method {
		case http.MethodGet:
			handleOncallGet(w, store, process)
		case http.MethodPost:
			handleOncallPost(w, r, store)
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

func handleOncallGet(w http.ResponseWriter, store *monitor.OnCallStore, process string) {
	if process != "" {
		e, ok := store.Active(process, time.Now())
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, e)
		return
	}
	writeJSON(w, store.All())
}

func handleOncallPost(w http.ResponseWriter, r *http.Request, store *monitor.OnCallStore) {
	var e monitor.OnCallEntry
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := store.Set(e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
