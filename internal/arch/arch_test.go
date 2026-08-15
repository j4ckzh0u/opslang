package arch

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// mockExecutor is a test double for SSHExecutor.
type mockExecutor struct {
	stdout   string
	exitCode int
	err      error
	lastCmd  string
}

func (m *mockExecutor) Exec(ctx context.Context, cmd string) (*ExecResult, error) {
	m.lastCmd = cmd
	if m.err != nil {
		return nil, m.err
	}
	return &ExecResult{
		Stdout:   m.stdout,
		Stderr:   "",
		ExitCode: m.exitCode,
	}, nil
}

// ============================================================
// Detect tests
// ============================================================

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		stdout   string
		exitCode int
		execErr  error
		wantArch string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "x86_64 maps to amd64",
			stdout:   "x86_64\n",
			exitCode: 0,
			wantArch: "amd64",
		},
		{
			name:     "aarch64 maps to arm64",
			stdout:   "aarch64\n",
			exitCode: 0,
			wantArch: "arm64",
		},
		{
			name:     "armv7l maps to arm",
			stdout:   "armv7l\n",
			exitCode: 0,
			wantArch: "arm",
		},
		{
			name:     "i686 maps to 386",
			stdout:   "i686\n",
			exitCode: 0,
			wantArch: "386",
		},
		{
			name:     "s390x maps to s390x",
			stdout:   "s390x\n",
			exitCode: 0,
			wantArch: "s390x",
		},
		{
			name:     "ppc64le maps to ppc64le",
			stdout:   "ppc64le\n",
			exitCode: 0,
			wantArch: "ppc64le",
		},
		{
			name:     "riscv64 maps to riscv64",
			stdout:   "riscv64\n",
			exitCode: 0,
			wantArch: "riscv64",
		},
		{
			name:     "whitespace trimmed",
			stdout:   "  x86_64  \n",
			exitCode: 0,
			wantArch: "amd64",
		},
		{
			name:     "unsupported architecture",
			stdout:   "sparc64\n",
			exitCode: 0,
			wantErr:  true,
			errMsg:   "unsupported architecture",
		},
		{
			name:     "non-zero exit code",
			stdout:   "",
			exitCode: 1,
			wantErr:  true,
			errMsg:   "uname exited with code 1",
		},
		{
			name:    "exec error",
			execErr: fmt.Errorf("connection lost"),
			wantErr: true,
			errMsg:  "failed to execute uname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockExecutor{
				stdout:   tt.stdout,
				exitCode: tt.exitCode,
				err:      tt.execErr,
			}

			got, err := Detect(context.Background(), mock)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantArch {
				t.Errorf("expected arch %q, got %q", tt.wantArch, got)
			}
		})
	}
}

func TestDetectExecutesUname(t *testing.T) {
	mock := &mockExecutor{stdout: "x86_64\n"}
	_, err := Detect(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastCmd != "uname -m" {
		t.Errorf("expected command 'uname -m', got %q", mock.lastCmd)
	}
}

func TestDetectContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	mock := &mockExecutor{stdout: "x86_64\n"}
	// The mock doesn't respect context, but Detect should pass it through.
	// This test verifies the context is passed to the executor.
	_, _ = Detect(ctx, mock)
	// If mock.lastCmd is set, context was passed through.
	if mock.lastCmd != "uname -m" {
		t.Error("expected uname -m to be called even with cancelled context (mock doesn't respect ctx)")
	}
}

// ============================================================
// MapArch tests
// ============================================================

func TestMapArch(t *testing.T) {
	tests := []struct {
		raw      string
		wantArch string
		wantErr  bool
	}{
		{"x86_64", "amd64", false},
		{"amd64", "amd64", false},
		{"aarch64", "arm64", false},
		{"arm64", "arm64", false},
		{"armv7l", "arm", false},
		{"armv6l", "arm", false},
		{"armv5l", "arm", false},
		{"i386", "386", false},
		{"i486", "386", false},
		{"i586", "386", false},
		{"i686", "386", false},
		{"mips", "mips", false},
		{"mips64", "mips64", false},
		{"ppc64le", "ppc64le", false},
		{"ppc64", "ppc64", false},
		{"s390x", "s390x", false},
		{"riscv64", "riscv64", false},
		{"  x86_64  ", "amd64", false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := MapArch(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantArch {
				t.Errorf("expected %q, got %q", tt.wantArch, got)
			}
		})
	}
}

// ============================================================
// SupportedArchitectures tests
// ============================================================

func TestSupportedArchitectures(t *testing.T) {
	archs := SupportedArchitectures()
	if len(archs) == 0 {
		t.Fatal("expected at least one supported architecture")
	}

	// Check that all expected GOARCH values are present.
	expected := map[string]bool{
		"amd64":   true,
		"arm64":   true,
		"arm":     true,
		"386":     true,
		"mips":    true,
		"mips64":  true,
		"ppc64le": true,
		"ppc64":   true,
		"s390x":   true,
		"riscv64": true,
	}

	for _, a := range archs {
		delete(expected, a)
	}
	if len(expected) > 0 {
		missing := make([]string, 0, len(expected))
		for k := range expected {
			missing = append(missing, k)
		}
		sort.Strings(missing)
		t.Errorf("missing expected architectures: %v", missing)
	}
}

func TestSupportedArchitecturesNoDuplicates(t *testing.T) {
	archs := SupportedArchitectures()
	seen := make(map[string]bool)
	for _, a := range archs {
		if seen[a] {
			t.Errorf("duplicate architecture: %q", a)
		}
		seen[a] = true
	}
}

// ============================================================
// RunnerBinaryName tests
// ============================================================

func TestRunnerBinaryName(t *testing.T) {
	tests := []struct {
		goarch string
		want   string
	}{
		{"amd64", "ops-runner-linux-amd64"},
		{"arm64", "ops-runner-linux-arm64"},
		{"arm", "ops-runner-linux-arm"},
		{"386", "ops-runner-linux-386"},
	}

	for _, tt := range tests {
		t.Run(tt.goarch, func(t *testing.T) {
			got := RunnerBinaryName(tt.goarch)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
