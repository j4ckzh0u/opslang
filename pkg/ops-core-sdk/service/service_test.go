package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestServiceStatusJSON verifies ServiceStatus serializes correctly to JSON.
func TestServiceStatusJSON(t *testing.T) {
	status := ServiceStatus{
		Name:        "nginx",
		ActiveState: "active",
		SubState:    "running",
		LoadState:   "loaded",
		MainPID:     1234,
		Description: "The NGINX HTTP Server",
		Enabled:     true,
		Active:      true,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal ServiceStatus: %v", err)
	}

	var decoded ServiceStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ServiceStatus: %v", err)
	}

	if decoded.Name != status.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, status.Name)
	}
	if decoded.ActiveState != status.ActiveState {
		t.Errorf("ActiveState: got %q, want %q", decoded.ActiveState, status.ActiveState)
	}
	if decoded.SubState != status.SubState {
		t.Errorf("SubState: got %q, want %q", decoded.SubState, status.SubState)
	}
	if decoded.LoadState != status.LoadState {
		t.Errorf("LoadState: got %q, want %q", decoded.LoadState, status.LoadState)
	}
	if decoded.MainPID != status.MainPID {
		t.Errorf("MainPID: got %d, want %d", decoded.MainPID, status.MainPID)
	}
	if decoded.Description != status.Description {
		t.Errorf("Description: got %q, want %q", decoded.Description, status.Description)
	}
	if decoded.Enabled != status.Enabled {
		t.Errorf("Enabled: got %v, want %v", decoded.Enabled, status.Enabled)
	}
	if decoded.Active != status.Active {
		t.Errorf("Active: got %v, want %v", decoded.Active, status.Active)
	}

	// Verify JSON tags are correct
	jsonStr := string(data)
	expectedFields := []string{
		`"name"`, `"active_state"`, `"sub_state"`, `"load_state"`,
		`"main_pid"`, `"description"`, `"enabled"`, `"active"`,
	}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output missing field %s: %s", field, jsonStr)
		}
	}
}

// TestServiceActionJSON verifies ServiceAction serializes correctly to JSON.
func TestServiceActionJSON(t *testing.T) {
	action := ServiceAction{
		Name:    "sshd",
		Action:  "start",
		Success: true,
		Message: "service \"sshd\" started successfully",
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal ServiceAction: %v", err)
	}

	var decoded ServiceAction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ServiceAction: %v", err)
	}

	if decoded.Name != action.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, action.Name)
	}
	if decoded.Action != action.Action {
		t.Errorf("Action: got %q, want %q", decoded.Action, action.Action)
	}
	if decoded.Success != action.Success {
		t.Errorf("Success: got %v, want %v", decoded.Success, action.Success)
	}
	if decoded.Message != action.Message {
		t.Errorf("Message: got %q, want %q", decoded.Message, action.Message)
	}

	// Verify JSON tags
	jsonStr := string(data)
	expectedFields := []string{`"name"`, `"action"`, `"success"`, `"message"`}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output missing field %s: %s", field, jsonStr)
		}
	}
}

