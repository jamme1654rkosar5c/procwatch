package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/robjkc/procwatch/internal/monitor"
)

// handleMetrics returns a Prometheus-compatible plain-text metrics exposition
// for all watched processes (up/down gauge + alert count).
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary := s.summaryBuilder.Build()

	var sb strings.Builder

	sb.WriteString("# HELP procwatch_process_up 1 if the process is currently running, 0 otherwise.\n")
	sb.WriteString("# TYPE procwatch_process_up gauge\n")
	for _, p := range summary.Processes {
		upVal := 0
		if p.Status == monitor.StatusUp {
			upVal = 1
		}
		sb.WriteString(fmt.Sprintf("procwatch_process_up{process=%q} %d\n", p.Name, upVal))
	}

	sb.WriteString("# HELP procwatch_process_alert_total Total number of alerts fired for the process.\n")
	sb.WriteString("# TYPE procwatch_process_alert_total counter\n")
	for _, p := range summary.Processes {
		sb.WriteString(fmt.Sprintf("procwatch_process_alert_total{process=%q} %d\n", p.Name, p.AlertCount))
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(sb.String()))
}
