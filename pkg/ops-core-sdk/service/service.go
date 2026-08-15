// Package service provides systemd service management operations.
// All functions use systemctl directly via os/exec (no shell invocation).
package service

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// systemctlBin is the path to the systemctl binary.
// It is a variable so tests can override it.
var systemctlBin = "systemctl"

// ServiceStatus represents the status of a systemd service.
type ServiceStatus struct {
	Name        string `json:"name"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	LoadState   string `json:"load_state"`
	MainPID     int    `json:"main_pid"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Active      bool   `json:"active"`
}

// ServiceAction represents the result of a service action (start, stop, restart, enable).
type ServiceAction struct {
	Name    string `json:"name"`
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Status returns the current status of a systemd service.
func Status(name string) (ServiceStatus, error) {
	status := ServiceStatus{Name: name}

	// Get properties via systemctl show
	props, err := runSystemctl("show", name,
		"--property=ActiveState,SubState,LoadState,MainPID,Description")
	if err != nil {
		return status, fmt.Errorf("failed to get status for service %q: %w", name, err)
	}

	parseProperties(props, &status)

	// Check if service is enabled
	enabled, err := isEnabled(name)
	if err != nil {
		// Non-fatal: we still return the status with Enabled=false
		status.Enabled = false
	} else {
		status.Enabled = enabled
	}

	status.Active = status.ActiveState == "active"

	return status, nil
}

// Start starts a systemd service.
func Start(name string) (ServiceAction, error) {
	return doAction(name, "start")
}

// Stop stops a systemd service.
func Stop(name string) (ServiceAction, error) {
	return doAction(name, "stop")
}

// Restart restarts a systemd service.
func Restart(name string) (ServiceAction, error) {
	return doAction(name, "restart")
}

// Enable enables a systemd service to start at boot.
func Enable(name string) (ServiceAction, error) {
	return doAction(name, "enable")
}

// Disable disables a systemd service from starting at boot.
func Disable(name string) (ServiceAction, error) {
	return doAction(name, "disable")
}

// doAction executes a systemctl action (start/stop/restart/enable) on a service.
func doAction(name, action string) (ServiceAction, error) {
	result := ServiceAction{
		Name:   name,
		Action: action,
	}

	output, err := runSystemctl(action, name)
	if err != nil {
		result.Success = false
		result.Message = strings.TrimSpace(output)
		if result.Message == "" {
			result.Message = err.Error()
		}
		return result, fmt.Errorf("failed to %s service %q: %w", action, name, err)
	}

	result.Success = true
	result.Message = fmt.Sprintf("service %q %sed successfully", name, action)
	return result, nil
}

// runSystemctl executes a systemctl command with the given arguments and returns
// the combined stdout output. It does NOT use a shell.
func runSystemctl(args ...string) (string, error) {
	cmd := exec.Command(systemctlBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("systemctl %s failed: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// isEnabled checks whether a service is enabled.
func isEnabled(name string) (bool, error) {
	cmd := exec.Command(systemctlBin, "is-enabled", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	// systemctl is-enabled exits 0 when enabled, non-zero otherwise
	if err != nil {
		// "disabled", "static", "masked", etc. are not errors per se, just not enabled
		return false, nil
	}

	return output == "enabled", nil
}

// parseProperties parses the key=value output from systemctl show into a ServiceStatus.
func parseProperties(output string, status *ServiceStatus) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}

		key := line[:idx]
		value := line[idx+1:]

		switch key {
		case "ActiveState":
			status.ActiveState = value
		case "SubState":
			status.SubState = value
		case "LoadState":
			status.LoadState = value
		case "MainPID":
			pid, err := strconv.Atoi(value)
			if err == nil {
				status.MainPID = pid
			}
		case "Description":
			status.Description = value
		}
	}
}
