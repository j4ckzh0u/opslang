package pkgng

import (
	"os"
	"testing"
)

func TestFindPkg(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/pkg"); err != nil {
		t.Skip("pkg not available, skipping")
	}
	path, err := findPkg()
	if err != nil {
		t.Fatalf("findPkg failed: %v", err)
	}
	if path == "" {
		t.Fatal("empty path returned")
	}
}

func TestInstallEmptyName(t *testing.T) {
	_, err := Install("", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveEmptyName(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInfoEmptyName(t *testing.T) {
	_, err := Info("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSearchEmptyName(t *testing.T) {
	_, err := Search("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/pkg"); err != nil {
		_, err := Update()
		if err == nil {
			t.Fatal("expected error when pkg not found")
		}
	}
}

func TestUpgradeNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/pkg"); err != nil {
		_, err := Upgrade("")
		if err == nil {
			t.Fatal("expected error when pkg not found")
		}
	}
}

func TestAutocleanNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/pkg"); err != nil {
		_, err := Autoclean()
		if err == nil {
			t.Fatal("expected error when pkg not found")
		}
	}
}

func TestListNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/pkg"); err != nil {
		_, err := List()
		if err == nil {
			t.Fatal("expected error when pkg not found")
		}
	}
}

func TestStatsNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/pkg"); err != nil {
		_, err := Stats()
		if err == nil {
			t.Fatal("expected error when pkg not found")
		}
	}
}
