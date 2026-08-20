// Package package provides a generic package manager wrapper.
// Auto-detects apt/yum/dnf/pacman/zypper/apk and delegates.
package package_mgr

import (
	"os/exec"
	"runtime"
	"strings"
)

// PkgResult is returned by all operations.
type PkgResult struct {
	Changed bool   `json:"changed"`
	Name    string `json:"name"`
	Action  string `json:"action"` // install/remove/update
	State   string `json:"state"`  // present/absent/latest
	Manager string `json:"manager"` // detected package manager
	Error   string `json:"error,omitempty"`
}

// InfoResult contains package info.
type InfoResult struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Manager string `json:"manager"`
	Exists  bool   `json:"exists"`
	Error   string `json:"error,omitempty"`
}

// Install installs a package using the detected package manager.
func Install(name string) PkgResult {
	return runAction(name, "install", "present")
}

// Remove removes a package.
func Remove(name string) PkgResult {
	return runAction(name, "remove", "absent")
}

// Update updates a package to latest version.
func Update(name string) PkgResult {
	return runAction(name, "update", "latest")
}

// Info returns package information.
func Info(name string) InfoResult {
	mgr := detectManager()
	if mgr == "" {
		return InfoResult{Name: name, Error: "no supported package manager found"}
	}
	var cmd *exec.Cmd
	switch mgr {
	case "apt":
		cmd = exec.Command("dpkg-query", "-W", "-f", "${Version}", name)
	case "yum", "dnf":
		cmd = exec.Command("rpm", "-q", "--qf", "%{VERSION}-%{RELEASE}", name)
	case "pacman":
		cmd = exec.Command("pacman", "-Qi", name)
	case "zypper":
		cmd = exec.Command("rpm", "-q", "--qf", "%{VERSION}-%{RELEASE}", name)
	case "apk":
		cmd = exec.Command("apk", "info", "-v", name)
	}
	if cmd == nil {
		return InfoResult{Name: name, Manager: mgr, Error: "unsupported manager"}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return InfoResult{Name: name, Manager: mgr, Exists: false}
	}
	version := strings.TrimSpace(string(out))
	// Clean up version string
	if idx := strings.Index(version, "\n"); idx >= 0 {
		version = version[:idx]
	}
	return InfoResult{Name: name, Version: version, Manager: mgr, Exists: true}
}

func runAction(name, action, state string) PkgResult {
	if runtime.GOOS != "linux" {
		return PkgResult{Name: name, Action: action, State: state, Error: "only supported on Linux"}
	}
	if name == "" {
		return PkgResult{Action: action, State: state, Error: "name is required"}
	}
	mgr := detectManager()
	if mgr == "" {
		return PkgResult{Name: name, Action: action, State: state, Error: "no supported package manager found"}
	}

	var cmd *exec.Cmd
	switch mgr {
	case "apt":
		switch action {
		case "install":
			cmd = exec.Command("apt-get", "install", "-y", name)
		case "remove":
			cmd = exec.Command("apt-get", "remove", "-y", name)
		case "update":
			cmd = exec.Command("apt-get", "install", "-y", "--only-upgrade", name)
		}
	case "yum":
		switch action {
		case "install":
			cmd = exec.Command("yum", "install", "-y", name)
		case "remove":
			cmd = exec.Command("yum", "remove", "-y", name)
		case "update":
			cmd = exec.Command("yum", "update", "-y", name)
		}
	case "dnf":
		switch action {
		case "install":
			cmd = exec.Command("dnf", "install", "-y", name)
		case "remove":
			cmd = exec.Command("dnf", "remove", "-y", name)
		case "update":
			cmd = exec.Command("dnf", "upgrade", "-y", name)
		}
	case "pacman":
		switch action {
		case "install":
			cmd = exec.Command("pacman", "-S", "--noconfirm", name)
		case "remove":
			cmd = exec.Command("pacman", "-R", "--noconfirm", name)
		case "update":
			cmd = exec.Command("pacman", "-S", "--noconfirm", name)
		}
	case "zypper":
		switch action {
		case "install":
			cmd = exec.Command("zypper", "install", "-y", name)
		case "remove":
			cmd = exec.Command("zypper", "remove", "-y", name)
		case "update":
			cmd = exec.Command("zypper", "update", "-y", name)
		}
	case "apk":
		switch action {
		case "install":
			cmd = exec.Command("apk", "add", name)
		case "remove":
			cmd = exec.Command("apk", "del", name)
		case "update":
			cmd = exec.Command("apk", "upgrade", name)
		}
	}
	if cmd == nil {
		return PkgResult{Name: name, Action: action, State: state, Manager: mgr, Error: "unsupported action"}
	}

	if _, err := cmd.CombinedOutput(); err != nil {
		return PkgResult{Name: name, Action: action, State: state, Manager: mgr, Error: err.Error()}
	}
	return PkgResult{Changed: true, Name: name, Action: action, State: state, Manager: mgr}
}

func detectManager() string {
	for _, mgr := range []string{"dnf", "yum", "apt-get", "pacman", "zypper", "apk"} {
		if _, err := exec.LookPath(mgr); err == nil {
			switch mgr {
			case "apt-get":
				return "apt"
			default:
				return mgr
			}
		}
	}
	return ""
}
