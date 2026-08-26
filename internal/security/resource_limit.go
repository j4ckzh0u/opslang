package security

import (
	"fmt"
	"strings"
)

// ResourceLimit defines resource constraints for spawned processes.
type ResourceLimit struct {
	CPUPercent int   `json:"cpu_percent"`
	MemoryMB   int64 `json:"memory_mb"`
}

// AppliedLimit represents an applied resource limit with a cleanup function.
type AppliedLimit struct {
	// Cleanup releases the resource limit
	Cleanup func()
	// Method describes how the limit was applied
	Method string
}

// Apply attempts to constrain the CURRENT process.
//
// Honest limitation: a running process cannot move itself into a systemd
// scope after startup, so Apply never actually constrains anything. It
// returns a no-op AppliedLimit plus an explanatory error directing callers
// to enforce limits at spawn time via ResourceLimit.SystemdRunPrefix (the
// remote executor does exactly that for ops-runner).
//
// The signature and "always non-nil AppliedLimit" contract are kept for
// backward compatibility with existing callers and tests.
func Apply(limit ResourceLimit) (*AppliedLimit, error) {
	return &AppliedLimit{
		Cleanup: func() {},
		Method:  "none",
	}, fmt.Errorf("cannot constrain the current process after startup; enforce limits at spawn time with SystemdRunPrefix")
}

// SystemdRunPrefix returns a command prefix that runs the following
// command inside a transient systemd scope enforcing these limits, e.g.
//
//	systemd-run --scope --quiet -p CPUQuota=80% -p MemoryMax=1024M -- <cmd>
//
// Zero-valued fields are omitted. The caller must verify systemd-run
// exists on the target host first (the remote executor probes with
// `command -v systemd-run` and warns when absent).
func (r ResourceLimit) SystemdRunPrefix() string {
	parts := []string{"systemd-run", "--scope", "--quiet"}
	if r.CPUPercent > 0 {
		parts = append(parts, "-p", fmt.Sprintf("CPUQuota=%d%%", r.CPUPercent))
	}
	if r.MemoryMB > 0 {
		parts = append(parts, "-p", fmt.Sprintf("MemoryMax=%dM", r.MemoryMB))
	}
	parts = append(parts, "--")
	return strings.Join(parts, " ") + " "
}

// DefaultResourceLimit returns sensible default resource limits.
func DefaultResourceLimit() ResourceLimit {
	return ResourceLimit{
		CPUPercent: 80,   // 80% CPU
		MemoryMB:   1024, // 1GB memory
	}
}

// String returns a human-readable representation of the resource limit.
func (r ResourceLimit) String() string {
	return fmt.Sprintf("CPU: %d%%, Memory: %dMB", r.CPUPercent, r.MemoryMB)
}
