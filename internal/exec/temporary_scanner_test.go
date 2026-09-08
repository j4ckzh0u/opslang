package exec

import (
	"context"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/sshx"
	"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/securityscan"
)

type fakeRemoteExecutor struct {
	result *sshx.ExecResult
	err    error
}

func (f fakeRemoteExecutor) ExecWithStdin(context.Context, string, []byte) (*sshx.ExecResult, error) {
	return f.result, f.err
}

func (f fakeRemoteExecutor) Exec(context.Context, string) (*sshx.ExecResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &sshx.ExecResult{}, nil
}

func validScanRequest() securityscan.ScanRequest {
	return securityscan.ScanRequest{Target: securityscan.ScanTarget{ID: "host-1"}, Scanners: []securityscan.Capability{securityscan.CapabilityVulnerability}}
}

func TestTemporaryScannerRejectsInvalidInput(t *testing.T) {
	scanner := &TemporaryScanner{Client: fakeRemoteExecutor{}}
	if _, err := scanner.Scan(context.Background(), validScanRequest()); err == nil {
		t.Fatal("expected empty path error")
	}
	if _, err := (&TemporaryScanner{RemotePath: "/tmp/scanner"}).Scan(context.Background(), validScanRequest()); err == nil {
		t.Fatal("expected nil client error")
	}
}

func TestTemporaryScannerRejectsOutputLimitAndJSON(t *testing.T) {
	request := validScanRequest()
	tooLarge := &TemporaryScanner{Client: fakeRemoteExecutor{result: &sshx.ExecResult{Stdout: "{}"}}, RemotePath: "/tmp/scanner", MaxOutput: 1}
	if _, err := tooLarge.Scan(context.Background(), request); err == nil {
		t.Fatal("expected output limit error")
	}
	invalidJSON := &TemporaryScanner{Client: fakeRemoteExecutor{result: &sshx.ExecResult{Stdout: "invalid"}}, RemotePath: "/tmp/scanner"}
	if _, err := invalidJSON.Scan(context.Background(), request); err == nil {
		t.Fatal("expected JSON decoding error")
	}
}

func TestTemporaryScannerReturnsReport(t *testing.T) {
	scanner := &TemporaryScanner{
		Client:     fakeRemoteExecutor{result: &sshx.ExecResult{Stdout: `{"backend":"test","status":"ok"}`}},
		RemotePath: "/tmp/scanner",
	}
	report, err := scanner.Scan(context.Background(), validScanRequest())
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if report.Backend != "test" || report.Status != "ok" {
		t.Fatalf("unexpected report: %+v", report)
	}
	if report.FinishedAt.IsZero() {
		t.Fatal("expected finished timestamp")
	}
}

func TestTemporaryScannerRejectsNonZeroExit(t *testing.T) {
	scanner := &TemporaryScanner{
		Client:     fakeRemoteExecutor{result: &sshx.ExecResult{ExitCode: 2, Stderr: "scan failed"}},
		RemotePath: "/tmp/scanner",
	}
	if _, err := scanner.Scan(context.Background(), validScanRequest()); err == nil {
		t.Fatal("expected non-zero exit error")
	}
}

func TestExecutorRunTemporaryScanValidatesInputs(t *testing.T) {
	var executor *Executor
	if _, err := executor.RunTemporaryScan(context.Background(), fakeRemoteExecutor{}, "/tmp/scanner", validScanRequest()); err == nil {
		t.Fatal("expected nil executor error")
	}
	if _, err := (&Executor{}).RunTemporaryScan(context.Background(), fakeRemoteExecutor{}, "/tmp/scanner", securityscan.ScanRequest{}); err == nil {
		t.Fatal("expected invalid request error")
	}
}

func TestTemporaryScannerScanAndCleanup(t *testing.T) {
	scanner := &TemporaryScanner{
		Client:     fakeRemoteExecutor{result: &sshx.ExecResult{Stdout: `{"backend":"test","status":"ok"}`}},
		RemotePath: "/tmp/scanner",
	}
	if _, err := scanner.ScanAndCleanup(context.Background(), validScanRequest()); err != nil {
		t.Fatalf("scan and cleanup failed: %v", err)
	}
}
