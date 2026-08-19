// Package composer provides PHP Composer management via CLI.
package composer

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is the common return for composer operations.
type Result struct {
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
	Output  string `json:"output,omitempty"`
}

// ProjectResult is returned by project-specific operations.
type ProjectResult struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Dir     string `json:"dir,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// VersionResult is returned by version.
type VersionResult struct {
	Version string `json:"version"`
	Success bool   `json:"success"`
}

func composer(args ...string) (string, error) {
	cmd := exec.Command("composer", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Install runs composer install in the given directory.
func Install(dir string, noDev bool) Result {
	if dir == "" {
		dir = "."
	}
	args := []string{"install", "--no-interaction", "--working-dir", dir}
	if noDev {
		args = append(args, "--no-dev")
	}
	out, err := composer(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("composer install failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Update runs composer update in the given directory.
func Update(dir string, noDev bool) Result {
	if dir == "" {
		dir = "."
	}
	args := []string{"update", "--no-interaction", "--working-dir", dir}
	if noDev {
		args = append(args, "--no-dev")
	}
	out, err := composer(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("composer update failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Require adds a package requirement.
func Require(dir, pkg, version string) Result {
	if dir == "" {
		dir = "."
	}
	if pkg == "" {
		return Result{Error: "package is required"}
	}
	dep := pkg
	if version != "" {
		dep = pkg + ":" + version
	}
	args := []string{"require", "--no-interaction", "--working-dir", dir, dep}
	out, err := composer(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("composer require failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Remove removes a package.
func Remove(dir, pkg string) Result {
	if dir == "" {
		dir = "."
	}
	if pkg == "" {
		return Result{Error: "package is required"}
	}
	args := []string{"remove", "--no-interaction", "--working-dir", dir, pkg}
	out, err := composer(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("composer remove failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// CreateProject creates a new composer project.
func CreateProject(dir, pkg, version string) ProjectResult {
	if dir == "" {
		return ProjectResult{Error: "directory is required"}
	}
	if pkg == "" {
		return ProjectResult{Error: "package is required"}
	}
	dep := pkg
	if version != "" {
		dep = pkg + ":" + version
	}
	args := []string{"create-project", "--no-interaction", dep, dir}
	out, err := composer(args...)
	if err != nil {
		return ProjectResult{Error: fmt.Sprintf("create-project failed: %s: %s", err, out)}
	}
	return ProjectResult{Dir: dir, Success: true, Changed: true}
}

// GlobalInstall installs a package globally.
func GlobalInstall(pkg, version string) Result {
	if pkg == "" {
		return Result{Error: "package is required"}
	}
	dep := pkg
	if version != "" {
		dep = pkg + ":" + version
	}
	args := []string{"global", "require", "--no-interaction", dep}
	out, err := composer(args...)
	if err != nil {
		return Result{Error: fmt.Sprintf("global require failed: %s: %s", err, out)}
	}
	return Result{Success: true, Changed: true, Output: strings.TrimSpace(out)}
}

// Version returns composer version.
func Version() VersionResult {
	out, err := composer("--version")
	if err != nil {
		return VersionResult{}
	}
	// e.g., "Composer version 2.6.5 2023-10-06 10:11:52"
	return VersionResult{Version: strings.TrimSpace(out), Success: true}
}
