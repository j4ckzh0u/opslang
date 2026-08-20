// Package debug provides Ansible debug module equivalent.
package debug

import "fmt"

// DebugResult is returned by Debug.
type DebugResult struct {
	Message string `json:"msg"`
	Changed bool   `json:"changed"`
	Var     string `json:"var,omitempty"`
}

// Debug outputs a message.
func Debug(msg string) DebugResult {
	return DebugResult{Message: msg, Changed: false}
}

// DebugVar outputs a named variable value.
func DebugVar(name, value string) DebugResult {
	return DebugResult{
		Message: fmt.Sprintf("%s = %s", name, value),
		Changed: false,
		Var:     name,
	}
}
