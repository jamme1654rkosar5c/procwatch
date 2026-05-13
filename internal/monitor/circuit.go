package monitor

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker for a process.
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"   // normal operation
	CircuitOpen     CircuitState = "open"     // alerts suppressed
	CircuitHalfOpen CircuitState = "half_open" // testing recovery
)

type circuitEntry struct {
	state     CircuitState
	failures  int
	openedAt  time.Time
	resetAfter time.Duration
}

// CircuitBreaker suppresses alerts for a process once failure count
// exceeds a threshold, reopening after a configurable reset window.
type CircuitBreaker struct {
	mu        sync.Mutex
	entries   map[string]*circuitEntry
	threshold int
	resetAfter time.Duration
	now       func() time.Time
}

// NewCircuitBreaker creates a CircuitBreaker that opens after threshold
// consecutive failures and resets after resetAfter duration.
func NewCircuitBreaker(threshold int, resetAfter time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		entries:    make(map[string]*circuitEntry),
		threshold:  threshold,
		resetAfter: resetAfter,
		now:        time.Now,
	}
}

// Record records a failure for the given process and returns the current state.
func (cb *CircuitBreaker) Record(process string) CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	e := cb.getOrCreate(process)
	cb.maybeReset(e)

	if e.state == CircuitOpen {
		return CircuitOpen
	}

	e.failures++
	if e.failures >= cb.threshold {
		e.state = CircuitOpen
		e.openedAt = cb.now()
	}
	return e.state
}

// State returns the current CircuitState for a process.
func (cb *CircuitBreaker) State(process string) CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	e := cb.getOrCreate(process)
	cb.maybeReset(e)
	return e.state
}

// Reset manually closes the circuit for a process.
func (cb *CircuitBreaker) Reset(process string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.entries, process)
}

// All returns a snapshot of all tracked process states.
func (cb *CircuitBreaker) All() map[string]CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	out := make(map[string]CircuitState, len(cb.entries))
	for k, e := range cb.entries {
		cb.maybeReset(e)
		out[k] = e.state
	}
	return out
}

func (cb *CircuitBreaker) getOrCreate(process string) *circuitEntry {
	if e, ok := cb.entries[process]; ok {
		return e
	}
	e := &circuitEntry{state: CircuitClosed, resetAfter: cb.resetAfter}
	cb.entries[process] = e
	return e
}

func (cb *CircuitBreaker) maybeReset(e *circuitEntry) {
	if e.state == CircuitOpen && cb.now().After(e.openedAt.Add(e.resetAfter)) {
		e.state = CircuitHalfOpen
		e.failures = 0
	}
}
