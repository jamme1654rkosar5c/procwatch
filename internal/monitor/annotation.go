package monitor

import (
	"sync"
	"time"
)

// Annotation holds a user-defined note attached to a process.
type Annotation struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AnnotationStore stores free-form annotations keyed by process name.
type AnnotationStore struct {
	mu   sync.RWMutex
	notes map[string]Annotation
}

// NewAnnotationStore returns an initialised AnnotationStore.
func NewAnnotationStore() *AnnotationStore {
	return &AnnotationStore{notes: make(map[string]Annotation)}
}

// Set creates or replaces the annotation for the given process.
func (s *AnnotationStore) Set(process, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	existing, ok := s.notes[process]
	if !ok {
		existing = Annotation{CreatedAt: now}
	}
	existing.Text = text
	existing.UpdatedAt = now
	s.notes[process] = existing
}

// Get returns the annotation for the given process and whether it exists.
func (s *AnnotationStore) Get(process string) (Annotation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.notes[process]
	return a, ok
}

// Delete removes the annotation for the given process.
func (s *AnnotationStore) Delete(process string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.notes, process)
}

// All returns a snapshot of all annotations keyed by process name.
func (s *AnnotationStore) All() map[string]Annotation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Annotation, len(s.notes))
	for k, v := range s.notes {
		out[k] = v
	}
	return out
}
