package api

import "github.com/madalinpopa/procwatch/internal/monitor"

// WithRunbookStoreExported exposes WithRunbookStore for black-box tests.
var WithRunbookStoreExported = WithRunbookStore

// NewRunbookStoreExported wraps monitor.NewRunbookStore for test helpers.
func NewRunbookStoreExported() *monitor.RunbookStore {
	return monitor.NewRunbookStore()
}
