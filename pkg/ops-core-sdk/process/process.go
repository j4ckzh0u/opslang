// Package process provides functions for managing processes on the system.
// All functions return structured data and do not invoke shell commands.
package process

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"

	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds information about a single process.
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Exe           string  `json:"exe"`
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

// List returns information about all running processes.
func List() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	result := make([]ProcessInfo, 0, len(procs))
	for _, p := range procs {
		info, err := getProcessInfo(p)
		if err != nil {
			// Skip processes that error out during info gathering
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// FindByName returns processes whose name contains the given string (case-insensitive).
func FindByName(name string) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	searchLower := strings.ToLower(name)
	result := make([]ProcessInfo, 0)

	for _, p := range procs {
		procName, err := p.Name()
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(procName), searchLower) {
			info, err := getProcessInfo(p)
			if err != nil {
				continue
			}
			result = append(result, info)
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
