package monitor

import (
	"context"
	"log"
	"time"

	"github.com/user/procwatch/internal/alert"
	"github.com/user/procwatch/internal/config"
)

// Watcher periodically checks all configured processes and fires alerts
// when a process is down or exceeds resource thresholds.
type Watcher struct {
	cfg    *config.Config
	sender *alert.Sender
	finder *Finder
	// prevStates tracks the last known state so we only alert on transitions.
	prevStates map[string]*ProcessState
}

// NewWatcher constructs a Watcher from the given config and alert sender.
func NewWatcher(cfg *config.Config, sender *alert.Sender) *Watcher {
	return &Watcher{
		cfg:        cfg,
		sender:     sender,
		finder:     &Finder{},
		prevStates: make(map[string]*ProcessState),
	}
}

// Run starts the watch loop and blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.cfg.PollInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("procwatch: starting; polling every %ds for %d process(es)",
		w.cfg.PollInterval, len(w.cfg.Processes))

	for {
		select {
		case <-ctx.Done():
			log.Println("procwatch: shutting down")
			return
		case <-ticker.C:
			w.tick()
		}
	}
}

// tick performs one poll cycle across all configured processes.
func (w *Watcher) tick() {
	for i := range w.cfg.Processes {
		proc := &w.cfg.Processes[i]
		state, err := w.finder.Find(proc)
		if err != nil {
			log.Printf("procwatch: error finding %s: %v", proc.Name, err)
			continue
		}
		w.evaluate(proc, state)
		w.prevStates[proc.Name] = state
	}
}

// evaluate compares current state against thresholds and previous state,
// sending an alert when a noteworthy event is detected.
func (w *Watcher) evaluate(proc *config.Process, state *ProcessState) {
	prev := w.prevStates[proc.Name]

	if !state.Running {
		// Only alert on the transition from running → down (or first observation).
		if prev == nil || prev.Running {
			w.sendAlert(proc.Name, EventDown, state)
		}
		return
	}

	if proc.MaxCPU > 0 && state.CPUPercent > proc.MaxCPU {
		w.sendAlert(proc.Name, EventHighCPU, state)
	}
	if proc.MaxMemMB > 0 && state.MemRSSMB > float64(proc.MaxMemMB) {
		w.sendAlert(proc.Name, EventHighMem, state)
	}
}

func (w *Watcher) sendAlert(name string, event Event, state *ProcessState) {
	payload := buildPayload(name, event, state)
	if err := w.sender.Send(payload); err != nil {
		log.Printf("procwatch: alert send failed for %s: %v", name, err)
	}
}
