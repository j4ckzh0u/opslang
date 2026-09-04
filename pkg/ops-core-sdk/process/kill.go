// Kill sends a signal to a process by PID.
package process

import (
	"fmt"
)

// KillResult is returned by Kill, reporting whether the signal was sent.
type KillResult struct {
	Pid    int    `json:"pid"`
	Signal string `json:"signal"`
	Sent   bool   `json:"sent"`
}

// signals maps signal names accepted by the DSL to syscall signal numbers.
// Kill sends the named signal (default "TERM") to the process with the
// given PID. It uses the Go runtime's signal support directly — no shell.
// Note: on Windows only KILL and INT are supported by the Go runtime.
func Kill(pid int, signalName string) (KillResult, error) {
	result := KillResult{Pid: pid, Signal: signalName}

	name := signalName
	if name == "" {
		name = "TERM"
	}
	result.Signal = name

	if !supportedSignal(name) {
		return result, fmt.Errorf("process.Kill: unknown signal %q (supported: TERM, KILL, HUP, INT, USR1, USR2)", signalName)
	}

	if err := killProcess(pid, name); err != nil {
		return result, fmt.Errorf("process.Kill: %w", err)
	}

	result.Sent = true
	return result, nil
}
