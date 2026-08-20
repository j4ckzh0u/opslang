// Package docker_network manages Docker networks.
package docker_network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// NetworkInfo contains information about a Docker network.
type NetworkInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Driver string `json:"driver"`
	Scope  string `json:"scope"`
}

// Result is the common return type for docker_network operations.
type Result struct {
	Network    *NetworkInfo `json:"network,omitempty"`
	Changed    bool         `json:"changed"`
	Error      string       `json:"error,omitempty"`
	DurationMs int64        `json:"duration_ms"`
}

func dockerCmd(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("docker %s: %s: %w", args[0], string(out), err)
	}
	return out, nil
}

// Inspect returns information about a Docker network.
func Inspect(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	out, err := dockerCmd("network", "inspect", name, "--format", "{{json .}}")
	if err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	var raw struct {
		Name   string `json:"Name"`
		ID     string `json:"Id"`
		Driver string `json:"Driver"`
		Scope  string `json:"Scope"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	info := &NetworkInfo{Name: raw.Name, ID: raw.ID, Driver: raw.Driver, Scope: raw.Scope}
	return Result{Network: info, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Create creates a Docker network (idempotent - skips if exists).
func Create(name, driver string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}
	if driver == "" {
		driver = "bridge"
	}

	// Check if exists
	_, err := dockerCmd("network", "inspect", name)
	if err == nil {
		// Already exists
		info := &NetworkInfo{Name: name, Driver: driver}
		return Result{Network: info, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	_, err = dockerCmd("network", "create", "--driver", driver, name)
	if err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	info := &NetworkInfo{Name: name, Driver: driver}
	return Result{Network: info, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Remove removes a Docker network (idempotent).
func Remove(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	// Check if exists
	_, err := dockerCmd("network", "inspect", name)
	if err != nil {
		// Already gone
		return Result{Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	_, err = dockerCmd("network", "rm", name)
	if err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	return Result{Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// List lists all Docker networks.
func List() ([]NetworkInfo, error) {
	out, err := dockerCmd("network", "ls", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	var networks []NetworkInfo
	lines := splitLines(string(out))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var raw struct {
			Name   string `json:"Name"`
			ID     string `json:"ID"`
			Driver string `json:"Driver"`
			Scope  string `json:"Scope"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		networks = append(networks, NetworkInfo{Name: raw.Name, ID: raw.ID, Driver: raw.Driver, Scope: raw.Scope})
	}
	return networks, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
