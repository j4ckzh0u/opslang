// Package pacman provides Arch Linux Pacman package management operations.
// Equivalent to Ansible's pacman module.
package pacman

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by mutating operations.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Package string `json:"package,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// InfoResult represents package information.
type InfoResult struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Arch    string `json:"arch,omitempty"`
	Size    string `json:"size,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// isNotFound checks whether an exec error indicates a missing binary.
func isNotFound(err error) bool {
	if _, ok := err.(*exec.Error); ok {
		return true
	}
	return false
}

// runPacman executes a pacman command.
func runPacman(args ...string) (string, error) {
	cmd := exec.Command("pacman", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return "", fmt.Errorf("pacman command not found: %w", err)
		}
		return output, fmt.Errorf("pacman %s: %v", strings.Join(args, " "), err)
	}
	return output, nil
}

// Install installs a package by name.
func Install(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	out, err := runPacman("--noconfirm", "-S", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Remove removes a package. If cascade is true, also removes dependent packages (-Rc).
func Remove(name string, cascade bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	flag := "-R"
	if cascade {
		flag = "-Rc"
	}
	out, err := runPacman("--noconfirm", flag, name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates a package. If name is empty, updates all packages (-Syu).
func Update(name string) (Result, error) {
	if name == "" {
		out, err := runPacman("--noconfirm", "-Syu")
		if err != nil {
			return Result{Status: "failed", Output: out, Error: err.Error()}, err
		}
		return Result{Status: "success", Changed: true, Output: out}, nil
	}
	out, err := runPacman("--noconfirm", "-S", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Upgrade performs a full system upgrade.
func Upgrade() (Result, error) {
	out, err := runPacman("--noconfirm", "-Syu")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Info returns information about a package.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{}, fmt.Errorf("name is required")
	}
	cmd := exec.Command("pacman", "-Qi", name)
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return InfoResult{}, fmt.Errorf("pacman command not found: %w", err)
		}
		// Try querying the sync database if not locally installed.
		cmd = exec.Command("pacman", "-Si", name)
		out, err = cmd.Output()
		if err != nil {
			return InfoResult{}, fmt.Errorf("pacman info: %w", err)
		}
	}
	result := InfoResult{Name: name}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Name":
			result.Name = val
		case "Version":
			result.Version = val
		case "Architecture":
			result.Arch = val
		case "Installed Size", "Download Size":
			result.Size = val
		case "Description":
			result.Summary = val
		}
	}
	return result, nil
}

// List lists installed packages.
func List() ([]InfoResult, error) {
	cmd := exec.Command("pacman", "-Q")
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("pacman command not found: %w", err)
		}
		return nil, fmt.Errorf("pacman list: %w", err)
	}
	packages := make([]InfoResult, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		packages = append(packages, InfoResult{
			Name:    fields[0],
			Version: fields[1],
		})
	}
	return packages, nil
}

// Search searches for packages matching name.
func Search(name string) ([]InfoResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	cmd := exec.Command("pacman", "-Ss", name)
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("pacman command not found: %w", err)
		}
		return nil, fmt.Errorf("pacman search: %w", err)
	}
	packages := make([]InfoResult, 0)
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, " ") {
			continue // skip description lines
		}
		// Format: repo/name version [installed]
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		nameParts := strings.SplitN(fields[0], "/", 2)
		pkgName := fields[0]
		if len(nameParts) == 2 {
			pkgName = nameParts[1]
		}
		pkg := InfoResult{
			Name:    pkgName,
			Version: fields[1],
		}
		// Next line may be the description.
		if i+1 < len(lines) && strings.HasPrefix(lines[i+1], " ") {
			pkg.Summary = strings.TrimSpace(lines[i+1])
			i++
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

// Clean cleans the package cache.
func Clean() (Result, error) {
	out, err := runPacman("--noconfirm", "-Sc")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// InstallFile installs a local package file.
func InstallFile(path string) (Result, error) {
	if path == "" {
		return Result{Status: "failed", Error: "path is required"}, fmt.Errorf("path is required")
	}
	out, err := runPacman("--noconfirm", "-U", path)
	if err != nil {
		return Result{Status: "failed", Package: path, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: path, Output: out}, nil
}

// RemoveOrphans removes orphaned packages.
func RemoveOrphans() (Result, error) {
	out, err := runPacman("--noconfirm", "-Rns")
	if err != nil {
		// pacman -Rns with no orphans returns an error — that is not a failure.
		if strings.Contains(out, "no targets specified") || strings.Contains(out, "requires:") {
			return Result{Status: "success", Changed: false, Output: out}, nil
		}
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// UpdateDatabase syncs package databases.
func UpdateDatabase() (Result, error) {
	out, err := runPacman("--noconfirm", "-Sy")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}
