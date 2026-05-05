package monitor

import (
	"sync"
	"time"
)

// ProcessSummary holds a rolled-up view of a single watched process.
type ProcessSummary struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // "up" | "down" | "unknown"
	PID         int32     `json:"pid,omitempty"`
	CPUPercent  float64   `json:"cpu_percent,omitempty"`
	MemoryMB    float64   `json:"memory_mb,omitempty"`
	LastSeen    time.Time `json:"last_seen,omitempty"`
	AlertCount  int       `json:"alert_count"`
	LastAlertAt time.Time `json:"last_alert_at,omitempty"`
}

// SummaryReport is returned by BuildSummary.
type SummaryReport struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Processes   []ProcessSummary `json:"processes"`
}

// SummaryBuilder composes a SummaryReport from the registry and history.
type SummaryBuilder struct {
	mu       sync.Mutex
	registry *StatusRegistry
	history  *History
}

// NewSummaryBuilder creates a SummaryBuilder backed by the given registry and history.
func NewSummaryBuilder(registry *StatusRegistry, history *History) *SummaryBuilder {
	return &SummaryBuilder{
		registry: registry,
		history:  history,
	}
}

// Build assembles a SummaryReport snapshot at the current moment.
func (sb *SummaryBuilder) Build() SummaryReport {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	statuses := sb.registry.All()
	summaries := make([]ProcessSummary, 0, len(statuses))

	for name, st := range statuses {
		ps := ProcessSummary{
			Name:   name,
			Status: st.Status,
		}
		if st.State != nil {
			ps.PID = st.State.PID
			ps.CPUPercent = st.State.CPUPercent
			ps.MemoryMB = float64(st.State.MemoryBytes) / (1024 * 1024)
			ps.LastSeen = st.State.CapturedAt
		}

		events := sb.history.ForProcess(name)
		ps.AlertCount = len(events)
		if len(events) > 0 {
			ps.LastAlertAt = events[len(events)-1].Timestamp
		}

		summaries = append(summaries, ps)
	}

	return SummaryReport{
		GeneratedAt: time.Now().UTC(),
		Processes:   summaries,
	}
}
