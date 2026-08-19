// Package pip manages Python packages via pip.
// Equivalent to ansible.builtin.pip module.
package pip

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
	Exists  bool   `json:"exists"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Install installs a Python package.
func Install(name string, version string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	executable := "pip3"

	pkg := name
	if version != "" {
		pkg = name + "==" + version
	}

	cmd := exec.Command(executable, "install", pkg)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("pip install: %v", err)}, err
	}

	changed := !strings.Contains(output, "already satisfied")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: output}, nil
}

// Uninstall uninstalls a Python package.
func Uninstall(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	executable := "pip3"

	cmd := exec.Command(executable, "uninstall", "-y", name)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("pip uninstall: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}, nil
}

// List lists installed Python packages.
func List() (Result, error) {
	executable := "pip3"

	cmd := exec.Command(executable, "list", "--format=columns")
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("pip list: %v", err)}, err
	}
	return Result{Status: "success", Output: strings.TrimSpace(string(out))}, nil
}

// Exists checks if a Python package is installed.
func Exists(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	executable := "pip3"

	cmd := exec.Command(executable, "show", name)
	out, err := cmd.Output()
	if err != nil {
		// pip show returns non-zero if package not found
		return Result{Status: "success", Exists: false, Package: name}, nil
	}
	output := strings.TrimSpace(string(out))
	return Result{Status: "success", Exists: output != "", Package: name, Output: output}, nil
}

// Freeze returns pip freeze output.
func Freeze() (Result, error) {
	executable := "pip3"

	cmd := exec.Command(executable, "freeze")
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("pip freeze: %v", err)}, err
	}
	return Result{Status: "success", Output: strings.TrimSpace(string(out))}, nil
}

// InstallRequirements installs from a requirements file.
func InstallRequirements(requirements string) (Result, error) {
	if requirements == "" {
		return Result{Status: "failed", Error: "requirements file is required"}, fmt.Errorf("requirements file is required")
	}
	executable := "pip3"

	cmd := exec.Command(executable, "install", "-r", requirements)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("pip install -r: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: "requirements:" + requirements, Output: output}, nil
}
