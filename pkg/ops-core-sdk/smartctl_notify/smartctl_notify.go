// Package smartctl_notify provides disk health monitoring via SMART.
package smartctl_notify

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DiskHealth represents the SMART health status of a disk.
type DiskHealth struct {
	Device        string `json:"device"`
	Healthy       bool   `json:"healthy"`
	Model         string `json:"model,omitempty"`
	Serial        string `json:"serial,omitempty"`
	Temperature   string `json:"temperature,omitempty"`
	PowerOnHours  string `json:"power_on_hours,omitempty"`
	Reallocated   string `json:"reallocated_sectors,omitempty"`
	Pending       string `json:"pending_sectors,omitempty"`
	Uncorrectable string `json:"uncorrectable_sectors,omitempty"`
}

// ActionResult represents the result of a smartctl action.
type ActionResult struct {
	Device     string `json:"device"`
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	DurationMs int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Check returns the SMART health status of a disk.
func Check(device string) (*DiskHealth, error) {
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}

	cmd := exec.Command("smartctl", "-H", "-A", "-i", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &DiskHealth{Device: device, Healthy: false}, nil
	}

	result := &DiskHealth{Device: device, Healthy: true}
	output := string(out)

	// Check overall health
	if strings.Contains(output, "FAILED") {
		result.Healthy = false
	}

	// Parse attributes
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Model Family:") || strings.HasPrefix(line, "Device Model:") || strings.HasPrefix(line, "Model Number:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result.Model = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Serial Number:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result.Serial = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Temperature:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result.Temperature = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Power_On_Hours") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				result.PowerOnHours = fields[9]
			}
		}
		if strings.HasPrefix(line, "Reallocated_Sector_Ct") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				result.Reallocated = fields[9]
			}
		}
		if strings.HasPrefix(line, "Current_Pending_Sector") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				result.Pending = fields[9]
			}
		}
		if strings.HasPrefix(line, "Offline_Uncorrectable") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				result.Uncorrectable = fields[9]
			}
		}
	}

	return result, nil
}

// ListDevices returns available disk devices.
func ListDevices() ([]string, error) {
	cmd := exec.Command("smartctl", "--scan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("smartctl scan: %s", string(out))
	}

	var devices []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			devices = append(devices, fields[0])
		}
	}
	return devices, nil
}

// ShortTest runs a SMART short self-test.
func ShortTest(device string) (*ActionResult, error) {
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}

	start := time.Now()
	cmd := exec.Command("smartctl", "-t", "short", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Device:     device,
			Changed:    false,
			Action:     "short_test",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(out),
			Error:      err.Error(),
		}, fmt.Errorf("smartctl short test: %s", string(out))
	}
	return &ActionResult{
		Device:     device,
		Changed:    true,
		Action:     "short_test",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(string(out)),
	}, nil
}

// LongTest runs a SMART extended self-test.
func LongTest(device string) (*ActionResult, error) {
	if device == "" {
		return nil, fmt.Errorf("device path is required")
	}

	start := time.Now()
	cmd := exec.Command("smartctl", "-t", "long", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Device:     device,
			Changed:    false,
			Action:     "long_test",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(out),
			Error:      err.Error(),
		}, fmt.Errorf("smartctl long test: %s", string(out))
	}
	return &ActionResult{
		Device:     device,
		Changed:    true,
		Action:     "long_test",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     strings.TrimSpace(string(out)),
	}, nil
}
