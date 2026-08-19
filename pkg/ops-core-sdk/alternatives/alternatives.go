// Package alternatives provides update-alternatives management operations.
package alternatives

import (
	"fmt"
	"os/exec"
	"strings"
)

// Alternative represents an alternative entry.
type Alternative struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Priority int    `json:"priority"`
	Status   string `json:"status"`
}

// ListResult represents the result of listing alternatives.
type ListResult struct {
	Alternatives []Alternative `json:"alternatives"`
}

// ActionResult represents the result of an alternatives action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// List returns all alternatives for a given name.
func List(name string) (*ListResult, error) {
	out, err := exec.Command("update-alternatives", "--list", name).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("update-alternatives list failed: %w", err)
	}

	result := &ListResult{Alternatives: make([]Alternative, 0)}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			alt := Alternative{
				Name: name,
				Path: line,
			}
			result.Alternatives = append(result.Alternatives, alt)
		}
	}

	return result, nil
}

// Display shows the current alternative for a name.
func Display(name string) (*Alternative, error) {
	out, err := exec.Command("update-alternatives", "--display", name).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("update-alternatives display failed: %w", err)
	}

	alt := &Alternative{Name: name}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		if strings.Contains(line, "link currently points to") {
			parts := strings.Split(line, "link currently points to")
			if len(parts) >= 2 {
				alt.Path = strings.TrimSpace(parts[1])
			}
		}
	}

	return alt, nil
}

// Set sets the alternative for a name.
func Set(name string, path string) (*ActionResult, error) {
	cmd := exec.Command("update-alternatives", "--set", name, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("update-alternatives set failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Set %s to %s", name, path),
	}, nil
}

// Install installs a new alternative.
func Install(name string, link string, path string, priority int) (*ActionResult, error) {
	cmd := exec.Command("update-alternatives", "--install", link, name, path, fmt.Sprintf("%d", priority))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("update-alternatives install failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Installed alternative %s -> %s (priority %d)", name, path, priority),
	}, nil
}

// Remove removes an alternative.
func Remove(name string, path string) (*ActionResult, error) {
	cmd := exec.Command("update-alternatives", "--remove", name, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("update-alternatives remove failed: %w (output: %s)", err, string(out))
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Removed alternative %s -> %s", name, path),
	}, nil
}
