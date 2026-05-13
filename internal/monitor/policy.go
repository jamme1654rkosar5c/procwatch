package monitor

import (
	"errors"
	"sync"
	"time"
)

// AlertPolicy defines routing and delivery rules for a process.
type AlertPolicy struct {
	Process    string        `json:"process"`
	MinSeverity string       `json:"min_severity"` // "info", "warn", "critical"
	Channels   []string      `json:"channels"`
	Cooldown   time.Duration `json:"cooldown_seconds"`
	CreatedAt  time.Time     `json:"created_at"`
}

// PolicyStore holds per-process alert policies.
type PolicyStore struct {
	mu       sync.RWMutex
	policies map[string]AlertPolicy
}

// NewPolicyStore returns an empty PolicyStore.
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{policies: make(map[string]AlertPolicy)}
}

// Set stores or replaces the policy for a process.
func (s *PolicyStore) Set(p AlertPolicy) error {
	if p.Process == "" {
		return errors.New("process name required")
	}
	if !validSeverity(p.MinSeverity) {
		return errors.New("min_severity must be info, warn, or critical")
	}
	p.CreatedAt = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[p.Process] = p
	return nil
}

// Get returns the policy for a process, if set.
func (s *PolicyStore) Get(process string) (AlertPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[process]
	return p, ok
}

// Delete removes the policy for a process.
func (s *PolicyStore) Delete(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.policies, process)
}

// All returns a copy of all policies.
func (s *PolicyStore) All() []AlertPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AlertPolicy, 0, len(s.policies))
	for _, p := range s.policies {
		out = append(out, p)
	}
	return out
}

func validSeverity(s string) bool {
	return s == "info" || s == "warn" || s == "critical"
}
