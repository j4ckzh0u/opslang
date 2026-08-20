package smartctl_notify

import "testing"

func TestCheckEmpty(t *testing.T) {
	_, err := Check("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestShortTestEmpty(t *testing.T) {
	_, err := ShortTest("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestLongTestEmpty(t *testing.T) {
	_, err := LongTest("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestCheckNonexistent(t *testing.T) {
	// Will return unhealthy if smartctl not installed or device doesn't exist
	result, _ := Check("/dev/nonexistent")
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestListDevices(t *testing.T) {
	// Will fail gracefully if smartctl not installed
	_, _ = ListDevices()
}
