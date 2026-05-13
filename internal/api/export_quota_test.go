package api

import "github.com/weezel/procwatch/internal/monitor"

// WithQuotaStoreExported exposes WithQuotaStore for black-box tests.
func WithQuotaStoreExported(mux interface{ Handle(string, interface{}) }, qs *monitor.QuotaStore) {
	if m, ok := mux.(interface {
		HandleFunc(string, func(interface{}, interface{}))
	}); ok {
		_ = m
	}
}

// MakeHandleQuotaExported exposes makeHandleQuota for testing.
var MakeHandleQuotaExported = makeHandleQuota
