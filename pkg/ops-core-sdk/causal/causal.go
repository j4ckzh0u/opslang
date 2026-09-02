// Package causal traces the read-only process ancestry of a Linux target.
// It intentionally uses /proc directly so compiled OpsLang binaries have no
// dependency on witr, Python, shell, or a target-side runtime installation.
package causal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	opsnet "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/net"
)

type Target struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type ProcessNode struct {
	PID              int32  `json:"pid"`
	PPID             int32  `json:"ppid"`
	Name             string `json:"name"`
	Executable       string `json:"executable,omitempty"`
	Command          string `json:"command,omitempty"`
	User             string `json:"user,omitempty"`
	StartTimeTicks   uint64 `json:"start_time_ticks,omitempty"`
	ContainerRuntime string `json:"container_runtime,omitempty"`
	ContainerID      string `json:"container_id,omitempty"`
}

type Edge struct {
	FromPID  int32  `json:"from_pid"`
	ToPID    int32  `json:"to_pid"`
	Relation string `json:"relation"`
}

type CausalTrace struct {
	Target      Target        `json:"target"`
	Nodes       []ProcessNode `json:"nodes"`
	Edges       []Edge        `json:"edges"`
	Warnings    []string      `json:"warnings,omitempty"`
	CollectedAt time.Time     `json:"collected_at"`
}

type PortConnection struct {
	PID         int32       `json:"pid"`
	ProcessName string      `json:"process_name,omitempty"`
	Protocol    string      `json:"protocol"`
	LocalAddr   string      `json:"local_addr"`
	RemoteAddr  string      `json:"remote_addr,omitempty"`
	Trace       CausalTrace `json:"trace"`
}

type FileTrace struct {
	PID         int32       `json:"pid"`
	FD          uint32      `json:"fd"`
	ProcessName string      `json:"process_name,omitempty"`
	Path        string      `json:"path"`
	Trace       CausalTrace `json:"trace"`
}

const maxAncestryDepth = 128

var containerPattern = regexp.MustCompile(`(?i)(?:/docker/|docker[-/]|containerd[-/]|crio[-/])([a-f0-9]{6,64})(?:\.scope)?$`)
var containerIDPattern = regexp.MustCompile(`(?i)^[a-f0-9]{6,64}$`)

func TracePID(pid int) (CausalTrace, error) {
	if pid <= 0 {
		return CausalTrace{}, fmt.Errorf("causal.trace_pid: pid must be positive")
	}
	if runtime.GOOS != "linux" {
		return CausalTrace{}, fmt.Errorf("causal.trace_pid: only supported on linux")
	}
	return tracePIDFromRoot("/proc", int32(pid))
}

// Find returns ancestry traces for processes whose name contains name.
func Find(name string) ([]CausalTrace, error) {
	if runtime.GOOS != "linux" {
		return []CausalTrace{}, nil
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("causal.find: read /proc: %w", err)
	}
	needle := strings.ToLower(name)
	pids := make([]int, 0)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		comm, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err == nil && strings.Contains(strings.ToLower(strings.TrimSpace(string(comm))), needle) {
			pids = append(pids, pid)
		}
	}
	sort.Ints(pids)
	traces := make([]CausalTrace, 0, len(pids))
	for _, pid := range pids {
		trace, err := TracePID(pid)
		if err != nil {
			continue
		}
		traces = append(traces, trace)
	}
	return traces, nil
}

// TracePort finds listening/connected sockets on port and traces each owner.
func TracePort(port int) ([]PortConnection, error) {
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("causal.trace_port: port must be between 1 and 65535")
	}
	if runtime.GOOS != "linux" {
		return []PortConnection{}, nil
	}
	connections, err := opsnet.Connections("inet")
	if err != nil {
		return nil, fmt.Errorf("causal.trace_port: %w", err)
	}
	return tracePortConnections("/proc", matchingPortConnections(connections, port))
}

// TraceFile finds processes with an open descriptor resolving to path.
func TraceFile(path string) ([]FileTrace, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "." || path == "" {
		return nil, fmt.Errorf("causal.trace_file: path must not be empty")
	}
	if runtime.GOOS != "linux" {
		return []FileTrace{}, nil
	}
	return traceFileFromRoot("/proc", path)
}

