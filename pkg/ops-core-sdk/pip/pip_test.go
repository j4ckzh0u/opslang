package pip

import (
	"testing"
)

func TestInstallEmptyName(t *testing.T) {
	r, err := Install("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestUninstallEmptyName(t *testing.T) {
	r, err := Uninstall("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestList(t *testing.T) {
	// List will succeed or fail based on pip3 availability — both are valid
	r, _ := List()
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestExistsEmptyName(t *testing.T) {
	r, err := Exists("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}

func TestExistsNonexistent(t *testing.T) {
	r, err := Exists("this_package_does_not_exist_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Exists {
		t.Error("expected exists=false for nonexistent package")
	}
}

func TestInstallRequirementsEmpty(t *testing.T) {
	r, err := InstallRequirements("")
	if err == nil {
		t.Fatal("expected error for empty requirements")
	}
	if r.Status != "failed" {
		t.Errorf("expected status=failed, got %s", r.Status)
	}
}
