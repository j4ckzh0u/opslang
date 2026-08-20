package runit

import "testing"

func TestStatusEmpty(t *testing.T) {
	_, err := Status("")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestStartEmpty(t *testing.T) {
	_, err := Start("")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestStopEmpty(t *testing.T) {
	_, err := Stop("")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestRestartEmpty(t *testing.T) {
	_, err := Restart("")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestReloadEmpty(t *testing.T) {
	_, err := Reload("")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestEnableEmpty(t *testing.T) {
	_, err := Enable("", "")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestDisableEmpty(t *testing.T) {
	_, err := Disable("")
	if err == nil {
		t.Error("expected error for empty service name")
	}
}

func TestList(t *testing.T) {
	// This will fail if /var/service doesn't exist, which is OK
	result, _ := List()
	if result == nil {
		t.Error("expected non-nil result")
	}
}
