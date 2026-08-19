// Package raw executes raw commands without SDK wrapping.
// Equivalent to ansible.builtin.raw module.
package raw

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result is returned by Execute.
type Result struct {
	Status     string `json:"status"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	Command    string `json:"command"`
	ElapsedMs  int64  `json:"elapsed_ms"`
	Error      string `json:"error,omitempty"`
}

// Execute runs a raw command string through /bin/sh -c.
// This is unlike process.exec which avoids shell; raw uses shell for
// compatibility with shell syntax (pipes, redirects, etc.).
func Execute(command string, timeout int) Result {
	if command == "" {
		return Result{Status: "failed", Error: "command is required"}
	}

	if timeout <= 0 {
		timeout = 30
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{
				Status:    "failed",
				Command:   command,
				ElapsedMs: elapsed,
				Error:     err.Error(),
			}
		}
	}

	result := Result{
		Status:    "success",
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  exitCode,
		Command:   command,
		ElapsedMs: elapsed,
	}

	if exitCode != 0 {
		result.Status = "failed"
		result.Error = fmt.Sprintf("exit code %d", exitCode)
	}

	return result
}

// ExecuteWithEnv runs a raw command with additional environment variables.
func ExecuteWithEnv(command string, timeout int, env map[string]string) Result {
	if command == "" {
		return Result{Status: "failed", Error: "command is required"}
	}
	if timeout <= 0 {
		timeout = 30
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)

	// Set environment
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{
				Status:    "failed",
				Command:   command,
				ElapsedMs: elapsed,
				Error:     err.Error(),
			}
		}
	}

	result := Result{
		Status:    "success",
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  exitCode,
		Command:   command,
		ElapsedMs: elapsed,
	}

	if exitCode != 0 {
		result.Status = "failed"
		result.Error = fmt.Sprintf("exit code %d", exitCode)
	}

	return result
}
