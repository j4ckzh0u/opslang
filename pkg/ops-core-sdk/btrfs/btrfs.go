// Package btrfs provides BTRFS filesystem management operations.
package btrfs

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// FilesystemInfo represents BTRFS filesystem information.
type FilesystemInfo struct {
	MountPoint string `json:"mount_point"`
	Device     string `json:"device"`
	Size       string `json:"size"`
	Used       string `json:"used"`
	Free       string `json:"free"`
	Profile    string `json:"profile,omitempty"`
	UUID       string `json:"uuid,omitempty"`
}

// SubvolumeInfo represents a BTRFS subvolume.
type SubvolumeInfo struct {
	ID       uint64 `json:"id"`
	Gen      uint64 `json:"generation"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"read_only"`
}

// ActionResult represents the result of a BTRFS action.
type ActionResult struct {
	Changed    bool   `json:"changed"`
	Action     string `json:"action"`
	DurationMs int64  `json:"duration_ms"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
}

// DeviceInfo represents a BTRFS device.
type DeviceInfo struct {
	Path  string `json:"path"`
	Size  string `json:"size"`
	Used  string `json:"used"`
	DeVID string `json:"devid"`
	UUID  string `json:"uuid"`
}

// ScrubStatusResult represents the result of a BTRFS scrub.
type ScrubStatusResult struct {
	Running    bool   `json:"running"`
	Status     string `json:"status"`
	Duration   string `json:"duration,omitempty"`
	DataScrubbed string `json:"data_scrubbed,omitempty"`
	ErrorCount int    `json:"error_count"`
}

// FilesystemList lists BTRFS filesystems.
func FilesystemList() ([]FilesystemInfo, error) {
	cmd := exec.Command("btrfs", "filesystem", "show")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("btrfs filesystem show: %s", string(out))
	}

	var results []FilesystemInfo
	current := FilesystemInfo{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Label:") || strings.HasPrefix(line, "no label") {
			if current.MountPoint != "" {
				results = append(results, current)
			}
			current = FilesystemInfo{}
			// Extract UUID
			if idx := strings.Index(line, "uuid:"); idx >= 0 {
				current.UUID = strings.TrimSpace(line[idx+5:])
			}
		} else if strings.HasPrefix(line, "Total devices") || strings.HasPrefix(line, "devid") {
			// Skip header lines
		} else if strings.Contains(line, "/dev/") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				current.Device = fields[len(fields)-1]
			}
		} else if strings.HasPrefix(line, "/") {
			current.MountPoint = line
		}
	}
	if current.MountPoint != "" {
		results = append(results, current)
	}

	return results, nil
}

// SubvolumeList lists subvolumes in a BTRFS filesystem.
func SubvolumeList(mountPoint string) ([]SubvolumeInfo, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	cmd := exec.Command("btrfs", "subvolume", "list", mountPoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("btrfs subvolume list: %s", string(out))
	}

	var results []SubvolumeInfo
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: ID <id> gen <gen> top level <top> path <path>
		var info SubvolumeInfo
		fields := strings.Fields(line)
		for i, f := range fields {
			switch f {
			case "ID":
				if i+1 < len(fields) {
					fmt.Sscanf(fields[i+1], "%d", &info.ID)
				}
			case "gen":
				if i+1 < len(fields) {
					fmt.Sscanf(fields[i+1], "%d", &info.Gen)
				}
			case "path":
				if i+1 < len(fields) {
					info.Path = strings.Join(fields[i+1:], " ")
				}
			}
		}
		if info.Path != "" {
			results = append(results, info)
		}
	}

	return results, nil
}

// SubvolumeCreate creates a BTRFS subvolume.
func SubvolumeCreate(path string) (*ActionResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "subvolume", "create", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "subvolume_create",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("create subvolume: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "subvolume_create",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     path,
	}, nil
}

// SubvolumeDelete deletes a BTRFS subvolume.
func SubvolumeDelete(path string) (*ActionResult, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "subvolume", "delete", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "subvolume_delete",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("delete subvolume: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "subvolume_delete",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     path,
	}, nil
}

// SnapshotCreate creates a snapshot of a BTRFS subvolume.
func SnapshotCreate(source, dest string, readOnly bool) (*ActionResult, error) {
	if source == "" || dest == "" {
		return nil, fmt.Errorf("source and dest are required")
	}

	start := time.Now()
	args := []string{"subvolume", "snapshot"}
	if readOnly {
		args = append(args, "-r")
	}
	args = append(args, source, dest)

	cmd := exec.Command("btrfs", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "snapshot_create",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("create snapshot: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "snapshot_create",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     fmt.Sprintf("%s -> %s", source, dest),
	}, nil
}

// ScrubStart starts a scrub on a BTRFS filesystem.
func ScrubStart(mountPoint string) (*ActionResult, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "scrub", "start", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "scrub_start",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("start scrub: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "scrub_start",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     string(output),
	}, nil
}

// ScrubStatus checks the status of a scrub.
func ScrubStatus(mountPoint string) (*ScrubStatusResult, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	cmd := exec.Command("btrfs", "scrub", "status", mountPoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("scrub status: %s", string(out))
	}

	result := &ScrubStatusResult{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Status:") {
			result.Status = strings.TrimSpace(strings.TrimPrefix(line, "Status:"))
			result.Running = strings.Contains(result.Status, "running")
		} else if strings.HasPrefix(line, "Duration:") {
			result.Duration = strings.TrimSpace(strings.TrimPrefix(line, "Duration:"))
		} else if strings.HasPrefix(line, "Data scrubbed:") {
			result.DataScrubbed = strings.TrimSpace(strings.TrimPrefix(line, "Data scrubbed:"))
		} else if strings.Contains(line, "error") || strings.Contains(line, "Error") {
			if idx := strings.Index(line, ":"); idx >= 0 {
				fmt.Sscanf(strings.TrimSpace(line[idx+1:]), "%d", &result.ErrorCount)
			}
		}
	}

	return result, nil
}

// DeviceAdd adds a device to a BTRFS filesystem.
func DeviceAdd(devicePath, mountPoint string) (*ActionResult, error) {
	if devicePath == "" || mountPoint == "" {
		return nil, fmt.Errorf("device path and mount point are required")
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "device", "add", devicePath, mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "device_add",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("add device: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "device_add",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     fmt.Sprintf("%s -> %s", devicePath, mountPoint),
	}, nil
}

// DeviceRemove removes a device from a BTRFS filesystem.
func DeviceRemove(devicePath, mountPoint string) (*ActionResult, error) {
	if devicePath == "" || mountPoint == "" {
		return nil, fmt.Errorf("device path and mount point are required")
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "device", "remove", devicePath, mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "device_remove",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("remove device: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "device_remove",
		DurationMs: time.Since(start).Milliseconds(),
		Output:     devicePath,
	}, nil
}

// BalanceStart starts a balance operation on a BTRFS filesystem.
func BalanceStart(mountPoint string) (*ActionResult, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "balance", "start", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "balance_start",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("start balance: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "balance_start",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// BalanceStatus checks the status of a balance operation.
func BalanceStatus(mountPoint string) (*ActionResult, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	cmd := exec.Command("btrfs", "balance", "status", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// "No balance found" returns error
		return &ActionResult{
			Changed:    false,
			Action:     "balance_status",
			Output:     string(output),
		}, nil
	}
	return &ActionResult{
		Changed:    false,
		Action:     "balance_status",
		Output:     string(output),
	}, nil
}

// QuotaEnable enables quota on a BTRFS filesystem.
func QuotaEnable(mountPoint string) (*ActionResult, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "quota", "enable", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "quota_enable",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("enable quota: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "quota_enable",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// QuotaDisable disables quota on a BTRFS filesystem.
func QuotaDisable(mountPoint string) (*ActionResult, error) {
	if mountPoint == "" {
		mountPoint = "/"
	}

	start := time.Now()
	cmd := exec.Command("btrfs", "quota", "disable", mountPoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ActionResult{
			Changed:    false,
			Action:     "quota_disable",
			DurationMs: time.Since(start).Milliseconds(),
			Output:     string(output),
			Error:      err.Error(),
		}, fmt.Errorf("disable quota: %s", string(output))
	}
	return &ActionResult{
		Changed:    true,
		Action:     "quota_disable",
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}
