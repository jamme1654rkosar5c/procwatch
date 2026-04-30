// Package monitor provides process discovery, resource sampling, and
// event evaluation for procwatch.
//
// # Overview
//
// A [Finder] resolves each configured process to a running OS process,
// collecting CPU and memory metrics via [buildState].
//
// A [Watcher] drives the polling loop: on each tick it calls [Finder.Find]
// for every configured process, compares the result against configured
// thresholds, and forwards [AlertPayload] values to an [alert.Sender]
// when a process goes down or exceeds a resource limit.
//
// Alert deduplication is handled by tracking the previous [ProcessState]
// per process name; a down alert is only fired on the running→down
// transition, preventing repeated notifications for a persistently
// crashed process.
package monitor
