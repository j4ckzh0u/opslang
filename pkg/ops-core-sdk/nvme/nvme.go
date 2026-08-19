// Package nvme provides NVMe device management.
package nvme

import (
	"fmt"
	"os/exec"
	"strings"
)

// DeviceInfo represents an NVMe device.
type DeviceInfo struct {
	Device     string `json:"device"`
	Model      string `json:"model,omitempty"`
	Serial     string `json:"serial,omitempty"`
	Firmware   string `json:"firmware,omitempty"`
	Namespaces int    `json:"namespaces,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ListResult is returned by device listing.
type ListResult struct {
	Devices []DeviceInfo `json:"devices"`
	Count   int          `json:"count"`
	Error   string       `json:"error,omitempty"`
}

// Result is returned by NVMe operations.
type Result struct {
	Device  string `json:"device,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

func nvme(args ...string) (string, error) {
	cmd := exec.Command("nvme", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// List lists NVMe devices.
func List() ListResult {
	out, err := nvme("list")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("nvme list failed: %s: %s", err, out)}
	}
	var devices []DeviceInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Node") || strings.HasPrefix(line, "---") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			devices = append(devices, DeviceInfo{
				Device:   fields[0],
				Model:    fields[2],
				Serial:   fields[3],
				Firmware: fields[4],
			})
		}
	}
	return ListResult{Devices: devices, Count: len(devices)}
}

// SmartLog returns SMART log for a device.
func SmartLog(device string) (string, error) {
	if device == "" {
		return "", fmt.Errorf("device is required")
	}
	out, err := nvme("smart-log", device)
	if err != nil {
		return "", fmt.Errorf("nvme smart-log failed: %w: %s", err, out)
	}
	return out, nil
}

// FirmwareLog returns firmware log for a device.
func FirmwareLog(device string) (string, error) {
	if device == "" {
		return "", fmt.Errorf("device is required")
	}
	out, err := nvme("fw-log", device)
	if err != nil {
		return "", fmt.Errorf("nvme fw-log failed: %w: %s", err, out)
	}
	return out, nil
}

// ErrorLog returns error log for a device.
func ErrorLog(device string) (string, error) {
	if device == "" {
		return "", fmt.Errorf("device is required")
	}
	out, err := nvme("error-log", device)
	if err != nil {
		return "", fmt.Errorf("nvme error-log failed: %w: %s", err, out)
	}
	return out, nil
}

// Version returns nvme-cli version.
func Version() (string, error) {
	out, err := nvme("version")
	if err != nil {
		return "", fmt.Errorf("nvme version failed: %w: %s", err, out)
	}
	return strings.TrimSpace(out), nil
}
