package monitor

import (
	"testing"
	"time"
)

func TestBuildSummary_EmptyRegistry(t *testing.T) {
	reg := NewStatusRegistry()
	hist := NewHistory(10)
	sb := NewSummaryBuilder(reg, hist)

	report := sb.Build()

	if len(report.Processes) != 0 {
		t.Fatalf("expected 0 processes, got %d", len(report.Processes))
	}
	if report.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
}

func TestBuildSummary_UpProcess(t *testing.T) {
	reg := NewStatusRegistry()
	hist := NewHistory(10)

	state := &ProcessState{
		PID:         1234,
		CPUPercent:  12.5,
		MemoryBytes: 1024 * 1024 * 50, // 50 MB
		CapturedAt:  time.Now(),
	}
	reg.Update("myapp", "up", state)

	sb := NewSummaryBuilder(reg, hist)
	report := sb.Build()

	if len(report.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(report.Processes))
	}
	ps := report.Processes[0]
	if ps.Name != "myapp" {
		t.Errorf("expected name myapp, got %s", ps.Name)
	}
	if ps.Status != "up" {
		t.Errorf("expected status up, got %s", ps.Status)
	}
	if ps.PID != 1234 {
		t.Errorf("expected PID 1234, got %d", ps.PID)
	}
	if ps.MemoryMB != 50.0 {
		t.Errorf("expected 50 MB, got %f", ps.MemoryMB)
	}
	if ps.AlertCount != 0 {
		t.Errorf("expected 0 alerts, got %d", ps.AlertCount)
	}
}

func TestBuildSummary_AlertCountFromHistory(t *testing.T) {
	reg := NewStatusRegistry()
	hist := NewHistory(10)

	reg.Update("svc", "down", nil)

	hist.Record(EventRecord{Process: "svc", Kind: "down", Timestamp: time.Now()})
	hist.Record(EventRecord{Process: "svc", Kind: "down", Timestamp: time.Now()})

	sb := NewSummaryBuilder(reg, hist)
	report := sb.Build()

	if len(report.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(report.Processes))
	}
	ps := report.Processes[0]
	if ps.AlertCount != 2 {
		t.Errorf("expected 2 alerts, got %d", ps.AlertCount)
	}
	if ps.LastAlertAt.IsZero() {
		t.Error("expected LastAlertAt to be set")
	}
}
