//go:build !windows

package process

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// startSleepProcess spawns a real child process that sleeps for a long time.
// It returns the PID and a channel that closes when the child has been
// reaped (without reaping, a dead child stays as a zombie and kill(pid, 0)
// keeps succeeding, which would mask a real termination).
func startSleepProcess(t *testing.T) (int, <-chan struct{}) {
	t.Helper()
	cmd := exec.Command("sleep", "300")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start child process: %v", err)
	}
	exited := make(chan struct{})
	go func() {
		_, _ = cmd.Process.Wait()
		close(exited)
	}()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		select {
		case <-exited:
		case <-time.After(2 * time.Second):
		}
	})
	return cmd.Process.Pid, exited
}

func TestKillTerminatesProcess(t *testing.T) {
	pid, exited := startSleepProcess(t)

	res, err := Kill(pid, "TERM")
	if err != nil {
		t.Fatalf("Kill error: %v", err)
	}
	if !res.Sent {
		t.Error("Sent = false, want true")
	}
	if res.Signal != "TERM" {
		t.Errorf("Signal = %q, want TERM", res.Signal)
	}

	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatalf("process %d still alive 2s after SIGTERM", pid)
	}
}

func TestKillDefaultSignalIsTerm(t *testing.T) {
	pid, exited := startSleepProcess(t)

	res, err := Kill(pid, "")
	if err != nil {
		t.Fatalf("Kill error: %v", err)
	}
	if res.Signal != "TERM" {
		t.Errorf("Signal = %q, want TERM (default)", res.Signal)
	}

	select {
	case <-exited:
	case <-time.After(2 * time.Second):
		t.Fatalf("process %d still alive after default-TERM kill", pid)
	}
}

func TestKillUnknownSignal(t *testing.T) {
	pid, exited := startSleepProcess(t)

	_, err := Kill(pid, "NOSUCHSIGNAL")
	if err == nil {
		t.Error("Kill with an unknown signal should fail")
	}

	// The failed kill must not have terminated the child.
	select {
	case <-exited:
		t.Fatal("child exited even though the Kill call failed")
	case <-time.After(200 * time.Millisecond):
		// still running, as expected
	}
}

func TestKillInvalidPid(t *testing.T) {
	// PID 300000000 is beyond typical pid_max on Linux and invalid on macOS.
	_, err := Kill(300000000, "TERM")
	if err == nil {
		t.Error("Kill on an invalid PID should fail")
	}
}

var _ = syscall.Kill // keep import when build tags change
