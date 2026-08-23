// Package process provides functions for managing processes on the system.
// All functions return structured data and do not invoke shell commands.
package process

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	gopsnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Per-process info collection is best-effort with a timeout: on some hosts
// gopsutil blocks indefinitely on a single process in an odd state (e.g. a
// zombie under macOS SIP). One such process must not hang a whole fleet
// probe, so slow processes are skipped instead of awaited forever.
const perProcessTimeout = 500 * time.Millisecond

// ProcessInfo holds information about a single process.
type ProcessInfo struct {
	PID  int32  `json:"pid"`
	Name string `json:"name"`
	Exe  string `json:"exe"`
	// ExeDir is the directory of the executable (path.Dir of Exe) — the
	// "where does this process's binary live" answer, without another
	// syscall.
	ExeDir        string  `json:"exe_dir"`
	Cwd           string  `json:"cwd"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	CreateTime    int64   `json:"create_time"`
	Username      string  `json:"username"`
}

// ExecResult holds the result of executing an external command.
type ExecResult struct {
	Command  string   `json:"command"`
	Args     []string `json:"args"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	ExitCode int      `json:"exit_code"`
	Pid      int      `json:"pid"`
}

// getProcessInfo gathers information about a single process.
func getProcessInfo(p *process.Process) (ProcessInfo, error) {
	info := ProcessInfo{
		PID: p.Pid,
	}

	if name, err := p.Name(); err == nil {
		info.Name = name
	}
	if exe, err := p.Exe(); err == nil {
		info.Exe = exe
		if dir := filepath.Dir(exe); dir != "." {
			info.ExeDir = dir
		}
	}
	if cwd, err := p.Cwd(); err == nil {
		info.Cwd = cwd
	}
	if statuses, err := p.Status(); err == nil && len(statuses) > 0 {
		info.Status = statuses[0]
	}
	if cpu, err := p.CPUPercent(); err == nil {
		info.CPUPercent = cpu
	}
	if mem, err := p.MemoryPercent(); err == nil {
		info.MemoryPercent = mem
	}
	if ct, err := p.CreateTime(); err == nil {
		info.CreateTime = ct
	}
	if user, err := p.Username(); err == nil {
		info.Username = user
	}

	return info, nil
}

// List returns information about all running processes. Collection is
// best-effort: processes whose info cannot be gathered (permission denied,
// exited mid-read, or a blocked syscall) are skipped rather than allowed to
// hang the caller. Output order matches the process table.
func List() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	infos := make([]ProcessInfo, len(procs))
	collectable := make([]bool, len(procs))
	idx := make(chan int)
	var wg sync.WaitGroup
	workers := 8
	if workers > len(procs) {
		workers = len(procs)
	}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range idx {
				info, err := getProcessInfoWithTimeout(procs[i])
				if err != nil {
					continue // skip processes that error out or time out
				}
				infos[i] = info
				collectable[i] = true
			}
		}()
	}
	for i := range procs {
		idx <- i
	}
	close(idx)
	wg.Wait()

	result := make([]ProcessInfo, 0, len(procs))
	for i := range procs {
		if collectable[i] {
			result = append(result, infos[i])
		}
	}
	return result, nil
}

// getProcessInfoWithTimeout runs getProcessInfo with a hard per-process
// timeout. The collecting goroutine may stay blocked on a syscall — that is
// the price of not hanging the caller; it cannot be preempted safely.
func getProcessInfoWithTimeout(p *process.Process) (ProcessInfo, error) {
	type outcome struct {
		info ProcessInfo
		err  error
	}
	ch := make(chan outcome, 1)
	go func() {
		info, err := getProcessInfo(p)
		ch <- outcome{info, err}
	}()
	select {
	case o := <-ch:
		return o.info, o.err
	case <-time.After(perProcessTimeout):
		return ProcessInfo{}, fmt.Errorf("info collection for pid %d timed out after %s", p.Pid, perProcessTimeout)
	}
}

// FindByName returns processes whose name contains the given string (case-insensitive).
// An empty name matches all processes with a non-empty name.
func FindByName(name string) ([]ProcessInfo, error) {
	allProcs, err := List()
	if err != nil {
		return nil, err
	}

	searchLower := strings.ToLower(name)
	result := make([]ProcessInfo, 0, len(allProcs))

	for _, p := range allProcs {
		if strings.Contains(strings.ToLower(p.Name), searchLower) {
			result = append(result, p)
		}
	}

	return result, nil
}

// FindByPort returns processes listening on the given TCP port.
func FindByPort(port int) ([]ProcessInfo, error) {
	conns, err := gopsnet.Connections("inet")
	if err != nil {
		return nil, err
	}

	pidSet := make(map[int32]struct{})
	for _, conn := range conns {
		if int(conn.Laddr.Port) == port && conn.Status == "LISTEN" {
			if conn.Pid != 0 {
				pidSet[conn.Pid] = struct{}{}
			}
		}
	}

	result := make([]ProcessInfo, 0, len(pidSet))
	for pid := range pidSet {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		info, err := getProcessInfo(p)
		if err != nil {
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// Exec executes the given command with arguments, capturing stdout and stderr.
// It does not invoke a shell - the command is executed directly.
func Exec(command string, args []string) (ExecResult, error) {
	result := ExecResult{
		Command: command,
		Args:    args,
	}

	cmd := exec.Command(command, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()

	if cmd.Process != nil {
		result.Pid = cmd.Process.Pid
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			} else {
				result.ExitCode = 1
			}
		} else {
			result.ExitCode = -1
		}
	} else {
		result.ExitCode = 0
	}

	return result, nil
}
