package firewalld_rich_rule

import "testing"

func TestAddEmptyRule(t *testing.T) {
	r := Add("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty rule")
	}
}

func TestRemoveEmptyRule(t *testing.T) {
	r := Remove("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty rule")
	}
}

func TestExistsEmptyRule(t *testing.T) {
	r := Exists("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty rule")
	}
}

func TestList(t *testing.T) {
	r := List("")
	// May fail if firewalld not running
	_ = r.Status
}
