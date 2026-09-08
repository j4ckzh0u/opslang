package securityscan

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const serviceScanPath = "/v1/security/scan"
const maxServiceResponseSize = 32 << 20

// ServiceBackend delegates scanning to a controller-owned scanner service.
type ServiceBackend struct {
	Config *ServiceConfig
	Client *http.Client
}

func (b *ServiceBackend) Name() string { return "scanner-service" }

func (b *ServiceBackend) Capabilities() []Capability {
	return []Capability{CapabilityVulnerability, CapabilitySBOM, CapabilityMisconfig, CapabilitySecret, CapabilityLicense}
}

func (b *ServiceBackend) Scan(ctx context.Context, request ScanRequest) (ScanReport, error) {
	if b == nil {
		return ScanReport{}, fmt.Errorf("scanner service backend is nil")
	}
	if err := ValidateScanRequest(request); err != nil {
		return ScanReport{}, err
	}
	config, err := b.Config.Clone()
	if err != nil {
		return ScanReport{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	config.Timeout = ServiceConfigTimeout(config.Timeout)
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}
	body, err := json.Marshal(request)
	if err != nil {
		return ScanReport{}, fmt.Errorf("encode scan request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(config.URL, "/")+serviceScanPath, bytes.NewReader(body))
	if err != nil {
		return ScanReport{}, fmt.Errorf("create scan request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+config.Token)
	httpRequest.Header.Set("Content-Type", "application/json")
	client := b.Client
	if client == nil {
		client, err = ServiceTLSClient(config)
		if err != nil {
			return ScanReport{}, err
		}
		defer client.CloseIdleConnections()
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := clientCopy.Do(httpRequest)
	if err != nil {
		return ScanReport{}, fmt.Errorf("call scanner service: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return ScanReport{}, fmt.Errorf("scanner service returned HTTP %d", response.StatusCode)
	}
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxServiceResponseSize+1))
	if err != nil {
		return ScanReport{}, fmt.Errorf("read scanner response: %w", err)
	}
	if len(responseBody) > maxServiceResponseSize {
		return ScanReport{}, fmt.Errorf("scanner response exceeds %d bytes", maxServiceResponseSize)
	}
	var report ScanReport
	if err := json.Unmarshal(responseBody, &report); err != nil {
		return ScanReport{}, fmt.Errorf("decode scanner response: %w", err)
	}
	switch report.Status {
	case StatusOK, StatusPartial, StatusUnsupported, StatusFailed:
	default:
		return ScanReport{}, fmt.Errorf("invalid scanner response status %q", report.Status)
	}
	if config.DatabaseSHA256 != "" && !strings.EqualFold(config.DatabaseSHA256, report.DatabaseSHA256) {
		return ScanReport{}, fmt.Errorf("scanner database sha256 mismatch")
	}
	if config.DatabaseVersion != "" && report.DatabaseVersion != config.DatabaseVersion {
		return ScanReport{}, fmt.Errorf("scanner database version mismatch: expected %q, got %q", config.DatabaseVersion, report.DatabaseVersion)
	}
	report.Backend = b.Name()
	report.Target = request.Target
	report.Finish()
	return report, nil
}

// ServiceTLSClient creates an HTTP client with an optional private CA.
func ServiceTLSClient(config *ServiceConfig) (*http.Client, error) {
	if err := ValidateServiceConfig(config); err != nil {
		return nil, err
	}
	if len(config.CA) == 0 {
		return &http.Client{Timeout: config.Timeout}, nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(config.CA) {
		return nil, fmt.Errorf("scanner service CA contains no certificates")
	}
	return &http.Client{Timeout: config.Timeout, Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS13}}}, nil
}

// ServiceConfigTimeout returns a safe default for clients without a timeout.
func ServiceConfigTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return 30 * time.Second
	}
	return timeout
}
