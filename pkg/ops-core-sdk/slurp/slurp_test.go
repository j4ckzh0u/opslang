package slurp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncodeEmptyPath(t *testing.T) {
	r := Encode("")
	if r.Status != "failed" {
		t.Error("expected failure for empty path")
	}
}

func TestEncodeNonExistent(t *testing.T) {
	r := Encode("/nonexistent/path")
	if r.Status != "failed" {
		t.Error("expected failure for non-existent file")
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")
	content := "hello world\nline 2\n"
	os.WriteFile(path, []byte(content), 0644)

	// Encode
	r := Encode(path)
	if r.Status != "success" {
		t.Fatalf("encode failed: %s", r.Error)
	}
	if r.Encoding != "base64" {
		t.Error("expected base64 encoding")
	}
	if r.Content == "" {
		t.Error("expected non-empty content")
	}

	// Decode to new file
	destPath := filepath.Join(tmpDir, "decoded.txt")
	r2 := Decode(r.Content, destPath)
	if r2.Status != "success" {
		t.Fatalf("decode failed: %s", r2.Error)
	}

	// Verify content
	data, _ := os.ReadFile(destPath)
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestDecodeEmpty(t *testing.T) {
	r := Decode("", "")
	if r.Status != "failed" {
		t.Error("expected failure for empty content")
	}
}

func TestDecodeInvalid(t *testing.T) {
	r := Decode("not-valid-base64!!!", "")
	if r.Status != "failed" {
		t.Error("expected failure for invalid base64")
	}
}
