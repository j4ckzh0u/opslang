// Package docker_volume manages Docker volumes.
package docker_volume

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// VolumeInfo contains information about a Docker volume.
type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// Result is the common return type for docker_volume operations.
type Result struct {
	Volume     *VolumeInfo `json:"volume,omitempty"`
	Changed    bool        `json:"changed"`
	Error      string      `json:"error,omitempty"`
	DurationMs int64       `json:"duration_ms"`
}

func dockerCmd(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("docker %s: %s: %w", args[0], string(out), err)
	}
	return out, nil
}

// Inspect returns information about a Docker volume.
func Inspect(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	out, err := dockerCmd("volume", "inspect", name, "--format", "{{json .}}")
	if err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	var raw struct {
		Name       string            `json:"Name"`
		Driver     string            `json:"Driver"`
		Mountpoint string            `json:"Mountpoint"`
		Labels     map[string]string `json:"Labels"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	info := &VolumeInfo{Name: raw.Name, Driver: raw.Driver, Mountpoint: raw.Mountpoint, Labels: raw.Labels}
	return Result{Volume: info, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Create creates a Docker volume (idempotent - skips if exists).
func Create(name, driver string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}
	if driver == "" {
		driver = "local"
	}

	// Check if exists
	_, err := dockerCmd("volume", "inspect", name)
	if err == nil {
		info := &VolumeInfo{Name: name, Driver: driver}
		return Result{Volume: info, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	_, err = dockerCmd("volume", "create", "--driver", driver, name)
	if err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	info := &VolumeInfo{Name: name, Driver: driver}
	return Result{Volume: info, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Remove removes a Docker volume (idempotent).
func Remove(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	// Check if exists
	_, err := dockerCmd("volume", "inspect", name)
	if err != nil {
		return Result{Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	_, err = dockerCmd("volume", "rm", name)
	if err != nil {
		return Result{Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}

	return Result{Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// List lists all Docker volumes.
func List() ([]VolumeInfo, error) {
	out, err := dockerCmd("volume", "ls", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	var volumes []VolumeInfo
	lines := splitLines(string(out))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var raw struct {
			Name       string `json:"Name"`
			Driver     string `json:"Driver"`
			Mountpoint string `json:"Mountpoint"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		volumes = append(volumes, VolumeInfo{Name: raw.Name, Driver: raw.Driver, Mountpoint: raw.Mountpoint})
	}
	return volumes, nil
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
