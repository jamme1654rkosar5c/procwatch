package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrebq/procwatch/internal/monitor"
)

// WithAlertBudget registers the /api/budget routes on the given mux.
func WithAlertBudget(mux *http.ServeMux, b *monitor.AlertBudget) {
	mux.HandleFunc("/api/budget", makeHandleBudget(b))
}

func makeHandleBudget(b *monitor.AlertBudget) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleBudgetGet(w, r, b)
		case http.MethodDelete:
			handleBudgetReset(w, r, b)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

type budgetResponse struct {
	Process   string `json:"process"`
	Remaining int    `json:"remaining"`
	CheckedAt string `json:"checked_at"`
}

func handleBudgetGet(w http.ResponseWriter, r *http.Request, b *monitor.AlertBudget) {
	process := r.URL.Query().Get("process")
	if process == "" {
		http.Error(w, "missing process query parameter", http.StatusBadRequest)
		return
	}
	resp := budgetResponse{
		Process:   process,
		Remaining: b.Remaining(process),
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleBudgetReset(w http.ResponseWriter, r *http.Request, b *monitor.AlertBudget) {
	var body struct {
		Process string `json:"process"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Process == "" {
		http.Error(w, "invalid or missing process in body", http.StatusBadRequest)
		return
	}
	if err := b.Reset(body.Process); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
