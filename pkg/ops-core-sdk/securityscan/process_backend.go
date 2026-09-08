package securityscan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ProcessBackend executes an externally managed scanner binary. The binary
// and its database are owned by the caller, keeping this SDK vendor-neutral.
type ProcessBackend struct {
	BinaryPath string
	Args       func(ScanRequest) ([]string, error)
	MaxOutput  int
}

func (b *ProcessBackend) Name() string { return "external-process" }

func (b *ProcessBackend) Capabilities() []Capability {
	return []Capability{CapabilityVulnerability, CapabilitySBOM, CapabilityMisconfig, CapabilitySecret, CapabilityLicense}
}

func (b *ProcessBackend) Scan(ctx context.Context, request ScanRequest) (ScanReport, error) {
	started := time.Now().UTC()
	report := ScanReport{Backend: b.Name(), Target: request.Target, Status: StatusFailed, StartedAt: started}
	if err := ValidateScanRequest(request); err != nil {
		return report, err
	}
	if strings.TrimSpace(b.BinaryPath) == "" {
		return report, fmt.Errorf("scanner binary path is required")
	}
	if b.Args == nil {
		return report, fmt.Errorf("scanner argument builder is required")
	}
	args, err := b.Args(request)
	if err != nil {
		return report, fmt.Errorf("build scanner arguments: %w", err)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}
	command := exec.CommandContext(ctx, b.BinaryPath, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &limitedBuffer{Buffer: &stdout, Limit: b.outputLimit()}
	command.Stderr = &limitedBuffer{Buffer: &stderr, Limit: b.outputLimit()}
	if err := command.Run(); err != nil {
		report.Errors = append(report.Errors, ScanError{Code: "scanner_execution", Message: scannerError(err, stderr.String())})
		report.Finish()
		return report, nil
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		report.Errors = append(report.Errors, ScanError{Code: "scanner_output", Message: fmt.Sprintf("decode scanner JSON: %v", err)})
		report.Finish()
		return report, nil
	}
	report.Backend = b.Name()
	report.Target = request.Target
	if report.Status == "" {
		report.Status = StatusOK
	}
	report.Finish()
	return report, nil
}

type limitedBuffer struct {
	*bytes.Buffer
	Limit int
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	if b.Limit > 0 && b.Len()+len(data) > b.Limit {
		return 0, fmt.Errorf("scanner output exceeds %d bytes", b.Limit)
	}
	return b.Buffer.Write(data)
}

func (b *ProcessBackend) outputLimit() int {
	if b.MaxOutput <= 0 {
		return 16 << 20
	}
	return b.MaxOutput
}

func scannerError(runErr error, stderr string) string {
	if strings.TrimSpace(stderr) == "" {
		return runErr.Error()
	}
	return fmt.Sprintf("%v: %s", runErr, strings.TrimSpace(stderr))
}
