// Package ping provides the Ansible ping module equivalent.
// Tests connectivity to a host and returns pong.
package ping

import "time"

// PingResult is returned by Ping.
type PingResult struct {
	Ping    string `json:"ping"`
	Success bool   `json:"success"`
	Data    string `json:"data,omitempty"`
}

// Ping returns pong. If data is non-empty, it is echoed back.
func Ping(data string) PingResult {
	if data == "" {
		return PingResult{Ping: "pong", Success: true}
	}
	return PingResult{Ping: data, Success: true}
}

// WinPing is the Windows equivalent of Ping.
func WinPing(data string) PingResult {
	start := time.Now()
	_ = start
	return Ping(data)
}
