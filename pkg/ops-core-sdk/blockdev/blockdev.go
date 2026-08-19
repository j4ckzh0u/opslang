// Package blockdev provides block device management operations.
package blockdev

import (
	"fmt"
	"os/exec"
	"strings"
)

// Device represents a block device.
type Device struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Type string `json:"type"`
}

// ListResult represents the result of listing devices.
type ListResult struct {
	Devices []Device `json:"devices"`
}

// DeviceInfo represents detailed device info.
type DeviceInfo struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	Model      string `json:"model"`
	Rotational string `json:"rotational"`
	ReadOnly   string `json:"readonly"`
}

// ActionResult represents the result of a blockdev action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// List returns all block devices.
func List() (*ListResult, error) {
	out, err := exec.Command("lsblk", "-d", "-n", "-o", "NAME,SIZE,TYPE").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}

	result := &ListResult{Devices: make([]Device, 0)}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			dev := Device{
				Name: fields[0],
				Size: fields[1],
				Type: fields[2],
			}
			result.Devices = append(result.Devices, dev)
		}
	}

	return result, nil
}

// Info returns detailed info about a block device.
func Info(device string) (*DeviceInfo, error) {
	out, err := exec.Command("lsblk", "-d", "-n", "-o", "NAME,SIZE,MODEL,ROTA,RO", device).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}

	fields := strings.Fields(string(out))
	if len(fields) < 5 {
		return nil, fmt.Errorf("unexpected lsblk output format")
	}

	info := &DeviceInfo{
		Name:       fields[0],
		Size:       fields[1],
		Model:      fields[2],
		Rotational: fields[3],
		ReadOnly:   fields[4],
	}

	return info, nil
}

// FlushBuffers flushes buffers for a block device.
func FlushBuffers(device string) (*ActionResult, error) {
	cmd := exec.Command("blockdev", "--flushbufs", device)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("blockdev flush failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Flushed buffers for %s", device),
	}, nil
}

// SetReadahead sets the readahead value for a block device.
func SetReadahead(device string, value int) (*ActionResult, error) {
	cmd := exec.Command("blockdev", "--setra", fmt.Sprintf("%d", value), device)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("blockdev setra failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Set readahead for %s to %d", device, value),
	}, nil
}
