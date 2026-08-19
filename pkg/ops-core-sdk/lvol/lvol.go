// Package lvol provides LVM logical volume management operations.
package lvol

import (
	"fmt"
	"os/exec"
	"strings"
)

// Volume represents an LVM logical volume.
type Volume struct {
	Name   string `json:"name"`
	VG     string `json:"vg"`
	Size   string `json:"size"`
	Path   string `json:"path"`
}

// ListResult represents the result of listing logical volumes.
type ListResult struct {
	Volumes []Volume `json:"volumes"`
}

// ActionResult represents the result of an LVM action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// VGInfo represents a volume group.
type VGInfo struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Free string `json:"free"`
}

// VGListResult represents the result of listing volume groups.
type VGListResult struct {
	VolumeGroups []VGInfo `json:"volume_groups"`
}

// List returns all logical volumes.
func List() (*ListResult, error) {
	cmd := exec.Command("lvs", "--noheadings", "-o", "lv_name,vg_name,lv_size,path", "--separator", "|")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lvs failed: %w (output: %s)", err, string(out))
	}

	result := &ListResult{Volumes: make([]Volume, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			result.Volumes = append(result.Volumes, Volume{
				Name: strings.TrimSpace(parts[0]),
				VG:   strings.TrimSpace(parts[1]),
				Size: strings.TrimSpace(parts[2]),
				Path: strings.TrimSpace(parts[3]),
			})
		}
	}
	return result, nil
}

// VGList returns all volume groups.
func VGList() (*VGListResult, error) {
	cmd := exec.Command("vgs", "--noheadings", "-o", "vg_name,vg_size,vg_free", "--separator", "|")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("vgs failed: %w (output: %s)", err, string(out))
	}

	result := &VGListResult{VolumeGroups: make([]VGInfo, 0)}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			result.VolumeGroups = append(result.VolumeGroups, VGInfo{
				Name: strings.TrimSpace(parts[0]),
				Size: strings.TrimSpace(parts[1]),
				Free: strings.TrimSpace(parts[2]),
			})
		}
	}
	return result, nil
}

// Create creates a new logical volume.
func Create(name string, vg string, size string) (*ActionResult, error) {
	cmd := exec.Command("lvcreate", "-n", name, "-L", size, vg)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lvcreate failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Created LV %s/%s (%s)", vg, name, size)}, nil
}

// Remove removes a logical volume.
func Remove(name string, vg string) (*ActionResult, error) {
	lvPath := fmt.Sprintf("%s/%s", vg, name)
	cmd := exec.Command("lvremove", "-f", lvPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lvremove failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Removed LV %s", lvPath)}, nil
}

// Resize resizes a logical volume.
func Resize(name string, vg string, size string) (*ActionResult, error) {
	lvPath := fmt.Sprintf("%s/%s", vg, name)
	cmd := exec.Command("lvresize", "-L", size, lvPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lvresize failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Resized LV %s to %s", lvPath, size)}, nil
}
