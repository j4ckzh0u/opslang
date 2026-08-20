// Package dnf provides Fedora/RHEL 8+ DNF package management operations.
// Equivalent to Ansible's dnf module.
package dnf

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of a dnf operation.
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
	Release      string `json:"release"`
	Architecture string `json:"architecture"`
	Size         string `json:"size"`
	Source       string `json:"source,omitempty"`
	Repo         string `json:"repo,omitempty"`
	Summary      string `json:"summary,omitempty"`
	URL          string `json:"url,omitempty"`
	License      string `json:"license,omitempty"`
}

// RepoResult represents repository information.
type RepoResult struct {
	Status     string `json:"status"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	State      string `json:"state"`
	Enabled    bool   `json:"enabled"`
	Revision   string `json:"revision,omitempty"`
	Size       string `json:"size,omitempty"`
}

func findDnf() (string, error) {
	for _, p := range []string{"/usr/bin/dnf", "/usr/local/bin/dnf"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("dnf"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("dnf not found")
}

func runCmd(args ...string) (string, error) {
	dnf, err := findDnf()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(dnf, args...)
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
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("dnf install: %v", err)}, err
	}
	changed := !strings.Contains(out, "already installed") && !strings.Contains(out, "Nothing to do")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: out}, nil
}

// Remove removes a package.
func Remove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("remove", "-y", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("dnf remove: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates a specific package or all packages if name is empty.
func Update(name string) (Result, error) {
	if name == "" {
		out, err := runCmd("upgrade", "-y")
		if err != nil {
			return Result{Status: "failed", Output: out, Error: fmt.Sprintf("dnf upgrade: %v", err)}, err
		}
		return Result{Status: "success", Changed: true, Output: out}, nil
	}
	out, err := runCmd("upgrade", "-y", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("dnf upgrade: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
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
	info := InfoResult{Status: "success", Package: name}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Version":
			info.Version = val
		case "Release":
			info.Release = val
		case "Arch":
			info.Architecture = val
		case "Size":
			info.Size = val
		case "Source":
			info.Source = val
		case "Repository":
			info.Repo = val
		case "Summary":
			info.Summary = val
		case "URL":
			info.URL = val
		case "License":
			info.License = val
		}
	}
	return info, nil
}

// List lists installed packages.
func List() ([]InfoResult, error) {
	out, err := runCmd("list", "--installed", "-q")
	if err != nil {
		return nil, fmt.Errorf("dnf list: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pkg := fields[0]
		verArch := fields[1]
		parts := strings.Split(verArch, ".")
		ver := ""
		arch := ""
		if len(parts) >= 2 {
			arch = parts[len(parts)-1]
			ver = strings.Join(parts[:len(parts)-1], ".")
		} else {
			ver = verArch
		}
		repo := ""
		if len(fields) >= 3 {
			repo = fields[2]
		}
		results = append(results, InfoResult{
			Status:       "success",
			Package:      pkg,
			Version:      ver,
			Architecture: arch,
			Repo:         repo,
		})
	}
	return results, nil
}

// Search searches for a package.
func Search(name string) ([]InfoResult, error) {
	if name == "" {
		return nil, fmt.Errorf("search name is required")
	}
	out, err := runCmd("search", name)
	if err != nil {
		return nil, fmt.Errorf("dnf search: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "=") || strings.HasPrefix(line, "Last") || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pkgParts := strings.Split(fields[0], ".")
		pkg := pkgParts[0]
		arch := ""
		if len(pkgParts) >= 2 {
			arch = pkgParts[len(pkgParts)-1]
		}
		results = append(results, InfoResult{
			Status:       "success",
			Package:      pkg,
			Version:      fields[1],
			Architecture: arch,
			Summary:      strings.Join(fields[2:], " "),
		})
	}
	return results, nil
}

// Clean cleans dnf cache.
func Clean() (Result, error) {
	out, err := runCmd("clean", "all")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("dnf clean: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// RepoList lists repositories.
func RepoList() ([]RepoResult, error) {
	out, err := runCmd("repolist", "-v")
	if err != nil {
		return nil, fmt.Errorf("dnf repolist: %w", err)
	}
	var results []RepoResult
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		state := "enabled"
		enabled := true
		if strings.Contains(line, "disabled") {
			state = "disabled"
			enabled = false
		}
		results = append(results, RepoResult{
			Status:  "success",
			ID:      fields[0],
			Name:    strings.Join(fields[1:], " "),
			State:   state,
			Enabled: enabled,
		})
	}
	return results, nil
}

// GroupList lists package groups.
func GroupList() ([]string, error) {
	out, err := runCmd("grouplist", "-q")
	if err != nil {
		return nil, fmt.Errorf("dnf grouplist: %w", err)
	}
	var groups []string
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "Last") {
			groups = append(groups, trimmed)
		}
	}
	return groups, nil
}

// GroupInstall installs a package group.
func GroupInstall(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "group name is required"}, fmt.Errorf("group name is required")
	}
	out, err := runCmd("groupinstall", "-y", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("dnf groupinstall: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// GroupRemove removes a package group.
func GroupRemove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "group name is required"}, fmt.Errorf("group name is required")
	}
	out, err := runCmd("groupremove", "-y", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("dnf groupremove: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// History shows dnf transaction history.
func History(count int) (string, error) {
	if count <= 0 {
		count = 10
	}
	args := []string{"history", "list"}
	out, err := runCmd(args...)
	if err != nil {
		return "", fmt.Errorf("dnf history: %w", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) > count+1 {
		lines = lines[:count+1]
	}
	return strings.Join(lines, "\n"), nil
}

// CheckUpdate checks for available updates.
func CheckUpdate() ([]InfoResult, error) {
	out, err := runCmd("check-update", "-q")
	if err != nil {
		// dnf check-update returns exit code 100 if updates available
		if out == "" {
			return nil, nil
		}
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, " ") || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pkg := strings.Split(fields[0], ".")[0]
		results = append(results, InfoResult{
			Status:  "success",
			Package: pkg,
			Version: fields[1],
		})
	}
	return results, nil
}

// ModuleList lists dnf modules.
func ModuleList() ([]string, error) {
	out, err := runCmd("module", "list", "-q")
	if err != nil {
		return nil, fmt.Errorf("dnf module list: %w", err)
	}
	return strings.Split(strings.TrimSpace(out), "\n"), nil
}

// ModuleEnable enables a dnf module stream.
func ModuleEnable(spec string) (Result, error) {
	if spec == "" {
		return Result{Status: "failed", Error: "module spec is required"}, fmt.Errorf("module spec is required")
	}
	out, err := runCmd("module", "enable", "-y", spec)
	if err != nil {
		return Result{Status: "failed", Package: spec, Output: out, Error: fmt.Sprintf("dnf module enable: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: spec, Output: out}, nil
}
