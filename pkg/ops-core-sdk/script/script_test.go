package script

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunEmpty(t *testing.T) {
	r := Run("", nil, "", "", "", 0, "")
	if r.Error == "" {
		t.Error("expected error for empty script")
	}
}

func TestRunNonExistent(t *testing.T) {
	r := Run("/nonexistent/script.sh", nil, "", "", "", 0, "")
	if r.Error == "" {
		t.Error("expected error for non-existent script")
	}
}

func TestRunEcho(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "test.sh")
	os.WriteFile(script, []byte("#!/bin/sh\necho hello\n"), 0755)

	r := Run(script, nil, "", "", "", 0, "")
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

func TestRunWithArgs(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "test.sh")
	os.WriteFile(script, []byte("#!/bin/sh\necho $1 $2\n"), 0755)

	r := Run(script, []string{"foo", "bar"}, "", "", "", 0, "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Stdout != "foo bar\n" {
		t.Errorf("unexpected: %q", r.Stdout)
	}
}
