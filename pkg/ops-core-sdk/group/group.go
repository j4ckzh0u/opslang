// Package group provides structured Linux group management operations for OpsLang.
// All functions return strongly-typed structs with JSON serialization support.
// Mutating operations (Add, Remove) call groupadd/groupdel directly via
// exec.Command — never through a shell wrapper.
package group

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// groupFile is the path to the group database file.
// It is a variable so tests can override it.
var groupFile = "/etc/group"

// groupaddBin is the path to the groupadd binary.
var groupaddBin = "groupadd"

// groupdelBin is the path to the groupdel binary.
var groupdelBin = "groupdel"

// GroupInfo represents a system group with its members.
type GroupInfo struct {
	GID     int      `json:"gid"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// AddResult is returned by Add, reporting whether a new group was created.
type AddResult struct {
	Changed bool   `json:"changed"`
	GID     int    `json:"gid"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove, reporting whether a group was deleted.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ExistsResult is returned by Exists, reporting whether a group is present.
type ExistsResult struct {
	Exists bool `json:"exists"`
}

// Info returns information about the group identified by name.
// It parses /etc/group for the matching entry.
func Info(name string) (GroupInfo, error) {
	f, err := os.Open(groupFile)
	if err != nil {
		return GroupInfo{}, fmt.Errorf("open %s: %w", groupFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}

		if fields[0] == name {
			gid, _ := strconv.Atoi(fields[2])
			members := parseMembers(fields[3])
			return GroupInfo{
				GID:     gid,
				Name:    fields[0],
				Members: members,
			}, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return GroupInfo{}, fmt.Errorf("reading %s: %w", groupFile, err)
	}

	return GroupInfo{}, fmt.Errorf("group %q not found", name)
}

// List reads /etc/group and returns a GroupInfo entry for every line.
func List() ([]GroupInfo, error) {
	f, err := os.Open(groupFile)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", groupFile, err)
	}
	defer f.Close()

	var groups []GroupInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}

		gid, _ := strconv.Atoi(fields[2])
		members := parseMembers(fields[3])

		groups = append(groups, GroupInfo{
			GID:     gid,
			Name:    fields[0],
			Members: members,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", groupFile, err)
	}
	return groups, nil
}

// Add creates a new group using groupadd.
// Recognised option keys:
//   - "gid"    — numeric GID for the new group
//   - "system" — "true" to create a system group
func Add(name string, opts map[string]string) (AddResult, error) {
	if name == "" {
		return AddResult{Error: "group name is required"}, fmt.Errorf("group name is required")
	}
	if opts == nil {
		opts = make(map[string]string)
	}

	// Idempotent like the Ansible group module: an existing group is left
	// untouched and reported with changed=false.
	if info, err := Info(name); err == nil {
		return AddResult{Changed: false, GID: info.GID}, nil
	}

	args := []string{}

	if gid, ok := opts["gid"]; ok && gid != "" {
		args = append(args, "-g", gid)
	}

	if opts["system"] == "true" {
		args = append(args, "-r")
	}

	args = append(args, name)

	cmd := exec.Command(groupaddBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return AddResult{
			Changed: false,
			Error:   fmt.Sprintf("groupadd failed: %s: %s", err, strings.TrimSpace(string(out))),
		}, fmt.Errorf("groupadd %s: %w", name, err)
	}

	// Read back the GID of the newly created group.
	gid := 0
	info, lookupErr := Info(name)
	if lookupErr == nil {
		gid = info.GID
	}

	return AddResult{
		Changed: true,
		GID:     gid,
	}, nil
}

// Remove deletes a group using groupdel.
func Remove(name string) (RemoveResult, error) {
	// Check existence first so we can report Changed=false for missing groups.
	if _, err := Info(name); err != nil {
		return RemoveResult{Changed: false}, nil
	}

	cmd := exec.Command(groupdelBin, name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return RemoveResult{
			Changed: false,
			Error:   fmt.Sprintf("groupdel failed: %s: %s", err, strings.TrimSpace(string(out))),
		}, fmt.Errorf("groupdel %s: %w", name, err)
	}

	return RemoveResult{Changed: true}, nil
}

// Exists reports whether the named group exists on the system.
// It parses /etc/group and does not return an error for a missing group;
// the boolean field indicates presence.
func Exists(name string) (ExistsResult, error) {
	f, err := os.Open(groupFile)
	if err != nil {
		return ExistsResult{}, fmt.Errorf("open %s: %w", groupFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) >= 1 && fields[0] == name {
			return ExistsResult{Exists: true}, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return ExistsResult{}, fmt.Errorf("reading %s: %w", groupFile, err)
	}

	return ExistsResult{Exists: false}, nil
}

// parseMembers splits the comma-separated member list from /etc/group.
// An empty field yields an empty slice (not a slice with one empty string).
func parseMembers(field string) []string {
	field = strings.TrimSpace(field)
	if field == "" {
		return []string{}
	}
	parts := strings.Split(field, ",")
	members := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			members = append(members, p)
		}
	}
	return members
}
