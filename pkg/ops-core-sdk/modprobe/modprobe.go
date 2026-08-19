// Package modprobe provides Linux kernel module management operations.
package modprobe

import (
	"fmt"
	"os/exec"
	"strings"
)

// Module represents a kernel module.
type Module struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Used string `json:"used_count"`
	By   string `json:"used_by"`
}

// ListResult represents the result of listing modules.
type ListResult struct {
	Modules []Module `json:"modules"`
}

// ActionResult represents the result of a modprobe action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// LoadedResult represents the result of checking if a module is loaded.
type LoadedResult struct {
	Loaded bool `json:"loaded"`
}

// List returns all loaded kernel modules.
func List() (*ListResult, error) {
	out, err := exec.Command("lsmod").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lsmod failed: %w", err)
	}

	result := &ListResult{Modules: make([]Module, 0)}
	lines := strings.Split(string(out), "\n")

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			mod := Module{
				Name: fields[0],
				Size: fields[1],
				Used: fields[2],
			}
			if len(fields) >= 4 {
				mod.By = strings.Join(fields[3:], " ")
			}
			result.Modules = append(result.Modules, mod)
		}
	}

	return result, nil
}

// Load loads a kernel module.
func Load(name string) (*ActionResult, error) {
	cmd := exec.Command("modprobe", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("modprobe load failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Loaded module %s", name),
	}, nil
}

// Unload unloads a kernel module.
func Unload(name string) (*ActionResult, error) {
	cmd := exec.Command("modprobe", "-r", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("modprobe unload failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Unloaded module %s", name),
	}, nil
}

// IsLoaded checks if a module is loaded.
func IsLoaded(name string) (*LoadedResult, error) {
	cmd := exec.Command("lsmod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lsmod failed: %w", err)
	}

	loaded := false
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			loaded = true
			break
		}
	}

	return &LoadedResult{Loaded: loaded}, nil
}
