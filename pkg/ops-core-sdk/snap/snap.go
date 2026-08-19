// Package snap manages snap packages.
// Equivalent to community.general.snap module.
package snap

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Package string `json:"package,omitempty"`
	Channel string `json:"channel,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status string   `json:"status"`
	Snaps  []string `json:"snaps"`
	Error  string   `json:"error,omitempty"`
}

// Install installs a snap package.
func Install(name string, channel string, classic bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	args := []string{"install", name}
	if channel != "" {
		args = append(args, "--channel="+channel)
	}
	if classic {
		args = append(args, "--classic")
	}

	cmd := exec.Command("snap", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("snap install: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Channel: channel, Output: output}, nil
}

// Remove removes a snap package.
func Remove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	cmd := exec.Command("snap", "remove", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("snap remove: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// Refresh updates a snap package.
func Refresh(name string, channel string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}

	args := []string{"refresh", name}
	if channel != "" {
		args = append(args, "--channel="+channel)
	}

	cmd := exec.Command("snap", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("snap refresh: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Channel: channel, Output: output}, nil
}

// List lists installed snaps.
func List() (ListResult, error) {
	cmd := exec.Command("snap", "list")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("snap list: %v", err)}, err
	}

	snaps := make([]string, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i == 0 { // skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line != "" {
			snaps = append(snaps, line)
		}
	}
	return ListResult{Status: "success", Snaps: snaps}, nil
}

// Enable enables a snap.
func Enable(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	cmd := exec.Command("snap", "enable", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("snap enable: %v: %s", err, strings.TrimSpace(string(out)))}, err
	}
	return Result{Status: "success", Changed: true, Package: name}, nil
}

// Disable disables a snap.
func Disable(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	cmd := exec.Command("snap", "disable", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("snap disable: %v: %s", err, strings.TrimSpace(string(out)))}, err
	}
	return Result{Status: "success", Changed: true, Package: name}, nil
}

// Get returns information about a snap.
func Get(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	cmd := exec.Command("snap", "info", name)
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("snap info: %v", err)}, err
	}
	return Result{Status: "success", Package: name, Output: strings.TrimSpace(string(out))}, nil
}

// Switch switches a snap to a different channel.
func Switch(name string, channel string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	if channel == "" {
		return Result{Status: "failed", Error: "channel is required"}, fmt.Errorf("channel is required")
	}
	cmd := exec.Command("snap", "switch", "--channel="+channel, name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("snap switch: %v: %s", err, strings.TrimSpace(string(out)))}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Channel: channel}, nil
}

// Changes returns recent snap changes.
func Changes() (ListResult, error) {
	cmd := exec.Command("snap", "changes")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("snap changes: %v", err)}, err
	}
	changes := make([]string, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i == 0 { // skip header
			continue
		}
		line = strings.TrimSpace(line)
		if line != "" {
			changes = append(changes, line)
		}
	}
	return ListResult{Status: "success", Snaps: changes}, nil
}
