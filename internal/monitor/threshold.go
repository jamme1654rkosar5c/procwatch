package monitor

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

// ThresholdAlert describes which threshold was breached.
type ThresholdAlert struct {
	Kind    string  // "cpu" or "mem"
	Value   float64 // observed value
	Limit   float64 // configured limit
	Message string
}

// checkThresholds inspects CPU and memory usage for the given process
// against the configured limits. It returns a non-nil ThresholdAlert when
// a limit is exceeded, or nil when everything is within bounds.
func checkThresholds(p *process.Process, cpuLimit, memLimitMB float64) (*ThresholdAlert, error) {
	if cpuLimit > 0 {
		cpuPct, err := p.CPUPercent()
		if err != nil {
			return nil, fmt.Errorf("cpu percent: %w", err)
		}
		if cpuPct > cpuLimit {
			return &ThresholdAlert{
				Kind:    "cpu",
				Value:   cpuPct,
				Limit:   cpuLimit,
				Message: fmt.Sprintf("CPU usage %.1f%% exceeds limit %.1f%%", cpuPct, cpuLimit),
			}, nil
		}
	}

	if memLimitMB > 0 {
		memInfo, err := p.MemoryInfo()
		if err != nil {
			return nil, fmt.Errorf("memory info: %w", err)
		}
		usedMB := float64(memInfo.RSS) / 1024 / 1024
		if usedMB > memLimitMB {
			return &ThresholdAlert{
				Kind:    "mem",
				Value:   usedMB,
				Limit:   memLimitMB,
				Message: fmt.Sprintf("Memory usage %.1f MB exceeds limit %.1f MB", usedMB, memLimitMB),
			}, nil
		}
	}

	return nil, nil
}
