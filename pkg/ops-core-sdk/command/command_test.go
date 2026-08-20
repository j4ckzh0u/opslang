package command

import (
	"testing"
	"time"
)

func TestRunEcho(t *testing.T) {
	r := Run([]string{"echo", "hello"}, "", "", "", 5*time.Second)
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.ExitCode != 0 {
		t.Errorf("expected exit 0, got %d", r.ExitCode)
	}
	if r.Stdout != "hello\n" {
		t.Errorf("unexpected stdout: %q", r.Stdout)
	}
}

func TestRunEmpty(t *testing.T) {
	r := Run(nil, "", "", "", 0)
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestRunFalse(t *testing.T) {
	r := Run([]string{"false"}, "", "", "", 5*time.Second)
	if r.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestShell(t *testing.T) {
	r := Shell([]string{"echo hello world"}, "", "", "", 5*time.Second, "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Stdout != "hello world\n" {
		t.Errorf("unexpected: %q", r.Stdout)
	}
}
