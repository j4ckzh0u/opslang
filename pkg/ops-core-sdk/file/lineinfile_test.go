package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLineInFile_PresentExactAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "line3", true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if content != "line1\nline2\nline3\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestLineInFile_PresentExactAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "line2", true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Changed {
		t.Fatal("expected changed=false when line already exists")
	}
}

func TestLineInFile_PresentRegexpReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("foo=old\nbar=keep\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "foo=new", true, "^foo=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if content != "foo=new\nbar=keep\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestLineInFile_PresentRegexpNoMatchAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("bar=keep\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "foo=new", true, "^foo=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if content != "bar=keep\nfoo=new\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestLineInFile_AbsentExact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "line2", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if content != "line1\nline3\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestLineInFile_AbsentRegexp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("foo=old\nbar=keep\nfoo=other\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "", false, "^foo=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Changed {
		t.Fatal("expected changed=true")
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if content != "bar=keep\n" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestLineInFile_AbsentNotPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := LineInFile(path, "line2", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Changed {
		t.Fatal("expected changed=false when line not present")
	}
}

func TestLineInFile_InvalidRegexp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")
	if err := os.WriteFile(path, []byte("line1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LineInFile(path, "line", true, "[invalid")
	if err == nil {
		t.Fatal("expected error for invalid regexp")
	}
}
