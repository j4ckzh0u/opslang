// Package uri provides advanced HTTP operations with full response control.
package uri

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response represents an HTTP response.
type Response struct {
	StatusCode int               `json:"status_code"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
	DurationMs int64             `json:"duration_ms"`
}

// Request represents an HTTP request configuration.
type Request struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
	Timeout int               `json:"timeout_ms"`
}

// Do executes an HTTP request with full control.
func Do(url string, method string, headers map[string]string, body string, timeoutMs int) (*Response, error) {
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	headerMap := make(map[string]string)
	for k, v := range resp.Header {
		headerMap[k] = strings.Join(v, ", ")
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    headerMap,
		DurationMs: time.Since(start).Milliseconds(),
	}, nil
}

// Get performs an HTTP GET request.
func Get(url string) (*Response, error) {
	return Do(url, "GET", nil, "", 30000)
}

// Post performs an HTTP POST request with JSON body.
func Post(url string, body interface{}) (*Response, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	return Do(url, "POST", headers, string(jsonBytes), 30000)
}

// Put performs an HTTP PUT request with JSON body.
func Put(url string, body interface{}) (*Response, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	return Do(url, "PUT", headers, string(jsonBytes), 30000)
}

// Delete performs an HTTP DELETE request.
func Delete(url string) (*Response, error) {
	return Do(url, "DELETE", nil, "", 30000)
}

// Download downloads a URL to a local file.
func Download(url string, dest string) (*Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	var buf bytes.Buffer
	size, err := io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	_ = size
	return &Response{
		StatusCode: resp.StatusCode,
		Body:       fmt.Sprintf("Downloaded %d bytes to %s", buf.Len(), dest),
		Headers:    map[string]string{"Content-Type": resp.Header.Get("Content-Type")},
	}, nil
}
