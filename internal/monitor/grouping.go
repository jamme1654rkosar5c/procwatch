package monitor

import (
	"fmt"
	"sync"
)

// GroupRegistry maps processes to named groups, allowing bulk operations
// and filtered views across processes that share a common group label.
type GroupRegistry struct {
	mu     sync.RWMutex
	groups map[string]map[string]struct{} // group -> set of process names
	procs  map[string]string              // process -> group
}

// NewGroupRegistry creates an empty GroupRegistry.
func NewGroupRegistry() *GroupRegistry {
	return &GroupRegistry{
		groups: make(map[string]map[string]struct{}),
		procs:  make(map[string]string),
	}
}

// Assign places a process into a group, removing it from any previous group.
func (r *GroupRegistry) Assign(process, group string) error {
	if process == "" {
		return fmt.Errorf("process name must not be empty")
	}
	if group == "" {
		return fmt.Errorf("group name must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from old group if present.
	if old, ok := r.procs[process]; ok && old != group {
		delete(r.groups[old], process)
		if len(r.groups[old]) == 0 {
			delete(r.groups, old)
		}
	}

	if r.groups[group] == nil {
		r.groups[group] = make(map[string]struct{})
	}
	r.groups[group][process] = struct{}{}
	r.procs[process] = group
	return nil
}

// Remove removes a process from its group entirely.
func (r *GroupRegistry) Remove(process string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.procs[process]; ok {
		delete(r.groups[g], process)
		if len(r.groups[g]) == 0 {
			delete(r.groups, g)
		}
		delete(r.procs, process)
	}
}

// GroupOf returns the group assigned to a process, or "" if unassigned.
func (r *GroupRegistry) GroupOf(process string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.procs[process]
}

// Members returns a copy of all process names in the given group.
func (r *GroupRegistry) Members(group string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	set, ok := r.groups[group]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	return out
}

// All returns a snapshot of every group and its members.
func (r *GroupRegistry) All() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string][]string, len(r.groups))
	for g, set := range r.groups {
		members := make([]string, 0, len(set))
		for p := range set {
			members = append(members, p)
		}
		out[g] = members
	}
	return out
}
