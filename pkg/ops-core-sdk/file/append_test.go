package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendCreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.log")
	res, err := Append(path, "hello\n")
	if err != nil {
		t.Fatalf("Append error: %v", err)
	}
	if res.Size != 6 {
		t.Errorf("Size = %d, want 6", res.Size)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello\n" {
		t.Errorf("content = %q, want %q", string(data), "hello\n")
	}
}

func TestAppendPreservesExistingContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "append.log")
	if err := os.WriteFile(path, []byte("first\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Append(path, "second\n")
	if err != nil {
		t.Fatalf("Append error: %v", err)
	}
	if res.Size != 13 {
		t.Errorf("Size = %d, want 13", res.Size)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "first\nsecond\n" {
		t.Errorf("content = %q, want %q", string(data), "first\nsecond\n")
	}
}

func TestAppendError(t *testing.T) {
	// Append to a path whose parent does not exist must fail.
	_, err := Append(filepath.Join(t.TempDir(), "missing-dir", "f.txt"), "x")
	if err == nil {
		t.Error("Append to a path with a missing parent should fail")
	}
}

func TestTemplateRendersVariables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.conf")
	content := "host={{ host }}\nport={{port}}\nuser={{ user }}\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Template(path, map[string]interface{}{
		"host": "web-01",
		"port": 8080,
	})
	if err != nil {
		t.Fatalf("Template error: %v", err)
	}

	want := "host=web-01\nport=8080\nuser={{ user }}\n"
	if res.Content != want {
		t.Errorf("Content = %q, want %q", res.Content, want)
	}
	if res.Size != int64(len(want)) {
		t.Errorf("Size = %d, want %d", res.Size, len(want))
	}

	// The source file must be untouched.
	src, _ := os.ReadFile(path)
	if string(src) != content {
		t.Errorf("source file was modified: %q", string(src))
	}
}

func TestTemplateMissingFile(t *testing.T) {
	_, err := Template(filepath.Join(t.TempDir(), "nope.conf"), nil)
	if err == nil {
		t.Error("Template on a missing file should fail")
	}
}

func TestTemplateNoPlaceholders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.txt")
	if err := os.WriteFile(path, []byte("plain text"), 0644); err != nil {
		t.Fatal(err)
	}
	res, err := Template(path, map[string]interface{}{"k": "v"})
	if err != nil {
		t.Fatalf("Template error: %v", err)
	}
	if res.Content != "plain text" {
		t.Errorf("Content = %q", res.Content)
	}
}
