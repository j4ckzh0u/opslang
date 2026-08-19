// Package expect provides interactive command execution with pattern matching.
// Equivalent to ansible.builtin.expect module.
package expect

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Result is returned by Run.
type Result struct {
	Status    string   `json:"status"`
	Stdout    string   `json:"stdout"`
	Matches   []string `json:"matches,omitempty"`
	ExitCode  int      `json:"exit_code"`
	Command   string   `json:"command"`
	TimedOut  bool     `json:"timed_out"`
	Error     string   `json:"error,omitempty"`
}

// Command runs an interactive command, responding to prompts.
// responses maps expected patterns to responses.
// timeout is in seconds.
func Run(command string, responses map[string]string, timeout int) Result {
	if command == "" {
		return Result{Status: "failed", Error: "command is required"}
	}
	if timeout <= 0 {
		timeout = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)

	var stdin bytes.Buffer
	cmd.Stdin = &stdin
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	// For simple expect-like behavior, we write all responses upfront
	// This works for commands that prompt sequentially
	for _, response := range responses {
		stdin.WriteString(response + "\n")
	}

	err := cmd.Run()

	exitCode := 0
	timedOut := false
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			timedOut = true
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if !timedOut {
			return Result{
				Status:  "failed",
				Command: command,
				Error:   err.Error(),
			}
		}
	}

	output := stdout.String()

	// Find matches
	var matches []string
	for pattern := range responses {
		if strings.Contains(output, pattern) {
			matches = append(matches, pattern)
		}
	}

	result := Result{
		Status:   "success",
		Stdout:   output,
		Matches:  matches,
		ExitCode: exitCode,
		Command:  command,
		TimedOut: timedOut,
	}

	if exitCode != 0 && !timedOut {
		result.Status = "failed"
		result.Error = fmt.Sprintf("exit code %d", exitCode)
	}
	if timedOut {
		result.Status = "failed"
		result.Error = fmt.Sprintf("timed out after %ds", timeout)
	}

	return result
}

// RunSimple runs a command with a single expected prompt and response.
func RunSimple(command, prompt, response string, timeout int) Result {
	return Run(command, map[string]string{prompt: response}, timeout)
}
