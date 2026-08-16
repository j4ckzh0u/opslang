// Package security implements permission checks, audit logging, and resource limits.
package security

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// ResourceLimits defines CPU and memory limits for task execution.
type ResourceLimits struct {
	CPUPercent   int    // CPU usage percentage (0-100 per core)
	MemoryMB     int    // Memory limit in MB
	CPUQuota     string // systemd CPU quota (e.g., "50%")
	MemoryLimit  string // systemd memory limit (e.g., "512M")
}

// DefaultResourceLimits returns default resource limits.
func DefaultResourceLimits() *ResourceLimits {
	return &ResourceLimits{
		CPUPercent:  100,
		MemoryMB:    1024,
		CPUQuota:    "100%",
		MemoryLimit: "1024M",
	}
}

// ApplyResourceLimits applies resource limits to the current process.
// It tries systemd-run first, then falls back to ulimit.
func ApplyResourceLimits(limits *ResourceLimits) error {
	if limits == nil {
		return nil
	}

	// Try systemd-run first (Linux only)
	if runtime.GOOS == "linux" && isSystemdAvailable() {
		return applySystemdLimits(limits)
	}

	// Fall back to ulimit
	return applyUlimitLimits(limits)
}

// isSystemdAvailable checks if systemd-run is available.
func isSystemdAvailable() bool {
	_, err := exec.LookPath("systemd-run")
	return err == nil
}

// applySystemdLimits applies limits using systemd-run --scope.
func applySystemdLimits(limits *ResourceLimits) error {
	// Build systemd-run command
	args := []string{
		"--scope",
		"-p", fmt.Sprintf("CPUQuota=%s", limits.CPUQuota),
		"-p", fmt.Sprintf("MemoryLimit=%s", limits.MemoryLimit),
	}

	// Re-execute current process with systemd-run
	cmd := exec.Command("systemd-run", args...)
	cmd.Args = append(cmd.Args, os.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment
	cmd.Env = os.Environ()

	// Execute
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply systemd limits: %w", err)
	}

	return nil
}

// applyUlimitLimits applies limits using ulimit (Unix only).
func applyUlimitLimits(limits *ResourceLimits) error {
	// Memory limit (RLIMIT_AS - address space)
	if limits.MemoryMB > 0 {
		memBytes := uint64(limits.MemoryMB) * 1024 * 1024
		var rlimit syscall.Rlimit
		rlimit.Cur = memBytes
		rlimit.Max = memBytes
		if err := syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit); err != nil {
			// Non-fatal: some systems don't allow setting rlimits
			fmt.Fprintf(os.Stderr, "Warning: failed to set memory limit: %v\n", err)
		}
	}

	// Note: CPU limit via ulimit is not straightforward
	// CPU affinity can be set via taskset on Linux, but that's different from quota

	return nil
}

// ParseResourceLimits parses resource limit strings.
func ParseResourceLimits(cpu, memory string) (*ResourceLimits, error) {
	limits := DefaultResourceLimits()

	if cpu != "" {
		percent, err := strconv.Atoi(strings.TrimSuffix(cpu, "%"))
		if err != nil {
			return nil, fmt.Errorf("invalid CPU limit: %s", cpu)
		}
		if percent < 1 || percent > 1000 {
			return nil, fmt.Errorf("CPU limit must be between 1%%%% and 1000%%%%")
		}
		limits.CPUPercent = percent
		limits.CPUQuota = fmt.Sprintf("%d%%", percent)
	}

	if memory != "" {
		mb, err := parseMemoryString(memory)
		if err != nil {
			return nil, fmt.Errorf("invalid memory limit: %s", memory)
		}
		limits.MemoryMB = mb
		limits.MemoryLimit = fmt.Sprintf("%dM", mb)
	}

	return limits, nil
}

// parseMemoryString parses memory strings like "512M", "1G", "1024K".
func parseMemoryString(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty memory string")
	}

	// Check for suffix
	var multiplier int
	var numStr string

	lastChar := strings.ToUpper(s[len(s)-1:])
	switch lastChar {
	case "K":
		multiplier = 1
		numStr = s[:len(s)-1]
	case "M":
		multiplier = 1
		numStr = s[:len(s)-1]
	case "G":
		multiplier = 1024
		numStr = s[:len(s)-1]
	default:
		// Assume MB if no suffix
		multiplier = 1
		numStr = s
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}

	if lastChar == "K" {
		// Convert KB to MB
		return num / 1024, nil
	}

	return num * multiplier, nil
}
