// Package user provides structured Linux user management operations for OpsLang.
// All functions return strongly-typed structs with JSON serialization support.
// Mutating operations (Add, Remove, Modify) call useradd/userdel/usermod directly
// via exec.Command — never through a shell wrapper.
package user

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

// UserInfo represents a system user account with full metadata.
type UserInfo struct {
	UID      int      `json:"uid"`
	GID      int      `json:"gid"`
	Username string   `json:"username"`
	Home     string   `json:"home"`
	Shell    string   `json:"shell"`
	Groups   []string `json:"groups"`
	System   bool     `json:"system"`
}

// AddResult is returned by Add, reporting whether a new user was created.
type AddResult struct {
	Changed bool   `json:"changed"`
	UID     int    `json:"uid"`
	Error   string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove, reporting whether a user was deleted.
type RemoveResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ModifyResult is returned by Modify, reporting whether a user was changed.
type ModifyResult struct {
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ExistsResult is returned by Exists, reporting whether a user account is present.
type ExistsResult struct {
	Exists bool `json:"exists"`
}

// Info returns detailed information about the user identified by username.
// It uses os/user.Lookup for the base record and resolves supplementary groups.
func Info(username string) (UserInfo, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return UserInfo{}, fmt.Errorf("user %q not found: %w", username, err)
	}

	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)

	groups, _ := u.GroupIds()
	groupNames := make([]string, 0, len(groups))
	for _, gidStr := range groups {
		g, lookupErr := user.LookupGroupId(gidStr)
		if lookupErr == nil {
			groupNames = append(groupNames, g.Name)
		}
	}

	system := uid < 1000
	if username == "root" {
		system = true
	}

	return UserInfo{
		UID:      uid,
		GID:      gid,
		Username: u.Username,
		Home:     u.HomeDir,
		Shell:    readShell(username),
		Groups:   groupNames,
		System:   system,
	}, nil
}

// List reads /etc/passwd and returns a UserInfo entry for every line.
func List() ([]UserInfo, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("open /etc/passwd: %w", err)
	}
	defer f.Close()

	var users []UserInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}
		uid, _ := strconv.Atoi(fields[2])
		gid, _ := strconv.Atoi(fields[3])

		system := uid < 1000
		if fields[0] == "root" {
			system = true
		}

		users = append(users, UserInfo{
			UID:      uid,
			GID:      gid,
			Username: fields[0],
			Home:     fields[5],
			Shell:    fields[6],
			Groups:   []string{},
			System:   system,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading /etc/passwd: %w", err)
	}
	return users, nil
}

// Add creates a new user account using useradd.
// Recognised option keys:
//   - "shell"        — login shell (default /bin/sh)
//   - "home"         — home directory
//   - "groups"       — comma-separated supplementary groups
//   - "system"       — "true" to create a system account
//   - "create_home"  — "true" (default) or "false" to skip home creation
func Add(username string, opts map[string]string) (AddResult, error) {
	if opts == nil {
		opts = make(map[string]string)
	}

	args := []string{}

	if shell, ok := opts["shell"]; ok && shell != "" {
		args = append(args, "-s", shell)
	}

	if home, ok := opts["home"]; ok && home != "" {
		args = append(args, "-d", home)
	}

	if groups, ok := opts["groups"]; ok && groups != "" {
		args = append(args, "-G", groups)
	}

	if opts["system"] == "true" {
		args = append(args, "-r")
	}

	createHome := opts["create_home"]
	if createHome == "" {
		createHome = "true"
	}
	if createHome == "true" && opts["system"] != "true" {
		args = append(args, "-m")
	}

	args = append(args, username)

	cmd := exec.Command("useradd", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return AddResult{
			Changed: false,
			Error:   fmt.Sprintf("useradd failed: %s: %s", err, strings.TrimSpace(string(out))),
		}, fmt.Errorf("useradd %s: %w", username, err)
	}

	// Read back the UID of the newly created user.
	u, lookupErr := user.Lookup(username)
	uid := 0
	if lookupErr == nil {
		uid, _ = strconv.Atoi(u.Uid)
	}

	return AddResult{
		Changed: true,
		UID:     uid,
	}, nil
}

// Remove deletes a user account using userdel.
// If removeHome is true the user's home directory and mail spool are also removed.
func Remove(username string, removeHome bool) (RemoveResult, error) {
	// Check existence first so we can report Changed=false for missing users.
	_, err := user.Lookup(username)
	if err != nil {
		return RemoveResult{Changed: false}, nil
	}

	args := []string{}
	if removeHome {
		args = append(args, "-r")
	}
	args = append(args, username)

	cmd := exec.Command("userdel", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		return RemoveResult{
			Changed: false,
			Error:   fmt.Sprintf("userdel failed: %s: %s", cmdErr, strings.TrimSpace(string(out))),
		}, fmt.Errorf("userdel %s: %w", username, cmdErr)
	}

	return RemoveResult{Changed: true}, nil
}

// Modify updates an existing user account using usermod.
// Recognised option keys:
//   - "shell"        — new login shell
//   - "home"         — new home directory
//   - "groups"       — new comma-separated supplementary group list (replaces existing)
//   - "login"        — new login name (rename user)
//   - "comment"      — GECOS comment
func Modify(username string, opts map[string]string) (ModifyResult, error) {
	// Check existence first.
	if _, err := user.Lookup(username); err != nil {
		return ModifyResult{Changed: false, Error: fmt.Sprintf("user %q not found", username)},
			fmt.Errorf("user %q not found: %w", username, err)
	}

	args := []string{}
	if v, ok := opts["shell"]; ok && v != "" {
		args = append(args, "-s", v)
	}
	if v, ok := opts["home"]; ok && v != "" {
		args = append(args, "-d", v)
	}
	if v, ok := opts["groups"]; ok && v != "" {
		args = append(args, "-G", v)
	}
	if v, ok := opts["login"]; ok && v != "" {
		args = append(args, "-l", v)
	}
	if v, ok := opts["comment"]; ok && v != "" {
		args = append(args, "-c", v)
	}

	if len(args) == 0 {
		return ModifyResult{Changed: false}, nil
	}

	args = append(args, username)

	cmd := exec.Command("usermod", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ModifyResult{
			Changed: false,
			Error:   fmt.Sprintf("usermod failed: %s: %s", err, strings.TrimSpace(string(out))),
		}, fmt.Errorf("usermod %s: %w", username, err)
	}

	return ModifyResult{Changed: true}, nil
}

// Exists reports whether the named user account exists on the system.
func Exists(username string) (ExistsResult, error) {
	_, err := user.Lookup(username)
	if err != nil {
		// Treat any lookup error as "does not exist" without returning an error,
		// so callers can use the bool field directly.
		return ExistsResult{Exists: false}, nil
	}
	return ExistsResult{Exists: true}, nil
}

// readShell tries to read the user's shell from /etc/passwd.
// Falls back to an empty string when the file cannot be read or the user is absent.
func readShell(username string) string {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) >= 7 && fields[0] == username {
			return fields[6]
		}
	}
	return ""
}
