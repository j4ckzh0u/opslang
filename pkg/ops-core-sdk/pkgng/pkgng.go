// Package pkgng provides FreeBSD pkg-ng package management operations.
// Equivalent to Ansible's pkgng module.
package pkgng

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of a pkg operation.
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
	License     string `json:"license,omitempty"`
	Installed   bool   `json:"installed"`
}

func findPkg() (string, error) {
	for _, p := range []string{"/usr/sbin/pkg", "/usr/local/sbin/pkg"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("pkg"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("pkg not found")
}

func runCmd(args ...string) (string, error) {
	pkg, err := findPkg()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(pkg, args...)
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
		pkg = name + "-" + version
	}
	out, err := runCmd("install", "-y", pkg)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("pkg install: %v", err)}, err
	}
	changed := !strings.Contains(out, "already installed")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: out}, nil
}

// Remove removes a package.
func Remove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("delete", "-y", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("pkg delete: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates the repository catalog.
func Update() (Result, error) {
	out, err := runCmd("update", "-f")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("pkg update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Upgrade upgrades all packages.
func Upgrade(name string) (Result, error) {
	args := []string{"upgrade", "-y"}
	if name != "" {
		args = append(args, name)
	}
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("pkg upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Autoclean removes cached packages.
func Autoclean() (Result, error) {
	out, err := runCmd("autoclean", "-y")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("pkg autoclean: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Info returns package information.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{Status: "failed"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("info", name)
	if err != nil {
		return InfoResult{Status: "success", Package: name}, nil
	}
	info := InfoResult{Status: "success", Package: name, Installed: true}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Package = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Version") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Version = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Comment") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Description = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "WWW") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Homepage = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "License") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.License = strings.TrimSpace(parts[1])
			}
		}
	}
	return info, nil
}

// List lists installed packages.
func List() ([]InfoResult, error) {
	out, err := runCmd("info", "-a")
	if err != nil {
		return nil, fmt.Errorf("pkg info -a: %w", err)
	}
	var results []InfoResult
	var current *InfoResult
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Name") {
			if current != nil {
				results = append(results, *current)
			}
			parts := strings.SplitN(line, ":", 2)
			name := ""
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
			}
			current = &InfoResult{Status: "success", Package: name, Installed: true}
		}
		if current != nil {
			if strings.HasPrefix(line, "Version") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					current.Version = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "Comment") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					current.Description = strings.TrimSpace(parts[1])
				}
			}
		}
	}
	if current != nil {
		results = append(results, *current)
	}
	return results, nil
}

// Search searches for a package in repositories.
func Search(name string) ([]InfoResult, error) {
	if name == "" {
		return nil, fmt.Errorf("search name is required")
	}
	out, err := runCmd("search", name)
	if err != nil {
		return nil, fmt.Errorf("pkg search: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "-") && !strings.HasPrefix(line, " ") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				pkgName := parts[0]
				idx := strings.LastIndex(pkgName, "-")
				if idx > 0 {
					results = append(results, InfoResult{
						Status:    "success",
						Package:   pkgName[:idx],
						Version:   pkgName[idx+1:],
						Installed: false,
					})
				}
			}
		}
	}
	return results, nil
}

// Stats returns package database statistics.
func Stats() (map[string]interface{}, error) {
	out, err := runCmd("stats")
	if err != nil {
		return nil, fmt.Errorf("pkg stats: %w", err)
	}
	stats := make(map[string]interface{})
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				stats[key] = val
			}
		}
	}
	return stats, nil
}
