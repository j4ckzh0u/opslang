package service

import (
	"fmt"
	"strings"
)

// EnsureResult is returned by Ensure and EnsureEnabled. Actions lists the
// systemctl verbs that actually executed, so an idempotent re-run reports an
// empty list with changed=false.
type EnsureResult struct {
	Name    string   `json:"name"`
	State   string   `json:"state"`
	Enabled bool     `json:"enabled"`
	Active  bool     `json:"active"`
	Changed bool     `json:"changed"`
	Actions []string `json:"actions"`
	Message string   `json:"message"`
	Error   string   `json:"error,omitempty"`
}

// Ensure converges the run state of a systemd unit, mirroring the Ansible
// service/systemd module:
//
//	state=started   active unit   -> no command (changed=false)
//	state=started   inactive unit -> systemctl start
//	state=stopped   inactive unit -> no command (changed=false)
//	state=stopped   active unit   -> systemctl stop
//	state=restarted any           -> systemctl restart (always changed)
//	state=reloaded  any           -> systemctl reload, falling back to
//	                                restart when the unit does not support
//	                                reload (same fallback Ansible performs)
func Ensure(name, state string) (EnsureResult, error) {
	if name == "" {
		return EnsureResult{State: state, Error: "service name is required"}, fmt.Errorf("service name is required")
	}
	state = strings.ToLower(strings.TrimSpace(state))
	res := EnsureResult{Name: name, State: state}

	switch state {
	case "started", "stopped", "restarted", "reloaded":
	default:
		res.Error = fmt.Sprintf("invalid state %q: must be started|stopped|restarted|reloaded", state)
		return res, fmt.Errorf("service.ensure: %s", res.Error)
	}

	status, err := Status(name)
	if err != nil {
		res.Error = err.Error()
		res.Message = "unable to read unit status"
		return res, err
	}
	res.Active = status.Active
	res.Enabled = status.Enabled

	var action string
	switch state {
	case "started":
		if status.Active {
			res.Message = fmt.Sprintf("service %q already active", name)
			return res, nil
		}
		action = "start"
	case "stopped":
		if !status.Active {
			res.Message = fmt.Sprintf("service %q already inactive", name)
			return res, nil
		}
		action = "stop"
	case "restarted":
		action = "restart"
	case "reloaded":
		action = "reload"
	}

	if _, err := doAction(name, action); err != nil {
		// A unit without a reload handler makes systemctl exit non-zero;
		// fall back to restart, exactly like Ansible's service module.
		if action == "reload" {
			if _, rerr := doAction(name, "restart"); rerr == nil {
				res.Changed = true
				res.Active = true
				res.Actions = []string{"reload", "restart"}
				res.Message = fmt.Sprintf("service %q does not support reload; restarted instead", name)
				return res, nil
			}
		}
		res.Error = err.Error()
		res.Message = fmt.Sprintf("systemctl %s failed", action)
		return res, err
	}

	res.Changed = true
	res.Actions = []string{action}
	if state == "started" || state == "restarted" || state == "reloaded" {
		res.Active = true
	} else {
		res.Active = false
	}
	res.Message = fmt.Sprintf("service %q converged to %s", name, state)
	return res, nil
}

// EnsureEnabled converges boot-time enablement of a systemd unit,
// mirroring the Ansible enabled=yes/no flag: an already-converged unit runs
// no command and reports changed=false.
func EnsureEnabled(name string, enabled bool) (EnsureResult, error) {
	if name == "" {
		return EnsureResult{Error: "service name is required"}, fmt.Errorf("service name is required")
	}

	res := EnsureResult{Name: name, Enabled: enabled}
	res.State = "enabled"
	if !enabled {
		res.State = "disabled"
	}

	status, err := Status(name)
	if err != nil {
		res.Error = err.Error()
		res.Message = "unable to read unit status"
		return res, err
	}
	res.Active = status.Active
	res.Enabled = status.Enabled

	if status.Enabled == enabled {
		res.Message = fmt.Sprintf("service %q already %s", name, res.State)
		return res, nil
	}

	action := "enable"
	if !enabled {
		action = "disable"
	}
	if _, err := doAction(name, action); err != nil {
		res.Error = err.Error()
		res.Message = fmt.Sprintf("systemctl %s failed", action)
		return res, err
	}

	res.Changed = true
	res.Enabled = enabled
	res.Actions = []string{action}
	res.Message = fmt.Sprintf("service %q converged to %s", name, res.State)
	return res, nil
}
