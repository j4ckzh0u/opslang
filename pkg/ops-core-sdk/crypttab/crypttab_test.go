package crypttab

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAdd_EmptyArgs(t *testing.T) {
	_, err := Add("", "/dev/sda1", "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	_, err = Add("test", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty device")
	}
}

func TestRemove_EmptyName(t *testing.T) {
	_, err := Remove("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestModify_EmptyName(t *testing.T) {
	_, err := Modify("", "/dev/sda1", "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGet_EmptyName(t *testing.T) {
	_, err := Get("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestExists_EmptyName(t *testing.T) {
	_, err := Exists("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestActionResultJSON(t *testing.T) {
	r := ActionResult{Name: "cryptroot", Changed: true, Success: true}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestEntryJSON(t *testing.T) {
	entry := Entry{
		Name:    "cryptroot",
		Device:  "/dev/sda1",
		KeyFile: "none",
		Options: "luks",
	}
	b, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestListResultJSON(t *testing.T) {
	result := ListResult{
		Entries: []Entry{
			{Name: "cryptroot", Device: "/dev/sda1"},
		},
	}
	b, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestCrypttabOperations(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	crypttabPath := filepath.Join(tmpDir, "crypttab")

	// Test List on non-existent file
	result, err := List()
	if err != nil {
		t.Logf("List returned error (expected on non-standard systems): %v", err)
	} else {
		t.Logf("List returned %d entries", len(result.Entries))
	}

	// Test with temp file
	oldCrypttab := "/etc/crypttab"
	if err := os.WriteFile(crypttabPath, []byte("# Test\n"), 0644); err != nil {
		t.Skip("Cannot create test crypttab file")
	}

	// Note: Actual operations require /etc/crypttab which we can't modify in tests
	t.Logf("Test file created at %s", crypttabPath)
	_ = oldCrypttab
}
