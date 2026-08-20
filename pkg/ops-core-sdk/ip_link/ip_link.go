// Package ip_link provides network interface management.
// Supports listing, configuring interfaces, MTU, MAC address on Linux.
package ip_link

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Interface represents a network interface.
type Interface struct {
	Name       string   `json:"name"`
	Index      int      `json:"index"`
	MTU        int      `json:"mtu"`
	State      string   `json:"state"` // UP or DOWN
	MAC        string   `json:"mac,omitempty"`
	Flags      []string `json:"flags,omitempty"`
	LinkType   string   `json:"link_type,omitempty"`
}

// InterfaceResult represents the result of interface operations.
type InterfaceResult struct {
	Success    bool        `json:"success"`
	Interfaces []Interface `json:"interfaces,omitempty"`
	Changed    bool        `json:"changed,omitempty"`
	Error      string      `json:"error,omitempty"`
	Duration   int64       `json:"duration_ms"`
}

// List returns all network interfaces.
func List() InterfaceResult {
	start := time.Now()

	out, err := exec.Command("ip", "link", "show").Output()
	if err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list interfaces: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	ifaces := parseInterfaces(string(out))
	return InterfaceResult{
		Success:    true,
		Interfaces: ifaces,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// Get returns information about a specific interface.
func Get(name string) InterfaceResult {
	start := time.Now()

	out, err := exec.Command("ip", "link", "show", name).Output()
	if err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get interface %s: %v", name, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	ifaces := parseInterfaces(string(out))
	return InterfaceResult{
		Success:    true,
		Interfaces: ifaces,
		Duration:   time.Since(start).Milliseconds(),
	}
}

// SetUp brings an interface up.
func SetUp(name string) InterfaceResult {
	start := time.Now()

	cmd := exec.Command("ip", "link", "set", name, "up")
	if err := cmd.Run(); err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to set %s up: %v", name, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return InterfaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// SetDown brings an interface down.
func SetDown(name string) InterfaceResult {
	start := time.Now()

	cmd := exec.Command("ip", "link", "set", name, "down")
	if err := cmd.Run(); err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to set %s down: %v", name, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return InterfaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// SetMTU sets the MTU of an interface.
func SetMTU(name string, mtu int) InterfaceResult {
	start := time.Now()

	if mtu < 68 || mtu > 65535 {
		return InterfaceResult{
			Success: false,
			Error:   fmt.Sprintf("invalid MTU %d (must be 68-65535)", mtu),
		}
	}

	cmd := exec.Command("ip", "link", "set", name, "mtu", strconv.Itoa(mtu))
	if err := cmd.Run(); err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to set MTU: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return InterfaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// SetMAC sets the MAC address of an interface.
// Interface must be down before changing MAC.
func SetMAC(name, mac string) InterfaceResult {
	start := time.Now()

	// Bring interface down first
	exec.Command("ip", "link", "set", name, "down").Run()

	cmd := exec.Command("ip", "link", "set", name, "address", mac)
	if err := cmd.Run(); err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to set MAC: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return InterfaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// SetName renames an interface.
// Interface must be down before renaming.
func SetName(oldName, newName string) InterfaceResult {
	start := time.Now()

	// Bring interface down first
	exec.Command("ip", "link", "set", oldName, "down").Run()

	cmd := exec.Command("ip", "link", "set", oldName, "name", newName)
	if err := cmd.Run(); err != nil {
		return InterfaceResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to rename: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return InterfaceResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

func parseInterfaces(output string) []Interface {
	var ifaces []Interface
	blocks := strings.Split(output, "\n\n")

	for _, block := range blocks {
		if block == "" {
			continue
		}
		iface := parseInterfaceBlock(block)
		if iface.Name != "" {
			ifaces = append(ifaces, iface)
		}
	}
	return ifaces
}

func parseInterfaceBlock(block string) Interface {
	iface := Interface{}
	lines := strings.Split(block, "\n")
	if len(lines) == 0 {
		return iface
	}

	// Parse first line: "2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 ..."
	firstLine := lines[0]
	parts := strings.Fields(firstLine)
	if len(parts) < 4 {
		return iface
	}

	// Index
	fmt.Sscanf(parts[0], "%d:", &iface.Index)

	// Name (remove trailing colon)
	iface.Name = strings.TrimSuffix(parts[1], ":")

	// Flags (from <...>)
	if strings.Contains(firstLine, "<") {
		flagStart := strings.Index(firstLine, "<")
		flagEnd := strings.Index(firstLine, ">")
		if flagStart >= 0 && flagEnd > flagStart {
			flagStr := firstLine[flagStart+1 : flagEnd]
			iface.Flags = strings.Split(flagStr, ",")
			iface.State = "DOWN"
			for _, f := range iface.Flags {
				if f == "UP" {
					iface.State = "UP"
					break
				}
			}
		}
	}

	// MTU
	for i, p := range parts {
		if p == "mtu" && i+1 < len(parts) {
			fmt.Sscanf(parts[i+1], "%d", &iface.MTU)
			break
		}
	}

	// Parse second line for MAC and link type
	if len(lines) > 1 {
		secondLine := lines[1]
		if strings.Contains(secondLine, "link/ether") {
			fields := strings.Fields(secondLine)
			for i, f := range fields {
				if f == "link/ether" && i+1 < len(fields) {
					iface.MAC = fields[i+1]
					iface.LinkType = "ether"
					break
				}
			}
		} else if strings.Contains(secondLine, "link/loopback") {
			iface.LinkType = "loopback"
		}
	}

	return iface
}
