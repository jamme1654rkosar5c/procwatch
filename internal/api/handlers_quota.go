package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/weezel/procwatch/internal/monitor"
)

// WithQuotaStore registers the /api/quota routes on the given mux.
func WithQuotaStore(mux *http.ServeMux, qs *monitor.QuotaStore) {
	mux.HandleFunc("/api/quota", makeHandleQuota(qs))
}

func makeHandleQuota(qs *monitor.QuotaStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleQuotaGet(w, r, qs)
		case http.MethodPost:
			handleQuotaPost(w, r, qs)
		case http.MethodDelete:
			handleQuotaReset(w, r, qs)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleQuotaGet(w http.ResponseWriter, _ *http.Request, qs *monitor.QuotaStore) {
	writeJSON(w, qs.All())
}

type quotaRequest struct {
	Process string `json:"process"`
	Limit   int    `json:"limit"`
	Window  string `json:"window"`
}

func handleQuotaPost(w http.ResponseWriter, r *http.Request, qs *monitor.QuotaStore) {
	var req quotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Window != "" {
		d, err := time.ParseDuration(req.Window)
		if err != nil || d <= 0 {
			http.Error(w, "invalid window duration", http.StatusBadRequest)
			return
		}
	}
	if err := qs.SetLimit(req.Process, req.Limit); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleQuotaReset(w http.ResponseWriter, r *http.Request, qs *monitor.QuotaStore) {
	process := r.URL.Query().Get("process")
	if process == "" {
		http.Error(w, "process query param required", http.StatusBadRequest)
		return
	}
	qs.Reset(process)
	w.WriteHeader(http.StatusNoContent)
}
