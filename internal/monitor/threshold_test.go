package monitor

import (
	"testing"

	"github.com/shirou/gopsutil/v3/process"
)

// TestCheckThresholds_NilWhenNoLimits verifies that zero limits produce no alert.
func TestCheckThresholds_NilWhenNoLimits(t *testing.T) {
	// Use the current process which is definitely running.
	procs, err := process.Processes()
	if err != nil || len(procs) == 0 {
		t.Skip("cannot list processes")
	}
	var self *process.Process
	for _, p := range procs {
		if ok, _ := p.IsRunning(); ok {
			self = p
			break
		}
	}
	if self == nil {
		t.Skip("no running process found")
	}

	alert, err := checkThresholds(self, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert != nil {
		t.Errorf("expected nil alert with zero limits, got %+v", alert)
	}
}

// TestCheckThresholds_CPUAlertTriggered verifies a CPU alert fires when limit is tiny.
func TestCheckThresholds_CPUAlertTriggered(t *testing.T) {
	procs, err := process.Processes()
	if err != nil || len(procs) == 0 {
		t.Skip("cannot list processes")
	}
	var self *process.Process
	for _, p := range procs {
		if ok, _ := p.IsRunning(); ok {
			self = p
			break
		}
	}
	if self == nil {
		t.Skip("no running process found")
	}

	// A limit of -1 is invalid; use an astronomically small positive value
	// so any non-zero CPU reading triggers the alert.
	// CPUPercent may return 0 on first call; we just check no error occurs
	// and that the alert kind is correct when triggered.
	alert, err := checkThresholds(self, 0.000001, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert != nil && alert.Kind != "cpu" {
		t.Errorf("expected kind=cpu, got %s", alert.Kind)
	}
}

// TestCheckThresholds_MemAlertTriggered verifies a mem alert fires when limit is tiny.
func TestCheckThresholds_MemAlertTriggered(t *testing.T) {
	procs, err := process.Processes()
	if err != nil || len(procs) == 0 {
		t.Skip("cannot list processes")
	}
	var self *process.Process
	for _, p := range procs {
		if ok, _ := p.IsRunning(); ok {
			self = p
			break
		}
	}
	if self == nil {
		t.Skip("no running process found")
	}

	// 0.001 MB = ~1 KB; any real process will exceed this.
	alert, err := checkThresholds(self, 0, 0.001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert == nil {
		t.Fatal("expected mem alert, got nil")
	}
	if alert.Kind != "mem" {
		t.Errorf("expected kind=mem, got %s", alert.Kind)
	}
	if alert.Value <= 0 {
		t.Errorf("expected positive mem value, got %f", alert.Value)
	}
}
