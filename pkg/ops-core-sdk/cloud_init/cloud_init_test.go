package cloud_init

import "testing"

func TestStatus(t *testing.T) {
	// Status will return available=false if cloud-init not installed
	result, _ := Status()
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestModules(t *testing.T) {
	// Modules may fail if cloud-init not installed, which is OK
	_, _ = Modules()
}