// TraceContainer finds processes belonging to a container cgroup ID.
func TraceContainer(id string) ([]CausalTrace, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	if len(id) < 6 || len(id) > 64 {
		return nil, fmt.Errorf("causal.trace_container: id must contain 6-64 hexadecimal characters")
	}
	if !containerIDPattern.MatchString(id) {
		return nil, fmt.Errorf("causal.trace_container: id must be hexadecimal")
	}
	if runtime.GOOS != "linux" {
		return []CausalTrace{}, nil
	}
	return traceContainerFromRoot("/proc", id)
}

func tracePIDFromRoot(root string, pid int32) (CausalTrace, error) {
	trace := CausalTrace{Target: Target{Kind: "pid", Value: strconv.Itoa(int(pid))}, CollectedAt: time.Now().UTC(), Nodes: []ProcessNode{}, Edges: []Edge{}, Warnings: []string{}}
	seen := map[int32]bool{}
	current := pid
	for depth := 0; depth < maxAncestryDepth && current > 0; depth++ {
		if seen[current] {
			trace.Warnings = append(trace.Warnings, fmt.Sprintf("cycle detected at pid %d", current))
			break
		}
		seen[current] = true
		node, err := readProcessNode(root, current)
		if err != nil {
			if len(trace.Nodes) == 0 {
				return CausalTrace{}, err
			}
			trace.Warnings = append(trace.Warnings, fmt.Sprintf("pid %d unavailable: %v", current, err))
			break
		}
		if len(trace.Nodes) > 0 {
			trace.Edges = append(trace.Edges, Edge{FromPID: trace.Nodes[len(trace.Nodes)-1].PID, ToPID: node.PID, Relation: "parent"})
		}
		trace.Nodes = append(trace.Nodes, node)
		if node.PPID <= 0 || node.PPID == node.PID {
			break
		}
		current = node.PPID
	}
	if len(trace.Nodes) == maxAncestryDepth {
		trace.Warnings = append(trace.Warnings, fmt.Sprintf("ancestry truncated at %d nodes", maxAncestryDepth))
	}
	return trace, nil
}

func matchingPortConnections(connections []opsnet.ConnectionInfo, port int) []PortConnection {
	matched := make([]PortConnection, 0)
	for _, conn := range connections {
		if parseAddrPort(conn.LocalAddr) != port || conn.Pid <= 0 {
			continue
		}
		matched = append(matched, PortConnection{PID: conn.Pid, ProcessName: conn.ProcessName, Protocol: conn.Protocol, LocalAddr: conn.LocalAddr, RemoteAddr: conn.RemoteAddr})
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].PID != matched[j].PID {
			return matched[i].PID < matched[j].PID
		}
		return matched[i].LocalAddr < matched[j].LocalAddr
	})
	return matched
}

func tracePortConnections(root string, connections []PortConnection) ([]PortConnection, error) {
	result := make([]PortConnection, 0, len(connections))
	for _, conn := range connections {
		trace, err := tracePIDFromRoot(root, conn.PID)
		if err != nil {
			continue
		}
		conn.Trace = trace
		result = append(result, conn)
	}
	return result, nil
}

func parseAddrPort(addr string) int {
	addr = strings.TrimSpace(addr)
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		port, _ := strconv.Atoi(addr[idx+1:])
		return port
	}
	return 0
}

func traceFileFromRoot(root, path string) ([]FileTrace, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("causal.trace_file: read proc: %w", err)
	}
	result := make([]FileTrace, 0)
	for _, entry := range entries {
		pidValue, err := strconv.Atoi(entry.Name())
		if err != nil || pidValue <= 0 {
			continue
		}
		fdDir := filepath.Join(root, entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fdEntry := range fds {
			fd, err := strconv.ParseUint(fdEntry.Name(), 10, 32)
			if err != nil {
				continue
			}
			link, err := os.Readlink(filepath.Join(fdDir, fdEntry.Name()))
			if err != nil {
				continue
			}
			link = strings.TrimSuffix(link, " (deleted)")
			if filepath.Clean(link) != path {
				continue
			}
			trace, err := tracePIDFromRoot(root, int32(pidValue))
			if err != nil {
				continue
			}
			result = append(result, FileTrace{PID: int32(pidValue), FD: uint32(fd), ProcessName: trace.Nodes[0].Name, Path: path, Trace: trace})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].PID != result[j].PID {
			return result[i].PID < result[j].PID
		}
		return result[i].FD < result[j].FD
	})
	return result, nil
}

