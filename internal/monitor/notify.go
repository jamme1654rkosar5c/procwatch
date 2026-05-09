package monitor

import (
	"sync"
	"time"
)

// NotifyRule defines a condition-based notification rule for a process.
type NotifyRule struct {
	Process   string
	EventType string // "down", "cpu", "mem", "flap", "recovered"
	Channels  []string
	CreatedAt time.Time
}

// NotifyRuleStore manages per-process notification routing rules.
type NotifyRuleStore struct {
	mu    sync.RWMutex
	rules map[string][]NotifyRule // keyed by process name
}

// NewNotifyRuleStore creates an empty NotifyRuleStore.
func NewNotifyRuleStore() *NotifyRuleStore {
	return &NotifyRuleStore{
		rules: make(map[string][]NotifyRule),
	}
}

// Add appends a notification rule for a process.
func (s *NotifyRuleStore) Add(rule NotifyRule) error {
	if rule.Process == "" {
		return errEmptyProcess
	}
	if rule.EventType == "" {
		return errEmptyEventType
	}
	rule.CreatedAt = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.Process] = append(s.rules[rule.Process], rule)
	return nil
}

// Get returns all rules for a given process.
func (s *NotifyRuleStore) Get(process string) []NotifyRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NotifyRule, len(s.rules[process]))
	copy(out, s.rules[process])
	return out
}

// ChannelsFor returns the union of channels for a process+eventType pair.
func (s *NotifyRuleStore) ChannelsFor(process, eventType string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := map[string]struct{}{}
	var out []string
	for _, r := range s.rules[process] {
		if r.EventType == eventType || r.EventType == "*" {
			for _, ch := range r.Channels {
				if _, ok := seen[ch]; !ok {
					seen[ch] = struct{}{}
					out = append(out, ch)
				}
			}
		}
	}
	return out
}

// Delete removes all rules for a process.
func (s *NotifyRuleStore) Delete(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rules, process)
}

// All returns a copy of all rules across all processes.
func (s *NotifyRuleStore) All() map[string][]NotifyRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]NotifyRule, len(s.rules))
	for k, v := range s.rules {
		cp := make([]NotifyRule, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

var errEmptyProcess = errString("process name must not be empty")
var errEmptyEventType = errString("event type must not be empty")

type errString string

func (e errString) Error() string { return string(e) }
