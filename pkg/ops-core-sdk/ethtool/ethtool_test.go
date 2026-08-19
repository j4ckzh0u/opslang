package ethtool

import "testing"

func TestShowEmpty(t *testing.T) {
	r := Show("")
	if r.Error == "" {
		t.Error("expected error for empty interface")
	}
}

func TestSetSpeedEmpty(t *testing.T) {
	r := SetSpeed("", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestSetDuplexEmpty(t *testing.T) {
	r := SetDuplex("", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestSetAutonegEmpty(t *testing.T) {
	r := SetAutoneg("", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}

func TestSetPauseEmpty(t *testing.T) {
	r := SetPause("", "", "")
	if r.Error == "" {
		t.Error("expected error for empty interface")
	}
}

func TestSetOffloadEmpty(t *testing.T) {
	r := SetOffload("", "", "")
	if r.Error == "" {
		t.Error("expected error for empty args")
	}
}
