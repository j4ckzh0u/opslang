// WaitForConnection — Ansible wait_for_connection equivalent.
package opsnet

import (
	"fmt"
	"net"
	"time"
)

// WaitForConnectionResult is returned by WaitForConnection.
type WaitForConnectionResult struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Connected  bool   `json:"connected"`
	DurationMs int64  `json:"duration_ms"`
	Attempts   int    `json:"attempts"`
}

// WaitForConnection waits until a TCP connection to host:port succeeds, or timeout expires.
func WaitForConnection(host string, port int, timeout int) (WaitForConnectionResult, error) {
	result := WaitForConnectionResult{Host: host, Port: port}
	start := time.Now()
	deadline := start.Add(time.Duration(timeout) * time.Second)
	addr := fmt.Sprintf("%s:%d", host, port)

	for time.Now().Before(deadline) {
		result.Attempts++
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			conn.Close()
			result.Connected = true
			result.DurationMs = time.Since(start).Milliseconds()
			return result, nil
		}
		time.Sleep(1 * time.Second)
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result, fmt.Errorf("net.WaitForConnection: timeout after %ds waiting for %s", timeout, addr)
}
