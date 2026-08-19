package monit

import "testing"

func TestStartEmpty(t *testing.T) {
	r := Start("")
	if r.Status != "failed" {
		t.Error("expected failure for empty service")
	}
}

func TestStopEmpty(t *testing.T) {
	r := Stop("")
	if r.Status != "failed" {
		t.Error("expected failure for empty service")
	}
}

func TestMonitorEmpty(t *testing.T) {
	r := Monitor("")
	if r.Status != "failed" {
		t.Error("expected failure for empty service")
	}
}

func TestUnmonitorEmpty(t *testing.T) {
	r := Unmonitor("")
	if r.Status != "failed" {
		t.Error("expected failure for empty service")
	}
}

func TestRestartEmpty(t *testing.T) {
	r := Restart("")
	if r.Status != "failed" {
		t.Error("expected failure for empty service")
	}
}
