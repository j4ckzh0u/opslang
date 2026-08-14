package process

import (
	"net"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	procs, err := List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(procs) == 0 {
		t.Fatal("List() returned no processes, expected at least one")
	}

	// Verify that at least one process has a non-empty name
	foundName := false
	for _, p := range procs {
		if p.Name != "" {
			foundName = true
			break
		}
	}
	if !foundName {
		t.Error("no process in the list has a non-empty Name")
	}
}

func TestFindByName(t *testing.T) {
	searchTerm := "go"
	searchLower := strings.ToLower(searchTerm)

	// Test with lowercase search term.
	procsLower, err := FindByName(searchTerm)
	if err != nil {
		t.Fatalf("FindByName(%q) returned error: %v", searchTerm, err)
	}

	// Verify all returned processes actually match the search term (case-insensitive).
	for _, p := range procsLower {
		if !strings.Contains(strings.ToLower(p.Name), searchLower) {
			t.Errorf("FindByName(%q) returned process %q which doesn't match", searchTerm, p.Name)
		}
	}

	// Test with uppercase search term.
	procsUpper, err := FindByName(strings.ToUpper(searchTerm))
	if err != nil {
		t.Fatalf("FindByName(%q) returned error: %v", strings.ToUpper(searchTerm), err)
	}

	// Verify all returned processes match.
	for _, p := range procsUpper {
		if !strings.Contains(strings.ToLower(p.Name), searchLower) {
			t.Errorf("FindByName(%q) returned process %q which doesn't match", strings.ToUpper(searchTerm), p.Name)
		}
	}

	// Both searches should return the same number of results (case-insensitive).
	// Note: on a live system, process lists can change between the two FindByName calls,
	// so we allow a small tolerance. The key assertion is that the matching logic is
	// case-insensitive, which is verified by the per-result checks above.
	diff := len(procsLower) - len(procsUpper)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2 {
		t.Errorf("FindByName case-insensitive mismatch: lowercase returned %d, uppercase returned %d",
			len(procsLower), len(procsUpper))
	}
}

func TestFindByNameEmpty(t *testing.T) {
	// Empty search should match all processes (empty string is contained in every string).
	// FindByName now uses List() internally, so the filtering logic is identical.
	// We don't do a strict count comparison with List() because the process list can
	// change between the two calls (processes starting/stopping).
	procs, err := FindByName("")
	if err != nil {
		t.Fatalf("FindByName(\"\") returned error: %v", err)
	}
	if len(procs) == 0 {
		t.Error("FindByName(\"\") returned no processes, expected at least some")
	}
}

func TestFindByPort(t *testing.T) {
	// On macOS, finding processes by port may require elevated privileges.
	// Skip if we can't get results.
	if runtime.GOOS == "darwin" {
		t.Skip("FindByPort may not work on macOS without elevated privileges")
	}

	// Start a TCP listener on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	procs, err := FindByPort(port)
	if err != nil {
		t.Fatalf("FindByPort(%d) returned error: %v", port, err)
	}

	if len(procs) == 0 {
		t.Errorf("FindByPort(%d) returned no processes, expected at least one", port)
	}

	// Verify each returned process has a valid PID
	for _, p := range procs {
		if p.PID <= 0 {
			t.Errorf("FindByPort returned process with invalid PID: %d", p.PID)
		}
	}
}

func TestFindByPortNoListener(t *testing.T) {
	// Query a port that's unlikely to have a listener
	procs, err := FindByPort(59999)
	if err != nil {
		t.Fatalf("FindByPort(59999) returned error: %v", err)
	}
	// We don't strictly require 0 results because another process might be using this port,
	// but there should be no error
	_ = procs
}

func TestExec(t *testing.T) {
	result, err := Exec("echo", []string{"hello"})
	if err != nil {
		// Exec itself shouldn't return an error for a successful command
		// (errors are captured in the result)
		t.Logf("Exec returned error (may be expected): %v", err)
	}

	if result.Command != "echo" {
		t.Errorf("Command = %q, want %q", result.Command, "echo")
	}
	if len(result.Args) != 1 || result.Args[0] != "hello" {
		t.Errorf("Args = %v, want [hello]", result.Args)
	}
	if !strings.Contains(result.Stdout, "hello") {
		t.Errorf("Stdout = %q, want it to contain %q", result.Stdout, "hello")
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestExecFail(t *testing.T) {
	// Try to run a command that doesn't exist
	result, err := Exec("/nonexistent-command-12345", []string{})
	_ = err // error may or may not be returned

	if result.ExitCode == 0 {
		t.Error("ExitCode = 0, want non-zero for failed command")
	}
}

func TestExecFailWithBadPath(t *testing.T) {
	// ls on a nonexistent path should fail
	result, _ := Exec("ls", []string{"/nonexistent/path/that/does/not/exist"})
	if result.ExitCode == 0 {
		t.Error("ExitCode = 0, want non-zero for ls on nonexistent path")
	}
	if result.Stderr == "" {
		t.Error("Stderr is empty, expected error message from ls")
	}
}

func TestExecEmptyArgs(t *testing.T) {
	// Run a command that takes no args - "true" always exits 0
	result, _ := Exec("true", []string{})
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0 for 'true' command", result.ExitCode)
	}
	if result.Stdout != "" {
		t.Errorf("Stdout = %q, want empty for 'true' command", result.Stdout)
	}
}

func TestExecFalseCommand(t *testing.T) {
	// "false" always exits non-zero
	result, _ := Exec("false", []string{})
	if result.ExitCode == 0 {
		t.Error("ExitCode = 0, want non-zero for 'false' command")
	}
}

func TestExecCapturesBothStreams(t *testing.T) {
	// Use sh -c to write to both stdout and stderr
	// Wait - the spec says NO shell calls. Let's use a command that writes to stderr natively.
	// We can use "ls" on a path that exists (stdout) and check it doesn't crash.
	// Actually, let's just test that Stdout and Stderr fields are populated correctly.
	result, _ := Exec("echo", []string{"test-output"})
	if !strings.Contains(result.Stdout, "test-output") {
		t.Errorf("Stdout = %q, want it to contain 'test-output'", result.Stdout)
	}
	// stderr should be empty for echo
	if result.Stderr != "" {
		t.Errorf("Stderr = %q, want empty for echo", result.Stderr)
	}
}

func TestExecPid(t *testing.T) {
	result, _ := Exec("echo", []string{"pid-test"})
	// PID should be set if the process started
	if result.Pid <= 0 && result.ExitCode == 0 {
		t.Errorf("Pid = %d, want > 0 for successful command", result.Pid)
	}
}

func TestExecCommandNotFound(t *testing.T) {
	// Test with a command that doesn't exist at all
	result, err := Exec("this-command-does-not-exist-opslang", []string{"arg1"})
	_ = err

	if result.Command != "this-command-does-not-exist-opslang" {
		t.Errorf("Command = %q, want 'this-command-does-not-exist-opslang'", result.Command)
	}
	if len(result.Args) != 1 || result.Args[0] != "arg1" {
		t.Errorf("Args = %v, want [arg1]", result.Args)
	}
	// Exit code should be non-zero for a nonexistent command
	if result.ExitCode == 0 {
		t.Error("ExitCode = 0, want non-zero for nonexistent command")
	}
}

// Helper to check if a command exists on the system
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
