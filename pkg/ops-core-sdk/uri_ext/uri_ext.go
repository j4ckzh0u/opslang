// Package uri_ext provides extended URI operations (PATCH, DELETE, HEAD, OPTIONS).
// Extends the basic uri package with additional HTTP methods.
package uri_ext

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Result is returned by all HTTP operations.
type Result struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	ElapsedMs  int64             `json:"elapsed_ms"`
	Error      string            `json:"error,omitempty"`
}

func doRequest(method, url string, body []byte, headers map[string]string, timeout int) Result {
	if url == "" {
		return Result{Status: "failed", Error: "url is required"}
	}
	if timeout <= 0 {
		timeout = 30
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return Result{Status: "failed", Method: method, URL: url,
			Error: fmt.Sprintf("create request: %v", err)}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return Result{Status: "failed", Method: method, URL: url, ElapsedMs: elapsed,
			Error: fmt.Sprintf("request: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{Status: "failed", Method: method, URL: url, StatusCode: resp.StatusCode,
			ElapsedMs: elapsed, Error: fmt.Sprintf("read body: %v", err)}
	}

	// Collect headers
	hdrs := make(map[string]string)
	for k := range resp.Header {
		hdrs[k] = resp.Header.Get(k)
	}

	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
	}

	return Result{
		Status:     status,
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    hdrs,
		Method:     method,
		URL:        url,
		ElapsedMs:  elapsed,
	}
}

// Patch performs an HTTP PATCH request.
func Patch(url string, body []byte, headers map[string]string, timeout int) Result {
	return doRequest("PATCH", url, body, headers, timeout)
}

// Delete performs an HTTP DELETE request.
func Delete(url string, headers map[string]string, timeout int) Result {
	return doRequest("DELETE", url, nil, headers, timeout)
}

// Head performs an HTTP HEAD request (no body returned).
func Head(url string, headers map[string]string, timeout int) Result {
	return doRequest("HEAD", url, nil, headers, timeout)
}

// Options performs an HTTP OPTIONS request.
func Options(url string, headers map[string]string, timeout int) Result {
	return doRequest("OPTIONS", url, nil, headers, timeout)
}
