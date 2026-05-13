package monitor

import (
	"sync"
	"time"
)

// SuppressionRule defines a rule that suppresses alerts for a process+event combination
// until a specified expiry time.
type SuppressionRule struct {
	Process   string    `json:"process"`
	EventType string    `json:"event_type"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SuppressionStore tracks active suppression rules keyed by process and event type.
type SuppressionStore struct {
	mu    sync.RWMutex
	rules map[string]map[string]SuppressionRule // process -> eventType -> rule
	now   func() time.Time
}

// NewSuppressionStore returns an initialised SuppressionStore.
func NewSuppressionStore() *SuppressionStore {
	return &SuppressionStore{
		rules: make(map[string]map[string]SuppressionRule),
		now:   time.Now,
	}
}

// Add inserts or replaces a suppression rule for the given process and event type.
func (s *SuppressionStore) Add(process, eventType, reason string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rules[process]; !ok {
		s.rules[process] = make(map[string]SuppressionRule)
	}
	s.rules[process][eventType] = SuppressionRule{
		Process:   process,
		EventType: eventType,
		Reason:    reason,
		ExpiresAt: s.now().Add(duration),
	}
}

// IsSuppressed reports whether alerts of eventType for process are currently suppressed.
func (s *SuppressionStore) IsSuppressed(process, eventType string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	byEvent, ok := s.rules[process]
	if !ok {
		return false
	}
	rule, ok := byEvent[eventType]
	if !ok {
		return false
	}
	return s.now().Before(rule.ExpiresAt)
}

// Remove deletes the suppression rule for the given process and event type.
func (s *SuppressionStore) Remove(process, eventType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if byEvent, ok := s.rules[process]; ok {
		delete(byEvent, eventType)
		if len(byEvent) == 0 {
			delete(s.rules, process)
		}
	}
}

// All returns a flat slice of all currently active (non-expired) suppression rules.
func (s *SuppressionStore) All() []SuppressionRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	var out []SuppressionRule
	for _, byEvent := range s.rules {
		for _, rule := range byEvent {
			if now.Before(rule.ExpiresAt) {
				out = append(out, rule)
			}
		}
	}
	return out
}
