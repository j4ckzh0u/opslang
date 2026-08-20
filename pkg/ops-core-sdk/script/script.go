// Package script provides Ansible script module equivalent.
// Runs a local script on the remote target.
package script

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ScriptResult is returned by Run.
type ScriptResult struct {
	Script   string `json:"script"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"rc"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// Run executes a script file with optional arguments.
func Run(scriptPath string, args []string, chdir string, creates string, removes string, timeout time.Duration, executable string) ScriptResult {
	if scriptPath == "" {
		return ScriptResult{Error: "script path is required", ExitCode: 1}
	}
	info, err := os.Stat(scriptPath)
	if err != nil {
		return ScriptResult{Error: "script not found: " + scriptPath, ExitCode: 1}
	}
	if info.IsDir() {
		return ScriptResult{Error: "script path is a directory", ExitCode: 1}
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmdArgs := append([]string{scriptPath}, args...)
	var cmd *exec.Cmd
	if executable != "" {
		cmd = exec.CommandContext(ctx, executable, cmdArgs...)
	} else {
		cmd = exec.CommandContext(ctx, scriptPath, args...)
	}
	if chdir != "" {
		cmd.Dir = chdir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result := ScriptResult{
		Script:  scriptPath + " " + strings.Join(args, " "),
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Changed: true,
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = runErr.Error()
		}
	}
	return result
}
