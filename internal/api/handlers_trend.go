package api

import (
	"net/http"

	"github.com/derekg/procwatch/internal/monitor"
)

// WithTrendAnalyzer registers the /api/trends endpoint on srv.
func WithTrendAnalyzer(srv *Server, ta *monitor.TrendAnalyzer) {
	srv.mux.HandleFunc("/api/trends", makeTrendHandler(ta))
}

func makeTrendHandler(ta *monitor.TrendAnalyzer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		process := r.URL.Query().Get("process")
		if process != "" {
			res := ta.Analyze(process)
			writeJSON(w, map[string]any{
				"process":   process,
				"direction": res.Direction,
				"delta":     res.Delta,
				"samples":   res.Samples,
			})
			return
		}

		all := ta.All()
		type row struct {
			Process   string                 `json:"process"`
			Direction monitor.TrendDirection `json:"direction"`
			Delta     float64                `json:"delta"`
			Samples   int                    `json:"samples"`
		}
		rows := make([]row, 0, len(all))
		for proc, res := range all {
			rows = append(rows, row{
				Process:   proc,
				Direction: res.Direction,
				Delta:     res.Delta,
				Samples:   res.Samples,
			})
		}
		writeJSON(w, rows)
	}
}
