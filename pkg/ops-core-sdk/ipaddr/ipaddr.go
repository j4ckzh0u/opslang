// Package ipaddr provides IP address management via ip command.
package ipaddr

import (
	"fmt"
	"os/exec"
	"strings"
)

// AddressInfo represents an IP address entry.
type AddressInfo struct {
	Interface string `json:"interface"`
	Family    string `json:"family"` // inet, inet6
	Address   string `json:"address"`
	PrefixLen string `json:"prefix_len"`
	Scope     string `json:"scope,omitempty"`
	Error     string `json:"error,omitempty"`
}

// LinkInfo represents a network link.
type LinkInfo struct {
	Name    string `json:"name"`
	Index   string `json:"index"`
	State   string `json:"state"`
	MAC     string `json:"mac,omitempty"`
	MTU     string `json:"mtu,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ListResult is returned by address listing.
type ListResult struct {
	Addresses []AddressInfo `json:"addresses"`
	Count     int           `json:"count"`
	Error     string        `json:"error,omitempty"`
}

// LinksResult is returned by link listing.
type LinksResult struct {
	Links []LinkInfo `json:"links"`
	Count int        `json:"count"`
	Error string     `json:"error,omitempty"`
}

// Result is returned by address operations.
type Result struct {
	Address   string `json:"address,omitempty"`
	Interface string `json:"interface,omitempty"`
	Success   bool   `json:"success"`
	Changed   bool   `json:"changed"`
	Error     string `json:"error,omitempty"`
}

func ip(args ...string) (string, error) {
	cmd := exec.Command("ip", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// List lists all IP addresses.
func List() ListResult {
	out, err := ip("-o", "addr", "show")
	if err != nil {
		return ListResult{Error: fmt.Sprintf("ip addr show failed: %s: %s", err, out)}
	}
	var addrs []AddressInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			addrs = append(addrs, AddressInfo{
				Interface: fields[1],
				Family:    fields[2],
				Address:   strings.Split(fields[3], "/")[0],
				PrefixLen: strings.Split(fields[3], "/")[1],
				Scope:     fields[4],
			})
		}
	}
	return ListResult{Addresses: addrs, Count: len(addrs)}
}

// ListInterface lists addresses for a specific interface.
func ListInterface(iface string) ListResult {
	if iface == "" {
		return ListResult{Error: "interface is required"}
	}
	out, err := ip("-o", "addr", "show", "dev", iface)
	if err != nil {
		return ListResult{Error: fmt.Sprintf("ip addr show dev %s failed: %s: %s", iface, err, out)}
	}
	var addrs []AddressInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			addrs = append(addrs, AddressInfo{
				Interface: fields[1],
				Family:    fields[2],
				Address:   strings.Split(fields[3], "/")[0],
				PrefixLen: strings.Split(fields[3], "/")[1],
			})
		}
	}
	return ListResult{Addresses: addrs, Count: len(addrs)}
}

// Add adds an IP address to an interface.
func Add(addr, iface string) Result {
	if addr == "" || iface == "" {
		return Result{Error: "address and interface are required"}
	}
	out, err := ip("addr", "add", addr, "dev", iface)
	if err != nil {
		return Result{Address: addr, Interface: iface, Error: fmt.Sprintf("ip addr add failed: %s: %s", err, out)}
	}
	return Result{Address: addr, Interface: iface, Success: true, Changed: true}
}

// Delete removes an IP address from an interface.
func Delete(addr, iface string) Result {
	if addr == "" || iface == "" {
		return Result{Error: "address and interface are required"}
	}
	out, err := ip("addr", "del", addr, "dev", iface)
	if err != nil {
		return Result{Address: addr, Interface: iface, Error: fmt.Sprintf("ip addr del failed: %s: %s", err, out)}
	}
	return Result{Address: addr, Interface: iface, Success: true, Changed: true}
}

// Flush removes all addresses from an interface.
func Flush(iface string) Result {
	if iface == "" {
		return Result{Error: "interface is required"}
	}
	out, err := ip("addr", "flush", "dev", iface)
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("ip addr flush failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}

// Links lists all network interfaces.
func Links() LinksResult {
	out, err := ip("-o", "link", "show")
	if err != nil {
		return LinksResult{Error: fmt.Sprintf("ip link show failed: %s: %s", err, out)}
	}
	var links []LinkInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 4 {
			link := LinkInfo{
				Index: strings.TrimSuffix(fields[0], ":"),
				Name:  strings.TrimSuffix(fields[1], ":"),
			}
			for i, f := range fields {
				if f == "state" && i+1 < len(fields) {
					link.State = fields[i+1]
				}
				if f == "mtu" && i+1 < len(fields) {
					link.MTU = fields[i+1]
				}
				if f == "link/ether" && i+1 < len(fields) {
					link.MAC = fields[i+1]
				}
			}
			links = append(links, link)
		}
	}
	return LinksResult{Links: links, Count: len(links)}
}

// LinkUp brings up an interface.
func LinkUp(iface string) Result {
	if iface == "" {
		return Result{Error: "interface is required"}
	}
	out, err := ip("link", "set", iface, "up")
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("link up failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}

// LinkDown brings down an interface.
func LinkDown(iface string) Result {
	if iface == "" {
		return Result{Error: "interface is required"}
	}
	out, err := ip("link", "set", iface, "down")
	if err != nil {
		return Result{Interface: iface, Error: fmt.Sprintf("link down failed: %s: %s", err, out)}
	}
	return Result{Interface: iface, Success: true, Changed: true}
}
