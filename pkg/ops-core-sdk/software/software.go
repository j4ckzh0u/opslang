// Package software collects installed software and currently running programs.
package software

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
)

type Package struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	Architecture     string   `json:"architecture,omitempty"`
	Manager          string   `json:"manager"`
	InstallLocation  string   `json:"install_location,omitempty"`
	InstalledFiles   []string `json:"installed_files,omitempty"`
	Publisher        string   `json:"publisher,omitempty"`
	UninstallCommand string   `json:"uninstall_command,omitempty"`
}

type RunningProgram struct {
	PID                 int32  `json:"pid"`
	Name                string `json:"name"`
	Version             string `json:"version,omitempty"`
	Executable          string `json:"executable,omitempty"`
	ExecutableDirectory string `json:"executable_directory,omitempty"`
	CommandLine         string `json:"command_line,omitempty"`
	PackageName         string `json:"package_name,omitempty"`
	PackageVersion      string `json:"package_version,omitempty"`
	User                string `json:"user,omitempty"`
	Status              string `json:"status,omitempty"`
}

type Error struct {
	Scope   string `json:"scope"`
	Item    string `json:"item,omitempty"`
	Message string `json:"message"`
}

type InventoryResult struct {
	Host            string           `json:"host"`
	OS              string           `json:"os"`
	OSVersion       string           `json:"os_version,omitempty"`
	Packages        []Package        `json:"packages"`
	RunningPrograms []RunningProgram `json:"running_programs"`
	Errors          []Error          `json:"errors,omitempty"`
	CollectedAt     time.Time        `json:"collected_at"`
}

// Inventory collects software metadata without changing the target machine.
// Per-item failures are retained in Errors so one unreadable process/package
// does not discard the rest of a host report.
func Inventory() (InventoryResult, error) {
	result := InventoryResult{Packages: make([]Package, 0), RunningPrograms: make([]RunningProgram, 0), Errors: make([]Error, 0), CollectedAt: time.Now().UTC()}
	info, err := host.Info()
	if err != nil {
		result.Errors = append(result.Errors, Error{Scope: "system", Message: err.Error()})
	} else {
		result.Host = info.Hostname
		result.OS = info.OS
		result.OSVersion = info.PlatformVersion
	}
	if result.OS == "" {
		result.OS = runtime.GOOS
	}

	packages, packageErrors := collectPackages()
	result.Packages = packages
	result.Errors = append(result.Errors, packageErrors...)
	programs, processErrors := collectRunningPrograms()
	enrichProgramPackages(programs, result.Packages)
	result.RunningPrograms = programs
	result.Errors = append(result.Errors, processErrors...)
	if len(result.Packages) == 0 && len(result.RunningPrograms) == 0 && len(result.Errors) > 0 {
		return result, fmt.Errorf("software inventory failed: %s", result.Errors[0].Message)
	}
	return result, nil
}

func enrichProgramPackages(programs []RunningProgram, packages []Package) {
	byFile := make(map[string]Package)
	for _, item := range packages {
		for _, file := range item.InstalledFiles {
			if file != "" {
				byFile[file] = item
			}
		}
	}
	for i := range programs {
		if item, ok := byFile[programs[i].Executable]; ok {
			programs[i].PackageName = item.Name
			programs[i].PackageVersion = item.Version
			programs[i].Version = item.Version
			continue
		}
		for _, item := range packages {
			location := strings.TrimRight(item.InstallLocation, `/\\`)
			executable := strings.ToLower(programs[i].Executable)
			locationLower := strings.ToLower(location)
			if location != "" && (strings.HasPrefix(executable, locationLower+"/") || strings.HasPrefix(executable, locationLower+`\`)) {
				programs[i].PackageName = item.Name
				programs[i].PackageVersion = item.Version
				programs[i].Version = item.Version
				break
			}
		}
	}
}

func collectRunningPrograms() ([]RunningProgram, []Error) {
	items := make([]RunningProgram, 0)
	errors := make([]Error, 0)
	procs, err := process.Processes()
	if err != nil {
		return items, []Error{{Scope: "processes", Message: err.Error()}}
	}
	for _, p := range procs {
		item := RunningProgram{PID: p.Pid}
		if value, err := p.Name(); err == nil {
			item.Name = value
		}
		if value, err := p.Exe(); err == nil {
			item.Executable = value
			item.ExecutableDirectory = executableDirectory(value)
		}
		if value, err := p.Cmdline(); err == nil {
			item.CommandLine = value
		}
		if value, err := p.Username(); err == nil {
			item.User = value
		}
		if values, err := p.Status(); err == nil && len(values) > 0 {
			item.Status = values[0]
		}
		if strings.TrimSpace(item.Name) == "" && strings.TrimSpace(item.Executable) == "" {
			errors = append(errors, Error{Scope: "process", Item: fmt.Sprint(p.Pid), Message: "process has no readable name or executable"})
			continue
		}
		items = append(items, item)
	}
	return items, errors
}

func executableDirectory(executable string) string {
	idx := strings.LastIndexAny(executable, `/\\`)
	if idx < 0 {
		return ""
	}
	return executable[:idx]
}
