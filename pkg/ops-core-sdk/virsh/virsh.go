// Package virsh provides libvirt/KVM virtualization management.
package virsh

import (
	"fmt"
	"os/exec"
	"strings"
)

// DomainInfo represents a VM domain.
type DomainInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
	State  string `json:"state"`
	Error  string `json:"error,omitempty"`
}

// Result is returned by domain operations.
type Result struct {
	Domain  string `json:"domain,omitempty"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by domain listing.
type ListResult struct {
	Domains []DomainInfo `json:"domains"`
	Count   int          `json:"count"`
	Error   string       `json:"error,omitempty"`
}

func virsh(args ...string) (string, error) {
	cmd := exec.Command("virsh", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Start starts a domain.
func Start(domain string) Result {
	if domain == "" {
		return Result{Error: "domain is required"}
	}
	out, err := virsh("start", domain)
	if err != nil {
		return Result{Domain: domain, Error: fmt.Sprintf("start failed: %s: %s", err, out)}
	}
	return Result{Domain: domain, Success: true, Changed: true}
}

// Stop stops a domain.
func Stop(domain string) Result {
	if domain == "" {
		return Result{Error: "domain is required"}
	}
	out, err := virsh("destroy", domain)
	if err != nil {
		return Result{Domain: domain, Error: fmt.Sprintf("destroy failed: %s: %s", err, out)}
	}
	return Result{Domain: domain, Success: true, Changed: true}
}

// Reboot reboots a domain.
func Reboot(domain string) Result {
	if domain == "" {
		return Result{Error: "domain is required"}
	}
	out, err := virsh("reboot", domain)
	if err != nil {
		return Result{Domain: domain, Error: fmt.Sprintf("reboot failed: %s: %s", err, out)}
	}
	return Result{Domain: domain, Success: true, Changed: true}
}

// Shutdown gracefully shuts down a domain.
func Shutdown(domain string) Result {
	if domain == "" {
		return Result{Error: "domain is required"}
	}
	out, err := virsh("shutdown", domain)
	if err != nil {
		return Result{Domain: domain, Error: fmt.Sprintf("shutdown failed: %s: %s", err, out)}
	}
	return Result{Domain: domain, Success: true, Changed: true}
}

// Suspend suspends a domain.
func Suspend(domain string) Result {
	if domain == "" {
		return Result{Error: "domain is required"}
	}
	out, err := virsh("suspend", domain)
	if err != nil {
		return Result{Domain: domain, Error: fmt.Sprintf("suspend failed: %s: %s", err, out)}
	}
	return Result{Domain: domain, Success: true, Changed: true}
}

// Resume resumes a suspended domain.
func Resume(domain string) Result {
	if domain == "" {
		return Result{Error: "domain is required"}
	}
	out, err := virsh("resume", domain)
	if err != nil {
		return Result{Domain: domain, Error: fmt.Sprintf("resume failed: %s: %s", err, out)}
	}
	return Result{Domain: domain, Success: true, Changed: true}
}

// List lists all domains.
func List() ListResult {
	out, err := virsh("list", "--all")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("list failed: %s: %s", err, out)}
	}
	var domains []DomainInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Id") || strings.HasPrefix(line, "---") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			domains = append(domains, DomainInfo{
				ID:    fields[0],
				Name:  fields[1],
				State: fields[2],
			})
		}
	}
	return ListResult{Domains: domains, Count: len(domains)}
}

// Info returns domain info.
func Info(domain string) (map[string]string, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	out, err := virsh("dominfo", domain)
	if err != nil {
		return nil, fmt.Errorf("dominfo failed: %w: %s", err, out)
	}
	info := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}
	return info, nil
}

// Version returns libvirt version.
func Version() (string, error) {
	out, err := virsh("version")
	if err != nil {
		return "", fmt.Errorf("virsh version failed: %w: %s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", nil
}
