// Package opspkg provides package management operations for apt/yum/dnf.
// Each function returns a strongly-typed struct, never raw text.
package opspkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PackageAction represents the result of an install or remove operation.
type PackageAction struct {
	Name    string `json:"name"`
	Action  string `json:"action"`
	Manager string `json:"manager"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Ensure makes a package present. It is idempotent: an installed package is
// reported with Changed=false and is not passed to the package manager again.
func Ensure(name string) (PackageAction, error) {
	if strings.TrimSpace(name) == "" {
		return PackageAction{Name: name, Action: "ensure", Success: false, Error: "package name is required"}, fmt.Errorf("package name is required")
	}
	info, err := Info(name)
	if err == nil && strings.Contains(strings.ToLower(info.Status), "installed") {
		return PackageAction{Name: name, Action: "ensure", Manager: info.Manager, Success: true, Changed: false, Message: "package already installed"}, nil
	}
	result, installErr := Install(name)
	result.Action = "ensure"
	if installErr != nil {
		result.Error = fmt.Sprintf("ensure %s failed: %v", name, installErr)
		return result, installErr
	}
	result.Message = "package installed"
	return result, nil
}

// PackageInfo represents metadata about an installed package.
type PackageInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	Manager      string `json:"manager"`
}

// managerEntry holds the name and path of a supported package manager.
type managerEntry struct {
	name string
	path string
}

// supportedManagers defines the detection order and binary paths.
var supportedManagers = []managerEntry{
	{name: "apt", path: "/usr/bin/apt-get"},
	{name: "yum", path: "/usr/bin/yum"},
	{name: "dnf", path: "/usr/bin/dnf"},
}

// detectManager checks which package manager binary exists and returns
// the manager name and its full path. Returns an error if none is found.
func detectManager() (name string, path string, err error) {
	for _, m := range supportedManagers {
		if info, statErr := os.Stat(m.path); statErr == nil && !info.IsDir() {
			return m.name, m.path, nil
		}
	}
	return "", "", fmt.Errorf("no supported package manager found (checked apt-get, yum, dnf)")
}

// runCommand executes a command with the given arguments and returns the
// combined output. It does NOT invoke a shell.
func runCommand(binary string, args ...string) (string, error) {
	cmd := exec.Command(binary, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Install installs a package using the detected package manager.
func Install(name string) (PackageAction, error) {
	mgrName, mgrPath, err := detectManager()
	if err != nil {
		return PackageAction{Name: name, Action: "install", Success: false, Message: err.Error(), Error: err.Error()}, err
	}

	output, execErr := runCommand(mgrPath, "install", "-y", name)
	action := PackageAction{
		Name:    name,
		Action:  "install",
		Manager: mgrName,
		Success: execErr == nil,
		Changed: execErr == nil,
		Message: strings.TrimSpace(output),
	}
	if execErr != nil {
		action.Error = fmt.Sprintf("install %s with %s failed: %s: %s", name, mgrName, execErr, strings.TrimSpace(output))
		return action, fmt.Errorf("install %s with %s failed: %w: %s", name, mgrName, execErr, output)
	}
	return action, nil
}

// Remove removes a package using the detected package manager.
func Remove(name string) (PackageAction, error) {
	mgrName, mgrPath, err := detectManager()
	if err != nil {
		return PackageAction{Name: name, Action: "remove", Success: false, Message: err.Error(), Error: err.Error()}, err
	}

	output, execErr := runCommand(mgrPath, "remove", "-y", name)
	action := PackageAction{
		Name:    name,
		Action:  "remove",
		Manager: mgrName,
		Success: execErr == nil,
		Changed: execErr == nil,
		Message: strings.TrimSpace(output),
	}
	if execErr != nil {
		action.Error = fmt.Sprintf("remove %s with %s failed: %s: %s", name, mgrName, execErr, strings.TrimSpace(output))
		return action, fmt.Errorf("remove %s with %s failed: %w: %s", name, mgrName, execErr, output)
	}
	return action, nil
}

// Info returns metadata for a single installed package.
func Info(name string) (PackageInfo, error) {
	mgrName, _, err := detectManager()
	if err != nil {
		return PackageInfo{Name: name, Manager: ""}, err
	}

	switch mgrName {
	case "apt":
		return infoApt(name)
	case "yum", "dnf":
		return infoRpm(name, mgrName)
	default:
		return PackageInfo{}, fmt.Errorf("unsupported manager: %s", mgrName)
	}
}

// List returns metadata for all installed packages.
func List() ([]PackageInfo, error) {
	mgrName, _, err := detectManager()
	if err != nil {
		return nil, err
	}

	switch mgrName {
	case "apt":
		return listApt()
	case "yum", "dnf":
		return listRpm(mgrName)
	default:
		return nil, fmt.Errorf("unsupported manager: %s", mgrName)
	}
}

// parseDpkgInfoLine parses a single line from dpkg-query -W -f output.
// Format: "${Status}|${Version}|${Architecture}|${Description}"
func parseDpkgInfoLine(name, line, mgr string) PackageInfo {
	parts := strings.SplitN(line, "|", 4)
	info := PackageInfo{Name: name, Manager: mgr}
	if len(parts) >= 1 {
		info.Status = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		info.Version = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		info.Architecture = strings.TrimSpace(parts[2])
	}
	if len(parts) >= 4 {
		info.Description = strings.TrimSpace(parts[3])
	}
	return info
}

// infoApt queries package info via dpkg-query.
func infoApt(name string) (PackageInfo, error) {
	output, err := runCommand("dpkg-query", "-W", "-f",
		"${Status}|${Version}|${Architecture}|${Description}", name)
	if err != nil {
		return PackageInfo{Name: name, Manager: "apt"}, fmt.Errorf("dpkg-query failed for %s: %w: %s", name, err, output)
	}
	return parseDpkgInfoLine(name, strings.TrimSpace(output), "apt"), nil
}

// parseRpmInfo parses the output of rpm -qi into a PackageInfo.
func parseRpmInfo(name, output, mgr string) PackageInfo {
	info := PackageInfo{Name: name, Manager: mgr}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "Name":
			info.Name = val
		case "Version":
			info.Version = val
		case "Architecture":
			info.Architecture = val
		case "Summary", "Description":
			if info.Description == "" {
				info.Description = val
			}
		case "Status":
			info.Status = val
		}
	}
	if info.Status == "" {
		info.Status = "installed"
	}
	return info
}

// infoRpm queries package info via rpm -qi.
func infoRpm(name, mgr string) (PackageInfo, error) {
	output, err := runCommand("rpm", "-qi", name)
	if err != nil {
		return PackageInfo{Name: name, Manager: mgr}, fmt.Errorf("rpm -qi failed for %s: %w: %s", name, err, output)
	}
	return parseRpmInfo(name, output, mgr), nil
}

// parseDpkgListLine parses a single line from dpkg-query -W -f (list mode).
// Format: "${Package}|${Version}|${Architecture}|${Description}|${Status}\n"
func parseDpkgListLine(line, mgr string) PackageInfo {
	parts := strings.SplitN(line, "|", 5)
	info := PackageInfo{Manager: mgr}
	if len(parts) >= 1 {
		info.Name = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		info.Version = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		info.Architecture = strings.TrimSpace(parts[2])
	}
	if len(parts) >= 4 {
		info.Description = strings.TrimSpace(parts[3])
	}
	if len(parts) >= 5 {
		info.Status = strings.TrimSpace(parts[4])
	}
	return info
}

// listApt lists all installed packages via dpkg-query.
func listApt() ([]PackageInfo, error) {
	output, err := runCommand("dpkg-query", "-W", "-f",
		"${Package}|${Version}|${Architecture}|${Description}|${Status}\n")
	if err != nil {
		return nil, fmt.Errorf("dpkg-query list failed: %w: %s", err, output)
	}
	return parseDpkgListOutput(output, "apt"), nil
}

// parseDpkgListOutput splits dpkg-query list output into PackageInfo entries.
func parseDpkgListOutput(output, mgr string) []PackageInfo {
	var packages []PackageInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		packages = append(packages, parseDpkgListLine(line, mgr))
	}
	return packages
}

// parseRpmListLine parses a single line from rpm -qa --queryformat.
// Format: "%{NAME}|%{VERSION}-%{RELEASE}|%{ARCH}|%{SUMMARY}|installed\n"
func parseRpmListLine(line, mgr string) PackageInfo {
	parts := strings.SplitN(line, "|", 5)
	info := PackageInfo{Manager: mgr}
	if len(parts) >= 1 {
		info.Name = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		info.Version = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 {
		info.Architecture = strings.TrimSpace(parts[2])
	}
	if len(parts) >= 4 {
		info.Description = strings.TrimSpace(parts[3])
	}
	if len(parts) >= 5 {
		info.Status = strings.TrimSpace(parts[4])
	}
	return info
}

// listRpm lists all installed packages via rpm -qa.
func listRpm(mgr string) ([]PackageInfo, error) {
	output, err := runCommand("rpm", "-qa", "--queryformat",
		"%{NAME}|%{VERSION}-%{RELEASE}|%{ARCH}|%{SUMMARY}|installed\n")
	if err != nil {
		return nil, fmt.Errorf("rpm -qa failed: %w: %s", err, output)
	}
	return parseRpmListOutput(output, mgr), nil
}

// parseRpmListOutput splits rpm -qa output into PackageInfo entries.
func parseRpmListOutput(output, mgr string) []PackageInfo {
	var packages []PackageInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		packages = append(packages, parseRpmListLine(line, mgr))
	}
	return packages
}
