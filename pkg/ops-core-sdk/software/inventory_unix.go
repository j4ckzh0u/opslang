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
	var rpmFiles map[string][]string
	if len(infos) > 0 && isRPMManager(infos[0].Manager) {
		var rpmErr error
		rpmFiles, rpmErr = allRPMPackageFiles()
		if rpmErr != nil {
			errors = append(errors, Error{Scope: "package_files", Message: rpmErr.Error()})
		}
	}
	for _, info := range infos {
		item := Package{Name: info.Name, Version: info.Version, Architecture: info.Architecture, Manager: info.Manager}
		var files []string
		var fileErr error
		if rpmFiles != nil && isRPMManager(info.Manager) {
			files = rpmFiles[rpmPackageKey(info.Name, info.Architecture)]
			if files == nil {
				fileErr = fmt.Errorf("rpm package file manifest not found for %q", info.Name)
			}
		} else {
			files, fileErr = packageFiles(info.Manager, info.Name, info.Architecture)
		}
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

func isRPMManager(manager string) bool {
	return manager == "yum" || manager == "dnf"
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

// RPM queryformat iterates FILENAMES once per file, allowing all package
// manifests to be collected in one process while preserving package mapping.
func allRPMPackageFiles() (map[string][]string, error) {
	const queryFormat = "[%{=NAME}|%{=ARCH}|%{FILENAMES}\n]"
	out, err := exec.Command("rpm", "-qa", "--queryformat", queryFormat).Output()
	if err != nil {
		return nil, fmt.Errorf("rpm package file inventory failed: %w", err)
	}
	return parseRPMFileListOutput(string(out)), nil
}

func rpmPackageKey(name, architecture string) string {
	return name + "\x00" + architecture
}

func parseRPMFileListOutput(output string) map[string][]string {
	files := make(map[string][]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) != 3 || parts[0] == "" || parts[2] == "" {
			continue
		}
		key := rpmPackageKey(parts[0], parts[1])
		files[key] = append(files[key], parts[2])
	}
	return files
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
