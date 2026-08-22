// Package user provides Unix user management operations.
package user

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// UserInfo represents a system user.
type UserInfo struct {
	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Comment  string `json:"comment"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
}

// ListResult is returned by List.
type ListResult struct {
	Users []UserInfo `json:"users"`
}

// ExistsResult is returned by Exists.
type ExistsResult struct {
	Exists bool `json:"exists"`
}

// InfoResult is returned by Info.
type InfoResult struct {
	User UserInfo `json:"user"`
}

// AddResult is returned by Add.
type AddResult struct {
	Changed  bool   `json:"changed"`
	Username string `json:"username"`
	UID      int    `json:"uid"`
	Error    string `json:"error,omitempty"`
}

// RemoveResult is returned by Remove.
type RemoveResult struct {
	Changed  bool   `json:"changed"`
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

// ModifyResult is returned by Modify.
type ModifyResult struct {
	Changed  bool   `json:"changed"`
	Username string `json:"username"`
	Error    string `json:"error,omitempty"`
}

// passwdFile is the user database read for List/Info/Exists. It is a
// variable so tests can point it at a fixture instead of /etc/passwd.
var passwdFile = "/etc/passwd"

// List returns all system users.
func List() (ListResult, error) {
	f, err := os.Open(passwdFile)
	if err != nil {
		return ListResult{}, fmt.Errorf("open passwd: %w", err)
	}
	defer f.Close()

	var users []UserInfo
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}

		uid, _ := strconv.Atoi(fields[2])
		gid, _ := strconv.Atoi(fields[3])

		users = append(users, UserInfo{
			Username: fields[0],
			UID:      uid,
			GID:      gid,
			Comment:  fields[4],
			Home:     fields[5],
			Shell:    fields[6],
		})
	}

	return ListResult{Users: users}, nil
}

// Exists checks if a user exists.
func Exists(username string) (ExistsResult, error) {
	if username == "" {
		return ExistsResult{}, fmt.Errorf("username is required")
	}

	users, err := List()
	if err != nil {
		return ExistsResult{}, err
	}

	for _, u := range users.Users {
		if u.Username == username {
			return ExistsResult{Exists: true}, nil
		}
	}

	return ExistsResult{Exists: false}, nil
}

// Info returns information about a specific user.
func Info(username string) (InfoResult, error) {
	if username == "" {
		return InfoResult{}, fmt.Errorf("username is required")
	}

	users, err := List()
	if err != nil {
		return InfoResult{}, err
	}

	for _, u := range users.Users {
		if u.Username == username {
			return InfoResult{User: u}, nil
		}
	}

	return InfoResult{}, fmt.Errorf("user %q not found", username)
}

// Add creates a new user. opts supports keys: shell, home, uid, groups, create_home.
func Add(username string, opts map[string]string) (AddResult, error) {
	if username == "" {
		return AddResult{Error: "username is required"}, fmt.Errorf("username is required")
	}

	exists, _ := Exists(username)
	if exists.Exists {
		return AddResult{Changed: false, Username: username}, nil
	}

	args := []string{username}
	// When a group with the same name already exists, useradd's default
	// "create a matching primary group" fails. Bind the user to the
	// existing group instead — the same thing the Ansible user module
	// does when group=<name> pre-exists.
	if gid, ok := sameNameGroupGID(username); ok {
		args = append(args, "-g", gid)
	}
	if v, ok := opts["shell"]; ok && v != "" {
		args = append(args, "-s", v)
	}
	if v, ok := opts["home"]; ok && v != "" {
		args = append(args, "-d", v)
	}
	if v, ok := opts["uid"]; ok && v != "" {
		args = append(args, "-u", v)
	}
	if v, ok := opts["groups"]; ok && v != "" {
		args = append(args, "-G", v)
	}
	if v, ok := opts["create_home"]; ok && (v == "true" || v == "1") {
		args = append(args, "-m")
	}

	cmd := exec.Command(useraddBin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return AddResult{Error: string(output)}, fmt.Errorf("useradd: %s", output)
	}

	// useradd succeeded: the user exists even if the read-back below
	// fails, so Changed must be true regardless.
	result := AddResult{Changed: true, Username: username}
	if info, err := Info(username); err == nil {
		result.UID = info.User.UID
	}
	return result, nil
}

// Remove deletes a user.
func Remove(username string, removeHome bool) (RemoveResult, error) {
	if username == "" {
		return RemoveResult{Error: "username is required"}, fmt.Errorf("username is required")
	}

	exists, _ := Exists(username)
	if !exists.Exists {
		return RemoveResult{Changed: false, Username: username}, nil
	}

	args := []string{username}
	if removeHome {
		args = append(args, "-r")
	}

	cmd := exec.Command(userdelBin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return RemoveResult{Error: string(output)}, fmt.Errorf("userdel: %s", output)
	}

	return RemoveResult{Changed: true, Username: username}, nil
}

// Modify changes user properties. opts supports keys: shell, home, uid.
func Modify(username string, opts map[string]string) (ModifyResult, error) {
	if username == "" {
		return ModifyResult{Error: "username is required"}, fmt.Errorf("username is required")
	}

	exists, _ := Exists(username)
	if !exists.Exists {
		return ModifyResult{Error: "user not found"}, fmt.Errorf("user %q not found", username)
	}

	var args []string
	if v, ok := opts["shell"]; ok && v != "" {
		args = append(args, "-s", v)
	}
	if v, ok := opts["home"]; ok && v != "" {
		args = append(args, "-d", v)
	}
	if v, ok := opts["uid"]; ok && v != "" {
		args = append(args, "-u", v)
	}

	if len(args) == 0 {
		return ModifyResult{Changed: false, Username: username}, nil
	}

	args = append(args, username)
	cmd := exec.Command(usermodBin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return ModifyResult{Error: string(output)}, fmt.Errorf("usermod: %s", output)
	}

	return ModifyResult{Changed: true, Username: username}, nil
}
