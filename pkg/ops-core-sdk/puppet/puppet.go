// Package puppet manages Puppet agent operations.
// Equivalent to community.general.puppet module.
package puppet

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StatusResult is returned by Status.
type StatusResult struct {
	Status      string `json:"status"`
	Running     bool   `json:"running"`
	LastRun     string `json:"last_run,omitempty"`
	Failed      bool   `json:"failed"`
	PuppetAgent string `json:"puppet_agent,omitempty"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
}

// Run executes a Puppet agent run.
func Run(environment string, tags []string) Result {
	args := []string{"agent", "--no-daemonize", "--verbose"}

	if environment != "" {
		args = append(args, "--environment="+environment)
	}
	for _, tag := range tags {
		args = append(args, "--tags="+tag)
	}

	cmd := exec.Command("puppet", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("puppet: %v", err)}
	}

	changed := strings.Contains(output, "Total resources: ") && !strings.Contains(output, "Changed: 0")
	return Result{Status: "success", Changed: changed, Output: output}
}

// RunNoop runs Puppet in noop (dry-run) mode.
func RunNoop(environment string, tags []string) Result {
	args := []string{"agent", "--no-daemonize", "--noop", "--verbose"}

	if environment != "" {
		args = append(args, "--environment="+environment)
	}
	for _, tag := range tags {
		args = append(args, "--tags="+tag)
	}

	cmd := exec.Command("puppet", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		return Result{Status: "failed", Output: output, Error: fmt.Sprintf("puppet noop: %v", err)}
	}
	return Result{Status: "success", Changed: false, Output: output}
}

// Status returns the Puppet agent status.
func Status() StatusResult {
	cmd := exec.Command("puppet", "agent", "--status")
	out, err := cmd.Output()
	if err != nil {
		return StatusResult{Status: "failed", Error: fmt.Sprintf("puppet status: %v", err)}
	}

	output := strings.TrimSpace(string(out))
	running := strings.Contains(output, "running")
	failed := strings.Contains(output, "failed")
	lastRun := ""
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "last_run") || strings.Contains(line, "Last Run") {
			lastRun = strings.TrimSpace(line)
			break
		}
	}

	return StatusResult{
		Status:  "success",
		Running: running,
		LastRun: lastRun,
		Failed:  failed,
		Output:  output,
	}
}

// Disable disables the Puppet agent (creates the disabled lockfile).
func Disable(message string) Result {
	args := []string{"agent", "--disable"}
	if message != "" {
		args = append(args, "--"+message)
	}

	cmd := exec.Command("puppet", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Output: strings.TrimSpace(string(out)), Error: fmt.Sprintf("puppet disable: %v", err)}
	}
	return Result{Status: "success", Changed: true, Output: "disabled"}
}

// Enable enables the Puppet agent (removes the disabled lockfile).
func Enable() Result {
	cmd := exec.Command("puppet", "agent", "--enable")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Output: strings.TrimSpace(string(out)), Error: fmt.Sprintf("puppet enable: %v", err)}
	}
	return Result{Status: "success", Changed: true, Output: "enabled"}
}

// Fact returns a Puppet fact value.
func Fact(name string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "fact name is required"}
	}

	cmd := exec.Command("facter", name)
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("facter: %v", err)}
	}
	return Result{Status: "success", Output: strings.TrimSpace(string(out))}
}

// ModuleList lists installed Puppet modules.
func ModuleList() Result {
	cmd := exec.Command("puppet", "module", "list")
	out, err := cmd.Output()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("puppet module list: %v", err)}
	}
	return Result{Status: "success", Output: strings.TrimSpace(string(out))}
}

// ModuleInstall installs a Puppet module.
func ModuleInstall(name string, version string) Result {
	if name == "" {
		return Result{Status: "failed", Error: "module name is required"}
	}

	args := []string{"module", "install", name}
	if version != "" {
		args = append(args, "--version="+version)
	}

	cmd := exec.Command("puppet", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Output: strings.TrimSpace(string(out)), Error: fmt.Sprintf("puppet module install: %v", err)}
	}
	return Result{Status: "success", Changed: true, Output: strings.TrimSpace(string(out))}
}
