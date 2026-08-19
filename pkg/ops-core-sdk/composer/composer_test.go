package composer

import "testing"

func TestInstallMissingDir(t *testing.T) {
	r := Install("/nonexistent/project", false)
	if r.Success {
		t.Error("expected failure for missing dir")
	}
	if r.Error == "" {
		t.Error("expected error message")
	}
}

func TestRequireMissingPackage(t *testing.T) {
	r := Require(".", "", "")
	if r.Success {
		t.Error("expected failure for empty package")
	}
}

func TestRemoveMissingPackage(t *testing.T) {
	r := Remove(".", "")
	if r.Success {
		t.Error("expected failure for empty package")
	}
}

func TestCreateProjectMissingDir(t *testing.T) {
	r := CreateProject("", "laravel/laravel", "")
	if r.Success {
		t.Error("expected failure for empty dir")
	}
}

func TestGlobalInstallMissingPackage(t *testing.T) {
	r := GlobalInstall("", "")
	if r.Success {
		t.Error("expected failure for empty package")
	}
}

func TestVersion(t *testing.T) {
	v := Version()
	// May fail if composer not installed, just check structure
	if !v.Success && v.Version != "" {
		t.Error("unexpected state")
	}
}
