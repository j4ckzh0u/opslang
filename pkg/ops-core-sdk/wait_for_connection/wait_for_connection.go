// Package wait_for_connection waits for a host to become reachable.
// Equivalent to ansible.builtin.wait_for_connection module.
package wait_for_connection

import (
	"fmt"
	"net"
	"time"
)

// Result is returned by Wait.
type Result struct {
	Status    string `json:"status"`
	Reachable bool   `json:"reachable"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Attempts  int    `json:"attempts"`
	ElapsedMs int64  `json:"elapsed_ms"`
	Error     string `json:"error,omitempty"`
}

// Wait waits for a host:port to become reachable.
// timeout and delay are in seconds.
func Wait(host string, port, timeout, delay int) Result {
	if host == "" {
		return Result{Status: "failed", Error: "host is required"}
	}
	if port <= 0 {
		port = 22 // default SSH
	}
	if timeout <= 0 {
		timeout = 300 // 5 minutes default
	}
	if delay <= 0 {
		delay = 5 // 5 seconds between attempts
	}

	start := time.Now()
	deadline := start.Add(time.Duration(timeout) * time.Second)
	attempts := 0

	for time.Now().Before(deadline) {
		attempts++
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
		if err == nil {
			conn.Close()
			return Result{
				Status:    "success",
				Reachable: true,
				Host:      host,
				Port:      port,
				Attempts:  attempts,
				ElapsedMs: time.Since(start).Milliseconds(),
			}
		}
		time.Sleep(time.Duration(delay) * time.Second)
	}

	return Result{
		Status:    "failed",
		Reachable: false,
		Host:      host,
		Port:      port,
		Attempts:  attempts,
		ElapsedMs: time.Since(start).Milliseconds(),
		Error:     fmt.Sprintf("timeout after %ds", timeout),
	}
}

// CheckOnce checks if a host:port is reachable once.
func CheckOnce(host string, port int) Result {
	if host == "" {
		return Result{Status: "failed", Error: "host is required"}
	}
	if port <= 0 {
		port = 22
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
	if err != nil {
		return Result{
			Status:    "success",
			Reachable: false,
			Host:      host,
			Port:      port,
			Attempts:  1,
			ElapsedMs: time.Since(start).Milliseconds(),
		}
	}
	conn.Close()

	return Result{
		Status:    "success",
		Reachable: true,
		Host:      host,
		Port:      port,
		Attempts:  1,
		ElapsedMs: time.Since(start).Milliseconds(),
	}
}
