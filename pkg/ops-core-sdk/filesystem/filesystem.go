// Package filesystem provides filesystem creation and management operations.
package filesystem

import (
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a filesystem action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// MkfsResult represents the result of creating a filesystem.
type MkfsResult struct {
	Changed  bool   `json:"changed"`
	Device   string `json:"device"`
	FSType   string `json:"fstype"`
	Label    string `json:"label,omitempty"`
	UUID     string `json:"uuid,omitempty"`
}

// ResizeResult represents the result of resizing a filesystem.
type ResizeResult struct {
	Changed bool   `json:"changed"`
	Device  string `json:"device"`
	Message string `json:"message"`
}

// Mkfs creates a filesystem on a device.
func Mkfs(device string, fsType string, label string) (*MkfsResult, error) {
	args := []string{"-t", fsType}
	if label != "" {
		switch fsType {
		case "ext2", "ext3", "ext4":
			args = append(args, "-L", label)
		case "xfs":
			args = append(args, "-L", label)
		default:
			args = append(args, "-L", label)
		}
	}
	args = append(args, device)

	cmd := exec.Command("mkfs", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("mkfs failed: %w (output: %s)", err, string(out))
	}

	// Try to get UUID
	uuid := ""
	if blkidOut, err := exec.Command("blkid", "-s", "UUID", "-o", "value", device).Output(); err == nil {
		uuid = strings.TrimSpace(string(blkidOut))
	}

	return &MkfsResult{
		Changed: true,
		Device:  device,
		FSType:  fsType,
		Label:   label,
		UUID:    uuid,
	}, nil
}

// ResizeExt4 resizes an ext2/3/4 filesystem.
func ResizeExt4(device string) (*ResizeResult, error) {
	cmd := exec.Command("resize2fs", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("resize2fs failed: %w (output: %s)", err, string(out))
	}
	return &ResizeResult{
		Changed: true,
		Device:  device,
		Message: strings.TrimSpace(string(out)),
	}, nil
}

// ResizeXFS resizes an XFS filesystem (grow only).
func ResizeXFS(mountpoint string) (*ResizeResult, error) {
	cmd := exec.Command("xfs_growfs", mountpoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("xfs_growfs failed: %w (output: %s)", err, string(out))
	}
	return &ResizeResult{
		Changed: true,
		Device:  mountpoint,
		Message: strings.TrimSpace(string(out)),
	}, nil
}

// Check checks filesystem integrity (fsck).
func Check(device string) (*ActionResult, error) {
	cmd := exec.Command("fsck", "-n", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// fsck returns non-zero for "errors found" which is informational
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() <= 1 {
				return &ActionResult{Changed: false, Message: strings.TrimSpace(string(out))}, nil
			}
		}
		return nil, fmt.Errorf("fsck failed: %w (output: %s)", err, string(out))
	}
	return &ActionResult{Changed: false, Message: strings.TrimSpace(string(out))}, nil
}
