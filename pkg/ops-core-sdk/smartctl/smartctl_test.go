package smartctl

import "testing"

func TestDeviceEmpty(t *testing.T) {
	r := Device("")
	if r.Error == "" {
		t.Error("expected error for empty device")
	}
}

func TestHealthEmpty(t *testing.T) {
	r := Health("")
	if r.Error == "" {
		t.Error("expected error for empty device")
	}
}

func TestAttributesEmpty(t *testing.T) {
	r := Attributes("")
	if r.Error == "" {
		t.Error("expected error for empty device")
	}
}

func TestJSONEmpty(t *testing.T) {
	_, err := JSON("")
	if err == nil {
		t.Error("expected error for empty device")
	}
}

func TestList(t *testing.T) {
	r := List()
	// May fail if smartctl not installed
	_ = r.Count
}
