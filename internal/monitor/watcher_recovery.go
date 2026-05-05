package monitor

import (
	"context"
	"log"

	"github.com/ncolesummers/procwatch/internal/config"
)

// handleRecovery checks whether a process that was previously marked as down
// has come back up, and if so fires a recovery webhook and resets state.
func (w *Watcher) handleRecovery(proc config.Process, state *processState) {
	if state == nil {
		return
	}

	ev, recovered := checkRecovery(proc.Name, state.Up, w.downSince)
	if !recovered {
		return
	}

	ev.PID = state.PID
	payload := buildRecoveryPayload(ev)

	go func() {
		if err := w.sender.Send(context.Background(), payload); err != nil {
			log.Printf("procwatch: recovery alert failed for %q: %v", proc.Name, err)
			return
		}
		log.Printf("procwatch: process %q recovered after %.1fs downtime",
			proc.Name, ev.Downtime.Seconds())
	}()

	// Clear the alerted flag so a future crash triggers a fresh down-alert.
	w.alerted[proc.Name] = false
}
