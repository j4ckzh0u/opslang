// Package sys provides structured system information operations for OpsLang.
// All functions use pure Go (CGO_ENABLED=0 compatible) and return strongly-typed
// structs with JSON serialization support. No shell calls are made.
package sys

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

// CPUUsage represents CPU utilization percentages.
// Percent is the overall non-idle usage (0-100).
// User, System, Idle are the per-state breakdowns that sum to ~100.
type CPUUsage struct {
	Percent float64 `json:"percent"`
	User    float64 `json:"user"`
	System  float64 `json:"system"`
	Idle    float64 `json:"idle"`
}

// MemoryInfo represents virtual memory statistics.
type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskUsage represents disk usage statistics for a given path.
type DiskUsage struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// LoadAvg represents system load averages for 1, 5, and 15 minute intervals.
type LoadAvg struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// HostnameInfo represents host identification information.
type HostnameInfo struct {
	Hostname string `json:"hostname"`
	FQDN     string `json:"fqdn"`
}

// UptimeInfo represents system uptime information.
type UptimeInfo struct {
	Uptime   uint64 `json:"uptime"`
	BootTime uint64 `json:"boot_time"`
}

// UserInfo represents a logged-in user session.
type UserInfo struct {
	User      string `json:"user"`
	Terminal  string `json:"terminal"`
	Host      string `json:"host"`
	StartTime uint64 `json:"start_time"`
}

// GetCPUUsage returns overall CPU utilization percentages.
//
// It uses cpu.Times(false) to get aggregate CPU times since boot and computes
// User/System/Idle/Percent breakdowns directly from those cumulative times.
// This ensures User+System+Idle ~ 100 and User+System ~ Percent.
func GetCPUUsage() (CPUUsage, error) {
	// Aggregate CPU times since boot
	times, err := cpu.Times(false)
	if err != nil {
		return CPUUsage{}, fmt.Errorf("failed to get CPU times: %w", err)
	}

	var result CPUUsage

	if len(times) > 0 {
		t := times[0]
		total := t.User + t.System + t.Idle + t.Nice + t.Iowait + t.Irq +
			t.Softirq + t.Steal + t.Guest + t.GuestNice
		if total > 0 {
			idle := t.Idle + t.Iowait
			active := total - idle

			result.Idle = math.Round(idle/total*100*100) / 100
			result.User = math.Round((t.User+t.Guest)/total*100*100) / 100
			result.System = math.Round((t.System+t.Irq+t.Softirq)/total*100*100) / 100
			result.Percent = math.Round(active/total*100*100) / 100
		}
	}

	return result, nil
}

// GetMemoryInfo returns virtual memory statistics.
func GetMemoryInfo() (MemoryInfo, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("failed to get virtual memory info: %w", err)
	}
	return MemoryInfo{
		Total:       v.Total,
		Available:   v.Available,
		Used:        v.Used,
		UsedPercent: math.Round(v.UsedPercent*100) / 100,
	}, nil
}

// GetDiskUsage returns disk usage statistics for the filesystem containing path.
// The path must exist on the local filesystem.
func GetDiskUsage(path string) (DiskUsage, error) {
	if path == "" {
		return DiskUsage{}, errors.New("path must not be empty")
	}
	u, err := disk.Usage(path)
	if err != nil {
		return DiskUsage{}, fmt.Errorf("failed to get disk usage for %q: %w", path, err)
	}
	return DiskUsage{
		Path:        u.Path,
		Total:       u.Total,
		Used:        u.Used,
		Free:        u.Free,
		UsedPercent: math.Round(u.UsedPercent*100) / 100,
	}, nil
}

// GetLoadAvg returns system load averages for 1, 5, and 15 minute intervals.
func GetLoadAvg() (LoadAvg, error) {
	avg, err := load.Avg()
	if err != nil {
		return LoadAvg{}, fmt.Errorf("failed to get load average: %w", err)
	}
	if avg == nil {
		return LoadAvg{}, errors.New("load average not available on this platform")
	}
	return LoadAvg{
		Load1:  avg.Load1,
		Load5:  avg.Load5,
		Load15: avg.Load15,
	}, nil
}

// Hostname returns the host name and its fully qualified domain name.
// If FQDN lookup fails, FQDN falls back to the short hostname.
func Hostname() (HostnameInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return HostnameInfo{}, fmt.Errorf("failed to get hostname: %w", err)
	}

	fqdn := hostname
	addrs, err := net.LookupAddr(hostname)
	if err == nil && len(addrs) > 0 {
		name := strings.TrimSuffix(addrs[0], ".")
		if name != "" {
			fqdn = name
		}
	}

	return HostnameInfo{
		Hostname: hostname,
		FQDN:     fqdn,
	}, nil
}

// Uptime returns system uptime in seconds and the boot time as a UNIX timestamp.
func Uptime() (UptimeInfo, error) {
	up, err := host.Uptime()
	if err != nil {
		return UptimeInfo{}, fmt.Errorf("failed to get uptime: %w", err)
	}
	boot, err := host.BootTime()
	if err != nil {
		return UptimeInfo{}, fmt.Errorf("failed to get boot time: %w", err)
	}
	return UptimeInfo{
		Uptime:   up,
		BootTime: boot,
	}, nil
}

// Users returns information about currently logged-in users.
// Returns an empty slice (not nil) when no users are logged in,
// ensuring consistent JSON serialization as [] rather than null.
func Users() ([]UserInfo, error) {
	users, err := host.Users()
	if err != nil {
		return nil, fmt.Errorf("failed to get user list: %w", err)
	}
	result := make([]UserInfo, 0, len(users))
	for _, u := range users {
		result = append(result, UserInfo{
			User:      u.User,
			Terminal:  u.Terminal,
			Host:      u.Host,
			StartTime: uint64(u.Started),
		})
	}
	return result, nil
}
