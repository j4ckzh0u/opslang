// Package assert provides Ansible assert module equivalent.
package assert

import "fmt"

// AssertResult is returned by Assert.
type AssertResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"msg"`
	Success bool   `json:"success"`
	Failed  bool   `json:"failed"`
}

// Assert checks condition; if false, returns failed result.
func Assert(condition bool, successMsg, failMsg string) AssertResult {
	if condition {
		if successMsg == "" {
			successMsg = "Assertion passed"
		}
		return AssertResult{
			Changed: false,
			Message: successMsg,
			Success: true,
			Failed:  false,
		}
	}
	if failMsg == "" {
		failMsg = "Assertion failed"
	}
	return AssertResult{
		Changed: false,
		Message: failMsg,
		Success: false,
		Failed:  true,
	}
}

// AssertEqual compares two values for equality.
func AssertEqual[T comparable](a, b T, msg string) AssertResult {
	if msg == "" {
		msg = fmt.Sprintf("assert %v == %v", a, b)
	}
	return Assert(a == b, msg, msg)
}
