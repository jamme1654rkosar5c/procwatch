package monitor

import (
	"fmt"
	"time"
)

// Event describes the kind of alert being raised.
type Event string

const (
	EventDown    Event = "process_down"
	EventHighCPU Event = "high_cpu"
	EventHighMem Event = "high_memory"
)

// AlertPayload is the JSON body sent to the webhook.
type AlertPayload struct {
	Timestamp   time.Time    `json:"timestamp"`
	Event       Event        `json:"event"`
	ProcessName string       `json:"process_name"`
	Details     string       `json:"details"`
	State       *ProcessState `json:"state,omitempty"`
}

// buildPayload constructs an AlertPayload for the given event.
func buildPayload(name string, event Event, state *ProcessState) AlertPayload {
	var details string
	switch event {
	case EventDown:
		details = fmt.Sprintf("process %q is not running", name)
	case EventHighCPU:
		details = fmt.Sprintf("process %q CPU %.1f%% exceeds threshold", name, state.CPUPercent)
	case EventHighMem:
		details = fmt.Sprintf("process %q memory %.1f MB exceeds threshold", name, state.MemRSSMB)
	default:
		details = fmt.Sprintf("unknown event for process %q", name)
	}
	return AlertPayload{
		Timestamp:   time.Now().UTC(),
		Event:       event,
		ProcessName: name,
		Details:     details,
		State:       state,
	}
}
