// Package apt provides Debian/Ubuntu package management operations.
// Equivalent to Ansible's apt module.
package apt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of an apt operation.
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
	Architecture string `json:"architecture"`
	Description string `json:"description"`
	Status2     string `json:"package_status"` // installed/not-installed
	Source      string `json:"source,omitempty"`
	Priority    string `json:"priority,omitempty"`
	Section     string `json:"section,omitempty"`
	Maintainer  string `json:"maintainer,omitempty"`
	Size        string `json:"size,omitempty"`
}

// PolicyResult represents apt-cache policy output.
type PolicyResult struct {
	Status          string `json:"status"`
	Package         string `json:"package"`
	Installed       string `json:"installed"`
	Candidate       string `json:"candidate"`
	VersionTable    string `json:"version_table,omitempty"`
	Error           string `json:"error,omitempty"`
}

func findAptGet() (string, error) {
	for _, p := range []string{"/usr/bin/apt-get", "/usr/local/bin/apt-get"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("apt-get"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("apt-get not found")
}

func findAptCache() (string, error) {
	for _, p := range []string{"/usr/bin/apt-cache", "/usr/local/bin/apt-cache"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("apt-cache"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("apt-cache not found")
}

func findDpkg() (string, error) {
	for _, p := range []string{"/usr/bin/dpkg", "/usr/local/bin/dpkg"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("dpkg"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("dpkg not found")
}

func runCmd(binary string, args ...string) (string, error) {
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Install installs a package via apt-get.
func Install(name string, version string, updateCache bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}

	if updateCache {
		_, _ = runCmd(aptGet, "update", "-q")
	}

	pkg := name
	if version != "" {
		pkg = name + "=" + version
	}

	out, err := runCmd(aptGet, "install", "-y", "-q", pkg)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apt-get install: %v", err)}, err
	}

	changed := !strings.Contains(out, "already the newest version") && !strings.Contains(out, "is already installed")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: out}, nil
}

// Remove removes a package via apt-get. If purge is true, also removes config files.
func Remove(name string, purge bool) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}

	subcmd := "remove"
	if purge {
		subcmd = "purge"
	}

	out, err := runCmd(aptGet, subcmd, "-y", "-q", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apt-get %s: %v", subcmd, err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Upgrade upgrades a specific package, or all packages if name is empty.
func Upgrade(name string) (Result, error) {
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}

	if name == "" {
		out, err := runCmd(aptGet, "upgrade", "-y", "-q")
		if err != nil {
			return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apt-get upgrade: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Output: out}, nil
	}

	out, err := runCmd(aptGet, "install", "-y", "-q", "--only-upgrade", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apt-get upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// UpdateCache runs apt-get update to refresh package lists.
func UpdateCache() (Result, error) {
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "update", "-q")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apt-get update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// FullUpgrade runs apt-get upgrade.
func FullUpgrade() (Result, error) {
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "upgrade", "-y", "-q")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apt-get upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DistUpgrade runs apt-get dist-upgrade.
func DistUpgrade() (Result, error) {
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "dist-upgrade", "-y", "-q")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apt-get dist-upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Autoremove removes unused dependencies.
func Autoremove() (Result, error) {
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "autoremove", "-y", "-q")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apt-get autoremove: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Clean cleans the local repository of retrieved package files.
func Clean() (Result, error) {
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "clean", "-q")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("apt-get clean: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Info returns package information via dpkg -s and apt-cache show.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{Status: "failed"}, fmt.Errorf("package name is required")
	}
	dpkg, err := findDpkg()
	if err != nil {
		return InfoResult{Status: "failed"}, err
	}

	out, err := runCmd(dpkg, "-s", name)
	if err != nil {
		return InfoResult{Status: "success", Package: name, Status2: "not-installed"}, nil
	}

	info := InfoResult{Status: "success", Package: name, Status2: "installed"}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "Version":
			info.Version = parts[1]
		case "Architecture":
			info.Architecture = parts[1]
		case "Description":
			info.Description = parts[1]
		case "Status":
			// keep original status
		case "Priority":
			info.Priority = parts[1]
		case "Section":
			info.Section = parts[1]
		case "Maintainer":
			info.Maintainer = parts[1]
		case "Installed-Size":
			info.Size = parts[1]
		}
	}
	return info, nil
}

// List lists all installed packages via dpkg-query.
func List() ([]InfoResult, error) {
	dpkg, err := findDpkg()
	if err != nil {
		return nil, err
	}

	out, err := runCmd(dpkg, "-l")
	if err != nil {
		return nil, fmt.Errorf("dpkg -l: %w", err)
	}

	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		// skip header lines
		if !strings.HasPrefix(line, "ii") && !strings.HasPrefix(line, "rc") && !strings.HasPrefix(line, "hi") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		status := "installed"
		if strings.HasPrefix(line, "rc") {
			status = "removed-config-remaining"
		}
		results = append(results, InfoResult{
			Status:       "success",
			Package:      fields[1],
			Version:      fields[2],
			Architecture: fields[3],
			Description:  strings.Join(fields[4:], " "),
			Status2:      status,
		})
	}
	return results, nil
}

// Policy returns apt-cache policy output for a package.
func Policy(name string) (PolicyResult, error) {
	if name == "" {
		return PolicyResult{Status: "failed"}, fmt.Errorf("package name is required")
	}
	aptCache, err := findAptCache()
	if err != nil {
		return PolicyResult{Status: "failed"}, err
	}

	out, err := runCmd(aptCache, "policy", name)
	if err != nil {
		return PolicyResult{Status: "failed", Package: name, Error: fmt.Sprintf("apt-cache policy: %v", err)}, err
	}

	result := PolicyResult{Status: "success", Package: name}
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Installed:") {
			result.Installed = strings.TrimSpace(strings.TrimPrefix(trimmed, "Installed:"))
		} else if strings.HasPrefix(trimmed, "Candidate:") {
			result.Candidate = strings.TrimSpace(strings.TrimPrefix(trimmed, "Candidate:"))
		} else if strings.HasPrefix(trimmed, "Version table:") || strings.HasPrefix(trimmed, "***") || strings.HasPrefix(trimmed, "500") {
			result.VersionTable += line + "\n"
		}
	}
	return result, nil
}

// MarkAuto marks a package as automatically installed.
func MarkAuto(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "markauto", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apt-get markauto: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// MarkManual marks a package as manually installed.
func MarkManual(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	aptGet, err := findAptGet()
	if err != nil {
		return Result{Status: "failed", Error: err.Error()}, err
	}
	out, err := runCmd(aptGet, "unmarkauto", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("apt-get unmarkauto: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}
