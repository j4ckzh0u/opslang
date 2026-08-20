// Package fail provides Ansible fail module equivalent.
package fail

import "fmt"

// FailResult is returned when execution should fail.
type FailResult struct {
	Message string `json:"msg"`
	Failed  bool   `json:"failed"`
}

// Fail returns a failure result with the given message.
func Fail(message string) FailResult {
	if message == "" {
		message = "Task failed as requested"
	}
	return FailResult{
		Message: message,
		Failed:  true,
	}
}

// FailF formats a failure message.
func FailF(format string, args ...interface{}) FailResult {
	return Fail(fmt.Sprintf(format, args...))
}
