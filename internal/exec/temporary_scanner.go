package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/sshx"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/securityscan"
)

// TemporaryScanner runs a staged scanner with its request on stdin.
type TemporaryScanner struct {
	Client     RemoteCommandExecutor
	RemotePath string
	MaxOutput  int
}

type RemoteCommandExecutor interface {
	ExecWithStdin(context.Context, string, []byte) (*sshx.ExecResult, error)
}

type RemoteCommandCleanup interface {
	Exec(context.Context, string) (*sshx.ExecResult, error)
}

// Cleanup removes the staged scanner when the remote executor supports
// command execution. Cleanup remains explicit so callers can retain a binary
// for diagnostics when required.
func (s *TemporaryScanner) Cleanup(ctx context.Context) error {
	if s == nil || s.Client == nil {
		return fmt.Errorf("temporary scanner client is nil")
	}
	cleaner, ok := s.Client.(RemoteCommandCleanup)
	if !ok {
		return fmt.Errorf("remote executor does not support cleanup")
	}
	if strings.TrimSpace(s.RemotePath) == "" {
		return fmt.Errorf("temporary scanner path is empty")
	}
	result, err := cleaner.Exec(ctx, sshx.JoinCommand("rm", "-f", s.RemotePath))
	if err != nil {
		return fmt.Errorf("cleanup temporary scanner: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("cleanup temporary scanner exited with code %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	return nil
}

// ScanAndCleanup runs a scan and always attempts cleanup afterward.
func (s *TemporaryScanner) ScanAndCleanup(ctx context.Context, request securityscan.ScanRequest) (securityscan.ScanReport, error) {
	report, scanErr := s.Scan(ctx, request)
	cleanupErr := s.Cleanup(ctx)
	if scanErr != nil && cleanupErr != nil {
		return report, fmt.Errorf("scan failed: %v; cleanup failed: %w", scanErr, cleanupErr)
	}
	if scanErr != nil {
		return report, scanErr
	}
	if cleanupErr != nil {
		return report, cleanupErr
	}
	return report, nil
}

func (s *TemporaryScanner) Scan(ctx context.Context, request securityscan.ScanRequest) (securityscan.ScanReport, error) {
	if s == nil || s.Client == nil {
		return securityscan.ScanReport{}, fmt.Errorf("temporary scanner client is nil")
	}
	if strings.TrimSpace(s.RemotePath) == "" {
		return securityscan.ScanReport{}, fmt.Errorf("temporary scanner path is empty")
	}
	if err := securityscan.ValidateScanRequest(request); err != nil {
		return securityscan.ScanReport{}, err
	}
	if ctx == nil {
		return securityscan.ScanReport{}, fmt.Errorf("temporary scanner context is nil")
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return securityscan.ScanReport{}, fmt.Errorf("encode temporary scan request: %w", err)
	}
	result, err := s.Client.ExecWithStdin(ctx, sshx.JoinCommand(s.RemotePath, "--json"), payload)
	if err != nil {
		return securityscan.ScanReport{}, fmt.Errorf("execute temporary scanner: %w", err)
	}
	if s.MaxOutput > 0 && len(result.Stdout) > s.MaxOutput {
		return securityscan.ScanReport{}, fmt.Errorf("temporary scanner output exceeds %d bytes", s.MaxOutput)
	}
	if result.ExitCode != 0 {
		return securityscan.ScanReport{}, fmt.Errorf("temporary scanner exited with code %d: %s", result.ExitCode, strings.TrimSpace(result.Stderr))
	}
	var report securityscan.ScanReport
	if err := json.Unmarshal([]byte(result.Stdout), &report); err != nil {
		return securityscan.ScanReport{}, fmt.Errorf("decode temporary scanner output: %w", err)
	}
	report.Finish()
	return report, nil
}
