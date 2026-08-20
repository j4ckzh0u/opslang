// Package ip_neighbor provides ARP/neighbor table management.
// Supports listing, adding, deleting, and flushing neighbor entries on Linux.
package ip_neighbor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Neighbor represents an ARP/neighbor entry.
type Neighbor struct {
	IP     string `json:"ip"`
	Dev    string `json:"dev"`
	MAC    string `json:"mac,omitempty"`
	State  string `json:"state"`
	Router bool   `json:"router,omitempty"`
}

// NeighborResult represents the result of neighbor operations.
type NeighborResult struct {
	Success   bool        `json:"success"`
	Neighbors []Neighbor  `json:"neighbors,omitempty"`
	Changed   bool        `json:"changed,omitempty"`
	Error     string      `json:"error,omitempty"`
	Duration  int64       `json:"duration_ms"`
}

// List returns all neighbor entries.
func List() NeighborResult {
	start := time.Now()

	out, err := exec.Command("ip", "neigh", "show").Output()
	if err != nil {
		return NeighborResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list neighbors: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	neighbors := parseNeighbors(string(out))
	return NeighborResult{
		Success:   true,
		Neighbors: neighbors,
		Duration:  time.Since(start).Milliseconds(),
	}
}

// ListDev returns neighbor entries for a specific device.
func ListDev(dev string) NeighborResult {
	start := time.Now()

	out, err := exec.Command("ip", "neigh", "show", "dev", dev).Output()
	if err != nil {
		return NeighborResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list neighbors for %s: %v", dev, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	neighbors := parseNeighbors(string(out))
	return NeighborResult{
		Success:   true,
		Neighbors: neighbors,
		Duration:  time.Since(start).Milliseconds(),
	}
}

// Add adds a neighbor entry.
func Add(ip, dev, mac string) NeighborResult {
	start := time.Now()

	args := []string{"neigh", "add", ip, "dev", dev, "lladdr", mac}
	cmd := exec.Command("ip", args...)
	if err := cmd.Run(); err != nil {
		return NeighborResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to add neighbor: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NeighborResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete removes a neighbor entry.
func Delete(ip, dev string) NeighborResult {
	start := time.Now()

	args := []string{"neigh", "del", ip}
	if dev != "" {
		args = append(args, "dev", dev)
	}
	cmd := exec.Command("ip", args...)
	if err := cmd.Run(); err != nil {
		return NeighborResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete neighbor: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NeighborResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Flush removes all neighbor entries for a device.
func Flush(dev string) NeighborResult {
	start := time.Now()

	args := []string{"neigh", "flush", "dev", dev}
	cmd := exec.Command("ip", args...)
	if err := cmd.Run(); err != nil {
		return NeighborResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to flush neighbors: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return NeighborResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

func parseNeighbors(output string) []Neighbor {
	var neighbors []Neighbor
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		neigh := parseNeighborLine(line)
		if neigh.IP != "" {
			neighbors = append(neighbors, neigh)
		}
	}
	return neighbors
}

func parseNeighborLine(line string) Neighbor {
	neigh := Neighbor{}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return neigh
	}

	// First field is IP
	neigh.IP = fields[0]

	// Parse remaining fields
	for i := 1; i < len(fields); i++ {
		switch fields[i] {
		case "dev":
			if i+1 < len(fields) {
				neigh.Dev = fields[i+1]
				i++
			}
		case "lladdr":
			if i+1 < len(fields) {
				neigh.MAC = fields[i+1]
				i++
			}
		case "FAILED", "INCOMPLETE", "STALE", "DELAY", "PROBE", "REACHABLE", "PERMANENT", "NOARP":
			neigh.State = fields[i]
		case "router":
			neigh.Router = true
		}
	}

	return neigh
}
