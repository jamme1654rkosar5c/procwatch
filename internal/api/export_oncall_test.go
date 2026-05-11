package api

import "github.com/yourorg/procwatch/internal/monitor"

// WithOnCallStoreExported exposes WithOnCallStore for black-box tests.
func WithOnCallStoreExported(srv *Server, store *monitor.OnCallStore) {
	WithOnCallStore(srv, store)
}
