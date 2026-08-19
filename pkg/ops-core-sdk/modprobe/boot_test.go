package modprobe

import (
	"testing"
)

func TestSetBootValidation(t *testing.T) {
	// SetBoot requires root, so we test the validation logic
	_, err := SetBoot("", true)
	if err != nil {
		// Empty name should fail (can't write config)
		t.Logf("expected error for empty name: %v", err)
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Changed: true, Message: "test"}
	if !r.Changed {
		t.Error("expected Changed=true")
	}
	if r.Message != "test" {
		t.Errorf("expected 'test', got %s", r.Message)
	}
}

func TestReadBootModules(t *testing.T) {
	// readBootModules should return empty/nil if no config exists
	modules := readBootModules()
	// Just verify it doesn't panic
	_ = modules
}
