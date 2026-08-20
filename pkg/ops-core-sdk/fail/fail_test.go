package fail

import "testing"

func TestFail(t *testing.T) {
	r := Fail("boom")
	if !r.Failed {
		t.Error("expected failed")
	}
	if r.Message != "boom" {
		t.Errorf("expected boom, got %s", r.Message)
	}
}

func TestFailDefault(t *testing.T) {
	r := Fail("")
	if r.Message == "" {
		t.Error("expected default message")
	}
}

func TestFailF(t *testing.T) {
	r := FailF("error %d", 42)
	if r.Message != "error 42" {
		t.Errorf("expected 'error 42', got %s", r.Message)
	}
}
