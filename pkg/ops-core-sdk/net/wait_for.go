package opsnet

import (
	"fmt"
	goNet "net"
	"time"
)

// WaitForResult holds the result of a WaitFor operation.
type WaitForResult struct {
	Success    bool   `json:"success"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// dialTimeout is the per-attempt connection timeout.
const dialTimeout = 2 * time.Second

// retryInterval is the delay between connection attempts.
const retryInterval = 1 * time.Second

// WaitFor tries to connect to host:port until timeout (in seconds) is exceeded.
// Returns success=true if a TCP connection was established within the timeout.
func WaitFor(host string, port int, timeout int) (WaitForResult, error) {
	if host == "" {
		return WaitForResult{}, fmt.Errorf("opsnet: WaitFor host must not be empty")
	}
	if port < 1 || port > 65535 {
		return WaitForResult{}, fmt.Errorf("opsnet: WaitFor port must be between 1 and 65535, got %d", port)
	}
	if timeout < 1 {
		return WaitForResult{}, fmt.Errorf("opsnet: WaitFor timeout must be at least 1 second, got %d", timeout)
	}

	addr := goNet.JoinHostPort(host, fmt.Sprintf("%d", port))
	timeoutDur := time.Duration(timeout) * time.Second
	deadline := time.Now().Add(timeoutDur)
	start := time.Now()
	var lastErr error

	for time.Now().Before(deadline) {
		conn, err := goNet.DialTimeout("tcp", addr, dialTimeout)
		if err == nil {
			conn.Close()
			elapsed := time.Since(start)
			return WaitForResult{
				Success:    true,
				DurationMs: elapsed.Milliseconds(),
			}, nil
		}
		lastErr = err

		// Sleep until next attempt, but don't sleep past the deadline.
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		sleepDur := retryInterval
		if sleepDur > remaining {
			sleepDur = remaining
		}
		time.Sleep(sleepDur)
	}

	elapsed := time.Since(start)
	errMsg := fmt.Sprintf("timeout after %ds waiting for %s", timeout, addr)
	if lastErr != nil {
		errMsg = fmt.Sprintf("timeout after %ds waiting for %s: %v", timeout, addr, lastErr)
	}
	return WaitForResult{
		Success:    false,
		DurationMs: elapsed.Milliseconds(),
		Error:      errMsg,
	}, nil
}
