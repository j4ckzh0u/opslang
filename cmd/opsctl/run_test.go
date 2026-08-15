package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCommand_BasicScript(t *testing.T) {
	// Create a temp script file.
	dir := t.TempDir()
	script := filepath.Join(dir, "test.ops")
	content := `let x = 42
print("hello " + str(x))
report { value: x }
`
	if err := os.WriteFile(script, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset flags.
	runOutputJSON = false
	runVerbose = false

	err := runRunCommand(script)
	if err != nil {
		t.Fatalf("runRunCommand failed: %v", err)
	}
}

func TestRunCommand_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "test.ops")
	content := `let x = 10
report { num: x }
`
	if err := os.WriteFile(script, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	runOutputJSON = true
	defer func() { runOutputJSON = false }()

	err := runRunCommand(script)
	if err != nil {
		t.Fatalf("runRunCommand JSON failed: %v", err)
	}
}

func TestRunCommand_FileNotFound(t *testing.T) {
	err := runRunCommand("/nonexistent/script.ops")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to read") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCommand_ParseError(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "bad.ops")
	content := `let x = @@@ invalid syntax`
	if err := os.WriteFile(script, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	runOutputJSON = false
	err := runRunCommand(script)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestRunCommand_Verbose(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "test.ops")
	content := `print("verbose test")`
	if err := os.WriteFile(script, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	runVerbose = true
	defer func() { runVerbose = false }()

	err := runRunCommand(script)
	if err != nil {
		t.Fatalf("runRunCommand verbose failed: %v", err)
	}
}

func TestRunCommand_EmptyScript(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "empty.ops")
	if err := os.WriteFile(script, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	runOutputJSON = false
	err := runRunCommand(script)
	if err != nil {
		t.Fatalf("empty script should not error: %v", err)
	}
}

func TestRunCommand_AllExamples(t *testing.T) {
	// Test that all example scripts parse and run without error.
	examples := []string{
		"check_cpu.ops",
		"check_memory.ops",
		"file_ops.ops",
		"deploy_app.ops",
		"ensure_service.ops",
		"variables.ops",
		"control_flow.ops",
		"functions.ops",
		"disk_check.ops",
		"network_check.ops",
		"process_monitor.ops",
		"log_analyzer.ops",
		"config_generator.ops",
	}

	for _, name := range examples {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", name)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("example %s not found", name)
			}

			runOutputJSON = false
			runVerbose = false
			if err := runRunCommand(path); err != nil {
				t.Errorf("example %s failed: %v", name, err)
			}
		})
	}
}

func TestFormatOutputData(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"hello", "hello"},
		{42, "42"},
		{map[string]interface{}{"a": 1}, `{"a":1}`},
	}

	for _, tt := range tests {
		result := formatOutputData(tt.input)
		if result != tt.expected {
			t.Errorf("formatOutputData(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
