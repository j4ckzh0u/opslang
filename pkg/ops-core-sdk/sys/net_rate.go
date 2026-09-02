package sys

import (
	"fmt"
	"sort"
	"time"

	gopsnet "github.com/shirou/gopsutil/v4/net"
)

// NetIOCounters is the portable subset of interface counters used for rate
// sampling. Counters are cumulative since boot.
type NetIOCounters struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"bytes_sent"`
	BytesRecv uint64 `json:"bytes_recv"`
}

// NetInterfaceRate reports traffic observed during a sampling window.
type NetInterfaceRate struct {
	Name          string  `json:"name"`
	BytesSent     uint64  `json:"bytes_sent"`
	BytesRecv     uint64  `json:"bytes_recv"`
	BitsPerSecond float64 `json:"bits_per_second"`
	WindowSeconds int     `json:"window_seconds"`
}

// NetRate is an aggregate receive/transmit rate over a measured window.
type NetRate struct {
	WindowSeconds int                `json:"window_seconds"`
	BytesSent     uint64             `json:"bytes_sent"`
	BytesRecv     uint64             `json:"bytes_recv"`
	BitsPerSecond float64            `json:"bits_per_second"`
	Interfaces    []NetInterfaceRate `json:"interfaces"`
}

// GetNetRate samples cumulative interface counters for seconds. A three
// second window is suitable for an interactive health check; callers may use
// 300 for a five-minute average when a longer observation is required.
func GetNetRate(seconds int) (NetRate, error) {
	if seconds <= 0 || seconds > 3600 {
		return NetRate{}, fmt.Errorf("sys.net.rate: seconds must be between 1 and 3600")
	}
	before, err := readNetCounters()
	if err != nil {
		return NetRate{}, fmt.Errorf("sys.net.rate: initial sample: %w", err)
	}
	start := time.Now()
	time.Sleep(time.Duration(seconds) * time.Second)
	after, err := readNetCounters()
	if err != nil {
		return NetRate{}, fmt.Errorf("sys.net.rate: final sample: %w", err)
	}
	window := time.Since(start).Seconds()
	if window <= 0 {
		window = float64(seconds)
	}
	return calculateNetRateWithDuration(before, after, seconds, window), nil
}

func readNetCounters() (map[string]NetIOCounters, error) {
	stats, err := gopsnet.IOCounters(true)
	if err != nil {
		return nil, err
	}
	result := make(map[string]NetIOCounters, len(stats))
	for _, stat := range stats {
		result[stat.Name] = NetIOCounters{Name: stat.Name, BytesSent: stat.BytesSent, BytesRecv: stat.BytesRecv}
	}
	return result, nil
}

func calculateNetRate(before, after map[string]NetIOCounters, seconds int) NetRate {
	return calculateNetRateWithDuration(before, after, seconds, float64(seconds))
}

func calculateNetRateWithDuration(before, after map[string]NetIOCounters, seconds int, duration float64) NetRate {
	if duration <= 0 {
		duration = 1
	}
	names := make([]string, 0, len(after))
	for name := range after {
		names = append(names, name)
	}
	sort.Strings(names)
	result := NetRate{WindowSeconds: seconds, Interfaces: make([]NetInterfaceRate, 0, len(names))}
	for _, name := range names {
		now := after[name]
		old := before[name]
		if now.BytesSent < old.BytesSent || now.BytesRecv < old.BytesRecv {
			continue // interface restarted or counter wrapped during sampling
		}
		sent := now.BytesSent - old.BytesSent
		recv := now.BytesRecv - old.BytesRecv
		result.BytesSent += sent
		result.BytesRecv += recv
		result.Interfaces = append(result.Interfaces, NetInterfaceRate{
			Name: name, BytesSent: sent, BytesRecv: recv,
			BitsPerSecond: float64(sent+recv) * 8 / duration,
			WindowSeconds: seconds,
		})
	}
	result.BitsPerSecond = float64(result.BytesSent+result.BytesRecv) * 8 / duration
	return result
}
