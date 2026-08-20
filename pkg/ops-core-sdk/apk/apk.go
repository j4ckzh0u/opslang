// Package apk provides Alpine Linux APK package management operations.
// Equivalent to Ansible's apk module.
package apk

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of an apk operation.
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
	Status       string `json:"status"`
	Package      string `json:"package"`
	Version      string `json:"version"`
	Architecture string `json:"architecture,omitempty"`
	Size         string `json:"size,omitempty"`
	Description  string `json:"description,omitempty"`
	URL          string `json:"url,omitempty"`
	License      string `json:"license,omitempty"`
	Repository   string `json:"repository,omitempty"`
}

func findApk() (string, error) {
	for _, p := range []string{"/sbin/apk", "/usr/bin/apk"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("apk"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("apk not found")
}

func runCmd(args ...string) (string, error) {
	apk, err := findApk()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(apk, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Install installs a package.
func Install(name string, version string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	pkg := name
	if version != "" {
		pkg = name + "=" + version
	}
	out, err := runCmd("add", pkg)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apk add: %v", err)}, err
	}
	changed := !strings.Contains(out, "already installed")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: out}, nil
}

// Remove removes a package.
func Remove(name string, purge bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	args := []string{"del"}
	if purge {
		args = append(args, "--purge")
	}
	args = append(args, name)
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apk del: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates package database.
func Update() (Result, error) {
	out, err := runCmd("update")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apk update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Upgrade upgrades all packages or a specific package.
func Upgrade(name string) (Result, error) {
	if name == "" {
		out, err := runCmd("upgrade")
		if err != nil {
			return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apk upgrade: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Output: out}, nil
	}
	out, err := runCmd("add", "-u", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apk upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Info returns package information.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{Status: "failed"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("info", "-a", name)
	if err != nil {
		return InfoResult{Status: "success", Package: name}, nil
	}
	info := InfoResult{Status: "success", Package: name}
	lines := strings.Split(out, "\n")
	if len(lines) > 0 {
		header := lines[0]
		parts := strings.Fields(header)
		if len(parts) > 0 {
			pkgVer := parts[0]
			idx := strings.LastIndex(pkgVer, "-")
			if idx > 0 {
				info.Package = pkgVer[:idx]
				info.Version = pkgVer[idx+1:]
			}
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "webpage:") {
			info.URL = strings.TrimSpace(strings.TrimPrefix(line, "webpage:"))
		}
		if strings.HasPrefix(line, "license:") {
			info.License = strings.TrimSpace(strings.TrimPrefix(line, "license:"))
		}
		if strings.HasPrefix(line, "description:") {
			info.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return info, nil
}

// List lists installed packages.
func List() ([]InfoResult, error) {
	out, err := runCmd("info", "-v")
	if err != nil {
		return nil, fmt.Errorf("apk info: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		pkgVer := strings.TrimSpace(line)
		idx := strings.LastIndex(pkgVer, "-")
		if idx < 0 {
			continue
		}
		name := pkgVer[:idx]
		version := pkgVer[idx+1:]
		// Handle arch suffix like "name-version-arch"
		idx2 := strings.LastIndex(name, "-")
		if idx2 > 0 {
			version = name[idx2+1:] + "-" + version
			name = name[:idx2]
		}
		results = append(results, InfoResult{
			Status:  "success",
			Package: name,
			Version: version,
		})
	}
	return results, nil
}

// Search searches for a package.
func Search(name string) ([]InfoResult, error) {
	if name == "" {
		return nil, fmt.Errorf("search name is required")
	}
	out, err := runCmd("search", "-v", name)
	if err != nil {
		return nil, fmt.Errorf("apk search: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pkgVer := fields[0]
		idx := strings.LastIndex(pkgVer, "-")
		if idx < 0 {
			continue
		}
		results = append(results, InfoResult{
			Status:      "success",
			Package:     pkgVer[:idx],
			Version:     pkgVer[idx+1:],
			Description: strings.Join(fields[1:], " "),
		})
	}
	return results, nil
}

// Cache cleans apk cache.
func Cache() (Result, error) {
	out, err := runCmd("cache", "clean")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apk cache clean: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// UpgradeAvailable checks if upgrades are available.
func UpgradeAvailable() (bool, error) {
	out, err := runCmd("version", "-l", "<")
	if err != nil {
		// No upgrades available returns exit code 0 with no output
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

// Repository lists configured repositories.
func Repository() ([]string, error) {
	f, err := os.Open("/etc/apk/repositories")
	if err != nil {
		return nil, fmt.Errorf("cannot read /etc/apk/repositories: %w", err)
	}
	defer f.Close()
	data, err := os.ReadFile("/etc/apk/repositories")
	if err != nil {
		return nil, fmt.Errorf("read repositories: %w", err)
	}
	var repos []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			repos = append(repos, trimmed)
		}
	}
	return repos, nil
}
