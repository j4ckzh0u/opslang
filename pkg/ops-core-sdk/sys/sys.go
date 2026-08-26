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
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
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

// CPUInfo represents CPU hardware information.
type CPUInfo struct {
	VendorID  string  `json:"vendor_id"`
	ModelName string  `json:"model_name"`
	Cores     int32   `json:"cores"`
	Mhz       float64 `json:"mhz"`
	CacheSize int32   `json:"cache_size"`
}

// CPUCount represents logical and physical CPU counts.
type CPUCount struct {
	Logical  int `json:"logical"`
	Physical int `json:"physical"`
}

// DiskPartition represents a real data-bearing mount with capacity info.
// Pseudo filesystems (tmpfs, overlay, squashfs, proc, ...) are filtered
// out by GetDiskPartitions; size fields stay zero when the mount could
// not be stat'ed (e.g. a stale NFS export).
type DiskPartition struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Opts        string  `json:"opts"`
	TotalBytes  uint64  `json:"total_bytes,omitempty"`
	UsedBytes   uint64  `json:"used_bytes,omitempty"`
	FreeBytes   uint64  `json:"free_bytes,omitempty"`
	UsedPercent float64 `json:"used_percent,omitempty"`
}

// HostInfoResult represents detailed host/OS information.
type HostInfoResult struct {
	Hostname        string `json:"hostname"`
	Uptime          uint64 `json:"uptime"`
	BootTime        uint64 `json:"boot_time"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
}

// NetInterface represents a network interface (defined in sys to avoid circular dependency).
type NetInterface struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardware_addr"`
	MTU          int      `json:"mtu"`
	Up           bool     `json:"up"`
	Addresses    []string `json:"addresses"`
}

// sampleInterval is the delay between the two CPU samples taken by
// GetCPUUsage. 500ms gives a meaningful "current utilization" window
// (about 50 scheduler ticks on a typical 100Hz kernel) without making
// scripts noticeably slow.
const sampleInterval = 500 * time.Millisecond

// GetCPUUsage returns current CPU utilization percentages measured over a
// short sampling window (two cpu.Times snapshots delta). Computing from
// cumulative since-boot counters would report a lifetime average that
// barely moves on long-running servers and never triggers alerts.
func GetCPUUsage() (CPUUsage, error) {
	return GetCPUUsageInterval(sampleInterval)
}

// GetCPUUsageInterval is GetCPUUsage with an explicit sampling window.
// A zero or negative interval yields the since-boot average (single sample).
func GetCPUUsageInterval(interval time.Duration) (CPUUsage, error) {
	first, err := cpu.Times(false)
	if err != nil || len(first) == 0 {
		return CPUUsage{}, fmt.Errorf("failed to get CPU times: %w", err)
	}

	if interval <= 0 {
		return computeUsageSinceBoot(first[0]), nil
	}

	time.Sleep(interval)

	second, err := cpu.Times(false)
	if err != nil || len(second) == 0 {
		return CPUUsage{}, fmt.Errorf("failed to get CPU times: %w", err)
	}

	usage, ok := computeUsageDelta(first[0], second[0])
	if !ok {
		// No measurable delta (clock resolution smaller than the window);
		// fall back to the cumulative average rather than reporting 0.
		return computeUsageSinceBoot(second[0]), nil
	}
	return usage, nil
}

// totalTimes sums all CPU time fields.
func totalTimes(t cpu.TimesStat) float64 {
	return t.User + t.System + t.Idle + t.Nice + t.Iowait + t.Irq +
		t.Softirq + t.Steal + t.Guest + t.GuestNice
}

// computeUsageDelta derives utilization from two snapshots.
// Returns ok=false when the elapsed total is not measurably positive.
func computeUsageDelta(t1, t2 cpu.TimesStat) (CPUUsage, bool) {
	dTotal := totalTimes(t2) - totalTimes(t1)
	if dTotal <= 0 {
		return CPUUsage{}, false
	}

	pct := func(v float64) float64 {
		return math.Round(v/dTotal*100*100) / 100
	}

	dIdle := (t2.Idle - t1.Idle) + (t2.Iowait - t1.Iowait)

	return CPUUsage{
		Idle:    pct(dIdle),
		User:    pct((t2.User - t1.User) + (t2.Guest - t1.Guest)),
		System:  pct((t2.System - t1.System) + (t2.Irq - t1.Irq) + (t2.Softirq - t1.Softirq)),
		Percent: math.Round((dTotal-dIdle)/dTotal*100*100) / 100,
	}, true
}

// computeUsageSinceBoot derives utilization from cumulative counters.
func computeUsageSinceBoot(t cpu.TimesStat) CPUUsage {
	total := totalTimes(t)
	if total <= 0 {
		return CPUUsage{}
	}
	idle := t.Idle + t.Iowait

	pct := func(v float64) float64 {
		return math.Round(v/total*100*100) / 100
	}

	return CPUUsage{
		Idle:    pct(idle),
		User:    pct(t.User + t.Guest),
		System:  pct(t.System + t.Irq + t.Softirq),
		Percent: pct(total - idle),
	}
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

// GetNetInterfaces returns information about network interfaces.
// Uses standard library net.Interfaces() directly to avoid circular dependency.
func GetNetInterfaces() ([]NetInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list net interfaces: %w", err)
	}
	result := make([]NetInterface, 0, len(ifaces))
	for _, iface := range ifaces {
		info := NetInterface{
			Name: iface.Name,
			MTU:  iface.MTU,
			Up:   iface.Flags&net.FlagUp != 0,
		}
		if iface.HardwareAddr != nil {
			info.HardwareAddr = iface.HardwareAddr.String()
		}
		addrs, err := iface.Addrs()
		if err == nil {
			info.Addresses = make([]string, 0, len(addrs))
			for _, addr := range addrs {
				info.Addresses = append(info.Addresses, addr.String())
			}
		} else {
			info.Addresses = []string{}
		}
		result = append(result, info)
	}
	return result, nil
}

// GetCPUCount returns logical and physical CPU counts.
func GetCPUCount() (CPUCount, error) {
	logical, err := cpu.Counts(true)
	if err != nil {
		return CPUCount{}, fmt.Errorf("failed to get logical CPU count: %w", err)
	}
	physical, err := cpu.Counts(false)
	if err != nil {
		return CPUCount{}, fmt.Errorf("failed to get physical CPU count: %w", err)
	}
	return CPUCount{Logical: logical, Physical: physical}, nil
}

// GetCPUInfo returns CPU hardware information.
func GetCPUInfo() ([]CPUInfo, error) {
	infos, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}
	result := make([]CPUInfo, 0, len(infos))
	for _, info := range infos {
		result = append(result, CPUInfo{
			VendorID:  info.VendorID,
			ModelName: info.ModelName,
			Cores:     info.Cores,
			Mhz:       info.Mhz,
			CacheSize: info.CacheSize,
		})
	}
	return result, nil
}

// GetDiskPartitions has moved to disk.go with semantic filtering; see
// IsRealDataMount for the inclusion rules.

// GetHostInfo returns detailed host/OS information.
func GetHostInfo() (HostInfoResult, error) {
	info, err := host.Info()
	if err != nil {
		return HostInfoResult{}, fmt.Errorf("failed to get host info: %w", err)
	}
	return HostInfoResult{
		Hostname:        info.Hostname,
		Uptime:          info.Uptime,
		BootTime:        info.BootTime,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
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
