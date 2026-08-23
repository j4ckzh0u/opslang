package opsnet

import (
	"fmt"
	"strings"

	gopsnet "github.com/shirou/gopsutil/v4/net"
	gopsproc "github.com/shirou/gopsutil/v4/process"
)

// ConnectionInfo describes one socket with its owning process, the pure-Go
// equivalent of `ss -tuanp` / `netstat -tuanp` without invoking either.
type ConnectionInfo struct {
	Fd uint32 `json:"fd"`
	// Protocol is "tcp" or "udp" (unix sockets are not enumerated).
	Protocol string `json:"protocol"`
	// LocalAddr is "ip:port"; the address is bracketed for IPv6.
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	Status     string `json:"status"`
	Pid        int32  `json:"pid"`
	// ProcessName is resolved from the process table; empty when the
	// socket cannot be attributed to a process (kernel-owned sockets or
	// insufficient permission to read another user's /proc/<pid>/fd).
	ProcessName string `json:"process_name"`
	Uid         int32  `json:"uid"`
}

// connectionsKind enumerates the gopsutil connection kinds we accept.
var connectionsKind = map[string]bool{
	"inet": true, "inet4": true, "inet6": true,
	"tcp": true, "tcp4": true, "tcp6": true,
	"udp": true, "udp4": true, "udp6": true,
}

// Connections enumerates sockets with owning-process attribution. kind
// selects the address-family/protocol filter: "inet" (default, IPv4+IPv6
// TCP+UDP), "tcp", "tcp4", "udp", ... On Linux the socket table itself is
// world-readable, but pid attribution for OTHER users' sockets requires
// root: those entries come back with pid=0 and an empty process_name
// instead of being hidden, so callers can count them honestly.
func Connections(kind string) ([]ConnectionInfo, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" {
		kind = "inet"
	}
	if !connectionsKind[kind] {
		return nil, fmt.Errorf("opsnet.connections: invalid kind %q (use inet|inet4|inet6|tcp|tcp4|tcp6|udp|udp4|udp6)", kind)
	}

	stats, err := gopsnet.Connections(kind)
	if err != nil {
		return nil, fmt.Errorf("opsnet.connections: %w", err)
	}

	names := connectionProcessNames(stats)
	result := make([]ConnectionInfo, 0, len(stats))
	for _, st := range stats {
		proto := "tcp"
		if st.Type == 2 { // SOCK_DGRAM
			proto = "udp"
		}
		c := ConnectionInfo{
			Fd:       st.Fd,
			Protocol: proto,
			Status:   st.Status,
			Pid:      st.Pid,
		}
		c.LocalAddr = formatAddr(st.Laddr.IP, st.Laddr.Port)
		if st.Raddr.Port != 0 || st.Raddr.IP != "" {
			c.RemoteAddr = formatAddr(st.Raddr.IP, st.Raddr.Port)
		}
		if len(st.Uids) > 0 {
			c.Uid = st.Uids[0]
		}
		if st.Pid > 0 {
			c.ProcessName = names[st.Pid]
		}
		result = append(result, c)
	}
	return result, nil
}

// formatAddr renders an ip:port pair, bracketing IPv6 literals the way
// ss/netstat do.
func formatAddr(ip string, port uint32) string {
	if strings.Contains(ip, ":") && !strings.HasPrefix(ip, "[") {
		return fmt.Sprintf("[%s]:%d", ip, port)
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

// connectionProcessNames resolves pid -> process name in bulk. It re-reads
// the process table rather than calling process.List to avoid an import
// cycle (process imports net for FindByPort).
func connectionProcessNames(stats []gopsnet.ConnectionStat) map[int32]string {
	pids := make(map[int32]struct{}, 16)
	for _, st := range stats {
		if st.Pid > 0 {
			pids[st.Pid] = struct{}{}
		}
	}
	if len(pids) == 0 {
		return map[int32]string{}
	}

	names := make(map[int32]string, len(pids))
	for pid := range pids {
		if name, err := gopsProcName(pid); err == nil && name != "" {
			names[pid] = name
		}
	}
	return names
}

// gopsProcName returns the process name for a pid via /proc, best-effort.
func gopsProcName(pid int32) (string, error) {
	p, err := gopsproc.NewProcess(pid)
	if err != nil {
		return "", err
	}
	return p.Name()
}
