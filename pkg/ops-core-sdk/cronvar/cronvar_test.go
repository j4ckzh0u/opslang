package cronvar

import "testing"

func TestPresentEmptyName(t *testing.T) {
	r := Present("", "value", "", "", "")
	if r.Error == "" {
		t.Error("expected error for empty name")
	}
}

func TestAbsentEmptyName(t *testing.T) {
	r := Absent("", "")
	if r.Error == "" {
		t.Error("expected error for empty name")
	}
}

func TestGetEmptyName(t *testing.T) {
	r := Get("", "")
	if r.Error == "" {
		t.Error("expected error for empty name")
	}
}
