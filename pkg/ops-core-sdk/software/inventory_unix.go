//go:build !windows

package software

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	opspkg "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pkg"
)

func collectPackages() ([]Package, []Error) {
	if runtime.GOOS != "linux" {
		return []Package{}, []Error{{Scope: "packages", Message: "installed package inventory is supported on Linux targets"}}
	}
	infos, err := opspkg.List()
	if err != nil {
		return []Package{}, []Error{{Scope: "packages", Message: err.Error()}}
	}
	items := make([]Package, 0, len(infos))
	errors := make([]Error, 0)
	for _, info := range infos {
		item := Package{Name: info.Name, Version: info.Version, Architecture: info.Architecture, Manager: info.Manager}
		files, fileErr := packageFiles(info.Manager, info.Name, info.Architecture)
		if fileErr != nil {
			errors = append(errors, Error{Scope: "package_files", Item: info.Name, Message: fileErr.Error()})
		} else {
			item.InstalledFiles = files
			item.InstallLocation = commonInstallLocation(files)
		}
		items = append(items, item)
	}
	return items, errors
}

func packageFiles(manager, name, architecture string) ([]string, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("package name must not be empty")
	}
	if manager == "apt" {
		return dpkgPackageFiles("/var/lib/dpkg/info", name, architecture)
	}

	var command string
	var args []string
	switch manager {
	case "yum", "dnf":
		command, args = "rpm", []string{"-ql", name}
	default:
		return nil, fmt.Errorf("unsupported package manager %q", manager)
	}
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%s package file query failed: %w", manager, err)
	}
	return parseInstalledFiles(string(out)), nil
}

// dpkg stores package file manifests locally, so reading them avoids spawning
// one dpkg-query process for every installed package.
func dpkgPackageFiles(infoDir, name, architecture string) ([]string, error) {
	candidates := []string{filepath.Join(infoDir, name+".list")}
	if architecture != "" && !strings.HasSuffix(name, ":"+architecture) {
		candidates = append(candidates, filepath.Join(infoDir, name+":"+architecture+".list"))
	}
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return parseInstalledFiles(string(data)), nil
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read dpkg package file manifest %q: %w", candidate, err)
		}
	}
	return nil, fmt.Errorf("dpkg package file manifest not found for %q", name)
}

func parseInstalledFiles(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if value := strings.TrimSpace(line); value != "" {
			files = append(files, value)
		}
	}
	return files
}

func commonInstallLocation(files []string) string {
	if len(files) == 0 {
		return ""
	}
	first := files[0]
	if strings.HasPrefix(first, "/") {
		parts := strings.Split(strings.Trim(first, "/"), "/")
		if len(parts) >= 2 {
			return "/" + parts[0] + "/" + parts[1]
		}
	}
	return ""
}
