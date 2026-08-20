// Package command provides Ansible command/shell module equivalent.
// Runs commands without shell interpolation by default.
package command

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// CommandResult is returned by Run.
type CommandResult struct {
	Command  string        `json:"cmd"`
	Stdout   string        `json:"stdout"`
	Stderr   string        `json:"stderr"`
	ExitCode int           `json:"rc"`
	Duration time.Duration `json:"duration_ms"`
	Changed  bool          `json:"changed"`
	Error    string        `json:"error,omitempty"`
}

// Run executes args without shell.
func Run(args []string, chdir string, creates string, removes string, timeout time.Duration) CommandResult {
	if len(args) == 0 {
		return CommandResult{Error: "args is required", ExitCode: 1}
	}
	if creates != "" {
		if _, err := exec.LookPath(creates); err == nil {
			return CommandResult{Changed: false}
		}
	}
	if chdir != "" {
		if err := exec.Command("test", "-d", chdir).Run(); err != nil {
			return CommandResult{Error: "chdir: " + chdir + " does not exist", ExitCode: 1}
		}
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	start := time.Now()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if chdir != "" {
		cmd.Dir = chdir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	dur := time.Since(start)

	result := CommandResult{
		Command:  strings.Join(args, " "),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: dur,
		Changed:  true,
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}
	return result
}

// Shell runs args with shell interpolation (/bin/sh -c).
func Shell(args []string, chdir string, creates string, removes string, timeout time.Duration, executable string) CommandResult {
	if len(args) == 0 {
		return CommandResult{Error: "args is required", ExitCode: 1}
	}
	cmdLine := strings.Join(args, " ")
	if executable == "" {
		executable = "/bin/sh"
	}
	return Run([]string{executable, "-c", cmdLine}, chdir, creates, removes, timeout)
}
