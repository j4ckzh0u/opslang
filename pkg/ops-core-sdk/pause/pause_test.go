package pause

import "testing"

func TestSecondsInvalid(t *testing.T) {
	r := Seconds(0)
	if r.Status != "failed" {
		t.Error("expected failure for zero duration")
	}
}

func TestSecondsNegative(t *testing.T) {
	r := Seconds(-1)
	if r.Status != "failed" {
		t.Error("expected failure for negative duration")
	}
}

func TestSecondsShort(t *testing.T) {
	r := Seconds(1)
	if r.Status != "success" {
		t.Errorf("expected success: %s", r.Error)
	}
	if r.DurationMs < 900 { // Allow some tolerance
		t.Errorf("expected ~1000ms, got %dms", r.DurationMs)
	}
}
