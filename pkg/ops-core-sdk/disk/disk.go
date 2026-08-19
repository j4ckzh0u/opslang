// Package disk provides disk/filesystem operations.
// Ansible filesystem + parted equivalents.
package disk

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// FilesystemResult is returned by FilesystemCreate.
type FilesystemResult struct {
	Device   string `json:"device"`
	FSType   string `json:"fs_type"`
	Created  bool   `json:"created"`
}

// FilesystemCreate creates a filesystem on a device.
func FilesystemCreate(device string, fsType string) (FilesystemResult, error) {
	result := FilesystemResult{Device: device, FSType: fsType}

	if fsType == "" {
		fsType = "ext4"
	}

	var cmdName string
	var args []string
	switch fsType {
	case "ext2", "ext3", "ext4":
		cmdName = "mkfs." + fsType
		args = []string{device}
	case "xfs":
		cmdName = "mkfs.xfs"
		args = []string{device}
	case "btrfs":
		cmdName = "mkfs.btrfs"
		args = []string{"-f", device}
	case "vfat", "fat32":
		cmdName = "mkfs.vfat"
		args = []string{device}
	case "swap":
		cmdName = "mkswap"
		args = []string{device}
	default:
		return result, fmt.Errorf("disk.FilesystemCreate: unsupported filesystem type %q", fsType)
	}

	cmd := exec.Command(cmdName, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return result, fmt.Errorf("disk.FilesystemCreate: %s failed: %s: %w", cmdName, string(out), err)
	}
	result.Created = true
	return result, nil
}

// Partition represents a disk partition.
type Partition struct {
	Device   string `json:"device"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Size     string `json:"size"`
	FSType   string `json:"fs_type"`
	Name     string `json:"name"`
}

// PartListResult is returned by PartList.
type PartListResult struct {
	Device     string      `json:"device"`
	Partitions []Partition `json:"partitions"`
	Count      int         `json:"count"`
}

// PartList lists partitions on a device using lsblk.
func PartList(device string) (PartListResult, error) {
	result := PartListResult{Device: device, Partitions: make([]Partition, 0)}

	cmd := exec.Command("lsblk", "-nlpbo", "NAME,START,SIZE,FSTYPE,LABEL", device)
	out, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("disk.PartList: lsblk failed: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	first := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Skip the device itself (first line from lsblk)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if first {
			first = false
			continue
		}

		p := Partition{
			Device: fields[0],
			Start:  fields[1],
			Size:   fields[2],
		}
		if len(fields) > 3 {
			p.FSType = fields[3]
		}
		if len(fields) > 4 {
			p.Name = fields[4]
		}
		result.Partitions = append(result.Partitions, p)
	}
	result.Count = len(result.Partitions)
	return result, nil
}