func traceContainerFromRoot(root, id string) ([]CausalTrace, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("causal.trace_container: read proc: %w", err)
	}
	pids := make([]int, 0)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		cgroup, err := os.ReadFile(filepath.Join(root, entry.Name(), "cgroup"))
		if err != nil {
			continue
		}
		_, foundID := parseContainerCgroup(string(cgroup))
		if strings.EqualFold(foundID, id) {
			pids = append(pids, pid)
		}
	}
	sort.Ints(pids)
	result := make([]CausalTrace, 0, len(pids))
	for _, pid := range pids {
		trace, err := tracePIDFromRoot(root, int32(pid))
		if err == nil {
			result = append(result, trace)
		}
	}
	return result, nil
}

func readProcessNode(root string, pid int32) (ProcessNode, error) {
	dir := filepath.Join(root, strconv.Itoa(int(pid)))
	stat, err := os.ReadFile(filepath.Join(dir, "stat"))
	if err != nil {
		return ProcessNode{}, err
	}
	parsedPID, ppid, start, ok := parseProcStat(string(stat))
	if !ok || parsedPID != pid {
		return ProcessNode{}, fmt.Errorf("invalid /proc/%d/stat", pid)
	}
	comm, _ := os.ReadFile(filepath.Join(dir, "comm"))
	name := strings.TrimSpace(string(comm))
	if name == "" {
		name = procCommFromStat(string(stat))
	}
	cmdline, _ := os.ReadFile(filepath.Join(dir, "cmdline"))
	command := strings.TrimSpace(strings.ReplaceAll(string(cmdline), "\x00", " "))
	exe, _ := os.Readlink(filepath.Join(dir, "exe"))
	status, _ := os.ReadFile(filepath.Join(dir, "status"))
	user := parseUID(string(status))
	cgroup, _ := os.ReadFile(filepath.Join(dir, "cgroup"))
	runtimeName, containerID := parseContainerCgroup(string(cgroup))
	return ProcessNode{PID: pid, PPID: ppid, Name: name, Executable: exe, Command: command, User: user, StartTimeTicks: start, ContainerRuntime: runtimeName, ContainerID: containerID}, nil
}

func parseProcStat(stat string) (pid, ppid int32, start uint64, ok bool) {
	close := strings.LastIndex(stat, ")")
	if close < 0 {
		return 0, 0, 0, false
	}
	fields := strings.Fields(stat[close+1:])
	if len(fields) <= 19 {
		return 0, 0, 0, false
	}
	pid64, err1 := strconv.ParseInt(strings.TrimSpace(stat[:strings.Index(stat, " ")]), 10, 32)
	ppid64, err2 := strconv.ParseInt(fields[1], 10, 32)
	start64, err3 := strconv.ParseUint(fields[19], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return int32(pid64), int32(ppid64), start64, true
}

func procCommFromStat(stat string) string {
	start := strings.Index(stat, " (")
	end := strings.LastIndex(stat, ")")
	if start >= 0 && end > start+2 {
		return stat[start+2 : end]
	}
	return ""
}

func parseUID(status string) string {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				return fields[1]
			}
		}
	}
	return ""
}

func parseContainerCgroup(cgroup string) (runtimeName, id string) {
	for _, line := range strings.Split(cgroup, "\n") {
		path := line
		if idx := strings.LastIndex(line, ":"); idx >= 0 {
			path = line[idx+1:]
		}
		match := containerPattern.FindStringSubmatch(strings.TrimSpace(path))
		if len(match) == 2 {
			runtimeName = "docker"
			lower := strings.ToLower(path)
			if strings.Contains(lower, "containerd") {
				runtimeName = "containerd"
			} else if strings.Contains(lower, "crio") {
				runtimeName = "cri-o"
			}
			return runtimeName, match[1]
		}
	}
	return "", ""
}
