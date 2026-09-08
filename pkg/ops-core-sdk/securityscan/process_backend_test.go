package securityscan

import (
	"context"
	"runtime"
	"testing"
)

func TestProcessBackendRejectsMissingConfiguration(t *testing.T) {
	backend := &ProcessBackend{}
	_, err := backend.Scan(context.Background(), ScanRequest{Target: ScanTarget{ID: "host"}, Scanners: []Capability{CapabilitySBOM}})
	if err == nil {
		t.Fatal("expected missing binary error")
	}
}

func TestProcessBackendReturnsStructuredExecutionError(t *testing.T) {
	binary := "sh"
	if runtime.GOOS == "windows" {
		binary = "cmd"
	}
	backend := &ProcessBackend{BinaryPath: binary, Args: func(ScanRequest) ([]string, error) {
		if runtime.GOOS == "windows" {
			return []string{"/C", "exit", "1"}, nil
		}
		return []string{"-c", "exit 1"}, nil
	}}
	report, err := backend.Scan(context.Background(), ScanRequest{Target: ScanTarget{ID: "host"}, Scanners: []Capability{CapabilitySBOM}})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if report.Status != StatusFailed || len(report.Errors) != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
}
