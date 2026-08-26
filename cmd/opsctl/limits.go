// limits.go: shared helpers for resource-limit flags and arch-cache
// warnings used by both `opsctl deploy` and `opsctl exec`.
package main

import (
	"fmt"
	"os"

	"github.com/j4ckzh0u/opslang/internal/arch"
	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/security"
)

// resourceLimitFromFlags builds a *security.ResourceLimit from CLI flag
// values. Both unset (<=0) means no limiting: a ResourceLimit with all
// zero fields would wrap the runner in a pointless bare scope.
func resourceLimitFromFlags(cpuPercent int, memMB int64) *security.ResourceLimit {
	if cpuPercent <= 0 && memMB <= 0 {
		return nil
	}
	return &security.ResourceLimit{CPUPercent: cpuPercent, MemoryMB: memMB}
}

// archCacheForRun returns one shared architecture cache for the whole
// command run so multiple task steps reuse detections instead of
// re-reading the cache file per executor. A load failure is printed once
// as a warning; detection still works, it just cannot reuse earlier
// results.
func archCacheForRun() *arch.Cache {
	cache, err := opsexec.NewDefaultArchCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: architecture cache unavailable (%v); hosts will be re-probed this run\n", err)
	}
	return cache
}