// TestParseProperties verifies the parsing of systemctl show output.
func TestParseProperties(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ServiceStatus
	}{
		{
			name: "active running service",
			input: `ActiveState=active
SubState=running
LoadState=loaded
MainPID=1234
Description=The NGINX HTTP Server`,
			expected: ServiceStatus{
				ActiveState: "active",
				SubState:    "running",
				LoadState:   "loaded",
				MainPID:     1234,
				Description: "The NGINX HTTP Server",
			},
		},
		{
			name: "inactive dead service",
			input: `ActiveState=inactive
SubState=dead
LoadState=loaded
MainPID=0
Description=OpenSSH server daemon`,
			expected: ServiceStatus{
				ActiveState: "inactive",
				SubState:    "dead",
				LoadState:   "loaded",
				MainPID:     0,
				Description: "OpenSSH server daemon",
			},
		},
		{
			name: "service with no main PID",
			input: `ActiveState=active
SubState=running
LoadState=loaded
MainPID=0
Description=D-Bus System Message Bus`,
			expected: ServiceStatus{
				ActiveState: "active",
				SubState:    "running",
				LoadState:   "loaded",
				MainPID:     0,
				Description: "D-Bus System Message Bus",
			},
		},
		{
			name:     "empty output",
			input:    "",
			expected: ServiceStatus{},
		},
		{
			name: "output with blank lines and extra properties",
			input: `ActiveState=active

Id=nginx.service
SubState=running
SomeOtherProperty=ignored
LoadState=loaded
MainPID=5678
Description=nginx web server
`,
			expected: ServiceStatus{
				ActiveState: "active",
				SubState:    "running",
				LoadState:   "loaded",
				MainPID:     5678,
				Description: "nginx web server",
			},
		},
		{
			name: "invalid PID value",
			input: `ActiveState=active
SubState=running
LoadState=loaded
MainPID=notanumber
Description=test`,
			expected: ServiceStatus{
				ActiveState: "active",
				SubState:    "running",
				LoadState:   "loaded",
				MainPID:     0,
				Description: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status ServiceStatus
			parseProperties(tt.input, &status)

			if status.ActiveState != tt.expected.ActiveState {
				t.Errorf("ActiveState: got %q, want %q", status.ActiveState, tt.expected.ActiveState)
			}
			if status.SubState != tt.expected.SubState {
				t.Errorf("SubState: got %q, want %q", status.SubState, tt.expected.SubState)
			}
			if status.LoadState != tt.expected.LoadState {
				t.Errorf("LoadState: got %q, want %q", status.LoadState, tt.expected.LoadState)
			}
			if status.MainPID != tt.expected.MainPID {
				t.Errorf("MainPID: got %d, want %d", status.MainPID, tt.expected.MainPID)
			}
			if status.Description != tt.expected.Description {
				t.Errorf("Description: got %q, want %q", status.Description, tt.expected.Description)
			}
		})
	}
}

// TestStatusMissingSystemctl verifies that Status returns an error when systemctl is not available.
func TestStatusMissingSystemctl(t *testing.T) {
	// Override systemctlBin to a non-existent path
	original := systemctlBin
	systemctlBin = "/nonexistent/path/to/systemctl"
	defer func() { systemctlBin = original }()

	_, err := Status("nginx")
	if err == nil {
		t.Fatal("expected error when systemctl is missing, got nil")
	}

	if !strings.Contains(err.Error(), "failed to get status") {
		t.Errorf("error message should mention 'failed to get status': %v", err)
	}
}

// TestActionMissingSystemctl verifies that action functions return errors when systemctl is missing.
func TestActionMissingSystemctl(t *testing.T) {
	original := systemctlBin
	systemctlBin = "/nonexistent/path/to/systemctl"
	defer func() { systemctlBin = original }()

	tests := []struct {
		name   string
		fn     func(string) (ServiceAction, error)
		action string
	}{
		{"Start", Start, "start"},
		{"Stop", Stop, "stop"},
		{"Restart", Restart, "restart"},
		{"Enable", Enable, "enable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn("test-service")
			if err == nil {
				t.Fatalf("expected error from %s when systemctl is missing, got nil", tt.name)
			}
			if result.Success {
				t.Errorf("%s: expected Success=false, got true", tt.name)
			}
			if result.Name != "test-service" {
				t.Errorf("%s: Name: got %q, want %q", tt.name, result.Name, "test-service")
			}
			if result.Action != tt.action {
				t.Errorf("%s: Action: got %q, want %q", tt.name, result.Action, tt.action)
			}
			if result.Message == "" {
				t.Errorf("%s: expected non-empty error message", tt.name)
			}
		})
	}
}

// TestIsEnabledMissingSystemctl verifies isEnabled returns false when systemctl is missing.
func TestIsEnabledMissingSystemctl(t *testing.T) {
	original := systemctlBin
	systemctlBin = "/nonexistent/path/to/systemctl"
	defer func() { systemctlBin = original }()

	enabled, err := isEnabled("nginx")
	// isEnabled swallows the error and returns false
	if err != nil {
		t.Errorf("isEnabled should not return error for missing systemctl, got: %v", err)
	}
	if enabled {
		t.Error("isEnabled should return false when systemctl is missing")
	}
}

