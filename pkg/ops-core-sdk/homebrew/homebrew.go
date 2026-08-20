// Package homebrew provides macOS Homebrew package management operations.
// Equivalent to Ansible's homebrew module.
package homebrew

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of a brew operation.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Package string `json:"package,omitempty"`
	Version string `json:"version,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// InfoResult represents package information.
type InfoResult struct {
	Status      string `json:"status"`
	Package     string `json:"package"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	Installed   bool   `json:"installed"`
	Outdated    bool   `json:"outdated"`
}

// TapResult represents a tap.
type TapResult struct {
	Status string `json:"status"`
	Name   string `json:"name"`
	Path   string `json:"path,omitempty"`
}

func findBrew() (string, error) {
	for _, p := range []string{
		"/opt/homebrew/bin/brew",
		"/usr/local/bin/brew",
		"/home/linuxbrew/.linuxbrew/bin/brew",
	} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("brew"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("brew not found")
}

func runCmd(args ...string) (string, error) {
	brew, err := findBrew()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(brew, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Install installs a formula or cask.
func Install(name string, cask bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	args := []string{"install"}
	if cask {
		args = []string{"install", "--cask"}
	}
	args = append(args, name)
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("brew install: %v", err)}, err
	}
	changed := !strings.Contains(out, "already installed")
	return Result{Status: "success", Changed: changed, Package: name, Output: out}, nil
}

// Remove removes a formula or cask.
func Remove(name string, cask bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	args := []string{"uninstall"}
	if cask {
		args = []string{"uninstall", "--cask"}
	}
	args = append(args, name)
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("brew uninstall: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Upgrade upgrades a formula/cask or all.
func Upgrade(name string) (Result, error) {
	if name == "" {
		out, err := runCmd("upgrade")
		if err != nil {
			return Result{Status: "failed", Output: out, Error: fmt.Sprintf("brew upgrade: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Output: out}, nil
	}
	out, err := runCmd("upgrade", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("brew upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates Homebrew itself.
func Update() (Result, error) {
	out, err := runCmd("update")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("brew update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Info returns package information.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{Status: "failed"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("info", "--json=v2", name)
	if err != nil {
		return InfoResult{Status: "success", Package: name}, nil
	}
	info := InfoResult{Status: "success", Package: name}
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "\"version\"") || strings.Contains(line, "\"version\"") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				info.Version = strings.Trim(strings.TrimSpace(parts[1]), "\",")
			}
		}
		if strings.HasPrefix(trimmed, "\"desc\"") || strings.Contains(line, "\"desc\"") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				info.Description = strings.Trim(strings.TrimSpace(parts[1]), "\",")
			}
		}
		if strings.HasPrefix(trimmed, "\"homepage\"") || strings.Contains(line, "\"homepage\"") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				info.Homepage = strings.Trim(strings.TrimSpace(parts[1]), "\",")
			}
		}
	}
	return info, nil
}

// List lists installed formulas.
func List() ([]string, error) {
	out, err := runCmd("list", "--formula")
	if err != nil {
		return nil, fmt.Errorf("brew list: %w", err)
	}
	var packages []string
	for _, line := range strings.Split(out, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			packages = append(packages, trimmed)
		}
	}
	return packages, nil
}

// ListCasks lists installed casks.
func ListCasks() ([]string, error) {
	out, err := runCmd("list", "--cask")
	if err != nil {
		return nil, fmt.Errorf("brew list --cask: %w", err)
	}
	var packages []string
	for _, line := range strings.Split(out, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			packages = append(packages, trimmed)
		}
	}
	return packages, nil
}

// Outdated lists outdated packages.
func Outdated() ([]string, error) {
	out, err := runCmd("outdated")
	if err != nil {
		return nil, fmt.Errorf("brew outdated: %w", err)
	}
	var packages []string
	for _, line := range strings.Split(out, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			packages = append(packages, trimmed)
		}
	}
	return packages, nil
}

// Clean cleans old versions from the cellar.
func Clean() (Result, error) {
	out, err := runCmd("cleanup")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("brew cleanup: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Tap adds a tap.
func Tap(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "tap name is required"}, fmt.Errorf("tap name is required")
	}
	out, err := runCmd("tap", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("brew tap: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Untap removes a tap.
func Untap(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "tap name is required"}, fmt.Errorf("tap name is required")
	}
	out, err := runCmd("untap", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("brew untap: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// ListTaps lists current taps.
func ListTaps() ([]string, error) {
	out, err := runCmd("tap")
	if err != nil {
		return nil, fmt.Errorf("brew tap: %w", err)
	}
	var taps []string
	for _, line := range strings.Split(out, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			taps = append(taps, trimmed)
		}
	}
	return taps, nil
}

// Doctor runs brew doctor.
func Doctor() (string, error) {
	out, err := runCmd("doctor")
	if err != nil {
		return out, nil // doctor returns non-zero but output is still useful
	}
	return out, nil
}
