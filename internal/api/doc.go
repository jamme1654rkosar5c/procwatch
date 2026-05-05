// Package api exposes a lightweight HTTP interface for procwatch.
//
// Endpoints:
//
//	GET /healthz   – liveness probe, always returns {"status":"ok"}
//	GET /status    – current status of all watched processes
//	GET /history   – full event history; optional ?process=<name> filter
//	GET /summary   – aggregated summary (up/down counts, alert totals)
//
// The server is intended to run alongside the watcher loop and shares
// the same in-memory StatusRegistry, History, and SummaryBuilder
// instances, so no additional storage is required.
package api
