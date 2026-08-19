// Package multipath manages device-mapper multipathing.
// Equivalent to community.general.multipath module.
package multipath

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
	Action  string `json:"action,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PathResult is returned by path operations.
type PathResult struct {
	Status   string     `json:"status"`
	Paths    []PathInfo `json:"paths"`
	Count    int        `json:"count"`
	Error    string     `json:"error,omitempty"`
}

// PathInfo represents a multipath device path.
type PathInfo struct {
	HostChannelID string `json:"host_channel_id"`
	Device        string `json:"device"`
	Status        string `json:"status"`
	Size          string `json:"size,omitempty"`
}

// MapResult is returned by ListMaps.
type MapResult struct {
	Status string   `json:"status"`
	Maps   []string `json:"maps"`
	Error  string   `json:"error,omitempty"`
}

// FlushResult is returned by Flush.
type FlushResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Device  string `json:"device"`
	Error   string `json:"error,omitempty"`
}

// Reconfigure reloads multipath configuration.
func Reconfigure() Result {
	cmd := exec.Command("multipathd", "reconfigure")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("multipathd reconfigure: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Action: "reconfigure"}
}

// ListPaths returns all multipath paths.
func ListPaths() PathResult {
	cmd := exec.Command("multipath", "-ll")
	out, err := cmd.Output()
	if err != nil {
		return PathResult{Status: "failed", Error: fmt.Sprintf("multipath: %v", err)}
	}

	paths := make([]PathInfo, 0)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "sd") && strings.Contains(line, "active") || strings.Contains(line, "faulty") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				paths = append(paths, PathInfo{
					Device: parts[0],
					Status: "unknown",
				})
			}
		}
	}

	return PathResult{Status: "success", Paths: paths, Count: len(paths)}
}

// ListMaps returns all multipath maps.
func ListMaps() MapResult {
	cmd := exec.Command("multipath", "-ll")
	out, err := cmd.Output()
	if err != nil {
		return MapResult{Status: "failed", Error: fmt.Sprintf("multipath: %v", err)}
	}

	maps := make([]string, 0)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "dm-") && !strings.HasPrefix(line, " ") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				maps = append(maps, parts[0])
			}
		}
	}
	return MapResult{Status: "success", Maps: maps}
}

// AddMap adds a device to multipath management.
func AddMap(device string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device is required"}
	}

	cmd := exec.Command("multipathd", "add", "map", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("multipathd add: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Action: "add_map"}
}

// RemoveMap removes a device from multipath management.
func RemoveMap(device string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device is required"}
	}

	cmd := exec.Command("multipathd", "remove", "map", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("multipathd remove: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Action: "remove_map"}
}

// Flush removes all multipath maps.
func Flush() FlushResult {
	cmd := exec.Command("multipath", "-F")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return FlushResult{Status: "failed", Error: fmt.Sprintf("multipath flush: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return FlushResult{Status: "success", Changed: true, Device: "all"}
}
