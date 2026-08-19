// Package dmsetup manages device-mapper devices.
// Equivalent to community.general.dmsetup module.
package dmsetup

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Device  string `json:"device,omitempty"`
	Info    string `json:"info,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DeviceInfo represents device-mapper device information.
type DeviceInfo struct {
	Name      string `json:"name"`
	DevNumber string `json:"dev_number"`
	Targets   int    `json:"targets"`
	Events    string `json:"events,omitempty"`
}

// ListResult is returned by List.
type ListResult struct {
	Status  string       `json:"status"`
	Devices []DeviceInfo `json:"devices"`
	Error   string       `json:"error,omitempty"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	Status string     `json:"status"`
	Device DeviceInfo `json:"device"`
	Table  string     `json:"table,omitempty"`
	Info   string     `json:"info,omitempty"`
	Error  string     `json:"error,omitempty"`
}

// Create creates a new device-mapper device.
func Create(name string, table string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}
	if table == "" {
		return Result{Status: "failed", Error: "table is required"}
	}

	cmd := exec.Command("dmsetup", "create", name)
	cmd.Stdin = strings.NewReader(table)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("dmsetup create: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: name}
}

// Remove removes a device-mapper device.
func Remove(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}

	cmd := exec.Command("dmsetup", "remove", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("dmsetup remove: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: name}
}

// RemoveAll removes all device-mapper devices.
func RemoveAll() Result {
	cmd := exec.Command("dmsetup", "remove_all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("dmsetup remove_all: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true}
}

// List returns all device-mapper devices.
func List() ListResult {
	cmd := exec.Command("dmsetup", "ls")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("dmsetup ls: %v", err)}
	}

	devices := make([]DeviceInfo, 0)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "No devices") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			devices = append(devices, DeviceInfo{
				Name:      parts[0],
				DevNumber: strings.Trim(parts[1], "()"),
			})
		}
	}
	return ListResult{Status: "success", Devices: devices}
}

// Info returns information about a specific device.
func Info(name string) InfoResult {
	if name == "" {
		return InfoResult{Status: "failed", Error: "name is required"}
	}

	cmd := exec.Command("dmsetup", "info", name)
	out, err := cmd.Output()
	if err != nil {
		return InfoResult{Status: "failed", Error: fmt.Sprintf("dmsetup info: %v", err)}
	}

	dev := DeviceInfo{Name: name}
	info := strings.TrimSpace(string(out))

	// Get table
	cmd = exec.Command("dmsetup", "table", name)
	tableOut, _ := cmd.Output()

	return InfoResult{
		Status: "success",
		Device: dev,
		Table:  strings.TrimSpace(string(tableOut)),
		Info:   info,
	}
}

// Suspend suspends a device-mapper device.
func Suspend(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}

	cmd := exec.Command("dmsetup", "suspend", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("dmsetup suspend: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: name, Info: "suspended"}
}

// Resume resumes a suspended device-mapper device.
func Resume(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}

	cmd := exec.Command("dmsetup", "resume", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("dmsetup resume: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: name, Info: "resumed"}
}
