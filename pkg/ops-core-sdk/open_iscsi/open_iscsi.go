// Package open_iscsi manages iSCSI initiator configuration.
// Equivalent to community.general.open_iscsi module.
package open_iscsi

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Target  string `json:"target,omitempty"`
	Portal  string `json:"portal,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NodeResult is returned by discovery.
type NodeResult struct {
	Status string   `json:"status"`
	Nodes  []string `json:"nodes"`
	Error  string   `json:"error,omitempty"`
}

// SessionResult is returned by session queries.
type SessionResult struct {
	Status   string   `json:"status"`
	Sessions []string `json:"sessions"`
	Error    string   `json:"error,omitempty"`
}

// Discover discovers iSCSI targets on a portal.
func Discover(portal string, port int) NodeResult {
	if portal == "" {
		return NodeResult{Status: "failed", Error: "portal is required"}
	}
	if port <= 0 {
		port = 3260
	}

	cmd := exec.Command("iscsiadm", "-m", "discovery", "-t", "sendtargets", "-p", fmt.Sprintf("%s:%d", portal, port))
	out, err := cmd.Output()
	if err != nil {
		return NodeResult{Status: "failed", Error: fmt.Sprintf("iscsiadm discover: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	nodes := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			nodes = append(nodes, line)
		}
	}
	return NodeResult{Status: "success", Nodes: nodes}
}

// Login connects to an iSCSI target.
func Login(target string, portal string) Result {
	if target == "" {
		return Result{Status: "failed", Error: "target is required"}
	}
	if portal == "" {
		return Result{Status: "failed", Error: "portal is required"}
	}

	cmd := exec.Command("iscsiadm", "-m", "node", "-T", target, "-p", portal, "--login")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("iscsiadm login: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Target: target, Portal: portal}
}

// Logout disconnects from an iSCSI target.
func Logout(target string, portal string) Result {
	if target == "" {
		return Result{Status: "failed", Error: "target is required"}
	}
	if portal == "" {
		return Result{Status: "failed", Error: "portal is required"}
	}

	cmd := exec.Command("iscsiadm", "-m", "node", "-T", target, "-p", portal, "--logout")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("iscsiadm logout: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Target: target, Portal: portal}
}

// ListSessions returns active iSCSI sessions.
func ListSessions() SessionResult {
	cmd := exec.Command("iscsiadm", "-m", "session")
	out, err := cmd.Output()
	if err != nil {
		// Exit code 21 means no active sessions
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 21 {
			return SessionResult{Status: "success", Sessions: []string{}}
		}
		return SessionResult{Status: "failed", Error: fmt.Sprintf("iscsiadm session: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	sessions := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return SessionResult{Status: "success", Sessions: sessions}
}

// ListNodes returns all known iSCSI nodes.
func ListNodes() NodeResult {
	cmd := exec.Command("iscsiadm", "-m", "node")
	out, err := cmd.Output()
	if err != nil {
		return NodeResult{Status: "failed", Error: fmt.Sprintf("iscsiadm node: %v", err)}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	nodes := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			nodes = append(nodes, line)
		}
	}
	return NodeResult{Status: "success", Nodes: nodes}
}

// SetStartup sets the startup mode for a node.
func SetStartup(target string, portal string, startup string) Result {
	if target == "" || portal == "" || startup == "" {
		return Result{Status: "failed", Error: "target, portal, and startup are required"}
	}

	cmd := exec.Command("iscsiadm", "-m", "node", "-T", target, "-p", portal,
		"--op=update", "-n", "node.startup", "-v", startup)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{Status: "failed", Error: fmt.Sprintf("iscsiadm set-startup: %v: %s", err, strings.TrimSpace(string(out)))}
	}
	return Result{Status: "success", Changed: true, Target: target, Portal: portal}
}
