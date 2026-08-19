// Package composer manages PHP Composer packages.
// Equivalent to community.general.composer module.
package composer

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

// Install runs composer install in the given directory.
func Install(directory string, noDev bool) Result {
	if directory == "" {
		return Result{Status: "failed", Error: "directory is required"}
	}

	args := []string{"install", "--no-interaction"}
	if noDev {
		args = append(args, "--no-dev")
	}

	cmd := exec.Command("composer", args...)
	cmd.Dir = directory
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("composer install: %v", err)}
	}
	return Result{Status: "success", Changed: true, Output: output}
}

// Update runs composer update in the given directory.
func Update(directory string, noDev bool) Result {
	if directory == "" {
		return Result{Status: "failed", Error: "directory is required"}
	}

	args := []string{"update", "--no-interaction"}
	if noDev {
		args = append(args, "--no-dev")
	}

	cmd := exec.Command("composer", args...)
	cmd.Dir = directory
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("composer update: %v", err)}
	}
	return Result{Status: "success", Changed: true, Output: output}
}

// Require adds a specific package.
func Require(directory string, name string, version string) Result {
	if directory == "" || name == "" {
		return Result{Status: "failed", Error: "directory and package name are required"}
	}

	pkg := name
	if version != "" {
		pkg = name + ":" + version
	}

	cmd := exec.Command("composer", "require", "--no-interaction", pkg)
	cmd.Dir = directory
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("composer require: %v", err)}
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}
}

// Remove removes a package.
func Remove(directory string, name string) Result {
	if directory == "" || name == "" {
		return Result{Status: "failed", Error: "directory and package name are required"}
	}

	cmd := exec.Command("composer", "remove", "--no-interaction", name)
	cmd.Dir = directory
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("composer remove: %v", err)}
	}
	return Result{Status: "success", Changed: true, Package: name, Output: output}
}

// CreateProject creates a new Composer project.
func CreateProject(directory string, pkg string, version string) Result {
	if pkg == "" {
		return Result{Status: "failed", Error: "package name is required"}
	}

	composerPkg := pkg
	if version != "" {
		composerPkg = pkg + ":" + version
	}

	args := []string{"create-project", "--no-interaction", composerPkg}
	if directory != "" {
		args = append(args, directory)
	}

	cmd := exec.Command("composer", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("composer create-project: %v", err)}
	}
	return Result{Status: "success", Changed: true, Package: pkg, Output: output}
}

// GlobalInstall installs a package globally.
func GlobalInstall(pkg string, version string) Result {
	if pkg == "" {
		return Result{Status: "failed", Error: "package name is required"}
	}

	composerPkg := pkg
	if version != "" {
		composerPkg = pkg + ":" + version
	}

	cmd := exec.Command("composer", "global", "require", "--no-interaction", composerPkg)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("composer global require: %v", err)}
	}
	return Result{Status: "success", Changed: true, Package: pkg, Output: output}
}

// Version returns the Composer version.
func Version() Result {
	cmd := exec.Command("composer", "--version")
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("composer --version: %v", err)}
	}
	return Result{Status: "success", Version: strings.TrimSpace(string(out))}
}
