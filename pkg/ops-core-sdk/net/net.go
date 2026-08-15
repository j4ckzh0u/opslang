// Package opsnet provides pure-Go network and HTTP operations for OpsLang.
// All functions return strongly-typed structs and use only Go stdlib — no shell calls.
package opsnet

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// HTTPResponse holds the result of an HTTP request.
type HTTPResponse struct {
	StatusCode    int               `json:"status_code"`
	Status        string            `json:"status"`
	Body          string            `json:"body"`
	Headers       map[string]string `json:"headers"`
	ContentLength int64             `json:"content_length"`
}

// TCPResult holds the result of a TCP connectivity check.
type TCPResult struct {
	Host      string  `json:"host"`
	Port      int     `json:"port"`
	Connected bool    `json:"connected"`
	LatencyMs float64 `json:"latency_ms"`
}

// DNSResult holds the result of a DNS lookup.
type DNSResult struct {
	Domain    string   `json:"domain"`
	Addresses []string `json:"addresses"`
	CNAME     string   `json:"cname"`
}

// InterfaceInfo holds information about a single network interface.
type InterfaceInfo struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardware_addr"`
	MTU          int      `json:"mtu"`
	Up           bool     `json:"up"`
	Addresses    []string `json:"addresses"`
}

// defaultHTTPTimeout is the timeout used for HTTP requests.
const defaultHTTPTimeout = 30 * time.Second

// defaultTCPTimeout is the timeout used for TCP connectivity checks.
const defaultTCPTimeout = 5 * time.Second

// newDefaultClient creates an http.Client with sensible defaults.
func newDefaultClient() *http.Client {
	return &http.Client{
		Timeout: defaultHTTPTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
}

// HTTPGet performs an HTTP GET request and returns a structured response.
// The client uses a 30-second timeout.
func HTTPGet(url string) (HTTPResponse, error) {
	if url == "" {
		return HTTPResponse{}, fmt.Errorf("opsnet: HTTPGet url must not be empty")
	}

	client := newDefaultClient()
	resp, err := client.Get(url)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("opsnet: HTTPGet %q failed: %w", url, err)
	}
	defer resp.Body.Close()

	return buildHTTPResponse(resp)
}

// HTTPPost performs an HTTP POST request with a JSON body and returns a structured response.
// The client uses a 30-second timeout. Content-Type is set to application/json.
func HTTPPost(url string, body string) (HTTPResponse, error) {
	if url == "" {
		return HTTPResponse{}, fmt.Errorf("opsnet: HTTPPost url must not be empty")
	}

	client := newDefaultClient()
	resp, err := client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("opsnet: HTTPPost %q failed: %w", url, err)
	}
	defer resp.Body.Close()

	return buildHTTPResponse(resp)
}

// buildHTTPResponse reads the response and builds an HTTPResponse struct.
func buildHTTPResponse(resp *http.Response) (HTTPResponse, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{}, fmt.Errorf("opsnet: failed to read response body: %w", err)
	}

	headers := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ", ")
	}

	return HTTPResponse{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Body:          string(bodyBytes),
		Headers:       headers,
		ContentLength: resp.ContentLength,
	}, nil
}

// TCPConnect checks TCP connectivity to host:port and measures latency.
// Uses a 5-second dial timeout.
func TCPConnect(host string, port int) (TCPResult, error) {
	if host == "" {
		return TCPResult{}, fmt.Errorf("opsnet: TCPConnect host must not be empty")
	}
	if port < 1 || port > 65535 {
		return TCPResult{}, fmt.Errorf("opsnet: TCPConnect port must be between 1 and 65535, got %d", port)
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, defaultTCPTimeout)
	latency := time.Since(start)

	result := TCPResult{
		Host:      host,
		Port:      port,
		LatencyMs: float64(latency.Microseconds()) / 1000.0,
	}

	if err != nil {
		result.Connected = false
		// Return the result with Connected=false rather than an error,
		// so callers can use the struct directly.
		return result, nil
	}
	defer conn.Close()

	result.Connected = true
	return result, nil
}

// DNSLookup performs DNS resolution for the given domain.
// Returns both A/AAAA addresses and the CNAME record.
func DNSLookup(domain string) (DNSResult, error) {
	if domain == "" {
		return DNSResult{}, fmt.Errorf("opsnet: DNSLookup domain must not be empty")
	}

	result := DNSResult{
		Domain:    domain,
		Addresses: []string{},
		CNAME:     "",
	}

	addrs, err := net.LookupHost(domain)
	if err != nil {
		return result, fmt.Errorf("opsnet: DNSLookup host %q failed: %w", domain, err)
	}
	result.Addresses = addrs

	cname, err := net.LookupCNAME(domain)
	if err == nil && cname != "" && cname != domain+"." {
		result.CNAME = strings.TrimSuffix(cname, ".")
	}

	return result, nil
}

// Interfaces returns information about all network interfaces on the system.
func Interfaces() ([]InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("opsnet: failed to list interfaces: %w", err)
	}

	result := make([]InterfaceInfo, 0, len(ifaces))
	for _, iface := range ifaces {
		info := InterfaceInfo{
			Name: iface.Name,
			MTU:  iface.MTU,
			Up:   iface.Flags&net.FlagUp != 0,
		}

		if iface.HardwareAddr != nil {
			info.HardwareAddr = iface.HardwareAddr.String()
		}

		addrs, err := iface.Addrs()
		if err == nil {
			info.Addresses = make([]string, 0, len(addrs))
			for _, addr := range addrs {
				info.Addresses = append(info.Addresses, addr.String())
			}
		} else {
			info.Addresses = []string{}
		}

		result = append(result, info)
	}

	return result, nil
}
