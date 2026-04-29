package monitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessState holds the observed state of a watched process.
type ProcessState struct {
	Name    string
	PID     int32
	Running bool
	CPU     float64 // percent
	MemMB   float64
}

// Finder locates a process by name or PID file path.
type Finder struct{}

// Find returns the ProcessState for the given name or pidfile.
// name is checked first; if empty, pidFile is used.
func (f *Finder) Find(name, pidFile string) (*ProcessState, error) {
	if name != "" {
		return findByName(name)
	}
	if pidFile != "" {
		return findByPIDFile(pidFile)
	}
	return nil, fmt.Errorf("no locator provided: set name or pid_file")
}

func findByName(name string) (*ProcessState, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("listing processes: %w", err)
	}
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}
		if strings.EqualFold(n, name) {
			return buildState(name, p)
		}
	}
	return &ProcessState{Name: name, Running: false}, nil
}

func findByPIDFile(path string) (*ProcessState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &ProcessState{Name: path, Running: false}, nil
	}
	pid, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parsing pid file %s: %w", path, err)
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return &ProcessState{Name: path, Running: false}, nil
	}
	name, _ := p.Name()
	return buildState(name, p)
}

func buildState(name string, p *process.Process) (*ProcessState, error) {
	running, _ := p.IsRunning()
	cpu, _ := p.CPUPercent()
	mem, _ := p.MemoryInfo()
	var memMB float64
	if mem != nil {
		memMB = float64(mem.RSS) / 1024 / 1024
	}
	return &ProcessState{
		Name:    name,
		PID:     p.Pid,
		Running: running,
		CPU:     cpu,
		MemMB:   memMB,
	}, nil
}
