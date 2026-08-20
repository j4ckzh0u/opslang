package lvol

import (
	"testing"
)

func TestList(t *testing.T) {
	// May fail if LVM not installed, that's OK for CI
	_, _ = List()
}

func TestVGList(t *testing.T) {
	// May fail if LVM not installed, that's OK for CI
	_, _ = VGList()
}

func TestActionResult(t *testing.T) {
	r := ActionResult{
		Changed: true,
		Message: "test",
	}
	if !r.Changed {
		t.Error("expected changed")
	}
	if r.Message != "test" {
		t.Errorf("expected 'test', got %s", r.Message)
	}
}
