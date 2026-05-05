package monitor

import (
	"fmt"
	"time"
)

// RecoveryEvent holds metadata about a process that recovered after being down.
type RecoveryEvent struct {
	ProcessName string
	PID         int32
	DownAt      time.Time
	RecoveredAt time.Time
	Downtime    time.Duration
}

// checkRecovery compares the current process state against the watcher's
// last-known-down map and returns a RecoveryEvent when a process has come
// back up. It also cleans up the tracking state so future ticks are correct.
func checkRecovery(name string, isUp bool, downSince map[string]time.Time) (*RecoveryEvent, bool) {
	downAt, wasDown := downSince[name]
	if !wasDown {
		return nil, false
	}
	if !isUp {
		// Still down — nothing to report yet.
		return nil, false
	}

	now := time.Now()
	event := &RecoveryEvent{
		ProcessName: name,
		DownAt:      downAt,
		RecoveredAt: now,
		Downtime:    now.Sub(downAt),
	}
	delete(downSince, name)
	return event, true
}

// buildRecoveryPayload converts a RecoveryEvent into the alert payload map
// that the webhook sender expects.
func buildRecoveryPayload(ev *RecoveryEvent) map[string]interface{} {
	return map[string]interface{}{
		"event":        "recovered",
		"process":      ev.ProcessName,
		"pid":          ev.PID,
		"down_at":      ev.DownAt.UTC().Format(time.RFC3339),
		"recovered_at": ev.RecoveredAt.UTC().Format(time.RFC3339),
		"downtime_sec": fmt.Sprintf("%.1f", ev.Downtime.Seconds()),
	}
}
