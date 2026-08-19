package xattr

import (
	"os"
	"runtime"
	"testing"
)

func TestGet_RequiresPathAndName(t *testing.T) {
	_, err := Get("", "test")
	if err == nil {
		t.Error("expected error for empty path")
	}
	_, err = Get("/tmp/test", "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestSet_RequiresPathAndName(t *testing.T) {
	_, err := Set("", "test", "value")
	if err == nil {
		t.Error("expected error for empty path")
	}
	_, err = Set("/tmp/test", "", "value")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRemove_RequiresPathAndName(t *testing.T) {
	_, err := Remove("", "test")
	if err == nil {
		t.Error("expected error for empty path")
	}
	_, err = Remove("/tmp/test", "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestList_RequiresPath(t *testing.T) {
	_, err := List("")
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestSetAndGet(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("xattr operations only tested on Linux")
	}

	tmpFile := "/tmp/xattr_test_file"
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	// Set
	result, err := Set(tmpFile, "user.test", "hello")
	if err != nil {
		t.Skip("xattr may not be supported on this filesystem:", err)
	}
	if !result.Changed {
		t.Error("expected changed=true")
	}

	// Get
	getResult, err := Get(tmpFile, "user.test")
	if err != nil {
		t.Fatal(err)
	}
	if getResult.Value != "hello" {
		t.Errorf("expected 'hello', got '%s'", getResult.Value)
	}
	if getResult.Size != 5 {
		t.Errorf("expected size 5, got %d", getResult.Size)
	}

	// List
	listResult, err := List(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, attr := range listResult.Attributes {
		if attr == "user.test" {
			found = true
		}
	}
	if !found {
		t.Error("expected user.test in attribute list")
	}

	// Idempotent set
	result2, err := Set(tmpFile, "user.test", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if result2.Changed {
		t.Error("expected changed=false for same value")
	}

	// Remove
	rmResult, err := Remove(tmpFile, "user.test")
	if err != nil {
		t.Fatal(err)
	}
	if !rmResult.Changed {
		t.Error("expected changed=true on remove")
	}

	// Remove non-existent
	rmResult2, err := Remove(tmpFile, "user.nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if rmResult2.Changed {
		t.Error("expected changed=false for non-existent attr")
	}
}

func TestList_NonexistentFile(t *testing.T) {
	_, err := List("/nonexistent/file/xyz")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestResultFields(t *testing.T) {
	r := Result{
		Path:    "/tmp/test",
		Name:    "user.test",
		Value:   "hello",
		Changed: true,
		Message: "set",
	}
	if r.Path != "/tmp/test" {
		t.Error("path mismatch")
	}
	if r.Name != "user.test" {
		t.Error("name mismatch")
	}
}

func TestGetResultFields(t *testing.T) {
	r := GetResult{
		Path:  "/tmp/test",
		Name:  "user.test",
		Value: "hello",
		Size:  5,
	}
	if r.Size != 5 {
		t.Error("size mismatch")
	}
}

func TestListResultFields(t *testing.T) {
	r := ListResult{
		Path:       "/tmp/test",
		Attributes: []string{"user.a", "user.b"},
		Count:      2,
	}
	if r.Count != 2 {
		t.Error("count mismatch")
	}
}
