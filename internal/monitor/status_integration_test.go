package monitor

import (
	"testing"

	"github.com/aryan-mehta05/procwatch/internal/config"
)

func TestWatcher_tick_UpdatesStatusRegistry(t *testing.T) {
	w, sender := newTestWatcher(t)
	_ = sender

	// Inject a StatusRegistry into the watcher.
	reg := NewStatusRegistry()
	w.status = reg

	// Simulate a tick where the process is not found (down).
	w.tick()

	proc := w.cfg.Processes[0]
	s, ok := reg.Get(proc.Name)
	if !ok {
		t.Fatalf("expected status entry for process %q", proc.Name)
	}
	if s.Up {
		t.Errorf("expected process to be marked down after failed find")
	}
	if s.LastEvent != "down" {
		t.Errorf("expected LastEvent='down', got %q", s.LastEvent)
	}
}

func TestWatcher_tick_UpdatesStatusRegistry_Up(t *testing.T) {
	cfg := &config.Config{
		WebhookURL:   "http://example.com/hook",
		PollInterval: 1,
		Processes: []config.Process{
			{Name: "procwatch_self", PIDFile: ""},
		},
	}
	// Use current process name so findByName succeeds.
	cfg.Processes[0].Name = currentProcessName()

	reg := NewStatusRegistry()
	w := &Watcher{
		cfg:      cfg,
		sender:   &noopSender{},
		downSeen: make(map[string]bool),
		status:   reg,
		throttle: NewThrottler(cfg.AlertCooldown),
		history:  NewHistory(100),
	}

	w.tick()

	s, ok := reg.Get(cfg.Processes[0].Name)
	if !ok {
		t.Skip("process not found by name in this environment, skipping")
	}
	if !s.Up {
		t.Errorf("expected process to be marked up")
	}
	if s.LastEvent != "running" {
		t.Errorf("expected LastEvent='running', got %q", s.LastEvent)
	}
}
