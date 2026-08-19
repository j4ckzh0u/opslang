// Package supervisor provides supervisorctl management.
package supervisor

import (
	"fmt"
	"os/exec"
	"strings"
)

// ProcessInfo represents a supervisor process.
type ProcessInfo struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Status string `json:"status,omitempty"`
	Uptime string `json:"uptime,omitempty"`
}

// Result is returned by process operations.
type Result struct {
	Process string `json:"process,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// StatusResult is returned by status queries.
type StatusResult struct {
	Processes []ProcessInfo `json:"processes"`
	Count     int           `json:"count"`
	Error     string        `json:"error,omitempty"`
}

func supervisorctl(args ...string) (string, error) {
	cmd := exec.Command("supervisorctl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Start starts a process.
func Start(name string) Result {
	if name == "" {
		return Result{Error: "process name is required"}
	}
	out, err := supervisorctl("start", name)
	if err != nil {
		return Result{Process: name, Error: fmt.Sprintf("start failed: %s: %s", err, out)}
	}
	return Result{Process: name, Success: true, Changed: true}
}

// Stop stops a process.
func Stop(name string) Result {
	if name == "" {
		return Result{Error: "process name is required"}
	}
	out, err := supervisorctl("stop", name)
	if err != nil {
		return Result{Process: name, Error: fmt.Sprintf("stop failed: %s: %s", err, out)}
	}
	return Result{Process: name, Success: true, Changed: true}
}

// Restart restarts a process.
func Restart(name string) Result {
	if name == "" {
		return Result{Error: "process name is required"}
	}
	out, err := supervisorctl("restart", name)
	if err != nil {
		return Result{Process: name, Error: fmt.Sprintf("restart failed: %s: %s", err, out)}
	}
	return Result{Process: name, Success: true, Changed: true}
}

// Reload reloads supervisor configuration.
func Reload() Result {
	out, err := supervisorctl("reload")
	if err != nil {
		return Result{Error: fmt.Sprintf("reload failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true}
}

// Status returns status of all processes.
func Status() StatusResult {
	out, err := supervisorctl("status")
	if err != nil {
		return StatusResult{Error: fmt.Sprintf("status failed: %s: %s", err, out)}
	}
	var procs []ProcessInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "name          STATE    uptime info"
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			p := ProcessInfo{
				Name:  fields[0],
				State: fields[1],
			}
			if len(fields) >= 3 {
				p.Uptime = fields[2]
			}
			if len(fields) >= 4 {
				p.Status = strings.Join(fields[3:], " ")
			}
			procs = append(procs, p)
		}
	}
	return StatusResult{Processes: procs, Count: len(procs)}
}

// ClearLog clears the log for a process.
func ClearLog(name string) Result {
	if name == "" {
		return Result{Error: "process name is required"}
	}
	out, err := supervisorctl("clearlog", name)
	if err != nil {
		return Result{Process: name, Error: fmt.Sprintf("clearlog failed: %s: %s", err, out)}
	}
	return Result{Process: name, Success: true, Changed: true}
}

// Reread rereads the supervisor configuration.
func Reread() Result {
	out, err := supervisorctl("reread")
	if err != nil {
		return Result{Error: fmt.Sprintf("reread failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true}
}

// Update updates the supervisor process group.
func Update(name string) Result {
	args := []string{"update"}
	if name != "" {
		args = append(args, name)
	}
	out, err := supervisorctl(args...)
	if err != nil {
		return Result{Process: name, Error: fmt.Sprintf("update failed: %s: %s", err, out)}
	}
	return Result{Process: name, Success: true, Changed: true}
}
