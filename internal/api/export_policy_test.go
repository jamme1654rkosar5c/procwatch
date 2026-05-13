package api

import "github.com/shawnflorida/procwatch/internal/monitor"

// NewPolicyStoreExported exposes NewPolicyStore for use in external test packages.
func NewPolicyStoreExported() *monitor.PolicyStore {
	return monitor.NewPolicyStore()
}

// WithPolicyStoreExported exposes WithPolicyStore for use in external test packages.
var WithPolicyStoreExported = WithPolicyStore
