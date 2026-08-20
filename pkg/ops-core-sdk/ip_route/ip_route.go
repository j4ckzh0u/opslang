// Package ip_route provides IP routing table management.
// Supports listing, adding, deleting routes on Linux via netlink.
package ip_route

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// Route represents a network route entry.
type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway,omitempty"`
	Dev         string `json:"dev,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Table       string `json:"table,omitempty"`
	Metric      int    `json:"metric,omitempty"`
}

// RouteResult represents the result of route operations.
type RouteResult struct {
	Success  bool    `json:"success"`
	Routes   []Route `json:"routes,omitempty"`
	Changed  bool    `json:"changed,omitempty"`
	Error    string  `json:"error,omitempty"`
	Duration int64   `json:"duration_ms"`
}

// List returns all routes in the main routing table.
func List() RouteResult {
	start := time.Now()

	out, err := exec.Command("ip", "route", "show").Output()
	if err != nil {
		return RouteResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list routes: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	routes := parseRoutes(string(out))
	return RouteResult{
		Success:  true,
		Routes:   routes,
		Duration: time.Since(start).Milliseconds(),
	}
}

// ListTable returns routes from a specific routing table.
func ListTable(table string) RouteResult {
	start := time.Now()

	out, err := exec.Command("ip", "route", "show", "table", table).Output()
	if err != nil {
		return RouteResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to list routes in table %s: %v", table, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	routes := parseRoutes(string(out))
	return RouteResult{
		Success:  true,
		Routes:   routes,
		Duration: time.Since(start).Milliseconds(),
	}
}

// AddConfig holds route addition configuration.
type AddConfig struct {
	Destination string
	Gateway     string
	Dev         string
	Metric      int
	Table       string
}

// Add adds a new route to the routing table.
func Add(cfg AddConfig) RouteResult {
	start := time.Now()

	if cfg.Destination == "" {
		return RouteResult{
			Success: false,
			Error:   "destination is required",
		}
	}

	args := []string{"route", "add"}

	// Parse destination (could be "default" or CIDR)
	if cfg.Destination == "default" {
		args = append(args, "default")
	} else {
		// Validate CIDR
		if _, _, err := net.ParseCIDR(cfg.Destination); err != nil {
			// Try as plain IP
			if net.ParseIP(cfg.Destination) == nil {
				return RouteResult{
					Success: false,
					Error:   fmt.Sprintf("invalid destination: %s", cfg.Destination),
				}
			}
			args = append(args, cfg.Destination)
		} else {
			args = append(args, cfg.Destination)
		}
	}

	if cfg.Gateway != "" {
		args = append(args, "via", cfg.Gateway)
	}
	if cfg.Dev != "" {
		args = append(args, "dev", cfg.Dev)
	}
	if cfg.Metric > 0 {
		args = append(args, "metric", fmt.Sprintf("%d", cfg.Metric))
	}
	if cfg.Table != "" {
		args = append(args, "table", cfg.Table)
	}

	cmd := exec.Command("ip", args...)
	if err := cmd.Run(); err != nil {
		return RouteResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to add route: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return RouteResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Delete removes a route from the routing table.
func Delete(destination, table string) RouteResult {
	start := time.Now()

	if destination == "" {
		return RouteResult{
			Success: false,
			Error:   "destination is required",
		}
	}

	args := []string{"route", "del", destination}
	if table != "" {
		args = append(args, "table", table)
	}

	cmd := exec.Command("ip", args...)
	if err := cmd.Run(); err != nil {
		return RouteResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to delete route: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return RouteResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Flush removes all routes from a table or device.
func Flush(dev, table string) RouteResult {
	start := time.Now()

	args := []string{"route", "flush"}
	if dev != "" {
		args = append(args, "dev", dev)
	}
	if table != "" {
		args = append(args, "table", table)
	}

	cmd := exec.Command("ip", args...)
	if err := cmd.Run(); err != nil {
		return RouteResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to flush routes: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	return RouteResult{
		Success:  true,
		Changed:  true,
		Duration: time.Since(start).Milliseconds(),
	}
}

// Get returns the route to a specific destination.
func Get(destination string) RouteResult {
	start := time.Now()

	out, err := exec.Command("ip", "route", "get", destination).Output()
	if err != nil {
		return RouteResult{
			Success:  false,
			Error:    fmt.Sprintf("failed to get route to %s: %v", destination, err),
			Duration: time.Since(start).Milliseconds(),
		}
	}

	routes := parseRoutes(string(out))
	return RouteResult{
		Success:  true,
		Routes:   routes,
		Duration: time.Since(start).Milliseconds(),
	}
}

func parseRoutes(output string) []Route {
	var routes []Route
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		route := parseRouteLine(line)
		if route.Destination != "" {
			routes = append(routes, route)
		}
	}
	return routes
}

func parseRouteLine(line string) Route {
	r := Route{}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return r
	}

	r.Destination = parts[0]

	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "via":
			if i+1 < len(parts) {
				r.Gateway = parts[i+1]
				i++
			}
		case "dev":
			if i+1 < len(parts) {
				r.Dev = parts[i+1]
				i++
			}
		case "scope":
			if i+1 < len(parts) {
				r.Scope = parts[i+1]
				i++
			}
		case "table":
			if i+1 < len(parts) {
				r.Table = parts[i+1]
				i++
			}
		case "metric":
			if i+1 < len(parts) {
				fmt.Sscanf(parts[i+1], "%d", &r.Metric)
				i++
			}
		}
	}
	return r
}
