package sys_persist

import "testing"

func TestSetEmptyName(t *testing.T) {
	_, err := Set("", "1")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestSetEmptyValue(t *testing.T) {
	_, err := Set("net.ipv4.ip_forward", "")
	if err == nil {
		t.Error("expected error for empty value")
	}
}

func TestGetEmptyName(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRemoveEmptyName(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestList(t *testing.T) {
	// List should work even if /etc/sysctl.d doesn't exist
	result, _ := List()
	if result == nil {
		t.Error("expected non-nil result")
	}
}
