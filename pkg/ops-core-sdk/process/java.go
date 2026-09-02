package process

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// JavaLibrary identifies a JAR referenced by a Java process.
type JavaLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path"`
}

// JavaApp describes a running Java process and its classpath libraries.
type JavaApp struct {
	PID              int32         `json:"pid"`
	User             string        `json:"user,omitempty"`
	Command          string        `json:"command"`
	Executable       string        `json:"executable,omitempty"`
	ContainerRuntime string        `json:"container_runtime,omitempty"`
	ContainerID      string        `json:"container_id,omitempty"`
	Libraries        []JavaLibrary `json:"libraries"`
}

var javaVersionPattern = regexp.MustCompile(`[-_]v?([0-9][0-9A-Za-z.+~-]*)$`)
var containerPattern = regexp.MustCompile(`(?i)(?:/docker/|docker[-/]|containerd[-/]|crio[-/])([a-f0-9]{6,64})(?:\.scope)?$`)

// JavaApps discovers Java processes from /proc. It is read-only and works
// for host processes as well as processes in Docker/containerd cgroups.
func JavaApps() ([]JavaApp, error) {
	if runtime.GOOS != "linux" {
		return []JavaApp{}, nil
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("process.java_apps: read /proc: %w", err)
	}
	apps := make([]JavaApp, 0)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		cmdlineBytes, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "cmdline"))
		if err != nil || len(cmdlineBytes) == 0 {
			continue
		}
		command := strings.ReplaceAll(string(cmdlineBytes), "\x00", " ")
		command = strings.TrimSpace(command)
		if !isJavaCommand(command) {
			continue
		}
		exe, _ := os.Readlink(filepath.Join("/proc", entry.Name(), "exe"))
		user := ""
		if p, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "status")); err == nil {
			user = uidToUser(parseUID(string(p)))
		}
		cgroup, _ := os.ReadFile(filepath.Join("/proc", entry.Name(), "cgroup"))
		runtimeName, containerID := parseContainerCgroup(string(cgroup))
		apps = append(apps, JavaApp{PID: int32(pid), User: user, Command: command, Executable: exe,
			ContainerRuntime: runtimeName, ContainerID: containerID, Libraries: parseJavaLibraries(command)})
	}
	sort.Slice(apps, func(i, j int) bool { return apps[i].PID < apps[j].PID })
	return apps, nil
}

func isJavaCommand(command string) bool {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return false
	}
	base := filepath.Base(fields[0])
	return base == "java" || strings.HasPrefix(base, "java-") || strings.HasSuffix(base, "/java")
}

func parseJavaLibraries(command string) []JavaLibrary {
	seen := map[string]bool{}
	result := make([]JavaLibrary, 0)
	for _, field := range strings.Fields(command) {
		for _, part := range strings.FieldsFunc(field, func(r rune) bool { return r == ':' || r == ';' }) {
			part = strings.Trim(part, "'\"")
			if !strings.HasSuffix(strings.ToLower(part), ".jar") || seen[part] {
				continue
			}
			seen[part] = true
			base := strings.TrimSuffix(filepath.Base(part), filepath.Ext(part))
			name, version := base, ""
			if match := javaVersionPattern.FindStringSubmatch(base); len(match) == 2 {
				version = match[1]
				name = strings.TrimSuffix(base, match[0])
			}
			result = append(result, JavaLibrary{Name: name, Version: version, Path: part})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func parseContainerCgroup(cgroup string) (string, string) {
	for _, line := range strings.Split(cgroup, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if match := containerPattern.FindStringSubmatch(line); len(match) == 2 {
			runtimeName := "docker"
			lower := strings.ToLower(line)
			if strings.Contains(lower, "containerd") {
				runtimeName = "containerd"
			} else if strings.Contains(lower, "crio") {
				runtimeName = "cri-o"
			}
			return runtimeName, match[1]
		}
	}
	return "", ""
}

func parseUID(status string) string {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				return fields[1]
			}
		}
	}
	return ""
}

func uidToUser(uid string) string {
	if uid == "" {
		return ""
	}
	// Avoid a passwd database dependency in the common path; JavaApp keeps
	// the UID-derived value empty when the platform does not expose a name.
	return uid
}
