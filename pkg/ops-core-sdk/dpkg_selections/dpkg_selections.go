// Package dpkg_selections provides Debian dpkg package selection management.
// Equivalent to Ansible's dpkg_selections module.
package dpkg_selections

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of a selection operation.
type Result struct {
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Package  string `json:"package"`
	Previous string `json:"previous,omitempty"`
	Current  string `json:"current"`
	Error    string `json:"error,omitempty"`
}

// Selection represents a package selection entry.
type Selection struct {
	Package string `json:"package"`
	State   string `json:"state"` // install, hold, deinstall, purge
}

func findDpkg() (string, error) {
	for _, p := range []string{"/usr/bin/dpkg", "/usr/sbin/dpkg"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("dpkg"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("dpkg not found")
}

// SetSelection sets the dpkg selection state for a package.
// Valid states: install, hold, deinstall, purge
func SetSelection(name string, state string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	validStates := map[string]bool{"install": true, "hold": true, "deinstall": true, "purge": true}
	if !validStates[state] {
		return Result{Status: "failed", Error: fmt.Sprintf("invalid state %q, must be one of: install, hold, deinstall, purge", state)}, fmt.Errorf("invalid state %q", state)
	}
	dpkg, err := findDpkg()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}

	// Get current selection
	current := getCurrentSelection(dpkg, name)

	if current == state {
		return Result{Status: "success", Changed: false, Package: name, Current: state}, nil
	}

	// Set selection via dpkg --set-selections
	cmd := exec.Command(dpkg, "--set-selections")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s %s\n", name, state))
	_, err = cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Package: name, Error: fmt.Sprintf("dpkg --set-selections: %v", err)}, err
	}

	return Result{Status: "success", Changed: true, Package: name, Previous: current, Current: state}, nil
}

// GetSelection returns the current dpkg selection state for a package.
func GetSelection(name string) (Selection, error) {
	if name == "" {
		return Selection{}, fmt.Errorf("package name is required")
	}
	dpkg, err := findDpkg()
	if err != nil {
		return Selection{}, err
	}
	state := getCurrentSelection(dpkg, name)
	return Selection{Package: name, State: state}, nil
}

// ListSelections returns all dpkg selections.
func ListSelections() ([]Selection, error) {
	dpkg, err := findDpkg()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(dpkg, "--get-selections")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("dpkg --get-selections: %w", err)
	}
	var selections []Selection
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			selections = append(selections, Selection{
				Package: fields[0],
				State:   fields[1],
			})
		}
	}
	return selections, nil
}

// Hold sets a package on hold.
func Hold(name string) (Result, error) {
	return SetSelection(name, "hold")
}

// Unhold sets a package back to install (removes hold).
func Unhold(name string) (Result, error) {
	return SetSelection(name, "install")
}

func getCurrentSelection(dpkg, name string) string {
	cmd := exec.Command(dpkg, "--get-selections", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) >= 2 {
		return fields[1]
	}
	return ""
}
