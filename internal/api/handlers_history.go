package api

import (
	"net/http"
	"strings"

	"github.com/user/procwatch/internal/monitor"
)

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional ?process=<name> filter
	processName := strings.TrimSpace(r.URL.Query().Get("process"))

	var events []monitor.HistoryEntry
	if processName != "" {
		events = s.history.ForProcess(processName)
	} else {
		events = s.history.All()
	}

	if events == nil {
		events = []monitor.HistoryEntry{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(events),
		"events":  events,
	})
}
