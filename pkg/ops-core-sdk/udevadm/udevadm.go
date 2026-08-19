// Package udevadm provides functions for udev device manager administration.
package udevadm

import (
	"fmt"
	"os/exec"
	"strings"
)

// ControlResult represents the result of udevadm control operation.
type ControlResult struct {
	Action string `json:"action"`
	Status string `json:"status"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// TriggerResult represents the result of udevadm trigger operation.
type TriggerResult struct {
	Subsystem string `json:"subsystem,omitempty"`
	Action    string `json:"action"`
	Count     int    `json:"count"`
	Error     string `json:"error,omitempty"`
}

// SettleResult represents the result of udevadm settle operation.
type SettleResult struct {
	Timeout int    `json:"timeout"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// InfoResult represents the result of udevadm info operation.
type InfoResult struct {
	Query    string            `json:"query"`
	Name     string            `json:"name,omitempty"`
	Attrs    map[string]string `json:"attributes,omitempty"`
	Env      map[string]string `json:"environment,omitempty"`
	Error    string            `json:"error,omitempty"`
}

// MonitorResult represents the result of udevadm monitor operation.
type MonitorResult struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// Control manages the udev daemon.
// action: reload, reload-rules, stop-exec-queue, start-exec-queue, log-priority
func Control(action string) ControlResult {
	if action == "" {
		return ControlResult{Error: "action is required"}
	}

	cmd := exec.Command("udevadm", "control", "--"+action)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ControlResult{
			Action: action,
			Status: "failed",
			Error:  fmt.Sprintf("udevadm control failed: %s: %s", err, string(out)),
		}
	}

	return ControlResult{
		Action: action,
		Status: "success",
		Output: strings.TrimSpace(string(out)),
	}
}

// Trigger triggers device events.
// subsystem: optional subsystem filter (e.g., "block", "net")
func Trigger(subsystem string) TriggerResult {
	args := []string{"trigger"}
	if subsystem != "" {
		args = append(args, "--subsystem-match="+subsystem)
	}

	cmd := exec.Command("udevadm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return TriggerResult{
			Subsystem: subsystem,
			Action:    "trigger",
			Error:     fmt.Sprintf("udevadm trigger failed: %s: %s", err, string(out)),
		}
	}

	return TriggerResult{
		Subsystem: subsystem,
		Action:    "trigger",
		Count:     1,
	}
}

// Settle waits for pending udev events.
// timeout: maximum time to wait in seconds
func Settle(timeout int) SettleResult {
	if timeout <= 0 {
		timeout = 120
	}

	args := []string{"settle", "--timeout=" + fmt.Sprintf("%d", timeout)}
	cmd := exec.Command("udevadm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return SettleResult{
			Timeout: timeout,
			Status:  "failed",
			Error:   fmt.Sprintf("udevadm settle failed: %s: %s", err, string(out)),
		}
	}

	return SettleResult{
		Timeout: timeout,
		Status:  "success",
	}
}

// Info queries udev database for device information.
// query: query type (name, path, property, all)
// device: device path or name
func Info(query, device string) InfoResult {
	if device == "" {
		return InfoResult{Error: "device is required"}
	}
	if query == "" {
		query = "all"
	}

	args := []string{"info", "--query=" + query, "--name=" + device}
	cmd := exec.Command("udevadm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return InfoResult{
			Query: query,
			Error: fmt.Sprintf("udevadm info failed: %s: %s", err, string(out)),
		}
	}

	result := InfoResult{
		Query: query,
		Name:  device,
	}

	// Parse output based on query type
	output := strings.TrimSpace(string(out))
	if query == "property" || query == "all" {
		result.Env = make(map[string]string)
		for _, line := range strings.Split(output, "\n") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				result.Env[parts[0]] = parts[1]
			}
		}
	} else {
		result.Attrs = map[string]string{"output": output}
	}

	return result
}

// Monitor monitors kernel uevents and udev events.
// Returns captured output (in real usage would be streaming).
func Monitor() MonitorResult {
	// Note: monitor is typically a long-running process
	// This implementation captures a snapshot
	cmd := exec.Command("udevadm", "monitor", "--environment", "--udev")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return MonitorResult{
			Error: fmt.Sprintf("udevadm monitor failed: %s: %s", err, string(out)),
		}
	}

	return MonitorResult{
		Output: strings.TrimSpace(string(out)),
	}
}
