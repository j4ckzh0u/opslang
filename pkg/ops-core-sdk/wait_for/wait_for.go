// Package wait_for provides wait/polling operations for ports, files, and URLs.
package wait_for

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// Result represents the result of a wait operation.
type Result struct {
	Success    bool   `json:"success"`
	Target     string `json:"target"`
	ElapsedMs  int64  `json:"elapsed_ms"`
	Message    string `json:"message"`
}

// Port waits until a TCP port is open or times out.
func Port(host string, port int, timeoutMs int) (*Result, error) {
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	target := fmt.Sprintf("%s:%d", host, port)
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	start := time.Now()

	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		dialTimeout := 2 * time.Second
		if dialTimeout > remaining {
			dialTimeout = remaining
		}
		conn, err := net.DialTimeout("tcp", target, dialTimeout)
		if err == nil {
			conn.Close()
			return &Result{
				Success:   true,
				Target:    target,
				ElapsedMs: time.Since(start).Milliseconds(),
				Message:   fmt.Sprintf("Port %s is open", target),
			}, nil
		}
		sleepDur := 500 * time.Millisecond
		if sleepDur > time.Until(deadline) {
			break
		}
		time.Sleep(sleepDur)
	}

	return &Result{
		Success:   false,
		Target:    target,
		ElapsedMs: time.Since(start).Milliseconds(),
		Message:   fmt.Sprintf("Timeout waiting for %s", target),
	}, nil
}

// File waits until a file exists or times out.
func File(path string, timeoutMs int) (*Result, error) {
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	start := time.Now()

	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return &Result{
				Success:   true,
				Target:    path,
				ElapsedMs: time.Since(start).Milliseconds(),
				Message:   fmt.Sprintf("File %s exists", path),
			}, nil
		}
		sleepDur := 500 * time.Millisecond
		if sleepDur > time.Until(deadline) {
			break
		}
		time.Sleep(sleepDur)
	}

	return &Result{
		Success:   false,
		Target:    path,
		ElapsedMs: time.Since(start).Milliseconds(),
		Message:   fmt.Sprintf("Timeout waiting for %s", path),
	}, nil
}

// URL waits until an HTTP URL returns a 2xx status or times out.
func URL(url string, timeoutMs int) (*Result, error) {
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	start := time.Now()

	for time.Now().Before(deadline) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		httpTimeout := 3 * time.Second
		if httpTimeout > remaining {
			httpTimeout = remaining
		}
		client := &http.Client{Timeout: httpTimeout}
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return &Result{
					Success:   true,
					Target:    url,
					ElapsedMs: time.Since(start).Milliseconds(),
					Message:   fmt.Sprintf("URL %s returned %d", url, resp.StatusCode),
				}, nil
			}
		}
		sleepDur := 1 * time.Second
		if sleepDur > time.Until(deadline) {
			break
		}
		time.Sleep(sleepDur)
	}

	return &Result{
		Success:   false,
		Target:    url,
		ElapsedMs: time.Since(start).Milliseconds(),
		Message:   fmt.Sprintf("Timeout waiting for %s", url),
	}, nil
}
