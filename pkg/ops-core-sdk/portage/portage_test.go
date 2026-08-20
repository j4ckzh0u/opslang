package portage

import (
	"os"
	"testing"
)

func TestFindEmerge(t *testing.T) {
	if _, err := os.Stat("/usr/bin/emerge"); err != nil {
		t.Skip("emerge not available, skipping")
	}
	path, err := findEmerge()
	if err != nil {
		t.Fatalf("findEmerge failed: %v", err)
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

func TestUpdateEmptyName(t *testing.T) {
	// Update with empty name updates @world, should not error on validation
	// but may error if emerge not found
	if _, err := os.Stat("/usr/bin/emerge"); err != nil {
		_, err := Update("", false)
		if err == nil {
			t.Fatal("expected error when emerge not found")
		}
		return
	}
	_, err := Update("", false)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
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

func TestMetadataEmptyName(t *testing.T) {
	_, err := Metadata("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSyncNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/bin/emerge"); err != nil {
		_, err := Sync()
		if err == nil {
			t.Fatal("expected error when emerge not found")
		}
	}
}

func TestDepcleanNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/bin/emerge"); err != nil {
		_, err := Depclean()
		if err == nil {
			t.Fatal("expected error when emerge not found")
		}
	}
}

func TestListNotAvailable(t *testing.T) {
	if _, err := os.Stat("/usr/bin/emerge"); err != nil {
		_, err := List()
		if err == nil {
			t.Fatal("expected error when emerge not found")
		}
	}
}
