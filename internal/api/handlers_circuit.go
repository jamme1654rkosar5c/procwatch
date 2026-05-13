package api

import (
	"encoding/json"
	"net/http"

	"github.com/robmorgan/procwatch/internal/monitor"
)

// WithCircuitBreaker registers the /api/circuit routes on the given mux.
func WithCircuitBreaker(mux *http.ServeMux, cb *monitor.CircuitBreaker) {
	mux.HandleFunc("/api/circuit", makeHandleCircuit(cb))
}

func makeHandleCircuit(cb *monitor.CircuitBreaker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleCircuitGet(w, r, cb)
		case http.MethodDelete:
			handleCircuitReset(w, r, cb)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

type circuitResponse struct {
	Process string `json:"process"`
	State   string `json:"state"`
}

func handleCircuitGet(w http.ResponseWriter, r *http.Request, cb *monitor.CircuitBreaker) {
	process := r.URL.Query().Get("process")
	if process != "" {
		state := cb.State(process)
		writeJSON(w, http.StatusOK, circuitResponse{Process: process, State: string(state)})
		return
	}
	all := cb.All()
	results := make([]circuitResponse, 0, len(all))
	for p, s := range all {
		results = append(results, circuitResponse{Process: p, State: string(s)})
	}
	writeJSON(w, http.StatusOK, results)
}

func handleCircuitReset(w http.ResponseWriter, r *http.Request, cb *monitor.CircuitBreaker) {
	var body struct {
		Process string `json:"process"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Process == "" {
		http.Error(w, "process required", http.StatusBadRequest)
		return
	}
	cb.Reset(body.Process)
	w.WriteHeader(http.StatusNoContent)
}
