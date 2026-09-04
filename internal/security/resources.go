// Package security implements permission checks, audit logging, and resource limits.
package security

import (
	"fmt"
	"strconv"
	"strings"
)

// ResourceLimits defines CPU and memory limits for task execution.
type ResourceLimits struct {
	CPUPercent  int    // CPU usage percentage (0-100 per core)
	MemoryMB    int    // Memory limit in MB
	CPUQuota    string // systemd CPU quota (e.g., "50%")
	MemoryLimit string // systemd memory limit (e.g., "512M")
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

// ApplyResourceLimits applies resource limits to the current process via ulimit.
//
// ponytail: systemd-run --scope is an orchestrator tool — it wraps a process
// from the outside. Using it to re-exec the current binary is a design flaw
// (spawns process trees, breaks in tests, requires root). ulimit via
// setrlimit(2) is the correct in-process mechanism. Add systemd back only
// when an external orchestrator wraps the process before it starts.
func ApplyResourceLimits(limits *ResourceLimits) error {
	if limits == nil {
		return nil
	}
	return applyPlatformResourceLimits(limits)
}

// applyUlimitLimits applies limits using ulimit (Unix only).
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
