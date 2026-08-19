// Package pip provides Python package management operations via pip.
package pip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PackageInfo represents a Python package.
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListResult is returned by List.
type ListResult struct {
	Packages []PackageInfo `json:"packages"`
}

// InstallResult is returned by Install.
type InstallResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UninstallResult is returned by Uninstall.
type UninstallResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ExistsResult is returned by Exists.
type ExistsResult struct {
	Exists  bool   `json:"exists"`
	Version string `json:"version,omitempty"`
}

// List returns all installed Python packages.
func List() (ListResult, error) {
	cmd := exec.Command("pip", "list", "--format=json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return ListResult{}, fmt.Errorf("pip list failed: %w: %s", err, stderr.String())
	}

	var packages []PackageInfo
	if err := json.Unmarshal(stdout.Bytes(), &packages); err != nil {
		return ListResult{}, fmt.Errorf("parse pip output: %w", err)
	}

	return ListResult{Packages: packages}, nil
}

// Exists checks if a Python package is installed.
func Exists(name string) (ExistsResult, error) {
	result, err := List()
	if err != nil {
		return ExistsResult{}, err
	}

	nameLower := strings.ToLower(name)
	for _, pkg := range result.Packages {
		if strings.ToLower(pkg.Name) == nameLower {
			return ExistsResult{Exists: true, Version: pkg.Version}, nil
		}
	}

	return ExistsResult{Exists: false}, nil
}

// Install installs a Python package.
func Install(name string, version string) (InstallResult, error) {
	if name == "" {
		return InstallResult{Error: "package name is required"}, fmt.Errorf("package name is required")
	}

	// Check if already installed with correct version
	if version != "" {
		exists, err := Exists(name)
		if err == nil && exists.Exists && exists.Version == version {
			return InstallResult{Changed: false}, nil
		}
	} else {
		exists, err := Exists(name)
		if err == nil && exists.Exists {
			return InstallResult{Changed: false}, nil
		}
	}

	// Build package spec
	packageSpec := name
	if version != "" {
		packageSpec = fmt.Sprintf("%s==%s", name, version)
	}

	cmd := exec.Command("pip", "install", packageSpec)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return InstallResult{Error: stderr.String()}, fmt.Errorf("pip install failed: %w: %s", err, stderr.String())
	}

	return InstallResult{Changed: true}, nil
}

// Uninstall uninstalls a Python package.
func Uninstall(name string) (UninstallResult, error) {
	if name == "" {
		return UninstallResult{Error: "package name is required"}, fmt.Errorf("package name is required")
	}

	// Check if installed
	exists, err := Exists(name)
	if err != nil {
		return UninstallResult{Error: err.Error()}, err
	}
	if !exists.Exists {
		return UninstallResult{Changed: false}, nil
	}

	cmd := exec.Command("pip", "uninstall", "-y", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return UninstallResult{Error: stderr.String()}, fmt.Errorf("pip uninstall failed: %w: %s", err, stderr.String())
	}

	return UninstallResult{Changed: true}, nil
}
