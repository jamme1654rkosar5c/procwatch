package monitor

import (
	"sync"
	"time"
)

// Snapshot captures a point-in-time view of a process's resource usage.
type Snapshot struct {
	Process   string    `json:"process"`
	CPUPct    float64   `json:"cpu_pct"`
	MemBytes  uint64    `json:"mem_bytes"`
	PID       int       `json:"pid"`
	Up        bool      `json:"up"`
	RecordedAt time.Time `json:"recorded_at"`
}

// SnapshotStore retains the most recent snapshot per process.
type SnapshotStore struct {
	mu    sync.RWMutex
	store map[string]Snapshot
}

// NewSnapshotStore returns an initialised SnapshotStore.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		store: make(map[string]Snapshot),
	}
}

// Record saves or replaces the snapshot for the given process.
func (s *SnapshotStore) Record(snap Snapshot) {
	snap.RecordedAt = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[snap.Process] = snap
}

// Get returns the latest snapshot for a process and whether it exists.
func (s *SnapshotStore) Get(process string) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.store[process]
	return snap, ok
}

// All returns a copy of all stored snapshots keyed by process name.
func (s *SnapshotStore) All() map[string]Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Snapshot, len(s.store))
	for k, v := range s.store {
		out[k] = v
	}
	return out
}

// Delete removes the snapshot for a process.
func (s *SnapshotStore) Delete(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, process)
}
