// Package lsb_release provides OS distribution information gathering.
package lsb_release

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// InfoResult contains OS distribution information.
type InfoResult struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Release     string `json:"release"`
	Codename    string `json:"codename"`
	DistID      string `json:"dist_id"`
	Kernel      string `json:"kernel"`
	Arch        string `json:"arch"`
}

// Get returns OS distribution information.
// It reads from /etc/os-release, lsb_release command, or falls back to runtime info.
func Get() (*InfoResult, error) {
	result := &InfoResult{
		Kernel: runtime.GOOS,
		Arch:   runtime.GOARCH,
	}

	// Try /etc/os-release first (most modern Linux distros)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		parseOSRelease(string(data), result)
		if result.ID != "" {
			return result, nil
		}
	}

	// Try /etc/lsb-release
	if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
		parseLSBRelease(string(data), result)
		if result.ID != "" {
			return result, nil
		}
	}

	// Try lsb_release command
	if out, err := exec.Command("lsb_release", "-a").CombinedOutput(); err == nil {
		parseLSBCommand(string(out), result)
		if result.ID != "" {
			return result, nil
		}
	}

	// Fallback to runtime
	result.ID = runtime.GOOS
	result.Description = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

	return result, nil
}

func parseOSRelease(data string, result *InfoResult) {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], "\"'")
		switch key {
		case "ID":
			result.ID = val
		case "PRETTY_NAME":
			result.Description = val
		case "VERSION_ID":
			result.Release = val
		case "VERSION_CODENAME":
			result.Codename = val
		}
	}
}

func parseLSBRelease(data string, result *InfoResult) {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], "\"'")
		switch key {
		case "DISTRIB_ID":
			result.ID = val
		case "DISTRIB_DESCRIPTION":
			result.Description = val
		case "DISTRIB_RELEASE":
			result.Release = val
		case "DISTRIB_CODENAME":
			result.Codename = val
		}
	}
}

func parseLSBCommand(output string, result *InfoResult) {
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Distributor ID":
			result.ID = val
		case "Description":
			result.Description = val
		case "Release":
			result.Release = val
		case "Codename":
			result.Codename = val
		}
	}
}
