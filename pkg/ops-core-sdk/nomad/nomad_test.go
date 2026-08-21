package nomad

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJobListContextParsesJSONAndPassesNamespace(t *testing.T) {
	cleanup := installFakeNomad(t, "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$NOMAD_ARGS_FILE\"\nprintf '[{\\\"id\\\":\\\"job-1\\\",\\\"name\\\":\\\"api\\\",\\\"status\\\":\\\"running\\\"}]'\n")
	defer cleanup()

	jobs, err := JobListContext(context.Background(), "prod")
	if err != nil {
		t.Fatalf("JobListContext() error = %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != "job-1" {
		t.Fatalf("unexpected jobs: %#v", jobs)
	}
	args, err := os.ReadFile(os.Getenv("NOMAD_ARGS_FILE"))
	if err != nil {
		t.Fatalf("read captured args: %v", err)
	}
	if got := strings.TrimSpace(string(args)); got != "job status -json -namespace prod" {
		t.Fatalf("args = %q", got)
	}
}

func TestJobRunContextRejectsEmptyFile(t *testing.T) {
	_, err := JobRunContext(context.Background(), "", "")
	if err == nil {
		t.Fatal("JobRunContext() should reject an empty job file")
	}
}

func TestNodeDrainContextReturnsCommandError(t *testing.T) {
	cleanup := installFakeNomad(t, "#!/bin/sh\nprintf 'permission denied' >&2\nexit 1\n")
	defer cleanup()

	result, err := NodeDrainContext(context.Background(), "node-1", true)
	if err == nil {
		t.Fatal("NodeDrainContext() should return command errors")
	}
	if !strings.Contains(result.Error, "permission denied") {
		t.Fatalf("result error = %q", result.Error)
	}
}

func TestNodeListContextHonorsCancellation(t *testing.T) {
	cleanup := installFakeNomad(t, "#!/bin/sh\nsleep 5\n")
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := NodeListContext(ctx)
	if err == nil {
		t.Fatal("NodeListContext() should fail after cancellation")
	}
}

func installFakeNomad(t *testing.T, script string) func() {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "nomad")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake nomad: %v", err)
	}
	argsFile := filepath.Join(dir, "args")
	oldPath := os.Getenv("PATH")
	oldArgs := os.Getenv("NOMAD_ARGS_FILE")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatalf("set PATH: %v", err)
	}
	if err := os.Setenv("NOMAD_ARGS_FILE", argsFile); err != nil {
		t.Fatalf("set NOMAD_ARGS_FILE: %v", err)
	}
	return func() {
		_ = os.Setenv("PATH", oldPath)
		_ = os.Setenv("NOMAD_ARGS_FILE", oldArgs)
	}
}
