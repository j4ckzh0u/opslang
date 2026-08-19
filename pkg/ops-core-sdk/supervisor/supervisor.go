// Package supervisor manages Supervisor processes.
// Equivalent to community.general.supervisorctl module.
package supervisor

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Service string `json:"service,omitempty"`
	Action  string `json:"action,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StatusResult is returned by Status.
type StatusResult struct {
	Status   string        `json:"status"`
	Services []ServiceInfo `json:"services"`
	Error    string        `json:"error,omitempty"`
}

// ServiceInfo represents a supervisor service.
type ServiceInfo struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	PID    string `json:"pid,omitempty"`
	Uptime string `json:"uptime,omitempty"`
}

// Start starts a supervisor process.
func Start(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}
	}
	cmd := exec.Command("supervisorctl", "start", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl start: %v", err)}
	}
	return Result{Status: "success", Changed: true, Service: name, Action: "start", Output: output}
}

// Stop stops a supervisor process.
func Stop(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}
	}
	cmd := exec.Command("supervisorctl", "stop", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl stop: %v", err)}
	}
	return Result{Status: "success", Changed: true, Service: name, Action: "stop", Output: output}
}

// Restart restarts a supervisor process.
func Restart(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}
	}
	cmd := exec.Command("supervisorctl", "restart", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl restart: %v", err)}
	}
	return Result{Status: "success", Changed: true, Service: name, Action: "restart", Output: output}
}

// Reload reloads the supervisor daemon configuration.
func Reload() Result {
	cmd := exec.Command("supervisorctl", "reload")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl reload: %v", err)}
	}
	return Result{Status: "success", Changed: true, Action: "reload", Output: output}
}

// Status returns the status of all supervisor processes.
func Status() StatusResult {
	cmd := exec.Command("supervisorctl", "status")
	out, err := cmd.Output()
	if err != nil {
		return StatusResult{Status: "failed", Error: fmt.Sprintf("supervisorctl status: %v", err)}
	}

	services := make([]ServiceInfo, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			svc := ServiceInfo{Name: parts[0], State: parts[1]}
			if len(parts) >= 3 {
				svc.PID = parts[2]
			}
			if len(parts) >= 4 {
				svc.Uptime = parts[3]
			}
			services = append(services, svc)
		}
	}
	return StatusResult{Status: "success", Services: services}
}

// ClearLog clears the log for a specific supervisor process.
func ClearLog(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}
	}
	cmd := exec.Command("supervisorctl", "clear", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl clear: %v", err)}
	}
	return Result{Status: "success", Changed: true, Service: name, Action: "clear_log", Output: output}
}

// Reread re-reads supervisor configuration without applying changes.
func Reread() Result {
	cmd := exec.Command("supervisorctl", "reread")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl reread: %v", err)}
	}
	return Result{Status: "success", Changed: true, Action: "reread", Output: output}
}

// Update re-reads config and applies changes for a specific process.
func Update(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}
	}
	cmd := exec.Command("supervisorctl", "update", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("supervisorctl update: %v", err)}
	}
	return Result{Status: "success", Changed: true, Service: name, Action: "update", Output: output}
}
