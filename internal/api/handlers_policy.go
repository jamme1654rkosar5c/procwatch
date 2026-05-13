package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shawnflorida/procwatch/internal/monitor"
)

// WithPolicyStore registers the /api/policies routes on the given mux.
func WithPolicyStore(mux *http.ServeMux, store *monitor.PolicyStore) {
	mux.HandleFunc("/api/policies", makeHandlePolicy(store))
}

func makeHandlePolicy(store *monitor.PolicyStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlePolicyGet(w, r, store)
		case http.MethodPost:
			handlePolicyPost(w, r, store)
		case http.MethodDelete:
			handlePolicyDelete(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handlePolicyGet(w http.ResponseWriter, r *http.Request, store *monitor.PolicyStore) {
	process := strings.TrimSpace(r.URL.Query().Get("process"))
	if process != "" {
		p, ok := store.Get(process)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, p)
		return
	}
	writeJSON(w, store.All())
}

func handlePolicyPost(w http.ResponseWriter, r *http.Request, store *monitor.PolicyStore) {
	var p monitor.AlertPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := store.Set(p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handlePolicyDelete(w http.ResponseWriter, r *http.Request, store *monitor.PolicyStore) {
	process := strings.TrimSpace(r.URL.Query().Get("process"))
	if process == "" {
		http.Error(w, "process query param required", http.StatusBadRequest)
		return
	}
	store.Delete(process)
	w.WriteHeader(http.StatusNoContent)
}
