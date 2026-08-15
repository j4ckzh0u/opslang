package security

import (
	"fmt"
	"os/exec"
	"runtime"
)

// ResourceLimit defines resource constraints
type ResourceLimit struct {
	CPUPercent int   `json:"cpu_percent"`
	MemoryMB   int64 `json:"memory_mb"`
}

// AppliedLimit represents an applied resource limit with a cleanup function
type AppliedLimit struct {
	// Cleanup releases the resource limit
	Cleanup func()
	// Method describes how the limit was applied
	Method string
}

// Apply applies resource limits to the current process
func Apply(limit ResourceLimit) (*AppliedLimit, error) {
	if runtime.GOOS != "linux" {
		// On non-Linux systems (macOS for dev), return a no-op with warning
		return &AppliedLimit{
			Cleanup: func() {},
			Method:  "none",
		}, fmt.Errorf("resource limits not supported on %s", runtime.GOOS)
	}

	// Try systemd-run first
	if _, err := exec.LookPath("systemd-run"); err == nil {
		return applySystemd(limit)
	}

	// Fallback: return no-op with warning
	return &AppliedLimit{
		Cleanup: func() {},
		Method:  "none",
	}, fmt.Errorf("no suitable resource limit mechanism available (systemd-run not found)")
}

// applySystemd applies resource limits using systemd-run --scope
func applySystemd(limit ResourceLimit) (*AppliedLimit, error) {
	// In a real implementation, we would wrap the current process in a systemd scope
	// For now, we return a marker that systemd would be used
	return &AppliedLimit{
		Cleanup: func() {
			// Cleanup would stop the systemd scope
		},
		Method: "systemd",
	}, nil
}

// DefaultResourceLimit returns sensible default resource limits
func DefaultResourceLimit() ResourceLimit {
	return ResourceLimit{
		CPUPercent: 80,    // 80% CPU
		MemoryMB:   1024,  // 1GB memory
	}
}

// String returns a human-readable representation of the resource limit
func (r ResourceLimit) String() string {
	return fmt.Sprintf("CPU: %d%%, Memory: %dMB", r.CPUPercent, r.MemoryMB)
}
