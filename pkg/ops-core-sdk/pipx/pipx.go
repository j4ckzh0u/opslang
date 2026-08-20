// Package pipx manages Python packages via pipx (isolated environments).
package pipx

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result is the common return type for pipx operations.
type Result struct {
	Package    string `json:"package"`
	Changed    bool   `json:"changed"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

func findPipx() (string, error) {
	path, err := exec.LookPath("pipx")
	if err != nil {
		return "", fmt.Errorf("pipx not found in PATH")
	}
	return path, nil
}

func runCmd(args ...string) (string, error) {
	bin, err := findPipx()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("pipx %s: %s: %w", strings.Join(args, " "), string(out), err)
	}
	return string(out), nil
}

// Install installs a Python package via pipx (idempotent).
func Install(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	// Check if already installed
	list, _ := List()
	for _, pkg := range list {
		if pkg == name {
			return Result{Package: name, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
		}
	}

	if _, err := runCmd("install", name); err != nil {
		return Result{Package: name, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Package: name, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Uninstall removes a Python package (idempotent).
func Uninstall(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	// Check if installed
	list, _ := List()
	found := false
	for _, pkg := range list {
		if pkg == name {
			found = true
			break
		}
	}
	if !found {
		return Result{Package: name, Changed: false, DurationMs: time.Since(start).Milliseconds()}, nil
	}

	if _, err := runCmd("uninstall", name); err != nil {
		return Result{Package: name, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Package: name, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// Upgrade upgrades a Python package.
func Upgrade(name string) (Result, error) {
	start := time.Now()
	if name == "" {
		return Result{Error: "name must not be empty", DurationMs: time.Since(start).Milliseconds()}, fmt.Errorf("name must not be empty")
	}

	if _, err := runCmd("upgrade", name); err != nil {
		return Result{Package: name, Error: err.Error(), DurationMs: time.Since(start).Milliseconds()}, err
	}
	return Result{Package: name, Changed: true, DurationMs: time.Since(start).Milliseconds()}, nil
}

// List returns installed pipx packages.
func List() ([]string, error) {
	out, err := runCmd("list", "--short")
	if err != nil {
		return nil, err
	}

	var packages []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			// pipx list --short returns package names, one per line
			packages = append(packages, line)
		}
	}
	return packages, nil
}
