package sys

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// HostnameSetResult is returned by HostnameSet, reporting whether the hostname changed.
type HostnameSetResult struct {
	Changed  bool   `json:"changed"`
	Hostname string `json:"hostname"`
	Error    string `json:"error,omitempty"`
}

// HostnameSet sets the system hostname.
// If the hostname is already set to the requested value, returns Changed: false.
// Uses the hostname command and writes to /etc/hostname on Linux.
func HostnameSet(name string) (HostnameSetResult, error) {
	if name == "" {
		return HostnameSetResult{}, fmt.Errorf("failed to set hostname: name must not be empty")
	}

	current, err := os.Hostname()
	if err != nil {
		return HostnameSetResult{}, fmt.Errorf("failed to get current hostname: %w", err)
	}

	if current == name {
		return HostnameSetResult{Changed: false, Hostname: name}, nil
	}

	// Use hostname command (works on both Linux and macOS)
	cmd := exec.Command("hostname", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return HostnameSetResult{}, fmt.Errorf("failed to set hostname: %s: %w", string(out), err)
	}

	// Write /etc/hostname on Linux
	if runtime.GOOS == "linux" {
		if err := os.WriteFile("/etc/hostname", []byte(name+"\n"), 0644); err != nil {
			return HostnameSetResult{
				Changed:  true,
				Hostname: name,
				Error:    fmt.Sprintf("hostname set in kernel but failed to write /etc/hostname: %v", err),
			}, nil
		}
	}

	return HostnameSetResult{Changed: true, Hostname: name}, nil
}
