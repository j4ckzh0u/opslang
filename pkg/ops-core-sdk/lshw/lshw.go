// Package lshw provides hardware listing via lshw.
package lshw

import (
	"fmt"
	"os/exec"
	"strings"
)

// HardwareInfo represents hardware component info.
type HardwareInfo struct {
	ID          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Product     string `json:"product,omitempty"`
	Vendor      string `json:"vendor,omitempty"`
	PhysID      string `json:"phys_id,omitempty"`
	BusInfo     string `json:"bus_info,omitempty"`
	Version     string `json:"version,omitempty"`
	Serial      string `json:"serial,omitempty"`
	Width       string `json:"width,omitempty"`
	Clock       string `json:"clock,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ListResult is returned by hardware listing.
type ListResult struct {
	Components []HardwareInfo `json:"components"`
	Count      int            `json:"count"`
	Error      string         `json:"error,omitempty"`
}

func lshw(args ...string) (string, error) {
	cmd := exec.Command("lshw", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Short returns a short hardware listing.
func Short() ListResult {
	out, err := lshw("-short")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("lshw -short failed: %s: %s", err, out)}
	}
	var components []HardwareInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "H/W") || strings.HasPrefix(line, "=") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			components = append(components, HardwareInfo{
				Description: fields[1],
				ID:          fields[0],
				Product:     strings.Join(fields[2:], " "),
			})
		}
	}
	return ListResult{Components: components, Count: len(components)}
}

// Class returns hardware of a specific class.
func Class(class string) (string, error) {
	if class == "" {
		return "", fmt.Errorf("class is required")
	}
	out, err := lshw("-C", class)
	if err != nil {
		return "", fmt.Errorf("lshw -C failed: %w: %s", err, out)
	}
	return out, nil
}

// JSON returns full hardware listing as JSON.
func JSON() (string, error) {
	out, err := lshw("-json")
	if err != nil {
		return "", fmt.Errorf("lshw -json failed: %w: %s", err, out)
	}
	return out, nil
}

// System returns system summary.
func System() (string, error) {
	out, err := lshw("-C", "system")
	if err != nil {
		return "", fmt.Errorf("lshw system failed: %w: %s", err, out)
	}
	return out, nil
}

// Memory returns memory information.
func Memory() (string, error) {
	out, err := lshw("-C", "memory")
	if err != nil {
		return "", fmt.Errorf("lshw memory failed: %w: %s", err, out)
	}
	return out, nil
}

// Disk returns disk information.
func Disk() (string, error) {
	out, err := lshw("-C", "disk")
	if err != nil {
		return "", fmt.Errorf("lshw disk failed: %w: %s", err, out)
	}
	return out, nil
}

// Network returns network interface information.
func Network() (string, error) {
	out, err := lshw("-C", "network")
	if err != nil {
		return "", fmt.Errorf("lshw network failed: %w: %s", err, out)
	}
	return out, nil
}
