// Package gem manages Ruby gems.
// Equivalent to ansible.builtin.gem module.
package gem

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Package string `json:"package,omitempty"`
	Version string `json:"version,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Install installs a Ruby gem.
func Install(name string, version string, userInstall bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "gem name is required"}, fmt.Errorf("gem name is required")
	}

	args := []string{"install", name, "--no-document"}
	if version != "" {
		args = append(args, "--version", version)
	}
	if userInstall {
		args = append(args, "--user-install")
	}

	cmd := exec.Command("gem", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("gem install: %v", err)}, err
	}

	changed := !strings.Contains(output, "already installed")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: output}, nil
}

// Uninstall uninstalls a Ruby gem.
func Uninstall(name string, force bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "gem name is required"}, fmt.Errorf("gem name is required")
	}

	args := []string{"uninstall", name, "--all"}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("gem", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("gem uninstall: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// Update updates a Ruby gem.
func Update(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "gem name is required"}, fmt.Errorf("gem name is required")
	}

	cmd := exec.Command("gem", "update", name, "--no-document")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("gem update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// Info returns information about a gem.
func Info(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "gem name is required"}, fmt.Errorf("gem name is required")
	}

	cmd := exec.Command("gem", "specification", name, "--remote", "--ruby")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("gem info: %v", err)}, err
	}
	return Result{Status: "success", Package: name, Output: output}, nil
}

// List lists installed gems.
func List() (Result, error) {
	cmd := exec.Command("gem", "list", "--local")
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("gem list: %v", err)}, err
	}
	return Result{Status: "success", Output: strings.TrimSpace(string(out))}, nil
}

// Version returns the gem command version.
func Version() (Result, error) {
	cmd := exec.Command("gem", "--version")
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("gem --version: %v", err)}, err
	}
	return Result{Status: "success", Version: strings.TrimSpace(string(out))}, nil
}
