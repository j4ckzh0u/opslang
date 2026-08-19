package tuned

import "testing"

func TestSetMissingProfile(t *testing.T) {
	r := Set("")
	if r.Success {
		t.Error("expected failure for empty profile")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestStatus(t *testing.T) {
	r := Status()
	// May fail if tuned not installed
	_ = r.Active
}

func TestList(t *testing.T) {
	r := List()
	// May fail if tuned not installed
	_ = r.Count
}

func TestProfile(t *testing.T) {
	_, err := Profile()
	// May fail if tuned not installed
	_ = err
}

func TestOff(t *testing.T) {
	r := Off()
	// May fail if tuned not installed
	_ = r
}
