package sys

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// MountResult is returned by Mount, reporting whether a mount was performed.
type MountResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UnmountResult is returned by Unmount, reporting whether an unmount was performed.
type UnmountResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// MountInfo represents a single entry from /proc/mounts.
type MountInfo struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	FSType     string `json:"fs_type"`
	Options    string `json:"options"`
}

// parseProcMounts reads and parses /proc/mounts, returning a list of MountInfo entries.
func parseProcMounts() ([]MountInfo, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/mounts: %w", err)
	}
	defer f.Close()

	var mounts []MountInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		mounts = append(mounts, MountInfo{
			Device:     fields[0],
			MountPoint: fields[1],
			FSType:     fields[2],
			Options:    fields[3],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read /proc/mounts: %w", err)
	}
	return mounts, nil
}

// isMounted checks whether a given mountpoint exists in /proc/mounts.
func isMounted(mountpoint string) (bool, error) {
	mounts, err := parseProcMounts()
	if err != nil {
		return false, err
	}
	for _, m := range mounts {
		if m.MountPoint == mountpoint {
			return true, nil
		}
	}
	return false, nil
}

// Mount mounts a device at the given mountpoint.
// If the mountpoint is already mounted, returns Changed: false.
// fsType may be empty to let the kernel auto-detect.
// opts may contain an "options" key whose value is a comma-separated mount option string.
func Mount(device string, mountpoint string, fsType string, opts map[string]string) (MountResult, error) {
	if device == "" {
		return MountResult{}, fmt.Errorf("failed to mount: device must not be empty")
	}
	if mountpoint == "" {
		return MountResult{}, fmt.Errorf("failed to mount: mountpoint must not be empty")
	}

	mounted, err := isMounted(mountpoint)
	if err != nil {
		return MountResult{}, fmt.Errorf("failed to check mount status: %w", err)
	}
	if mounted {
		return MountResult{Changed: false}, nil
	}

	args := []string{}
	if fsType != "" {
		args = append(args, "-t", fsType)
	}
	if optStr, ok := opts["options"]; ok && optStr != "" {
		args = append(args, "-o", optStr)
	}
	args = append(args, device, mountpoint)

	cmd := exec.Command("mount", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return MountResult{}, fmt.Errorf("failed to mount %s on %s: %w: %s", device, mountpoint, err, strings.TrimSpace(string(output)))
	}

	return MountResult{Changed: true}, nil
}

// Unmount unmounts the given mountpoint.
// If the mountpoint is not mounted, returns Changed: false.
func Unmount(mountpoint string) (UnmountResult, error) {
	if mountpoint == "" {
		return UnmountResult{}, fmt.Errorf("failed to unmount: mountpoint must not be empty")
	}

	mounted, err := isMounted(mountpoint)
	if err != nil {
		return UnmountResult{}, fmt.Errorf("failed to check mount status: %w", err)
	}
	if !mounted {
		return UnmountResult{Changed: false}, nil
	}

	cmd := exec.Command("umount", mountpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return UnmountResult{}, fmt.Errorf("failed to unmount %s: %w: %s", mountpoint, err, strings.TrimSpace(string(output)))
	}

	return UnmountResult{Changed: true}, nil
}

// ListMounts returns all currently mounted filesystems from /proc/mounts.
func ListMounts() ([]MountInfo, error) {
	mounts, err := parseProcMounts()
	if err != nil {
		return nil, err
	}
	if mounts == nil {
		mounts = []MountInfo{}
	}
	return mounts, nil
}
