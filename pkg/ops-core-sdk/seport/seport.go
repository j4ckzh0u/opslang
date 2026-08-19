// Package seport manages SELinux port type definitions.
// Requires SELinux to be enabled and semanage to be installed.
package seport

import (
	"fmt"
	"os/exec"
	"strings"
)

// ActionResult represents the result of a port management operation.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message,omitempty"`
}

// PortEntry represents a single SELinux port type definition.
type PortEntry struct {
	SELinuxPortType string `json:"selinux_port_type"`
	Protocol        string `json:"protocol"`
	PortNumber      string `json:"port_number"`
}

// ListResult represents the result of listing port definitions.
type ListResult struct {
	Ports []PortEntry `json:"ports"`
}

// Add adds a port type definition for a specific protocol/port combination.
func Add(sePortType, proto, port string) (*ActionResult, error) {
	if sePortType == "" {
		return nil, fmt.Errorf("seport_type is required")
	}
	if proto == "" {
		return nil, fmt.Errorf("protocol is required (tcp/udp)")
	}
	if port == "" {
		return nil, fmt.Errorf("port is required")
	}

	// Check if already exists with correct type
	entries, err := listEntries()
	if err == nil {
		for _, e := range entries {
			if e.Protocol == proto && e.PortNumber == port {
				if e.SELinuxPortType == sePortType {
					return &ActionResult{Changed: false, Message: "port already defined with correct type"}, nil
				}
				// Need to modify
				cmd := exec.Command("semanage", "port", "-m", "-t", sePortType, "-p", proto, port)
				out, err := cmd.CombinedOutput()
				if err != nil {
					return nil, fmt.Errorf("failed to modify port type: %s: %w", string(out), err)
				}
				return &ActionResult{Changed: true, Message: "port type modified"}, nil
			}
		}
	}

	cmd := exec.Command("semanage", "port", "-a", "-t", sePortType, "-p", proto, port)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to add port type: %s: %w", string(out), err)
	}
	return &ActionResult{Changed: true, Message: "port type added"}, nil
}

// Remove removes a port type definition.
func Remove(proto, port string) (*ActionResult, error) {
	if proto == "" {
		return nil, fmt.Errorf("protocol is required (tcp/udp)")
	}
	if port == "" {
		return nil, fmt.Errorf("port is required")
	}

	cmd := exec.Command("semanage", "port", "-d", "-p", proto, port)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to remove port type: %s: %w", string(out), err)
	}
	return &ActionResult{Changed: true, Message: "port type removed"}, nil
}

// List returns all SELinux port type definitions.
func List() (*ListResult, error) {
	entries, err := listEntries()
	if err != nil {
		return nil, err
	}
	return &ListResult{Ports: entries}, nil
}

// Get returns the port type definition for a specific protocol/port.
func Get(proto, port string) (*PortEntry, error) {
	if proto == "" {
		return nil, fmt.Errorf("protocol is required")
	}
	if port == "" {
		return nil, fmt.Errorf("port is required")
	}

	entries, err := listEntries()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.Protocol == proto && e.PortNumber == port {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("port %s/%s not found", port, proto)
}

func listEntries() ([]PortEntry, error) {
	cmd := exec.Command("semanage", "port", "-l")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list ports: %s: %w", string(out), err)
	}

	entries := make([]PortEntry, 0)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "SELinux") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		entries = append(entries, PortEntry{
			SELinuxPortType: fields[0],
			Protocol:        fields[1],
			PortNumber:      fields[2],
		})
	}
	return entries, nil
}
