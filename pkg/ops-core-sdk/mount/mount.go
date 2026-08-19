// Package mount provides filesystem mount management operations.
package mount

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// MountInfo represents a mounted filesystem.
type MountInfo struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	FSType     string `json:"fs_type"`
	Options    string `json:"options"`
}

// ListResult represents the result of listing mounts.
type ListResult struct {
	Mounts []MountInfo `json:"mounts"`
}

// ActionResult represents the result of a mount operation.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// FstabEntry represents an /etc/fstab entry.
type FstabEntry struct {
	Device     string `json:"device"`
	MountPoint string `json:"mount_point"`
	FSType     string `json:"fs_type"`
	Options    string `json:"options"`
	Dump       string `json:"dump"`
	Pass       string `json:"pass"`
}

// FstabResult represents the result of listing fstab entries.
type FstabResult struct {
	Entries []FstabEntry `json:"entries"`
}

// List returns all mounted filesystems.
func List() (*ListResult, error) {
	out, err := exec.Command("mount").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("mount command failed: %w", err)
	}

	result := &ListResult{Mounts: make([]MountInfo, 0)}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		// Format: device on mountpoint type fstype (options)
		parts := strings.Fields(line)
		if len(parts) >= 5 && parts[1] == "on" && parts[3] == "type" {
			info := MountInfo{
				Device:     parts[0],
				MountPoint: parts[2],
				FSType:     parts[4],
			}
			if len(parts) >= 6 {
				info.Options = strings.Trim(parts[5], "()")
			}
			result.Mounts = append(result.Mounts, info)
		}
	}

	return result, nil
}

// Mount mounts a filesystem.
func Mount(device string, mountpoint string, fstype string, options string) (*ActionResult, error) {
	args := []string{}
	if fstype != "" {
		args = append(args, "-t", fstype)
	}
	if options != "" {
		args = append(args, "-o", options)
	}
	args = append(args, device, mountpoint)

	cmd := exec.Command("mount", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mount failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Mounted %s on %s", device, mountpoint),
	}, nil
}

// Unmount unmounts a filesystem.
func Unmount(mountpoint string) (*ActionResult, error) {
	cmd := exec.Command("umount", mountpoint)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("umount failed: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Unmounted %s", mountpoint),
	}, nil
}

// Fstab returns all entries from /etc/fstab.
func Fstab() (*FstabResult, error) {
	file, err := os.Open("/etc/fstab")
	if err != nil {
		if os.IsNotExist(err) {
			return &FstabResult{Entries: []FstabEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to open /etc/fstab: %w", err)
	}
	defer file.Close()

	result := &FstabResult{Entries: make([]FstabEntry, 0)}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 4 {
			entry := FstabEntry{
				Device:     parts[0],
				MountPoint: parts[1],
				FSType:     parts[2],
				Options:    parts[3],
			}
			if len(parts) >= 5 {
				entry.Dump = parts[4]
			}
			if len(parts) >= 6 {
				entry.Pass = parts[5]
			}
			result.Entries = append(result.Entries, entry)
		}
	}

	return result, nil
}

// AddFstab adds an entry to /etc/fstab.
func AddFstab(device string, mountpoint string, fstype string, options string) (*ActionResult, error) {
	if options == "" {
		options = "defaults"
	}

	file, err := os.OpenFile("/etc/fstab", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/fstab: %w", err)
	}
	defer file.Close()

	entry := fmt.Sprintf("%s\t%s\t%s\t%s\t0\t0\n", device, mountpoint, fstype, options)
	if _, err := file.WriteString(entry); err != nil {
		return nil, fmt.Errorf("failed to write to /etc/fstab: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Added %s to /etc/fstab", device),
	}, nil
}

// RemoveFstab removes an entry from /etc/fstab by device or mountpoint.
func RemoveFstab(target string) (*ActionResult, error) {
	file, err := os.Open("/etc/fstab")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/fstab: %w", err)
	}
	defer file.Close()

	var lines []string
	found := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
			continue
		}

		// Check if this line matches the target
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 {
			if parts[0] == target || parts[1] == target {
				found = true
				continue // Skip this line (remove it)
			}
		}

		lines = append(lines, line)
	}

	if !found {
		return &ActionResult{Changed: false, Message: "Entry not found in /etc/fstab"}, nil
	}

	// Write back
	if err := os.WriteFile("/etc/fstab", []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to write /etc/fstab: %w", err)
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Removed %s from /etc/fstab", target),
	}, nil
}
