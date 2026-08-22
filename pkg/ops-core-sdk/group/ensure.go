package group

import (
	"fmt"
)

// EnsureResult is returned by Ensure and Absent.
type EnsureResult struct {
	Name string `json:"name"`
	// Present is the converged state: true after Ensure, false after Absent.
	Present bool   `json:"present"`
	Changed bool   `json:"changed"`
	Action  string `json:"action"`
	GID     int    `json:"gid,omitempty"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Ensure makes a group present, mirroring the Ansible group module with
// state=present. An existing group is reported with changed=false and no
// groupadd invocation; a missing group is created with the requested opts
// (gid, system). Renumbering the GID of an existing group is not supported —
// groupmod -g is almost never what an operator wants for drift control.
func Ensure(name string, opts map[string]string) (EnsureResult, error) {
	if name == "" {
		return EnsureResult{Action: "ensure", Error: "group name is required"}, fmt.Errorf("group name is required")
	}

	res := EnsureResult{Name: name, Present: true, Action: "ensure"}

	info, err := Info(name)
	if err == nil {
		res.GID = info.GID
		res.Message = "group already exists"
		return res, nil
	}

	addRes, err := Add(name, opts)
	if err != nil {
		res.Error = addRes.Error
		res.Message = "groupadd failed"
		return res, err
	}

	res.GID = addRes.GID
	res.Changed = true
	res.Message = "group created"
	return res, nil
}

// Absent makes sure a group does not exist, mirroring the Ansible group
// module with state=absent. An already-missing group reports changed=false
// with zero commands executed.
func Absent(name string) (EnsureResult, error) {
	if name == "" {
		return EnsureResult{Action: "absent", Error: "group name is required"}, fmt.Errorf("group name is required")
	}

	res := EnsureResult{Name: name, Present: false, Action: "absent"}

	if _, err := Info(name); err != nil {
		res.Message = "group already absent"
		return res, nil
	}

	if _, err := Remove(name); err != nil {
		res.Error = err.Error()
		res.Message = "groupdel failed"
		return res, err
	}

	res.Changed = true
	res.Message = "group removed"
	return res, nil
}
