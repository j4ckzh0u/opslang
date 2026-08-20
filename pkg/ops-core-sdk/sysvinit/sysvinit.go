// Package sysvinit provides System V init service management operations.
// Equivalent to Ansible's sysvinit module.
package sysvinit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// StatusResult represents service status.
type StatusResult struct {
	Status    string `json:"status"`
	Name      string `json:"name"`
	Running   bool   `json:"running"`
	Enabled   bool   `json:"enabled"`
	PID       int    `json:"pid,omitempty"`
	State     string `json:"state"` // running, stopped, unknown
	ScriptPath string `json:"script_path,omitempty"`
}

// Result represents a service operation result.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Name    string `json:"name"`
	Action  string `json:"action"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

func findScript(name string) (string, error) {
	candidates := []string{
		filepath.Join("/etc/init.d", name),
		filepath.Join("/etc/rc.d", name),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("init script for %q not found", name)
}

func runScript(script, action string) (string, error) {
	cmd := exec.Command(script, action)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Status returns the status of a service.
func Status(name string) (StatusResult, error) {
	if name == "" {
		return StatusResult{Status: "failed"}, fmt.Errorf("service name is required")
	}
	script, err := findScript(name)
	if err != nil {
		return StatusResult{Status: "failed", Name: name}, err
	}
	result := StatusResult{Name: name, ScriptPath: script, Status: "success"}

	out, err := runScript(script, "status")
	if err != nil {
		result.Running = false
		result.State = "stopped"
	} else {
		result.Running = true
		result.State = "running"
		lower := strings.ToLower(out)
		if strings.Contains(lower, "pid") || strings.Contains(lower, "running") {
			result.Running = true
		}
	}

	// Check if enabled (look for rc.d symlinks)
	for _, rcDir := range []string{"/etc/rc2.d", "/etc/rc3.d", "/etc/rc5.d", "/etc/rc.d/rc3.d"} {
		matches, _ := filepath.Glob(filepath.Join(rcDir, "S*"+name))
		if len(matches) > 0 {
			result.Enabled = true
			break
		}
	}

	return result, nil
}

// Start starts a service.
func Start(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}, fmt.Errorf("service name is required")
	}
	script, err := findScript(name)
	if err != nil {
		return Result{Status: "failed", Name: name, Error: err.Error()}, err
	}
	out, err := runScript(script, "start")
	if err != nil {
		return Result{Status: "failed", Name: name, Action: "start", Output: out, Error: fmt.Sprintf("start: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Name: name, Action: "start", Output: out}, nil
}

// Stop stops a service.
func Stop(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}, fmt.Errorf("service name is required")
	}
	script, err := findScript(name)
	if err != nil {
		return Result{Status: "failed", Name: name, Error: err.Error()}, err
	}
	out, err := runScript(script, "stop")
	if err != nil {
		return Result{Status: "failed", Name: name, Action: "stop", Output: out, Error: fmt.Sprintf("stop: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Name: name, Action: "stop", Output: out}, nil
}

// Restart restarts a service.
func Restart(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}, fmt.Errorf("service name is required")
	}
	script, err := findScript(name)
	if err != nil {
		return Result{Status: "failed", Name: name, Error: err.Error()}, err
	}
	out, err := runScript(script, "restart")
	if err != nil {
		return Result{Status: "failed", Name: name, Action: "restart", Output: out, Error: fmt.Sprintf("restart: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Name: name, Action: "restart", Output: out}, nil
}

// Reload reloads a service.
func Reload(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}, fmt.Errorf("service name is required")
	}
	script, err := findScript(name)
	if err != nil {
		return Result{Status: "failed", Name: name, Error: err.Error()}, err
	}
	out, err := runScript(script, "reload")
	if err != nil {
		return Result{Status: "failed", Name: name, Action: "reload", Output: out, Error: fmt.Sprintf("reload: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Name: name, Action: "reload", Output: out}, nil
}

// Enable enables a service at boot.
func Enable(name string, runlevels string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}, fmt.Errorf("service name is required")
	}
	if runlevels == "" {
		runlevels = "2345"
	}
	script, err := findScript(name)
	if err != nil {
		return Result{Status: "failed", Name: name, Error: err.Error()}, err
	}

	// Try chkconfig first, then update-rc.d
	if chkconfig, err := exec.LookPath("chkconfig"); err == nil {
		cmd := exec.Command(chkconfig, "--add", name)
		cmd.CombinedOutput()
		cmd = exec.Command(chkconfig, name, "on")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return Result{Status: "failed", Name: name, Action: "enable", Output: string(out), Error: fmt.Sprintf("chkconfig: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Name: name, Action: "enable", Output: string(out)}, nil
	}

	if updaterecd, err := exec.LookPath("update-rc.d"); err == nil {
		cmd := exec.Command(updaterecd, name, "defaults")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return Result{Status: "failed", Name: name, Action: "enable", Output: string(out), Error: fmt.Sprintf("update-rc.d: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Name: name, Action: "enable", Output: string(out)}, nil
	}

	// Manual symlink creation
	for _, rl := range runlevels {
		rcDir := fmt.Sprintf("/etc/rc%s.d", string(rl))
		if _, err := os.Stat(rcDir); err != nil {
			continue
		}
		link := filepath.Join(rcDir, "S20"+name)
		if _, err := os.Lstat(link); err == nil {
			continue
		}
		if err := os.Symlink(script, link); err != nil {
			return Result{Status: "failed", Name: name, Action: "enable", Error: fmt.Sprintf("symlink: %v", err)}, err
		}
	}

	return Result{Status: "success", Changed: true, Name: name, Action: "enable"}, nil
}

// Disable disables a service at boot.
func Disable(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "service name is required"}, fmt.Errorf("service name is required")
	}

	// Try chkconfig first
	if chkconfig, err := exec.LookPath("chkconfig"); err == nil {
		cmd := exec.Command(chkconfig, name, "off")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return Result{Status: "failed", Name: name, Action: "disable", Output: string(out), Error: fmt.Sprintf("chkconfig: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Name: name, Action: "disable", Output: string(out)}, nil
	}

	if updaterecd, err := exec.LookPath("update-rc.d"); err == nil {
		cmd := exec.Command(updaterecd, name, "remove")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return Result{Status: "failed", Name: name, Action: "disable", Output: string(out), Error: fmt.Sprintf("update-rc.d: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Name: name, Action: "disable", Output: string(out)}, nil
	}

	// Remove symlinks manually
	for _, dir := range []string{"/etc/rc0.d", "/etc/rc1.d", "/etc/rc2.d", "/etc/rc3.d", "/etc/rc4.d", "/etc/rc5.d", "/etc/rc6.d"} {
		matches, _ := filepath.Glob(filepath.Join(dir, "S*"+name))
		for _, m := range matches {
			os.Remove(m)
		}
	}

	return Result{Status: "success", Changed: true, Name: name, Action: "disable"}, nil
}

// List lists available init services.
func List() ([]string, error) {
	var services []string
	dirs := []string{"/etc/init.d", "/etc/rc.d"}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				services = append(services, e.Name())
			}
		}
		break // Only use the first found directory
	}
	return services, nil
}
