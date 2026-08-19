// Package mdadm manages software RAID (mdadm).
// Equivalent to community.general.mdadm module.
package mdadm

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
	Level   string `json:"level,omitempty"`
	Devices string `json:"devices,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DetailResult is returned by Detail.
type DetailResult struct {
	Status  string `json:"status"`
	Device  string `json:"device"`
	State   string `json:"state,omitempty"`
	Level   string `json:"level,omitempty"`
	RaidDisks int  `json:"raid_disks"`
	Info    string `json:"info,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ScanResult is returned by Scan.
type ScanResult struct {
	Status string   `json:"status"`
	Arrays []string `json:"arrays"`
	Error  string   `json:"error,omitempty"`
}

// Create creates a new RAID array.
func Create(device string, level string, devices []string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device is required"}
	}
	if level == "" {
		level = "1"
	}
	if len(devices) == 0 {
		return Result{Status: "failed", Error: "at least one device is required"}
	}

	args := []string{"--create", device, "--level=" + level, "--raid-devices=" + fmt.Sprintf("%d", len(devices))}
	args = append(args, devices...)

	cmd := exec.Command("mdadm", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("mdadm create: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Level: level, Devices: strings.Join(devices, ",")}
}

// Destroy stops and destroys a RAID array.
func Destroy(device string) Result {
	if device == "" {
		return Result{Status: "failed", Error: "device is required"}
	}

	// Stop the array
	cmd := exec.Command("mdadm", "--stop", device)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("mdadm stop: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	// Zero superblock on each device
	cmd = exec.Command("mdadm", "--zero-superblock", device)
	_, _ = cmd.CombinedOutput() // best effort

	return Result{Status: "success", Changed: true, Device: device}
}

// Detail returns information about a RAID array.
func Detail(device string) DetailResult {
	if device == "" {
		return DetailResult{Status: "failed", Error: "device is required"}
	}

	cmd := exec.Command("mdadm", "--detail", "--brief", device)
	out, err := cmd.Output()
	if err != nil {
		return DetailResult{Status: "failed", Device: device, Error: fmt.Sprintf("mdadm detail: %v", err)}
	}

	info := strings.TrimSpace(string(out))
	lines := strings.Split(info, "\n")
	state := ""
	level := ""
	raidDisks := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "metadata=") {
			// ARRAY /dev/md0 metadata=1.2 name=... UUID=...
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "level=") {
					level = strings.TrimPrefix(p, "level=")
				}
				if strings.HasPrefix(p, "num-devices=") {
					n := 0
					fmt.Sscanf(strings.TrimPrefix(p, "num-devices="), "%d", &n)
					raidDisks = n
				}
			}
		}
		if strings.HasPrefix(line, "state :") || strings.Contains(line, "state=") {
			state = strings.TrimSpace(strings.TrimPrefix(line, "state :"))
		}
	}

	return DetailResult{
		Status:    "success",
		Device:    device,
		State:     state,
		Level:     level,
		RaidDisks: raidDisks,
		Info:      info,
	}
}

// Scan scans for RAID arrays.
func Scan() ScanResult {
	cmd := exec.Command("mdadm", "--detail", "--scan")
	out, err := cmd.Output()
	if err != nil {
		return ScanResult{Status: "failed", Error: fmt.Sprintf("mdadm scan: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	arrays := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			arrays = append(arrays, line)
		}
	}
	return ScanResult{Status: "success", Arrays: arrays}
}

// Add adds a device to a RAID array.
func Add(device string, member string) Result {
	if device == "" || member == "" {
		return Result{Status: "failed", Error: "device and member are required"}
	}
	cmd := exec.Command("mdadm", "--add", device, member)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("mdadm add: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Devices: member}
}

// Remove removes a device from a RAID array.
func Remove(device string, member string) Result {
	if device == "" || member == "" {
		return Result{Status: "failed", Error: "device and member are required"}
	}

	// First mark as failed
	cmd := exec.Command("mdadm", "--fail", device, member)
	_, _ = cmd.CombinedOutput() // best effort

	// Then remove
	cmd = exec.Command("mdadm", "--remove", device, member)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("mdadm remove: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Device: device, Devices: member}
}
