package api

import (
	"encoding/json"
	"net/http"

	"github.com/captaincoordinates/procwatch/internal/monitor"
)

// WithRosterStore registers roster endpoints on the given mux.
func WithRosterStore(mux *http.ServeMux, r *monitor.RosterStore) {
	mux.HandleFunc("/api/roster", makeHandleRoster(r))
}

func makeHandleRoster(r *monitor.RosterStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			handleRosterGet(w, req, r)
		case http.MethodPost:
			handleRosterPost(w, req, r)
		case http.MethodDelete:
			handleRosterDelete(w, req, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleRosterGet(w http.ResponseWriter, _ *http.Request, r *monitor.RosterStore) {
	writeJSON(w, r.All())
}

type rosterPostRequest struct {
	Process string `json:"process"`
	Owner   string `json:"owner"`
}

func handleRosterPost(w http.ResponseWriter, req *http.Request, r *monitor.RosterStore) {
	var body rosterPostRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := r.Enroll(body.Process, body.Owner); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleRosterDelete(w http.ResponseWriter, req *http.Request, r *monitor.RosterStore) {
	process := req.URL.Query().Get("process")
	if process == "" {
		http.Error(w, "process query param required", http.StatusBadRequest)
		return
	}
	r.Remove(process)
	w.WriteHeader(http.StatusNoContent)
}
