package monitor

import (
	"sync"
	"time"
)

// SilenceRule suppresses alerts for a named process during a time window.
type SilenceRule struct {
	ProcessName string
	Until       time.Time
}

// Silencer manages active silence rules for processes.
type Silencer struct {
	mu    sync.RWMutex
	rules map[string]time.Time
	now   func() time.Time
}

// NewSilencer creates a new Silencer.
func NewSilencer() *Silencer {
	return &Silencer{
		rules: make(map[string]time.Time),
		now:   time.Now,
	}
}

// Silence suppresses alerts for the given process until the specified time.
func (s *Silencer) Silence(processName string, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[processName] = until
}

// Lift removes a silence rule for the given process.
func (s *Silencer) Lift(processName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rules, processName)
}

// IsSilenced returns true if the process currently has an active silence rule.
func (s *Silencer) IsSilenced(processName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	until, ok := s.rules[processName]
	if !ok {
		return false
	}
	if s.now().After(until) {
		return false
	}
	return true
}

// All returns a snapshot of all active silence rules (unexpired only).
func (s *Silencer) All() []SilenceRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	out := make([]SilenceRule, 0, len(s.rules))
	for name, until := range s.rules {
		if now.Before(until) {
			out = append(out, SilenceRule{ProcessName: name, Until: until})
		}
	}
	return out
}
