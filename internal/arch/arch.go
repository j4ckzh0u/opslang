// Package arch provides remote architecture detection by executing
// `uname -m` on the target host and mapping it to GOARCH values.
package arch

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// SSHExecutor is the interface for executing commands on a remote host.
// This allows for easy testing with mocks.
type SSHExecutor interface {
	Exec(ctx context.Context, cmd string) (*ExecResult, error)
}

// ExecResult holds the result of a remote command execution.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// Mapping from `uname -m` output to GOARCH values.
var archMap = map[string]string{
	"x86_64":  "amd64",
	"amd64":   "amd64",
	"aarch64": "arm64",
	"arm64":   "arm64",
	"armv7l":  "arm",
	"armv6l":  "arm",
	"armv5l":  "arm",
	"arm":     "arm",
	"i386":    "386",
	"i486":    "386",
	"i586":    "386",
	"i686":    "386",
	"mips":    "mips",
	"mips64":  "mips64",
	"ppc64le": "ppc64le",
	"ppc64":   "ppc64",
	"s390x":   "s390x",
	"riscv64": "riscv64",
}

// Detect runs `uname -m` on the remote host and returns the GOARCH value.
func Detect(ctx context.Context, executor SSHExecutor) (string, error) {
	result, err := executor.Exec(ctx, "uname -m")
	if err != nil {
		return "", fmt.Errorf("failed to execute uname: %w", err)
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("uname exited with code %d: %s", result.ExitCode, result.Stderr)
	}

	raw := strings.TrimSpace(result.Stdout)
	goarch, ok := archMap[raw]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %q", raw)
	}

	return goarch, nil
}

// MapArch converts a raw `uname -m` output to GOARCH.
func MapArch(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	goarch, ok := archMap[raw]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %q", raw)
	}
	return goarch, nil
}

// SupportedArchitectures returns all GOARCH values that can be detected, sorted.
func SupportedArchitectures() []string {
	seen := make(map[string]bool)
	var result []string
	for _, goarch := range archMap {
		if !seen[goarch] {
			seen[goarch] = true
			result = append(result, goarch)
		}
	}
	sort.Strings(result)
	return result
}

// RunnerBinaryName returns the expected runner binary name for a given GOARCH.
func RunnerBinaryName(goarch string) string {
	return fmt.Sprintf("ops-runner-linux-%s", goarch)
}
