// Package monit manages monit process monitoring.
// Equivalent to community.general.monit module.
package monit

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Service string `json:"service,omitempty"`
	Action  string `json:"action,omitempty"`
	Error   string `json:"error,omitempty"`
}

// StatusResult is returned by Status.
type StatusResult struct {
	Status   string `json:"status"`
	Services string `json:"services,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Start starts monitoring a service.
func Start(service string) Result {
	if service == "" {
		return Result{Status: "failed", Error: "service is required"}
	}

	cmd := exec.Command("monit", "start", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("monit start: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Service: service, Action: "start"}
}

// Stop stops monitoring a service.
func Stop(service string) Result {
	if service == "" {
		return Result{Status: "failed", Error: "service is required"}
	}

	cmd := exec.Command("monit", "stop", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("monit stop: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Service: service, Action: "stop"}
}

// Monitor enables monitoring of a service.
func Monitor(service string) Result {
	if service == "" {
		return Result{Status: "failed", Error: "service is required"}
	}

	cmd := exec.Command("monit", "monitor", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("monit monitor: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Service: service, Action: "monitor"}
}

// Unmonitor disables monitoring of a service.
func Unmonitor(service string) Result {
	if service == "" {
		return Result{Status: "failed", Error: "service is required"}
	}

	cmd := exec.Command("monit", "unmonitor", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("monit unmonitor: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Service: service, Action: "unmonitor"}
}

// Restart restarts a monitored service.
func Restart(service string) Result {
	if service == "" {
		return Result{Status: "failed", Error: "service is required"}
	}

	cmd := exec.Command("monit", "restart", service)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("monit restart: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Service: service, Action: "restart"}
}

// Status returns monit status summary.
func Status() StatusResult {
	cmd := exec.Command("monit", "summary")
	out, err := cmd.Output()
	if err != nil {
		return StatusResult{Status: "failed", Error: fmt.Sprintf("monit summary: %v", err)}
	}
	return StatusResult{Status: "success", Services: strings.TrimSpace(string(out))}
}

// Reload reloads monit configuration.
func Reload() Result {
	cmd := exec.Command("monit", "reload")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("monit reload: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Action: "reload"}
}
