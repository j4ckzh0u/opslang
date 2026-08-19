// Package meta provides meta operations for task control.
// Equivalent to ansible.builtin.meta module.
package meta

import "fmt"

// Result is returned by all functions.
type Result struct {
	Status  string `json:"status"`
	Action  string `json:"action"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// EndHost signals that the current host should be skipped for remaining tasks.
func EndHost() Result {
	return Result{
		Status:  "success",
		Action:  "end_host",
		Message: "host marked to be skipped for remaining tasks",
	}
}

// EndPlay signals that the entire play should be ended.
func EndPlay() Result {
	return Result{
		Status:  "success",
		Action:  "end_play",
		Message: "play marked to end",
	}
}

// ClearHostErrors clears all previously failed tasks for the current host.
func ClearHostErrors() Result {
	return Result{
		Status:  "success",
		Action:  "clear_host_errors",
		Message: "host errors cleared",
	}
}

// RefreshInventory refreshes the inventory.
func RefreshInventory() Result {
	return Result{
		Status:  "success",
		Action:  "refresh_inventory",
		Message: "inventory refresh requested",
	}
}

// FlushHandlers immediately triggers all pending handlers.
func FlushHandlers() Result {
	return Result{
		Status:  "success",
		Action:  "flush_handlers",
		Message: "handlers flushed",
	}
}

// ResetConnection resets the persistent connection.
func ResetConnection() Result {
	return Result{
		Status:  "success",
		Action:  "reset_connection",
		Message: "connection reset requested",
	}
}

// Noop performs no operation (useful for conditional tasks).
func Noop() Result {
	return Result{
		Status:  "success",
		Action:  "noop",
		Message: "no operation",
	}
}

// Fail fails the task with a custom message.
func Fail(message string) Result {
	if message == "" {
		message = "task failed"
	}
	return Result{
		Status:  "failed",
		Action:  "fail",
		Message: message,
		Error:   message,
	}
}

// Assert checks a condition and fails if false.
func Assert(condition bool, message string) Result {
	if !condition {
		if message == "" {
			message = "assertion failed"
		}
		return Result{
			Status:  "failed",
			Action:  "assert",
			Message: message,
			Error:   fmt.Sprintf("assertion failed: %s", message),
		}
	}
	return Result{
		Status:  "success",
		Action:  "assert",
		Message: "assertion passed",
	}
}

// Debug outputs a debug message.
func Debug(message string, vars map[string]interface{}) Result {
	return Result{
		Status:  "success",
		Action:  "debug",
		Message: message,
	}
}
