package user

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Binary names are variables so tests can point them at stub scripts and
// verify idempotency without a real Linux user database.
var (
	useraddBin = "useradd"
	usermodBin = "usermod"
	userdelBin = "userdel"
)

// EnsureResult is returned by Ensure and Absent.
type EnsureResult struct {
	Username string `json:"username"`
	// Present is the converged state: true after Ensure, false after Absent.
	Present bool   `json:"present"`
	Changed bool   `json:"changed"`
	Action  string `json:"action"`
	Shell   string `json:"shell,omitempty"`
	Home    string `json:"home,omitempty"`
	UID     int    `json:"uid,omitempty"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Ensure makes a user present with the requested attributes, converging
// drift instead of just creating. It mirrors the Ansible user module with
// state=present:
//
//	missing user        -> useradd (changed=true)
//	exists, shell drift -> usermod -s (changed=true)
//	exists, no drift    -> no command run (changed=false)
//
// opts supports the same keys as Add: shell, home, uid, groups, create_home.
// Only shell and home are converged on an existing user; Ansible behaves the
// same way for a single invocation.
func Ensure(username string, opts map[string]string) (EnsureResult, error) {
	if username == "" {
		return EnsureResult{Action: "ensure", Error: "username is required"}, fmt.Errorf("username is required")
	}

	res := EnsureResult{Username: username, Present: true, Action: "ensure"}

	info, err := Info(username)
	if err == nil {
		// User exists: converge shell/home only when they differ.
		res.Shell = info.User.Shell
		res.Home = info.User.Home
		res.UID = info.User.UID

		var drift map[string]string
		if v := opts["shell"]; v != "" && v != info.User.Shell {
			drift = map[string]string{"shell": v}
		}
		if v := opts["home"]; v != "" && v != info.User.Home {
			if drift == nil {
				drift = map[string]string{}
			}
			drift["home"] = v
		}
		if drift == nil {
			res.Message = "user already up to date"
			return res, nil
		}

		if _, err := Modify(username, drift); err != nil {
			res.Error = err.Error()
			res.Changed = false
			res.Message = "converge failed"
			return res, err
		}
		res.Shell = firstNonEmpty(drift["shell"], res.Shell)
		res.Home = firstNonEmpty(drift["home"], res.Home)
		res.Changed = true
		res.Message = "user attributes converged"
		return res, nil
	}

	addRes, err := Add(username, opts)
	if err != nil {
		res.Error = addRes.Error
		res.Message = "useradd failed"
		res.Changed = false
		return res, err
	}

	after, err := Info(username)
	if err == nil {
		res.Shell = after.User.Shell
		res.Home = after.User.Home
		res.UID = after.User.UID
	}
	res.Changed = true
	res.Message = "user created"
	return res, nil
}

// Absent makes sure a user does not exist. It refuses to remove the root
// account: that is almost never intended and unrecoverable on a remote host.
// An already-missing user is reported with changed=false and no command run.
func Absent(username string, removeHome bool) (EnsureResult, error) {
	if username == "" {
		return EnsureResult{Action: "absent", Error: "username is required"}, fmt.Errorf("username is required")
	}
	if username == "root" {
		return EnsureResult{Username: username, Present: true, Action: "absent", Error: "refusing to remove user root"},
			fmt.Errorf("refusing to remove user root")
	}

	res := EnsureResult{Username: username, Present: false, Action: "absent"}

	exists, _ := Exists(username)
	if !exists.Exists {
		res.Message = "user already absent"
		return res, nil
	}

	args := []string{username}
	if removeHome {
		args = append(args, "-r")
	}
	cmd := exec.Command(userdelBin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		res.Error = string(output)
		res.Message = "userdel failed"
		return res, fmt.Errorf("userdel: %s", output)
	}

	res.Changed = true
	res.Message = "user removed"
	return res, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// groupFile is the group database read by sameNameGroupGID. It is a
// variable so tests can point it at a fixture instead of /etc/group.
var groupFile = "/etc/group"

// sameNameGroupGID reports whether /etc/group contains a group named
// username and returns its GID as a string for useradd -g.
func sameNameGroupGID(username string) (string, bool) {
	f, err := os.Open(groupFile)
	if err != nil {
		return "", false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) >= 3 && fields[0] == username {
			return fields[2], true
		}
	}
	return "", false
}
