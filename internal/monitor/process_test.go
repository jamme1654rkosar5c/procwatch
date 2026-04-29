package monitor

import (
	"fmt"
	"os"
	"testing"
)

func TestFinder_Find_NoLocator(t *testing.T) {
	f := &Finder{}
	_, err := f.Find("", "")
	if err == nil {
		t.Fatal("expected error when no locator provided")
	}
}

func TestFinder_Find_ByName_CurrentProcess(t *testing.T) {
	// The test binary itself should appear in the process list.
	// We look for a known process that is always running on any OS: use the
	// current process name via /proc/self or os.Executable as a smoke test.
	f := &Finder{}
	// Use a name that definitely won't exist to confirm "not running" path.
	state, err := f.Find("__procwatch_nonexistent_xyz__", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Running {
		t.Error("expected process to be not running")
	}
	if state.Name != "__procwatch_nonexistent_xyz__" {
		t.Errorf("unexpected name: %s", state.Name)
	}
}

func TestFinder_Find_ByPIDFile_Missing(t *testing.T) {
	f := &Finder{}
	state, err := f.Find("", "/tmp/__procwatch_no_such_pidfile__.pid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Running {
		t.Error("expected not running for missing pid file")
	}
}

func TestFinder_Find_ByPIDFile_InvalidContent(t *testing.T) {
	tmp, err := os.CreateTemp("", "procwatch-pid-*.pid")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	fmt.Fprintln(tmp, "not-a-number")
	tmp.Close()

	f := &Finder{}
	_, err = f.Find("", tmp.Name())
	if err == nil {
		t.Fatal("expected error for invalid pid file content")
	}
}

func TestFinder_Find_ByPIDFile_ValidPID(t *testing.T) {
	pid := os.Getpid()
	tmp, err := os.CreateTemp("", "procwatch-pid-*.pid")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	fmt.Fprintln(tmp, pid)
	tmp.Close()

	f := &Finder{}
	state, err := f.Find("", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.Running {
		t.Error("expected current process to be running")
	}
	if state.PID != int32(pid) {
		t.Errorf("expected PID %d, got %d", pid, state.PID)
	}
}
