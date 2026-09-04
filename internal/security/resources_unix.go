//go:build !windows

package security

import (
	"fmt"
	"os"
	"syscall"
)

func applyPlatformResourceLimits(limits *ResourceLimits) error {
	if limits.MemoryMB <= 0 {
		return nil
	}
	memBytes := uint64(limits.MemoryMB) * 1024 * 1024
	var rlimit syscall.Rlimit
	rlimit.Cur = memBytes
	rlimit.Max = memBytes
	if err := syscall.Setrlimit(syscall.RLIMIT_AS, &rlimit); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set memory limit: %v\n", err)
	}
	return nil
}
