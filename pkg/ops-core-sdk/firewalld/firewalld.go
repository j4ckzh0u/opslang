// Package firewalld provides firewalld service management.
package firewalld

import (
	"fmt"
	"os/exec"
	"strings"
)

// Status represents firewalld status.
type Status struct {
	Running bool   `json:"running"`
	Enabled bool   `json:"enabled"`
	Default string `json:"default_zone"`
}

// GetResult is returned by Get.
type GetResult struct {
	Status Status `json:"status"`
}

// ServiceResult is returned by service operations.
type ServiceResult struct {
	Changed bool   `json:"changed"`
	State   string `json:"state"`
	Error   string `json:"error,omitempty"`
}

// ZoneResult is returned by zone operations.
type ZoneResult struct {
	Zones []string `json:"zones"`
}

// Get returns firewalld status.
func Get() (GetResult, error) {
	status := Status{}

	// Check if running
	cmd := exec.Command("systemctl", "is-active", "firewalld")
	out, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "active" {
		status.Running = true
	}

	// Check if enabled
	cmd = exec.Command("systemctl", "is-enabled", "firewalld")
	out, err = cmd.Output()
	if err == nil && strings.TrimSpace(string(out)) == "enabled" {
		status.Enabled = true
	}

	// Get default zone
	if status.Running {
		cmd = exec.Command("firewall-cmd", "--get-default-zone")
		out, err = cmd.Output()
		if err == nil {
			status.Default = strings.TrimSpace(string(out))
		}
	}

	return GetResult{Status: status}, nil
}

// Start starts the firewalld service.
func Start() (ServiceResult, error) {
	result, _ := Get()
	if result.Status.Running {
		return ServiceResult{Changed: false, State: "running"}, nil
	}

	cmd := exec.Command("systemctl", "start", "firewalld")
	if output, err := cmd.CombinedOutput(); err != nil {
		return ServiceResult{Error: string(output), State: "failed"}, fmt.Errorf("start firewalld: %s", output)
	}

	return ServiceResult{Changed: true, State: "running"}, nil
}

// Stop stops the firewalld service.
func Stop() (ServiceResult, error) {
	result, _ := Get()
	if !result.Status.Running {
		return ServiceResult{Changed: false, State: "stopped"}, nil
	}

	cmd := exec.Command("systemctl", "stop", "firewalld")
	if output, err := cmd.CombinedOutput(); err != nil {
		return ServiceResult{Error: string(output), State: "failed"}, fmt.Errorf("stop firewalld: %s", output)
	}

	return ServiceResult{Changed: true, State: "stopped"}, nil
}

// Restart restarts the firewalld service.
func Restart() (ServiceResult, error) {
	cmd := exec.Command("systemctl", "restart", "firewalld")
	if output, err := cmd.CombinedOutput(); err != nil {
		return ServiceResult{Error: string(output), State: "failed"}, fmt.Errorf("restart firewalld: %s", output)
	}

	return ServiceResult{Changed: true, State: "running"}, nil
}

// Enable enables firewalld to start at boot.
func Enable() (ServiceResult, error) {
	result, _ := Get()
	if result.Status.Enabled {
		return ServiceResult{Changed: false, State: "enabled"}, nil
	}

	cmd := exec.Command("systemctl", "enable", "firewalld")
	if output, err := cmd.CombinedOutput(); err != nil {
		return ServiceResult{Error: string(output), State: "failed"}, fmt.Errorf("enable firewalld: %s", output)
	}

	return ServiceResult{Changed: true, State: "enabled"}, nil
}

// Disable disables firewalld from starting at boot.
func Disable() (ServiceResult, error) {
	result, _ := Get()
	if !result.Status.Enabled {
		return ServiceResult{Changed: false, State: "disabled"}, nil
	}

	cmd := exec.Command("systemctl", "disable", "firewalld")
	if output, err := cmd.CombinedOutput(); err != nil {
		return ServiceResult{Error: string(output), State: "failed"}, fmt.Errorf("disable firewalld: %s", output)
	}

	return ServiceResult{Changed: true, State: "disabled"}, nil
}

// ListZones returns all available zones.
func ListZones() (ZoneResult, error) {
	cmd := exec.Command("firewall-cmd", "--get-zones")
	out, err := cmd.Output()
	if err != nil {
		return ZoneResult{}, fmt.Errorf("list zones: %w", err)
	}

	zones := strings.Fields(strings.TrimSpace(string(out)))
	return ZoneResult{Zones: zones}, nil
}

// Reload reloads firewalld configuration.
func Reload() (ServiceResult, error) {
	cmd := exec.Command("firewall-cmd", "--reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return ServiceResult{Error: string(output), State: "failed"}, fmt.Errorf("reload firewalld: %s", output)
	}

	return ServiceResult{Changed: true, State: "reloaded"}, nil
}
