// Package runit provides runit service management operations.
package runit

import (
	"fmt"
	"os/exec"
	"strings"
)

// StatusResult represents the result of checking service status.
type StatusResult struct {
	Service string `json:"service"`
	Running bool   `json:"running"`
	State   string `json:"state"`
}

// ActionResult represents the result of a service action.
type ActionResult struct {
	Service string `json:"service"`
	Changed bool   `json:"changed"`
	State   string `json:"state"`
	Error   string `json:"error,omitempty"`
}

// ListResult represents the result of listing services.
type ListResult struct {
	Services []string `json:"services"`
	Count    int      `json:"count"`
}

// Status checks the status of a runit service.
func Status(service string) (*StatusResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	cmd := exec.Command("sv", "status", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &StatusResult{
			Service: service,
			Running: false,
			State:   "unknown",
		}, nil
	}

	output := strings.TrimSpace(string(out))
	running := strings.Contains(output, "run:") || strings.Contains(output, "up")
	state := "stopped"
	if running {
		state = "running"
	}

	return &StatusResult{
		Service: service,
		Running: running,
		State:   state,
	}, nil
}

// Start starts a runit service.
func Start(service string) (*ActionResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Check current status
	status, _ := Status(service)
	if status != nil && status.Running {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "running",
		}, nil
	}

	cmd := exec.Command("sv", "start", service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "failed",
			Error:   string(output),
		}, fmt.Errorf("start runit service: %s", output)
	}

	return &ActionResult{
		Service: service,
		Changed: true,
		State:   "running",
	}, nil
}

// Stop stops a runit service.
func Stop(service string) (*ActionResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Check current status
	status, _ := Status(service)
	if status != nil && !status.Running {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "stopped",
		}, nil
	}

	cmd := exec.Command("sv", "stop", service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "failed",
			Error:   string(output),
		}, fmt.Errorf("stop runit service: %s", output)
	}

	return &ActionResult{
		Service: service,
		Changed: true,
		State:   "stopped",
	}, nil
}

// Restart restarts a runit service.
func Restart(service string) (*ActionResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	cmd := exec.Command("sv", "restart", service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "failed",
			Error:   string(output),
		}, fmt.Errorf("restart runit service: %s", output)
	}

	return &ActionResult{
		Service: service,
		Changed: true,
		State:   "running",
	}, nil
}

// Reload reloads a runit service.
func Reload(service string) (*ActionResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	cmd := exec.Command("sv", "reload", service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "failed",
			Error:   string(output),
		}, fmt.Errorf("reload runit service: %s", output)
	}

	return &ActionResult{
		Service: service,
		Changed: true,
		State:   "running",
	}, nil
}

// Enable enables a runit service (creates symlink in /var/service/).
func Enable(service, serviceDir string) (*ActionResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if serviceDir == "" {
		serviceDir = "/etc/sv"
	}

	// Check if already enabled
	if _, err := exec.LookPath("ls"); err == nil {
		cmd := exec.Command("ls", "-d", "/var/service/"+service)
		if err := cmd.Run(); err == nil {
			return &ActionResult{
				Service: service,
				Changed: false,
				State:   "enabled",
			}, nil
		}
	}

	// Create symlink
	cmd := exec.Command("ln", "-s", serviceDir+"/"+service, "/var/service/"+service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "failed",
			Error:   string(output),
		}, fmt.Errorf("enable runit service: %s", output)
	}

	return &ActionResult{
		Service: service,
		Changed: true,
		State:   "enabled",
	}, nil
}

// Disable disables a runit service (removes symlink from /var/service/).
func Disable(service string) (*ActionResult, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Check if enabled
	cmd := exec.Command("ls", "-d", "/var/service/"+service)
	if err := cmd.Run(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "disabled",
		}, nil
	}

	// Remove symlink
	cmd = exec.Command("rm", "/var/service/"+service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &ActionResult{
			Service: service,
			Changed: false,
			State:   "failed",
			Error:   string(output),
		}, fmt.Errorf("disable runit service: %s", output)
	}

	return &ActionResult{
		Service: service,
		Changed: true,
		State:   "disabled",
	}, nil
}

// List lists all runit services.
func List() (*ListResult, error) {
	cmd := exec.Command("ls", "/var/service")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ListResult{
			Services: []string{},
			Count:    0,
		}, nil
	}

	services := strings.Fields(strings.TrimSpace(string(out)))
	return &ListResult{
		Services: services,
		Count:    len(services),
	}, nil
}
