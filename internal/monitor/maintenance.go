package monitor

import (
	"sync"
	"time"
)

// MaintenanceWindow represents a scheduled period during which alerts are suppressed.
type MaintenanceWindow struct {
	Process string
	Start   time.Time
	End     time.Time
	Reason  string
}

// IsActive returns true if the window covers the given time.
func (w MaintenanceWindow) IsActive(t time.Time) bool {
	return !t.Before(w.Start) && t.Before(w.End)
}

// MaintenanceScheduler tracks scheduled maintenance windows per process.
type MaintenanceScheduler struct {
	mu      sync.RWMutex
	windows map[string][]MaintenanceWindow
	now     func() time.Time
}

// NewMaintenanceScheduler creates a new MaintenanceScheduler.
func NewMaintenanceScheduler() *MaintenanceScheduler {
	return &MaintenanceScheduler{
		windows: make(map[string][]MaintenanceWindow),
		now:     time.Now,
	}
}

// Add registers a maintenance window for the given process.
func (s *MaintenanceScheduler) Add(w MaintenanceWindow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[w.Process] = append(s.windows[w.Process], w)
}

// IsUnderMaintenance returns true if the process has an active window right now.
func (s *MaintenanceScheduler) IsUnderMaintenance(process string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	for _, w := range s.windows[process] {
		if w.IsActive(now) {
			return true
		}
	}
	return false
}

// Prune removes expired windows to prevent unbounded growth.
func (s *MaintenanceScheduler) Prune() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for proc, ws := range s.windows {
		var active []MaintenanceWindow
		for _, w := range ws {
			if now.Before(w.End) {
				active = append(active, w)
			}
		}
		if len(active) == 0 {
			delete(s.windows, proc)
		} else {
			s.windows[proc] = active
		}
	}
}

// All returns a snapshot of all currently registered windows.
func (s *MaintenanceScheduler) All() []MaintenanceWindow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []MaintenanceWindow
	for _, ws := range s.windows {
		out = append(out, ws...)
	}
	return out
}
