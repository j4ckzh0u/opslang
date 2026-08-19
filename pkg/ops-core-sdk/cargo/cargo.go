// Package cargo provides Rust Cargo package management via CLI.
package cargo

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is the common return for cargo operations.
type Result struct {
	Package string `json:"package,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
	Output  string `json:"output,omitempty"`
}

// VersionResult is returned by cargo version.
type VersionResult struct {
	Version string `json:"version"`
	Success bool   `json:"success"`
}

func cargo(args ...string) (string, error) {
	cmd := exec.Command("cargo", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Install installs a crate globally.
func Install(pkg, version string, force bool) Result {
	if pkg == "" {
		return Result{Error: "package is required"}
	}
	args := []string{"install"}
	if version != "" {
		args = append(args, "--version", version)
	}
	if force {
		args = append(args, "--force")
	}
	args = append(args, pkg)
	out, err := cargo(args...)
	if err != nil {
		return Result{Package: pkg, Error: fmt.Sprintf("cargo install failed: %s: %s", err, out)}
	}
	return Result{Package: pkg, Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Uninstall removes a globally installed crate.
func Uninstall(pkg string) Result {
	if pkg == "" {
		return Result{Error: "package is required"}
	}
	out, err := cargo("uninstall", pkg)
	if err != nil {
		return Result{Package: pkg, Error: fmt.Sprintf("cargo uninstall failed: %s: %s", err, out)}
	}
	return Result{Package: pkg, Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Update updates a globally installed crate (or all if empty).
func Update(pkg string) Result {
	args := []string{"install-update"}
	if pkg != "" {
		args = append(args, pkg)
	}
	// Fallback: cargo install --force for updates
	args = []string{"install", "--force"}
	if pkg != "" {
		args = append(args, pkg)
	} else {
		args = append(args, "--list")
		return Result{Success: true, Output: "use package name to update specific crate"}
	}
	out, err := cargo(args...)
	if err != nil {
		return Result{Package: pkg, Error: fmt.Sprintf("cargo update failed: %s: %s", err, out)}
	}
	return Result{Package: pkg, Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// List lists installed crates.
func List() ([]string, error) {
	out, err := cargo("install", "--list")
	if err != nil {
		return nil, fmt.Errorf("cargo list failed: %w: %s", err, out)
	}
	var crates []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		// Lines like "cargo-edit v0.12.0:"
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "─") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				crates = append(crates, parts[0])
			}
		}
	}
	return crates, nil
}

// Build builds the project in the given directory.
func Build(dir string, release bool) Result {
	if dir == "" {
		dir = "."
	}
	args := []string{"build", "--manifest-path", dir + "/Cargo.toml"}
	if release {
		args = append(args, "--release")
	}
	out, err := cargo(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("cargo build failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Test runs tests for the project in the given directory.
func Test(dir string) Result {
	if dir == "" {
		dir = "."
	}
	args := []string{"test", "--manifest-path", dir + "/Cargo.toml"}
	out, err := cargo(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("cargo test failed: %s: %s", err, out)}
	}
	return Result{Success: true, Output: strings.TrimSpace(out)}
}

// Version returns cargo version.
func Version() VersionResult {
	out, err := cargo("--version")
	if err != nil {
		return VersionResult{}
	}
	return VersionResult{Version: strings.TrimSpace(out), Success: true}
}
