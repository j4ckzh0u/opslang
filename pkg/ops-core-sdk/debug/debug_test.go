package debug

import "testing"

func TestDebug(t *testing.T) {
	r := Debug("hello")
	if r.Message != "hello" {
		t.Errorf("expected hello, got %s", r.Message)
	}
	if r.Changed {
		t.Error("debug should not change")
	}
}

func TestDebugVar(t *testing.T) {
	r := DebugVar("x", "42")
	if r.Var != "x" {
		t.Errorf("expected x, got %s", r.Var)
	}
}
