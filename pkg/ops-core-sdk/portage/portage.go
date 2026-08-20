// Package portage provides Gentoo Portage package management operations.
// Equivalent to Ansible's portage module.
package portage

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents the result of a portage operation.
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
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	License     string `json:"license,omitempty"`
	Installed   bool   `json:"installed"`
}

func findEmerge() (string, error) {
	for _, p := range []string{"/usr/bin/emerge", "/usr/sbin/emerge"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("emerge"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("emerge not found")
}

func runCmd(args ...string) (string, error) {
	emerge, err := findEmerge()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(emerge, args...)
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
		pkg = "=" + name + "-" + version
	}
	out, err := runCmd("--ask=n", pkg)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("emerge: %v", err)}, err
	}
	changed := !strings.Contains(out, "already installed")
	return Result{Status: "success", Changed: changed, Package: name, Version: version, Output: out}, nil
}

// Remove removes a package.
func Remove(name string) (Result, error) {
	if name == "" {
		return Result{Status: "failed", Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("--unmerge", "--ask=n", name)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("emerge --unmerge: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Update updates a package or all.
func Update(name string, deep bool) (Result, error) {
	args := []string{"--update", "--ask=n"}
	if deep {
		args = append(args, "--deep")
	}
	if name != "" {
		args = append(args, name)
	} else {
		args = append(args, "@world")
	}
	out, err := runCmd(args...)
	if err != nil {
		return Result{Status: "failed", Package: name, Output: out, Error: fmt.Sprintf("emerge --update: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Package: name, Output: out}, nil
}

// Sync syncs the Portage tree.
func Sync() (Result, error) {
	out, err := runCmd("--sync", "--ask=n")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("emerge --sync: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Info returns package information.
func Info(name string) (InfoResult, error) {
	if name == "" {
		return InfoResult{Status: "failed"}, fmt.Errorf("package name is required")
	}
	out, err := runCmd("--info", name)
	if err != nil {
		return InfoResult{Status: "success", Package: name}, nil
	}
	info := InfoResult{Status: "success", Package: name}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "*") {
			if info.Package == name {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					info.Package = strings.TrimPrefix(parts[0], "*")
				}
			}
		}
		if strings.Contains(line, "Homepage:") {
			info.Homepage = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "Description:") {
			info.Description = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "License:") {
			info.License = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}
	return info, nil
}

// List lists installed packages.
func List() ([]InfoResult, error) {
	out, err := runCmd("-ep", "@world")
	if err != nil {
		return nil, fmt.Errorf("emerge -ep @world: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "[eb") && !strings.Contains(line, "[I") {
			continue
		}
		fields := strings.Fields(line)
		for i, f := range fields {
			if strings.Contains(f, "/") && strings.Contains(f, "-") {
				parts := strings.Split(f, "/")
				if len(parts) == 2 {
					cat := parts[0]
					pkgVer := parts[1]
					idx := strings.LastIndex(pkgVer, "-")
					if idx > 0 {
						results = append(results, InfoResult{
							Status:    "success",
							Category:  cat,
							Package:   pkgVer[:idx],
							Version:   pkgVer[idx+1:],
							Installed: true,
						})
					}
				}
				break
			}
			if i > 10 {
				break
			}
		}
	}
	return results, nil
}

// Search searches for a package.
func Search(name string) ([]InfoResult, error) {
	if name == "" {
		return nil, fmt.Errorf("search name is required")
	}
	out, err := runCmd("--search", name)
	if err != nil {
		return nil, fmt.Errorf("emerge --search: %w", err)
	}
	var results []InfoResult
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "*") || strings.HasPrefix(line, " ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pkg := fields[0]
		if strings.Contains(pkg, "/") {
			parts := strings.Split(pkg, "/")
			if len(parts) == 2 {
				results = append(results, InfoResult{
					Status:   "success",
					Category: parts[0],
					Package:  parts[1],
				})
			}
		}
	}
	return results, nil
}

// Depclean removes unneeded packages.
func Depclean() (Result, error) {
	out, err := runCmd("--depclean", "--ask=n")
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("emerge --depclean: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// Metadata returns package metadata.
func Metadata(name string) (map[string]string, error) {
	if name == "" {
		return nil, fmt.Errorf("package name is required")
	}
	out, err := runCmd("--info", name)
	if err != nil {
		return nil, fmt.Errorf("emerge --info: %w", err)
	}
	meta := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				meta[key] = val
			}
		}
	}
	return meta, nil
}
