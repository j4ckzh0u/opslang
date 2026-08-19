package puppet

import "testing"

func TestRun(t *testing.T) {
	// May fail if puppet not installed
	r := Run("", nil)
	_ = r.Status
}

func TestRunNoop(t *testing.T) {
	r := RunNoop("", nil)
	_ = r.Status
}

func TestDisable(t *testing.T) {
	r := Disable("test")
	_ = r.Status
}

func TestFactEmpty(t *testing.T) {
	r := Fact("")
	if r.Status != "failed" {
		t.Error("expected failure for empty fact name")
	}
}

func TestModuleInstallEmpty(t *testing.T) {
	r := ModuleInstall("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty module name")
	}
}
