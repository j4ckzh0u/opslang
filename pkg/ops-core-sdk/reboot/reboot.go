// Package reboot manages system reboot operations.
// Equivalent to ansible.builtin.reboot module.
package reboot

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Result is returned by all reboot functions.
type Result struct {
	Status     string `json:"status"`
	Rebooted   bool   `json:"rebooted"`
	Command    string `json:"command,omitempty"`
	ElapsedMs  int64  `json:"elapsed_ms"`
	Error      string `json:"error,omitempty"`
}

// CheckResult is returned by Check.
type CheckResult struct {
	Status   string `json:"status"`
	Booted   bool   `json:"booted"`
	Uptime   string `json:"uptime,omitempty"`
	Since    string `json:"since,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Request initiates a system reboot.
// msg: optional message to logged-in users.
// delay: seconds to wait before rebooting.
func Request(msg string, delay int) Result {
	start := time.Now()

	if runtime.GOOS != "linux" {
		return Result{Status: "failed", Error: fmt.Sprintf("reboot not supported on %s", runtime.GOOS)}
	}

	// Try systemctl first, then fall back to reboot command
	var cmd *exec.Cmd
	if msg != "" {
		// Use wall to notify users, then reboot
		wallCmd := exec.Command("wall", fmt.Sprintf("System rebooting: %s", msg))
		wallCmd.Run() // ignore error if wall not available
	}

	if delay > 0 {
		cmd = exec.Command("shutdown", "-r", fmt.Sprintf("+%d", delay/60))
	} else {
		// Check if systemctl is available
		if _, err := exec.LookPath("systemctl"); err == nil {
			cmd = exec.Command("systemctl", "reboot")
		} else {
			cmd = exec.Command("reboot")
		}
	}

	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		// Reboot command often exits with signal, which is expected
		errStr := strings.TrimSpace(string(out))
		if errStr == "" {
			errStr = err.Error()
		}
		return Result{
			Status:    "failed",
			Command:   strings.Join(cmd.Args, " "),
			ElapsedMs: elapsed,
			Error:     errStr,
		}
	}

	return Result{
		Status:    "success",
		Rebooted:  true,
		Command:   strings.Join(cmd.Args, " "),
		ElapsedMs: elapsed,
	}
}

// DryRun returns what command would be executed without actually rebooting.
func DryRun(msg string, delay int) Result {
	var command string
	if delay > 0 {
		command = fmt.Sprintf("shutdown -r +%d", delay/60)
	} else {
		command = "systemctl reboot (or reboot)"
	}

	if msg != "" {
		command = fmt.Sprintf("wall '%s' && %s", msg, command)
	}

	return Result{
		Status:  "success",
		Command: command,
	}
}

// Check checks if the system has recently booted and returns uptime info.
func Check() CheckResult {
	if runtime.GOOS != "linux" {
		return CheckResult{Status: "failed", Error: fmt.Sprintf("not supported on %s", runtime.GOOS)}
	}

	// Get uptime from /proc/uptime
	out, err := exec.Command("uptime", "-s").Output()
	if err != nil {
		// Try alternative
		out2, err2 := exec.Command("uptime").Output()
		if err2 != nil {
			return CheckResult{Status: "failed", Error: fmt.Sprintf("uptime: %v", err2)}
		}
		return CheckResult{
			Status: "success",
			Booted: true,
			Uptime: strings.TrimSpace(string(out2)),
		}
	}

	bootTime := strings.TrimSpace(string(out))
	return CheckResult{
		Status: "success",
		Booted: true,
		Since:  bootTime,
	}
}
