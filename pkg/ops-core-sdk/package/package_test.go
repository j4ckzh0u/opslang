package package_mgr

import "testing"

func TestInstallEmpty(t *testing.T) {
	r := Install("")
	if r.Error == "" {
		t.Error("expected error for empty name")
	}
}

func TestRemoveEmpty(t *testing.T) {
	r := Remove("")
	if r.Error == "" {
		t.Error("expected error for empty name")
	}
}

func TestDetectManager(t *testing.T) {
	mgr := detectManager()
	// On macOS, no Linux package manager is available
	// On Linux, one should be found
	// Just verify it doesn't panic
	_ = mgr
}

func TestInfoEmpty(t *testing.T) {
	r := Info("")
	// Should return without panic, may have error on non-Linux
	_ = r
}
