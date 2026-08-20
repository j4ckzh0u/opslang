package sysvinit

import (
	"testing"
)

func TestStatusEmptyName(t *testing.T) {
	_, err := Status("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStartEmptyName(t *testing.T) {
	_, err := Start("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStopEmptyName(t *testing.T) {
	_, err := Stop("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRestartEmptyName(t *testing.T) {
	_, err := Restart("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestReloadEmptyName(t *testing.T) {
	_, err := Reload("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestEnableEmptyName(t *testing.T) {
	_, err := Enable("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDisableEmptyName(t *testing.T) {
	_, err := Disable("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStatusResultStruct(t *testing.T) {
	r := StatusResult{
		Status:  "success",
		Name:    "sshd",
		Running: true,
		Enabled: true,
		PID:     1234,
		State:   "running",
	}
	if r.Name != "sshd" {
		t.Error("name mismatch")
	}
	if !r.Running {
		t.Error("running should be true")
	}
	if r.PID != 1234 {
		t.Error("pid mismatch")
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Status:  "success",
		Changed: true,
		Name:    "sshd",
		Action:  "start",
		Output:  "started",
	}
	if r.Action != "start" {
		t.Error("action mismatch")
	}
	if !r.Changed {
		t.Error("changed should be true")
	}
}

func TestListNoInit(t *testing.T) {
	// On systems without /etc/init.d, should return empty or error gracefully
	result, _ := List()
	t.Logf("services found: %d", len(result))
}
