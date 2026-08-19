package udevadm

import "testing"

func TestControlEmpty(t *testing.T) {
	r := Control("")
	if r.Error == "" {
		t.Error("expected error for empty action")
	}
}

func TestTrigger(t *testing.T) {
	r := Trigger("")
	// May fail if udevadm not available
	_ = r.Action
}

func TestSettle(t *testing.T) {
	r := Settle(1)
	// May fail if udevadm not available
	_ = r.Status
}

func TestInfoEmpty(t *testing.T) {
	r := Info("", "")
	if r.Error == "" {
		t.Error("expected error for empty device")
	}
}

func TestMonitor(t *testing.T) {
	r := Monitor()
	// May fail if udevadm not available
	_ = r.Output
}
