// Package systemdanalyze provides systemd boot and service analysis.
package systemdanalyze

import (
	"fmt"
	"os/exec"
	"strings"
)

// TimeResult represents boot time analysis.
type TimeResult struct {
	Kernel    string `json:"kernel,omitempty"`
	Initrd    string `json:"initrd,omitempty"`
	Userspace string `json:"userspace,omitempty"`
	Total     string `json:"total,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ServiceInfo represents a service's boot time.
type ServiceInfo struct {
	Name     string `json:"name"`
	Time     string `json:"time"`
	Activated string `json:"activated,omitempty"`
}

// CriticalChainResult represents the critical boot chain.
type CriticalChainResult struct {
	Chain    string `json:"chain"`
	Services []ServiceInfo `json:"services,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BlameResult represents service blame ordering.
type BlameResult struct {
	Services []ServiceInfo `json:"services"`
	Count    int           `json:"count"`
	Error    string        `json:"error,omitempty"`
}

func systemdAnalyze(args ...string) (string, error) {
	cmd := exec.Command("systemd-analyze", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Time returns boot time summary.
func Time() TimeResult {
	out, err := systemdAnalyze("time")
	if err != nil {
		return TimeResult{Error: fmt.Sprintf("systemd-analyze time failed: %s: %s", err, out)}
	}
	result := TimeResult{}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Startup finished in") || strings.Contains(line, "Graphical target reached") {
			result.Total = strings.TrimSpace(line)
		}
		if strings.Contains(line, "(kernel)") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.Contains(p, "+") || strings.Contains(p, "s") {
					result.Kernel = p
					break
				}
			}
		}
		if strings.Contains(line, "(initrd)") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.Contains(p, "+") || strings.Contains(p, "s") {
					result.Initrd = p
					break
				}
			}
		}
		if strings.Contains(line, "(userspace)") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.Contains(p, "+") || strings.Contains(p, "s") {
					result.Userspace = p
					break
				}
			}
		}
	}
	return result
}

// Blame returns services ordered by boot time.
func Blame() BlameResult {
	out, err := systemdAnalyze("blame")
	if err != nil {
		return BlameResult{Error: fmt.Sprintf("systemd-analyze blame failed: %s: %s", err, out)}
	}
	var services []ServiceInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			services = append(services, ServiceInfo{
				Time: fields[0],
				Name: fields[1],
			})
		}
	}
	return BlameResult{Services: services, Count: len(services)}
}

// CriticalChain returns the critical boot chain.
func CriticalChain() CriticalChainResult {
	out, err := systemdAnalyze("critical-chain")
	if err != nil {
		return CriticalChainResult{Error: fmt.Sprintf("systemd-analyze critical-chain failed: %s: %s", err, out)}
	}
	return CriticalChainResult{Chain: strings.TrimSpace(out)}
}

// Security returns security score for units.
func Security(unit string) (string, error) {
	args := []string{"security"}
	if unit != "" {
		args = append(args, unit)
	}
	out, err := systemdAnalyze(args...)
	if err != nil {
		return "", fmt.Errorf("systemd-analyze security failed: %w: %s", err, out)
	}
	return out, nil
}

// Verify verifies unit files.
func Verify(unit string) (string, error) {
	args := []string{"verify"}
	if unit != "" {
		args = append(args, unit)
	}
	out, err := systemdAnalyze(args...)
	if err != nil {
		return "", fmt.Errorf("systemd-analyze verify failed: %w: %s", err, out)
	}
	return out, nil
}
