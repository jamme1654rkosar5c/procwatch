package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/wander/procwatch/internal/monitor"
)

// WithNotifyRuleStore registers the /api/notify routes on the given mux.
func WithNotifyRuleStore(mux *http.ServeMux, store *monitor.NotifyRuleStore) {
	mux.HandleFunc("/api/notify", makeHandleNotify(store))
}

func makeHandleNotify(store *monitor.NotifyRuleStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNotifyGet(w, r, store)
		case http.MethodPost:
			handleNotifyPost(w, r, store)
		case http.MethodDelete:
			handleNotifyDelete(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleNotifyGet(w http.ResponseWriter, r *http.Request, store *monitor.NotifyRuleStore) {
	process := strings.TrimSpace(r.URL.Query().Get("process"))
	if process != "" {
		writeJSON(w, store.Get(process))
		return
	}
	writeJSON(w, store.All())
}

type notifyRuleRequest struct {
	Process   string   `json:"process"`
	EventType string   `json:"event_type"`
	Channels  []string `json:"channels"`
}

func handleNotifyPost(w http.ResponseWriter, r *http.Request, store *monitor.NotifyRuleStore) {
	var req notifyRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	err := store.Add(monitor.NotifyRule{
		Process:   req.Process,
		EventType: req.EventType,
		Channels:  req.Channels,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handleNotifyDelete(w http.ResponseWriter, r *http.Request, store *monitor.NotifyRuleStore) {
	process := strings.TrimSpace(r.URL.Query().Get("process"))
	if process == "" {
		http.Error(w, "process query param required", http.StatusBadRequest)
		return
	}
	store.Delete(process)
	w.WriteHeader(http.StatusNoContent)
}
