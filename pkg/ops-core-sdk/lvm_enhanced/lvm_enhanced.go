// Package lvm_enhanced provides enhanced LVM operations.
// Complements the existing lvol package with pvcreate, vgextend, lvextend, etc.
package lvm_enhanced

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
	Size    string `json:"size,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by list operations.
type ListResult struct {
	Status string   `json:"status"`
	Items  []string `json:"items"`
	Error  string   `json:"error,omitempty"`
}

// PVCreate creates a physical volume.
func PVCreate(device string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device is required"}
	}

	cmd := exec.Command("pvcreate", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("pvcreate: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device}
}

// PVRemove removes a physical volume.
func PVRemove(device string, force bool) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device is required"}
	}

	args := []string{device}
	if force {
		args = []string{"-ff", "-y", device}
	}

	cmd := exec.Command("pvremove", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("pvremove: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device}
}

// PVList lists all physical volumes.
func PVList() ListResult {
	cmd := exec.Command("pvs", "--noheadings", "-o", "pv_name")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("pvs: %v", err)}
	}

	items := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	return ListResult{Status: "success", Items: items}
}

// VGCreate creates a volume group.
func VGCreate(name string, devices []string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}
	if len(devices) == 0 {
		return Result{Status: "failed", Error: "at least one device is required"}
	}

	args := append([]string{name}, devices...)
	cmd := exec.Command("vgcreate", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("vgcreate: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: name}
}

// VGRemove removes a volume group.
func VGRemove(name string, force bool) Result {
	if name == "" {
		return Result{Status: "failed", Error: "name is required"}
	}

	args := []string{name}
	if force {
		args = []string{"-ff", "-y", name}
	}

	cmd := exec.Command("vgremove", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("vgremove: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: name}
}

// VGExtend extends a volume group with a new physical volume.
func VGExtend(vgName string, device string) Result {
	if vgName == "" || device == "" {
		return Result{Status: "failed", Error: "vg name and device are required"}
	}

	cmd := exec.Command("vgextend", vgName, device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("vgextend: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: vgName}
}

// VGList lists all volume groups.
func VGList() ListResult {
	cmd := exec.Command("vgs", "--noheadings", "-o", "vg_name")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("vgs: %v", err)}
	}

	items := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	return ListResult{Status: "success", Items: items}
}

// LVExtend extends a logical volume by a given size.
func LVExtend(lvPath string, size string) Result {
	if lvPath == "" || size == "" {
		return Result{Status: "failed", Error: "lv path and size are required"}
	}

	cmd := exec.Command("lvextend", "-L", "+"+size, lvPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("lvextend: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: lvPath, Size: size}
}

// LVExtendAll extends a logical volume to use all free space.
func LVExtendAll(lvPath string) Result {
	if lvPath == "" {
		return Result{Status: "failed", Error: "lv path is required"}
	}

	cmd := exec.Command("lvextend", "-l", "+100%FREE", lvPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("lvextend: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: lvPath, Size: "all"}
}

// LVList lists all logical volumes.
func LVList() ListResult {
	cmd := exec.Command("lvs", "--noheadings", "-o", "lv_name,vg_name,lv_size")
	out, err := cmd.Output()
	if err != nil {
		return ListResult{Status: "failed", Error: fmt.Sprintf("lvs: %v", err)}
	}

	items := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			items = append(items, line)
		}
	}
	return ListResult{Status: "success", Items: items}
}
