// Package swap manages swap space on Linux systems.
// Equivalent to ansible.posix.mount (swap) and system swap management.
package swap

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// InfoResult is returned by Info.
type InfoResult struct {
	Status    string      `json:"status"`
	Total     uint64      `json:"total"`     // bytes
	Used      uint64      `json:"used"`      // bytes
	Free      uint64      `json:"free"`      // bytes
	Entries   []SwapEntry `json:"entries"`
	Error     string      `json:"error,omitempty"`
}

// SwapEntry represents one swap area.
type SwapEntry struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Size     uint64 `json:"size"`     // bytes
	Used     uint64 `json:"used"`     // bytes
	Priority int    `json:"priority"`
}

// CreateResult is returned by Create.
type CreateResult struct {
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	Device   string `json:"device"`
	Size     string `json:"size"`
	Error    string `json:"error,omitempty"`
}

// EnableDisableResult is returned by Enable/Disable.
type EnableDisableResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Device  string `json:"device"`
	Error   string `json:"error,omitempty"`
}

// Info returns swap space information.
func Info() InfoResult {
	if runtime.GOOS != "linux" {
		return InfoResult{Status: "failed", Error: fmt.Sprintf("not supported on %s", runtime.GOOS)}
	}

	f, err := os.Open("/proc/swaps")
	if err != nil {
		return InfoResult{Status: "failed", Error: fmt.Sprintf("read /proc/swaps: %v", err)}
	}
	defer f.Close()

	var entries []SwapEntry
	var total, used uint64

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum == 1 {
			continue // skip header
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}
		size, _ := strconv.ParseUint(fields[2], 10, 64)
		usedKb, _ := strconv.ParseUint(fields[3], 10, 64)
		priority, _ := strconv.Atoi(fields[4])

		entry := SwapEntry{
			Filename: fields[0],
			Type:     fields[1],
			Size:     size * 1024, // KB to bytes
			Used:     usedKb * 1024,
			Priority: priority,
		}
		entries = append(entries, entry)
		total += entry.Size
		used += entry.Used
	}

	return InfoResult{
		Status:  "success",
		Total:   total,
		Used:    used,
		Free:    total - used,
		Entries: entries,
	}
}

// Create creates a swap file of the given size (in MB) and enables it.
func Create(path string, sizeMB int) CreateResult {
	if runtime.GOOS != "linux" {
		return CreateResult{Status: "failed", Error: fmt.Sprintf("not supported on %s", runtime.GOOS)}
	}
	if path == "" || sizeMB <= 0 {
		return CreateResult{Status: "failed", Error: "path and positive size are required"}
	}

	// Create swap file with dd
	ddCmd := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", path),
		"bs=1M", fmt.Sprintf("count=%d", sizeMB))
	if out, err := ddCmd.CombinedOutput(); err != nil {
		return CreateResult{Status: "failed", Error: fmt.Sprintf("dd: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	// Set permissions
	if err := os.Chmod(path, 0600); err != nil {
		return CreateResult{Status: "failed", Error: fmt.Sprintf("chmod: %v", err)}
	}

	// mkswap
	mkCmd := exec.Command("mkswap", path)
	if out, err := mkCmd.CombinedOutput(); err != nil {
		return CreateResult{Status: "failed", Error: fmt.Sprintf("mkswap: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	// swapon
	onCmd := exec.Command("swapon", path)
	if out, err := onCmd.CombinedOutput(); err != nil {
		return CreateResult{Status: "failed", Error: fmt.Sprintf("swapon: %v: %s", err, strings.TrimSpace(string(out)))}
	}

	return CreateResult{
		Status:  "success",
		Changed: true,
		Device:  path,
		Size:    fmt.Sprintf("%dM", sizeMB),
	}
}

// Enable enables a swap device.
func Enable(device string) EnableDisableResult {
	if runtime.GOOS != "linux" {
		return EnableDisableResult{Status: "failed", Error: fmt.Sprintf("not supported on %s", runtime.GOOS)}
	}
	if device == "" {
		return EnableDisableResult{Status: "failed", Error: "device is required"}
	}

	cmd := exec.Command("swapon", device)
	if out, err := cmd.CombinedOutput(); err != nil {
		return EnableDisableResult{Status: "failed", Device: device,
			Error: fmt.Sprintf("swapon: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return EnableDisableResult{Status: "success", Changed: true, Device: device}
}

// Disable disables a swap device.
func Disable(device string) EnableDisableResult {
	if runtime.GOOS != "linux" {
		return EnableDisableResult{Status: "failed", Error: fmt.Sprintf("not supported on %s", runtime.GOOS)}
	}
	if device == "" {
		return EnableDisableResult{Status: "failed", Error: "device is required"}
	}

	cmd := exec.Command("swapoff", device)
	if out, err := cmd.CombinedOutput(); err != nil {
		return EnableDisableResult{Status: "failed", Device: device,
			Error: fmt.Sprintf("swapoff: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return EnableDisableResult{Status: "success", Changed: true, Device: device}
}
