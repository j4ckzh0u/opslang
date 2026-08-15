package sshx

import (
	"net"
)

// IsConnectionError checks if an error is a network/connection error.
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(net.Error); ok {
		return true
	}
	return false
}
