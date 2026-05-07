package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/procwatch/internal/monitor"
)

type dependencyRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type dependencyResponse struct {
	From         string   `json:"from"`
	Dependencies []string `json:"dependencies"`
	Dependents   []string `json:"dependents"`
}

// WithDependencyGraph injects a DependencyGraph into the Server.
func WithDependencyGraph(g *monitor.DependencyGraph) func(*Server) {
	return func(s *Server) {
		s.deps = g
	}
}

// handleDependency manages process dependency edges.
//
//	POST   /deps        – register a dependency edge {from, to}
//	DELETE /deps        – remove a dependency edge {from, to}
//	GET    /deps?process=<name> – list deps and dependents for a process
func (s *Server) handleDependency(w http.ResponseWriter, r *http.Request) {
	if s.deps == nil {
		http.Error(w, "dependency graph not configured", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		process := r.URL.Query().Get("process")
		if process == "" {
			all := s.deps.All()
			writeJSON(w, http.StatusOK, all)
			return
		}
		resp := dependencyResponse{
			From:         process,
			Dependencies: s.deps.Dependencies(process),
			Dependents:   s.deps.Dependents(process),
		}
		if resp.Dependencies == nil {
			resp.Dependencies = []string{}
		}
		if resp.Dependents == nil {
			resp.Dependents = []string{}
		}
		writeJSON(w, http.StatusOK, resp)

	case http.MethodPost:
		var req dependencyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.From == "" || req.To == "" {
			http.Error(w, "invalid request body: from and to required", http.StatusBadRequest)
			return
		}
		s.deps.Add(req.From, req.To)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		var req dependencyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.From == "" || req.To == "" {
			http.Error(w, "invalid request body: from and to required", http.StatusBadRequest)
			return
		}
		s.deps.Remove(req.From, req.To)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
