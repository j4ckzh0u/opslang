// Package zypper provides SUSE Zypper package management operations.
// Equivalent to Ansible's zypper module.
package zypper

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
	Status  string `json:"status,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// RepoInfo represents repository information.
type RepoInfo struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	Enabled string `json:"enabled,omitempty"`
	Type    string `json:"type,omitempty"`
}

// isNotFound checks whether an exec error indicates a missing binary.
func isNotFound(err error) bool {
	if _, ok := err.(*exec.Error); ok {
		return true
	}
	return false
}

// runZypper executes a zypper command with --non-interactive.
func runZypper(args ...string) (string, error) {
	fullArgs := append([]string{"--non-interactive"}, args...)
	cmd := exec.Command("zypper", fullArgs...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if isNotFound(err) {
			return "", fmt.Errorf("zypper command not found: %w", err)
		}
		return output, fmt.Errorf("zypper %s: %v", strings.Join(args, " "), err)
	}
	return output, nil
}

// Install installs a package by name, optionally at a specific version.
func Install(name string, version string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	pkg := name
	if version != "" {
		pkg = name + "=" + version
	}
	out, err := runZypper("install", "--auto-agree-with-licenses", pkg)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Remove removes a package by name.
func Remove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	out, err := runZypper("remove", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates a package. If name is empty, updates all packages.
func Update(name string) (Result, error) {
	args := []string{"update"}
	if name != "" {
		args = append(args, name)
	}
	out, err := runZypper(args...)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// DistUpgrade performs a distribution upgrade.
func DistUpgrade() (Result, error) {
	out, err := runZypper("dist-upgrade", "--auto-agree-with-licenses")
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
	cmd := exec.Command("zypper", "--non-interactive", "info", name)
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return InfoResult{}, fmt.Errorf("zypper command not found: %w", err)
		}
		return InfoResult{}, fmt.Errorf("zypper info: %w", err)
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
		case "Status":
			result.Status = val
		case "Summary":
			result.Summary = val
		}
	}
	return result, nil
}

// List lists installed packages.
func List() ([]InfoResult, error) {
	cmd := exec.Command("zypper", "--non-interactive", "search", "--installed-only", "--type", "package")
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("zypper command not found: %w", err)
		}
		return nil, fmt.Errorf("zypper list: %w", err)
	}
	packages := make([]InfoResult, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i < 2 || strings.HasPrefix(line, "-") {
			continue // skip headers and separators
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pkg := InfoResult{
			Status: fields[0],
			Name:   fields[2],
		}
		if len(fields) >= 4 {
			pkg.Version = fields[3]
		}
		if len(fields) >= 5 {
			pkg.Arch = fields[4]
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

// Clean cleans the zypper cache.
func Clean() (Result, error) {
	out, err := runZypper("clean", "--all")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// RepoList lists configured repositories.
func RepoList() ([]RepoInfo, error) {
	cmd := exec.Command("zypper", "--non-interactive", "repos", "--details")
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("zypper command not found: %w", err)
		}
		return nil, fmt.Errorf("zypper repos: %w", err)
	}
	repos := make([]RepoInfo, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i < 2 || strings.HasPrefix(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		repo := RepoInfo{
			Name: fields[1],
		}
		if len(fields) >= 4 {
			repo.URL = fields[3]
		}
		if len(fields) >= 5 {
			repo.Enabled = fields[4]
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

// RepoAdd adds a repository.
func RepoAdd(name string, url string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	if url == "" {
		return Result{Status: "failed", Error: "url is required"}, fmt.Errorf("url is required")
	}
	out, err := runZypper("addrepo", url, name)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// RepoRemove removes a repository.
func RepoRemove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	out, err := runZypper("removerepo", name)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Refresh refreshes all repositories.
func Refresh() (Result, error) {
	out, err := runZypper("refresh")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Search searches for packages matching name.
func Search(name string) ([]InfoResult, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	cmd := exec.Command("zypper", "--non-interactive", "search", name)
	out, err := cmd.Output()
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("zypper command not found: %w", err)
		}
		return nil, fmt.Errorf("zypper search: %w", err)
	}
	packages := make([]InfoResult, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i < 2 || strings.HasPrefix(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pkg := InfoResult{
			Name: fields[2],
		}
		if len(fields) >= 4 {
			pkg.Version = fields[3]
		}
		if len(fields) >= 5 {
			pkg.Arch = fields[4]
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}

// Patch applies available patches.
func Patch() (Result, error) {
	out, err := runZypper("patch", "--auto-agree-with-licenses")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// PatternInstall installs a zypper pattern.
func PatternInstall(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	out, err := runZypper("install", "--type", "pattern", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// PatternRemove removes a zypper pattern.
func PatternRemove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}, fmt.Errorf("name is required")
	}
	out, err := runZypper("remove", "--type", "pattern", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: err.Error()}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}
