// Package smartctl provides SMART disk health monitoring.
package smartctl

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// DeviceInfo represents SMART device information.
type DeviceInfo struct {
	Device       string `json:"device"`
	Model        string `json:"model,omitempty"`
	Serial       string `json:"serial,omitempty"`
	Firmware     string `json:"firmware,omitempty"`
	Capacity     string `json:"capacity,omitempty"`
	RotationRate int    `json:"rotation_rate,omitempty"`
	FormFactor   string `json:"form_factor,omitempty"`
	Error        string `json:"error,omitempty"`
}

// HealthResult represents SMART health status.
type HealthResult struct {
	Device     string `json:"device"`
	SmartPass  bool   `json:"smart_pass"`
	Healthy    bool   `json:"healthy"`
	Overall    string `json:"overall,omitempty"`
	Error      string `json:"error,omitempty"`
}

// AttributesResult represents SMART attributes.
type AttributesResult struct {
	Device     string      `json:"device"`
	Attributes []Attribute `json:"attributes"`
	Error      string      `json:"error,omitempty"`
}

// Attribute represents a single SMART attribute.
type Attribute struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Value     int    `json:"value"`
	Worst     int    `json:"worst"`
	Thresh    int    `json:"thresh"`
	WhenFailed string `json:"when_failed,omitempty"`
	RawValue  string `json:"raw_value,omitempty"`
}

// ListResult represents a list of devices.
type ListResult struct {
	Devices []string `json:"devices"`
	Count   int      `json:"count"`
	Error   string   `json:"error,omitempty"`
}

func smartctl(args ...string) (string, error) {
	cmd := exec.Command("smartctl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Device returns SMART device information.
func Device(device string) DeviceInfo {
	if device == "" {
		return DeviceInfo{Error: "device is required"}
	}
	out, err := smartctl("-i", device)
	if err != nil {
		return DeviceInfo{Device: device, Error: fmt.Sprintf("smartctl -i failed: %s: %s", err, out)}
	}
	info := DeviceInfo{Device: device}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Model Number:") {
			info.Model = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Serial Number:") {
			info.Serial = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Firmware Version:") {
			info.Firmware = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "User Capacity:") {
			info.Capacity = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Rotation Rate:") {
			val := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			fmt.Sscanf(val, "%d", &info.RotationRate)
		} else if strings.Contains(line, "Form Factor:") {
			info.FormFactor = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}
	return info
}

// Health returns SMART health status.
func Health(device string) HealthResult {
	if device == "" {
		return HealthResult{Error: "device is required"}
	}
	out, err := smartctl("-H", device)
	if err != nil {
		return HealthResult{Device: device, Error: fmt.Sprintf("smartctl -H failed: %s: %s", err, out)}
	}
	result := HealthResult{Device: device, SmartPass: true, Healthy: true}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "SMART overall-health self-assessment test result") {
			result.Overall = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if strings.Contains(strings.ToLower(result.Overall), "failed") {
				result.Healthy = false
			}
		}
		if strings.Contains(line, "SMART Health Status: OK") {
			result.SmartPass = true
		}
	}
	return result
}

// Attributes returns SMART attributes.
func Attributes(device string) AttributesResult {
	if device == "" {
		return AttributesResult{Error: "device is required"}
	}
	out, err := smartctl("-A", device)
	if err != nil {
		return AttributesResult{Device: device, Error: fmt.Sprintf("smartctl -A failed: %s: %s", err, out)}
	}
	var attrs []Attribute
	inTable := false
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ID#") {
			inTable = true
			continue
		}
		if inTable {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				attr := Attribute{}
				fmt.Sscanf(fields[0], "%d", &attr.ID)
				attr.Name = fields[1]
				fmt.Sscanf(fields[3], "%d", &attr.Value)
				fmt.Sscanf(fields[4], "%d", &attr.Worst)
				fmt.Sscanf(fields[5], "%d", &attr.Thresh)
				attr.WhenFailed = fields[6]
				attr.RawValue = fields[9]
				attrs = append(attrs, attr)
			} else {
				inTable = false
			}
		}
	}
	return AttributesResult{Device: device, Attributes: attrs}
}

// List lists available block devices that support SMART.
func List() ListResult {
	out, err := smartctl("--scan")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("smartctl --scan failed: %s: %s", err, out)}
	}
	var devices []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				devices = append(devices, fields[0])
			}
		}
	}
	return ListResult{Devices: devices, Count: len(devices)}
}

// JSON returns full SMART output as JSON.
func JSON(device string) (string, error) {
	if device == "" {
		return "", fmt.Errorf("device is required")
	}
	out, err := smartctl("-j", "-a", device)
	if err != nil {
		return "", fmt.Errorf("smartctl -j failed: %w: %s", err, out)
	}
	// Validate it's valid JSON
	var js json.RawMessage
	if json.Unmarshal([]byte(out), &js) != nil {
		return "", fmt.Errorf("invalid JSON output from smartctl")
	}
	return out, nil
}
