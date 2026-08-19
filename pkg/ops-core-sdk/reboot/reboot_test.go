package reboot

import "testing"

func TestDryRun(t *testing.T) {
	r := DryRun("maintenance", 60)
	if r.Status != "success" {
		t.Errorf("dry run failed: %s", r.Error)
	}
	if r.Command == "" {
		t.Error("expected a command")
	}
}

func TestDryRunNoDelay(t *testing.T) {
	r := DryRun("", 0)
	if r.Status != "success" {
		t.Errorf("dry run failed: %s", r.Error)
	}
	if r.Command == "" {
		t.Error("expected a command")
	}
}

func TestCheck(t *testing.T) {
	r := Check()
	// May fail on non-Linux
	if r.Status == "success" && !r.Booted {
		t.Error("expected booted=true on success")
	}
}

func TestRequestPermission(t *testing.T) {
	// Request will fail without root permissions
	r := Request("", 0)
	// Just verify it doesn't panic
	_ = r.Status
}
