//go:build !windows

package software

import (
	"fmt"
	"os/exec"
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
		files, fileErr := packageFiles(info.Manager, info.Name)
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

func packageFiles(manager, name string) ([]string, error) {
	var command string
	var args []string
	switch manager {
	case "apt":
		command, args = "dpkg-query", []string{"-L", name}
	case "yum", "dnf":
		command, args = "rpm", []string{"-ql", name}
	default:
		return nil, fmt.Errorf("unsupported package manager %q", manager)
	}
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%s package file query failed: %w", manager, err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if value := strings.TrimSpace(line); value != "" {
			files = append(files, value)
		}
	}
	return files, nil
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
