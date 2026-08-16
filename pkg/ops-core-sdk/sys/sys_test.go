package sys

import (
	"encoding/json"
	"math"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

// ---------- Struct JSON marshaling tests ----------

func TestCPUUsageJSON(t *testing.T) {
	v := CPUUsage{Percent: 12.5, User: 5.0, System: 3.0, Idle: 92.0}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	for _, key := range []string{"percent", "user", "system", "idle"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
	if m["percent"].(float64) != 12.5 {
		t.Errorf("percent = %v, want 12.5", m["percent"])
	}
}

func TestMemoryInfoJSON(t *testing.T) {
	v := MemoryInfo{Total: 16777216, Available: 8388608, Used: 8388608, UsedPercent: 50.0}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	for _, key := range []string{"total", "available", "used", "used_percent"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestDiskUsageJSON(t *testing.T) {
	v := DiskUsage{Path: "/", Total: 100, Used: 50, Free: 50, UsedPercent: 50.0}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if m["path"].(string) != "/" {
		t.Errorf("path = %v, want /", m["path"])
	}
	for _, key := range []string{"path", "total", "used", "free", "used_percent"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestLoadAvgJSON(t *testing.T) {
	v := LoadAvg{Load1: 1.0, Load5: 0.5, Load15: 0.25}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	for _, key := range []string{"load1", "load5", "load15"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestHostnameInfoJSON(t *testing.T) {
	v := HostnameInfo{Hostname: "myhost", FQDN: "myhost.example.com"}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if m["hostname"].(string) != "myhost" {
		t.Errorf("hostname = %v, want myhost", m["hostname"])
	}
	if m["fqdn"].(string) != "myhost.example.com" {
		t.Errorf("fqdn = %v, want myhost.example.com", m["fqdn"])
	}
}

func TestUptimeInfoJSON(t *testing.T) {
	v := UptimeInfo{Uptime: 3600, BootTime: 1700000000}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	for _, key := range []string{"uptime", "boot_time"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestUserInfoJSON(t *testing.T) {
	v := UserInfo{User: "root", Terminal: "pts/0", Host: "10.0.0.1", StartTime: 1700000000}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	for _, key := range []string{"user", "terminal", "host", "start_time"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

// TestUsersJSONEmptySlice verifies that an empty Users slice marshals to [] not null.
func TestUsersJSONEmptySlice(t *testing.T) {
	users := []UserInfo{}
	data, err := json.Marshal(users)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	if string(data) != "[]" {
		t.Errorf("empty slice marshaled to %s, want []", string(data))
	}
}

// ---------- Function tests ----------

func TestGetCPUUsage(t *testing.T) {
	result, err := GetCPUUsage()
	if err != nil {
		t.Fatalf("GetCPUUsage() error: %v", err)
	}

	if result.Percent < 0 || result.Percent > 100 {
		t.Errorf("Percent = %f, want [0, 100]", result.Percent)
	}
	if result.User < 0 {
		t.Errorf("User = %f, want >= 0", result.User)
	}
	if result.System < 0 {
		t.Errorf("System = %f, want >= 0", result.System)
	}
	if result.Idle < 0 || result.Idle > 100 {
		t.Errorf("Idle = %f, want [0, 100]", result.Idle)
	}
	if math.IsNaN(result.Percent) || math.IsInf(result.Percent, 0) {
		t.Errorf("Percent is NaN or Inf")
	}
	if math.IsNaN(result.User) || math.IsInf(result.User, 0) {
		t.Errorf("User is NaN or Inf")
	}
	if math.IsNaN(result.System) || math.IsInf(result.System, 0) {
		t.Errorf("System is NaN or Inf")
	}
	if math.IsNaN(result.Idle) || math.IsInf(result.Idle, 0) {
		t.Errorf("Idle is NaN or Inf")
	}

	// User + System should approximately equal Percent (non-idle portion)
	activeSum := result.User + result.System
	if activeSum > 0 && result.Percent > 0 {
		diff := math.Abs(activeSum-result.Percent) / result.Percent
		if diff > 0.05 {
			t.Errorf("User+System (%f) differs from Percent (%f) by %.1f%%",
				activeSum, result.Percent, diff*100)
		}
	}

	// Idle + (User+System) should approximately equal 100
	total := result.Idle + result.User + result.System
	if total > 0 {
		diff := math.Abs(total-100) / 100
		if diff > 0.05 {
			t.Errorf("Idle+User+System = %f, want ~100 (diff %.1f%%)",
				total, diff*100)
		}
	}

	t.Logf("CPU: Percent=%.1f%% User=%.1f%% System=%.1f%% Idle=%.1f%%",
		result.Percent, result.User, result.System, result.Idle)
}

func TestGetMemoryInfo(t *testing.T) {
	result, err := GetMemoryInfo()
	if err != nil {
		t.Fatalf("GetMemoryInfo() error: %v", err)
	}

	if result.Total == 0 {
		t.Error("Total memory is 0")
	}
	if result.Used > result.Total {
		t.Errorf("Used (%d) > Total (%d)", result.Used, result.Total)
	}
	if result.UsedPercent < 0 || result.UsedPercent > 100 {
		t.Errorf("UsedPercent = %f, want [0, 100]", result.UsedPercent)
	}

	// Available should not exceed Total
	if result.Available > result.Total {
		t.Errorf("Available (%d) > Total (%d)", result.Available, result.Total)
	}

	t.Logf("Memory: Total=%d Available=%d Used=%d UsedPercent=%.1f%%",
		result.Total, result.Available, result.Used, result.UsedPercent)
}

func TestGetDiskUsage(t *testing.T) {
	// Test with current working directory (always exists)
	result, err := GetDiskUsage(".")
	if err != nil {
		t.Fatalf("GetDiskUsage(\".\") error: %v", err)
	}

	if result.Total == 0 {
		t.Error("Total disk space is 0")
	}
	if result.Used > result.Total {
		t.Errorf("Used (%d) > Total (%d)", result.Used, result.Total)
	}
	if result.UsedPercent < 0 || result.UsedPercent > 100 {
		t.Errorf("UsedPercent = %f, want [0, 100]", result.UsedPercent)
	}
	if result.Path == "" {
		t.Error("Path is empty")
	}

	t.Logf("Disk: Path=%s Total=%d Used=%d Free=%d UsedPercent=%.1f%%",
		result.Path, result.Total, result.Used, result.Free, result.UsedPercent)
}

func TestGetDiskUsageWithTempDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "sys-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	result, err := GetDiskUsage(dir)
	if err != nil {
		t.Fatalf("GetDiskUsage(%q) error: %v", dir, err)
	}
	if result.Total == 0 {
		t.Error("Total disk space is 0")
	}
}

func TestGetDiskUsageEmptyPath(t *testing.T) {
	_, err := GetDiskUsage("")
	if err == nil {
		t.Error("GetDiskUsage(\"\") expected error, got nil")
	}
}

func TestGetDiskUsageInvalidPath(t *testing.T) {
	_, err := GetDiskUsage("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("GetDiskUsage with invalid path expected error, got nil")
	}
}

func TestGetLoadAvg(t *testing.T) {
	result, err := GetLoadAvg()
	if err != nil {
		t.Fatalf("GetLoadAvg() error: %v", err)
	}

	if result.Load1 < 0 {
		t.Errorf("Load1 = %f, want >= 0", result.Load1)
	}
	if result.Load5 < 0 {
		t.Errorf("Load5 = %f, want >= 0", result.Load5)
	}
	if result.Load15 < 0 {
		t.Errorf("Load15 = %f, want >= 0", result.Load15)
	}
	if math.IsNaN(result.Load1) || math.IsInf(result.Load1, 0) {
		t.Errorf("Load1 is NaN or Inf")
	}

	t.Logf("LoadAvg: 1m=%.2f 5m=%.2f 15m=%.2f",
		result.Load1, result.Load5, result.Load15)
}

func TestHostname(t *testing.T) {
	result, err := Hostname()
	if err != nil {
		t.Fatalf("Hostname() error: %v", err)
	}

	if result.Hostname == "" {
		t.Error("Hostname is empty")
	}
	if result.FQDN == "" {
		t.Error("FQDN is empty")
	}

	t.Logf("Hostname: %s, FQDN: %s", result.Hostname, result.FQDN)
}

func TestUptime(t *testing.T) {
	result, err := Uptime()
	if err != nil {
		t.Fatalf("Uptime() error: %v", err)
	}

	if result.Uptime == 0 {
		t.Error("Uptime is 0")
	}
	if result.BootTime == 0 {
		t.Error("BootTime is 0")
	}

	t.Logf("Uptime: %ds, BootTime: %d", result.Uptime, result.BootTime)
}

func TestUsers(t *testing.T) {
	result, err := Users()
	if err != nil {
		t.Fatalf("Users() error: %v", err)
	}

	// Must return non-nil slice (even if empty)
	if result == nil {
		t.Fatal("Users() returned nil, want non-nil slice")
	}

	for i, u := range result {
		if u.User == "" {
			t.Errorf("Users()[%d].User is empty", i)
		}
	}

	t.Logf("Users: %d logged in", len(result))
	for _, u := range result {
		t.Logf("  %s on %s from %s", u.User, u.Terminal, u.Host)
	}
}

// TestRoundTripJSON verifies that structs survive a JSON round-trip without data loss.
func TestRoundTripJSON(t *testing.T) {
	t.Run("CPUUsage", func(t *testing.T) {
		orig := CPUUsage{Percent: 42.5, User: 20.3, System: 10.2, Idle: 72.0}
		data, _ := json.Marshal(orig)
		var decoded CPUUsage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})

	t.Run("MemoryInfo", func(t *testing.T) {
		orig := MemoryInfo{Total: 1024, Available: 512, Used: 512, UsedPercent: 50.0}
		data, _ := json.Marshal(orig)
		var decoded MemoryInfo
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})

	t.Run("DiskUsage", func(t *testing.T) {
		orig := DiskUsage{Path: "/data", Total: 200, Used: 100, Free: 100, UsedPercent: 50.0}
		data, _ := json.Marshal(orig)
		var decoded DiskUsage
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})

	t.Run("LoadAvg", func(t *testing.T) {
		orig := LoadAvg{Load1: 1.5, Load5: 2.0, Load15: 1.0}
		data, _ := json.Marshal(orig)
		var decoded LoadAvg
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})

	t.Run("HostnameInfo", func(t *testing.T) {
		orig := HostnameInfo{Hostname: "host1", FQDN: "host1.example.com"}
		data, _ := json.Marshal(orig)
		var decoded HostnameInfo
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})

	t.Run("UptimeInfo", func(t *testing.T) {
		orig := UptimeInfo{Uptime: 86400, BootTime: 1700000000}
		data, _ := json.Marshal(orig)
		var decoded UptimeInfo
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})

	t.Run("UserInfo", func(t *testing.T) {
		orig := UserInfo{User: "admin", Terminal: "tty1", Host: "localhost", StartTime: 1700000000}
		data, _ := json.Marshal(orig)
		var decoded UserInfo
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("round-trip failed: %v", err)
		}
		if decoded != orig {
			t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
		}
	})
}

// ---------- Live system function tests ----------

func TestGetNetInterfaces(t *testing.T) {
	result, err := GetNetInterfaces()
	if err != nil {
		t.Fatalf("GetNetInterfaces() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("GetNetInterfaces() returned empty list, expected at least loopback")
	}
	// Verify struct fields are populated
	for _, iface := range result {
		if iface.Name == "" {
			t.Error("interface has empty Name")
		}
		if iface.MTU < 0 {
			t.Errorf("interface %q has non-positive MTU: %d", iface.Name, iface.MTU)
		}
	}
}

func TestGetCPUCount(t *testing.T) {
	result, err := GetCPUCount()
	if err != nil {
		t.Fatalf("GetCPUCount() error = %v", err)
	}
	if result.Logical <= 0 {
		t.Errorf("Logical CPU count = %d, want > 0", result.Logical)
	}
	if result.Physical <= 0 {
		t.Errorf("Physical CPU count = %d, want > 0", result.Physical)
	}
	if result.Physical > result.Logical {
		t.Errorf("Physical (%d) > Logical (%d), unexpected", result.Physical, result.Logical)
	}
}

func TestGetCPUInfo(t *testing.T) {
	result, err := GetCPUInfo()
	if err != nil {
		t.Fatalf("GetCPUInfo() error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("GetCPUInfo() returned empty list")
	}
	for _, info := range result {
		if info.ModelName == "" {
			t.Error("CPU info returned empty ModelName")
		}
		if info.Cores <= 0 {
			t.Errorf("Cores = %d, want > 0", info.Cores)
		}
	}
}

func TestGetDiskPartitions(t *testing.T) {
	result, err := GetDiskPartitions()
	if err != nil {
		t.Fatalf("GetDiskPartitions() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("GetDiskPartitions() returned empty list")
	}
	for _, p := range result {
		if p.Mountpoint == "" {
			t.Error("partition has empty Mountpoint")
		}
		if p.Fstype == "" {
			t.Error("partition has empty Fstype")
		}
	}
}

func TestGetHostInfo(t *testing.T) {
	result, err := GetHostInfo()
	if err != nil {
		t.Fatalf("GetHostInfo() error = %v", err)
	}
	if result.OS == "" {
		t.Error("GetHostInfo() returned empty OS")
	}
	if result.Platform == "" {
		t.Error("GetHostInfo() returned empty Platform")
	}
	if result.KernelArch == "" {
		t.Error("GetHostInfo() returned empty KernelArch")
	}
}

func TestHostnameJSON(t *testing.T) {
	orig := HostnameInfo{Hostname: "testhost", FQDN: "testhost.example.com"}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded HostnameInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded != orig {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, orig)
	}
}

// ---------- CPU usage sampling tests ----------

func TestComputeUsageDelta(t *testing.T) {
	// Deltas: user +60, system +20, idle +20 -> total 100 ticks.
	// Expected: Percent 80, User 60, System 20, Idle 20.
	t1 := cpu.TimesStat{User: 10, System: 5, Idle: 85}
	t2 := cpu.TimesStat{User: 70, System: 25, Idle: 105}

	usage, ok := computeUsageDelta(t1, t2)
	if !ok {
		t.Fatal("computeUsageDelta returned ok=false for a positive delta")
	}

	if usage.Percent != 80 {
		t.Errorf("Percent = %v, want 80", usage.Percent)
	}
	if usage.Idle != 20 {
		t.Errorf("Idle = %v, want 20", usage.Idle)
	}
	if usage.User != 60 {
		t.Errorf("User = %v, want 60", usage.User)
	}
	if usage.System != 20 {
		t.Errorf("System = %v, want 20", usage.System)
	}
}

func TestComputeUsageDeltaNoDelta(t *testing.T) {
	t1 := cpu.TimesStat{User: 10, System: 5, Idle: 80}
	if _, ok := computeUsageDelta(t1, t1); ok {
		t.Error("computeUsageDelta returned ok=true for identical samples")
	}
}

func TestGetCPUUsageMeasuresCurrentWindow(t *testing.T) {
	// Burn CPU on all cores and verify the reported usage reflects the
	// sampling window (high), not a since-boot lifetime average (which
	// would be far lower right after boot on an idle machine).
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
			}
		}()
	}

	usage, err := GetCPUUsageInterval(400 * time.Millisecond)
	close(stop)
	wg.Wait()

	if err != nil {
		t.Fatalf("GetCPUUsageInterval error: %v", err)
	}
	if usage.Percent < 40 {
		t.Errorf("Percent under full load = %v, want >= 40 (windowed sampling appears broken)", usage.Percent)
	}
}

func TestGetCPUUsageSinceBootFallback(t *testing.T) {
	usage, err := GetCPUUsageInterval(0)
	if err != nil {
		t.Fatalf("GetCPUUsageInterval(0) error: %v", err)
	}
	if usage.Percent < 0 || usage.Percent > 100 {
		t.Errorf("Percent = %v, out of range", usage.Percent)
	}
}
