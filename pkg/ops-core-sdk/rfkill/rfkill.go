// Package rfkill manages wireless device soft/hard blocks.
// Equivalent to community.general.rfkill module.
package rfkill

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Device   string `json:"device,omitempty"`
	Blocked  bool   `json:"blocked"`
	SoftBlock bool  `json:"soft_block"`
	HardBlock bool  `json:"hard_block"`
	Error    string `json:"error,omitempty"`
}

// DeviceInfo represents a wireless device.
type DeviceInfo struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Device    string `json:"device"`
	SoftBlock bool   `json:"soft_block"`
	HardBlock bool   `json:"hard_block"`
}

// ListResult is returned by List.
type ListResult struct {
	Status  string       `json:"status"`
	Devices []DeviceInfo `json:"devices"`
	Error   string       `json:"error,omitempty"`
}

// List returns all rfkill devices.
func List() ListResult {
	cmd := exec.Command("rfkill", "list")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("rfkill: %v", err)}
	}

	devices := make([]DeviceInfo, 0)
	blocks := strings.Split(strings.TrimSpace(string(out)), "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) == 0 {
			continue
		}

		dev := DeviceInfo{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, ":") || !strings.Contains(line, ":") {
				parts := strings.SplitN(lines[0], ":", 2)
				if len(parts) > 0 {
					dev.ID = strings.TrimSpace(parts[0])
				}
				if len(parts) > 1 {
					dev.Device = strings.TrimSpace(parts[1])
				}
				continue
			}
			kv := strings.SplitN(line, ":", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])

			switch key {
			case "Type":
				dev.Type = val
			case "Soft blocked":
				dev.SoftBlock = val == "Yes"
			case "Hard blocked":
				dev.HardBlock = val == "Yes"
			}
		}
		if dev.ID != "" {
			devices = append(devices, dev)
		}
	}

	return ListResult{Status: "success", Devices: devices}
}

// Block soft-blocks a device by index.
func Block(device string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device index is required"}
	}

	cmd := exec.Command("rfkill", "block", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("rfkill block: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Blocked: true, SoftBlock: true}
}

// Unblock soft-unblocks a device by index.
func Unblock(device string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device index is required"}
	}

	cmd := exec.Command("rfkill", "unblock", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("rfkill unblock: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Blocked: false, SoftBlock: false}
}

// BlockAll soft-blocks all devices of a type (wifi, bluetooth, etc.).
func BlockAll(deviceType string) Result {
	if deviceType == "" {
		return Result{Status: "failed", Error: "type is required"}
	}

	cmd := exec.Command("rfkill", "block", deviceType)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("rfkill block: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: deviceType, Blocked: true}
}

// UnblockAll soft-unblocks all devices of a type.
func UnblockAll(deviceType string) Result {
	if deviceType == "" {
		return Result{Status: "failed", Error: "type is required"}
	}

	cmd := exec.Command("rfkill", "unblock", deviceType)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("rfkill unblock: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: deviceType, Blocked: false}
}
