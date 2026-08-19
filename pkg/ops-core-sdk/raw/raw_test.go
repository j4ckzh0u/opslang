package raw

import (
	"strings"
	"testing"
)

func TestExecuteEmpty(t *testing.T) {
	r := Execute("", 5)
	if r.Status != "failed" {
		t.Error("expected failure for empty command")
	}
}

func TestExecuteSimple(t *testing.T) {
	r := Execute("echo hello", 5)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if strings.TrimSpace(r.Stdout) != "hello" {
		t.Errorf("expected 'hello', got '%s'", r.Stdout)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.ExitCode)
	}
}

func TestExecutePipe(t *testing.T) {
	r := Execute("echo -e 'b\na\nc' | sort", 5)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if !strings.Contains(r.Stdout, "a") {
		t.Errorf("pipe didn't work: %s", r.Stdout)
	}
}

func TestExecuteNonZero(t *testing.T) {
	r := Execute("exit 42", 5)
	if r.Status != "failed" {
		t.Error("expected failure for non-zero exit")
	}
	if r.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", r.ExitCode)
	}
}

func TestExecuteStderr(t *testing.T) {
	r := Execute("echo error >&2", 5)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if !strings.Contains(r.Stderr, "error") {
		t.Errorf("expected stderr 'error', got '%s'", r.Stderr)
	}
}

func TestExecuteWithEnv(t *testing.T) {
	r := ExecuteWithEnv("echo $TEST_VAR", 5, map[string]string{"TEST_VAR": "hello"})
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if !strings.Contains(r.Stdout, "hello") {
		t.Errorf("expected 'hello', got '%s'", r.Stdout)
	}
}
