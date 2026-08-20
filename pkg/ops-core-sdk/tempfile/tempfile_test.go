package tempfile

import (
	"os"
	"testing"
)

func TestCreateFile(t *testing.T) {
	r := CreateFile("test-", ".tmp", "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.State != "file" {
		t.Errorf("expected file state, got %s", r.State)
	}
	if _, err := os.Stat(r.Path); err != nil {
		t.Errorf("file should exist: %v", err)
	}
	os.Remove(r.Path)
}

func TestCreateDir(t *testing.T) {
	r := CreateDir("test-", "", "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.State != "directory" {
		t.Errorf("expected directory state, got %s", r.State)
	}
	info, err := os.Stat(r.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
	os.RemoveAll(r.Path)
}

func TestDelete(t *testing.T) {
	r := CreateFile("del-", ".tmp", "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	dr := Delete(r.Path)
	if dr.Error != "" {
		t.Fatal(dr.Error)
	}
	if _, err := os.Stat(r.Path); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestDeleteEmpty(t *testing.T) {
	r := Delete("")
	if r.Error == "" {
		t.Error("expected error for empty path")
	}
}
