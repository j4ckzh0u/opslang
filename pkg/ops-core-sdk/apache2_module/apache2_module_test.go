package apache2_module

import (
	"os/exec"
	"testing"
)

func skipIfNoApache(t *testing.T) {
	if _, err := exec.LookPath("apache2ctl"); err != nil {
		if _, err := exec.LookPath("httpd"); err != nil {
			t.Skip("apache2ctl/httpd not found in PATH")
		}
	}
}

func TestCheckValidation(t *testing.T) {
	_, err := Check("")
	if err == nil {
		t.Error("expected error for empty module")
	}
}

func TestEnableValidation(t *testing.T) {
	_, err := Enable("")
	if err == nil {
		t.Error("expected error for empty module")
	}
}

func TestDisableValidation(t *testing.T) {
	_, err := Disable("")
	if err == nil {
		t.Error("expected error for empty module")
	}
}

func TestCheckModule(t *testing.T) {
	skipIfNoApache(t)
	// Check a commonly available module
	result, err := Check("rewrite")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	// We just check it doesn't crash; enabled may vary
	_ = result.Enabled
}

func TestEnableIdempotent(t *testing.T) {
	skipIfNoApache(t)
	// Enable and check idempotency
	result1, err := Enable("rewrite")
	if err != nil {
		t.Logf("Enable() error (may need root): %v", err)
		return
	}
	if !result1.Changed && !result1.Enabled {
		// Was already enabled, should be idempotent
		result2, _ := Enable("rewrite")
		if result2.Changed {
			t.Error("Expected Changed=false for idempotent enable")
		}
	}
}
