// Package yarn manages JavaScript packages via Yarn.
// Equivalent to community.general.yarn module.
package yarn

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status    string `json:"status"`
	Changed   bool   `json:"changed"`
	Package   string `json:"package,omitempty"`
	Version   string `json:"version,omitempty"`
	Global    bool   `json:"global"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status   string   `json:"status"`
	Packages []string `json:"packages"`
	Global   bool     `json:"global"`
	Error    string   `json:"error,omitempty"`
}

// Install installs a JavaScript package via Yarn.
func Install(name string, version string, global bool) Result {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}
	}

	pkg := name
	if version != "" {
		pkg = name + "@" + version
	}

	args := []string{"add", pkg}
	if global {
		args = append(args, "--global")
	}

	cmd := exec.Command("yarn", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("yarn add: %v", err)}
	}
	return Result{Status: "success", Changed: true, Package: name, Version: version, Global: global, Output: output}
}

// Remove removes a JavaScript package.
func Remove(name string, global bool) Result {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}
	}

	args := []string{"remove", name}
	if global {
		args = append(args, "--global")
	}

	cmd := exec.Command("yarn", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("yarn remove: %v", err)}
	}
	return Result{Status: "success", Changed: true, Package: name, Global: global, Output: output}
}

// Global installs all dependencies from package.json.
func Global(directory string) Result {
	if directory == "" {
		return Result{Status: "failed", Error: "directory is required"}
	}

	cmd := exec.Command("yarn", "install")
	cmd.Dir = directory
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("yarn install: %v", err)}
	}
	return Result{Status: "success", Changed: true, Output: output}
}

// List lists installed packages.
func List(global bool) ListResult {
	args := []string{"list", "--json"}
	if global {
		args = append(args, "--global")
	}

	cmd := exec.Command("yarn", args...)
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("yarn list: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	packages := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, line)
		}
	}
	return ListResult{Status: "success", Packages: packages, Global: global}
}
