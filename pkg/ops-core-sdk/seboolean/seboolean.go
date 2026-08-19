// Package seboolean provides SELinux boolean management operations.
package seboolean

import (
	"fmt"
	"os/exec"
	"strings"
)

// BoolInfo represents an SELinux boolean.
type BoolInfo struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Default string `json:"default"`
}

// ListResult represents the result of listing booleans.
type ListResult struct {
	Booleans []BoolInfo `json:"booleans"`
}

// ActionResult represents the result of a boolean action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// List returns all SELinux booleans.
func List() (*ListResult, error) {
	cmd := exec.Command("getsebool", "-a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("getsebool failed: %w (output: %s)", err, string(out))
	}

	result := &ListResult{Booleans: make([]BoolInfo, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "boolean_name --> on/off"
		parts := strings.Split(line, " --> ")
		if len(parts) == 2 {
			result.Booleans = append(result.Booleans, BoolInfo{
				Name:  strings.TrimSpace(parts[0]),
				State: strings.TrimSpace(parts[1]),
			})
		}
	}
	return result, nil
}

// Get returns the state of a specific SELinux boolean.
func Get(name string) (*BoolInfo, error) {
	cmd := exec.Command("getsebool", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("getsebool failed: %w (output: %s)", err, string(out))
	}

	parts := strings.Split(strings.TrimSpace(string(out)), " --> ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected getsebool output: %s", string(out))
	}

	return &BoolInfo{
		Name:  strings.TrimSpace(parts[0]),
		State: strings.TrimSpace(parts[1]),
	}, nil
}

// Set sets the state of an SELinux boolean.
func Set(name string, state bool, persistent bool) (*ActionResult, error) {
	value := "off"
	if state {
		value = "on"
	}

	args := []string{"-P"}
	if !persistent {
		args = []string{}
	}
	args = append(args, name+"="+value)

	cmd := exec.Command("setsebool", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("setsebool failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Set %s to %s", name, value)}, nil
}
