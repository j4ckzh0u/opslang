package cargo

import "testing"

func TestInstallMissingPackage(t *testing.T) {
	r := Install("", "", false)
	if r.Success {
		t.Error("expected failure for empty package")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestUninstallMissingPackage(t *testing.T) {
	r := Uninstall("")
	if r.Success {
		t.Error("expected failure for empty package")
	}
}

func TestBuild(t *testing.T) {
	r := Build("/nonexistent/project", false)
	if r.Success {
		t.Error("expected failure for missing project")
	}
}

func TestTest(t *testing.T) {
	r := Test("/nonexistent/project")
	if r.Success {
		t.Error("expected failure for missing project")
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	// May fail if cargo not installed
	if !v.Success && v.Version != "" {
		t.Error("unexpected state")
	}
}

func TestList(t *testing.T) {
	_, err := List()
	// May fail if cargo not installed, just check no panic
	_ = err
}