// TestRunWithFakeSystemctl tests the full flow using a fake systemctl script.
func TestRunWithFakeSystemctl(t *testing.T) {
	// Create a temporary directory for our fake systemctl
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "systemctl")

	// Create a fake systemctl script that simulates responses
	script := `#!/bin/sh
case "$1" in
  show)
    echo "ActiveState=active"
    echo "SubState=running"
    echo "LoadState=loaded"
    echo "MainPID=9999"
    echo "Description=Fake Service"
    ;;
  is-enabled)
    echo "enabled"
    exit 0
    ;;
  start|stop|restart|enable)
    exit 0
    ;;
  *)
    echo "Unknown command: $1" >&2
    exit 1
    ;;
esac
`
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create fake systemctl: %v", err)
	}

	original := systemctlBin
	systemctlBin = fakeBin
	defer func() { systemctlBin = original }()

	// Test Status
	t.Run("Status with fake systemctl", func(t *testing.T) {
		status, err := Status("fake-service")
		if err != nil {
			t.Fatalf("Status returned error: %v", err)
		}
		if status.Name != "fake-service" {
			t.Errorf("Name: got %q, want %q", status.Name, "fake-service")
		}
		if status.ActiveState != "active" {
			t.Errorf("ActiveState: got %q, want %q", status.ActiveState, "active")
		}
		if status.SubState != "running" {
			t.Errorf("SubState: got %q, want %q", status.SubState, "running")
		}
		if status.LoadState != "loaded" {
			t.Errorf("LoadState: got %q, want %q", status.LoadState, "loaded")
		}
		if status.MainPID != 9999 {
			t.Errorf("MainPID: got %d, want %d", status.MainPID, 9999)
		}
		if status.Description != "Fake Service" {
			t.Errorf("Description: got %q, want %q", status.Description, "Fake Service")
		}
		if !status.Enabled {
			t.Error("Enabled: got false, want true")
		}
		if !status.Active {
			t.Error("Active: got false, want true")
		}
	})

	// Test actions
	actions := []struct {
		name   string
		fn     func(string) (ServiceAction, error)
		action string
	}{
		{"Start", Start, "start"},
		{"Stop", Stop, "stop"},
		{"Restart", Restart, "restart"},
		{"Enable", Enable, "enable"},
	}

	for _, tt := range actions {
		t.Run(tt.name+" with fake systemctl", func(t *testing.T) {
			result, err := tt.fn("fake-service")
			if err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}
			if !result.Success {
				t.Errorf("%s: expected Success=true, got false", tt.name)
			}
			if result.Action != tt.action {
				t.Errorf("%s: Action: got %q, want %q", tt.name, result.Action, tt.action)
			}
			if result.Name != "fake-service" {
				t.Errorf("%s: Name: got %q, want %q", tt.name, result.Name, "fake-service")
			}
		})
	}
}

// TestRunWithFakeSystemctlDisabled tests Status when service is disabled.
func TestRunWithFakeSystemctlDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "systemctl")

	script := `#!/bin/sh
case "$1" in
  show)
    echo "ActiveState=inactive"
    echo "SubState=dead"
    echo "LoadState=loaded"
    echo "MainPID=0"
    echo "Description=Disabled Service"
    ;;
  is-enabled)
    echo "disabled"
    exit 1
    ;;
  *)
    exit 1
    ;;
esac
`
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create fake systemctl: %v", err)
	}

	original := systemctlBin
	systemctlBin = fakeBin
	defer func() { systemctlBin = original }()

	status, err := Status("disabled-service")
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}

	if status.Active {
		t.Error("Active: got true, want false for inactive service")
	}
	if status.Enabled {
		t.Error("Enabled: got true, want false for disabled service")
	}
	if status.ActiveState != "inactive" {
		t.Errorf("ActiveState: got %q, want %q", status.ActiveState, "inactive")
	}
}

// TestRunWithFakeSystemctlActionFailure tests action failure path.
func TestRunWithFakeSystemctlActionFailure(t *testing.T) {
	tmpDir := t.TempDir()
	fakeBin := filepath.Join(tmpDir, "systemctl")

	script := `#!/bin/sh
case "$1" in
  start)
    echo "Failed to start foo.service: Unit foo.service not found." >&2
    exit 5
    ;;
  *)
    exit 1
    ;;
esac
`
	if err := os.WriteFile(fakeBin, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to create fake systemctl: %v", err)
	}

	original := systemctlBin
	systemctlBin = fakeBin
	defer func() { systemctlBin = original }()

	result, err := Start("foo")
	if err == nil {
		t.Fatal("expected error from Start when service not found")
	}
	if result.Success {
		t.Error("Success: got true, want false")
	}
	if result.Action != "start" {
		t.Errorf("Action: got %q, want %q", result.Action, "start")
	}
	if result.Message == "" {
		t.Error("Message: expected non-empty error message on failure")
	}
	if !strings.Contains(result.Message, "not found") {
		t.Errorf("Message: expected to contain 'not found', got %q", result.Message)
	}
}
