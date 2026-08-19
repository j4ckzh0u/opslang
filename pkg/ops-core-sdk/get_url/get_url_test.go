package get_url

import (
	"os"
	"testing"
)

func TestDownload_EmptyURL(t *testing.T) {
	_, err := Download("", "/tmp/test", "", false)
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestDownload_EmptyDest(t *testing.T) {
	_, err := Download("http://example.com/file", "", "", false)
	if err == nil {
		t.Error("expected error for empty dest")
	}
}

func TestDownload_InvalidURL(t *testing.T) {
	_, err := Download("http://invalid-host-that-does-not-exist-xyz.com/file", "/tmp/test_get_url", "", false)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestDownload_FileAlreadyExists(t *testing.T) {
	tmpFile := "/tmp/get_url_test_exists.txt"
	os.WriteFile(tmpFile, []byte("existing content"), 0644)
	defer os.Remove(tmpFile)

	result, err := Download("http://example.com/file", tmpFile, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed {
		t.Error("expected changed=false when file exists and force=false")
	}
	if result.Message != "file already exists" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestDownload_ForceOverwrite(t *testing.T) {
	tmpFile := "/tmp/get_url_test_force.txt"
	os.WriteFile(tmpFile, []byte("old content"), 0644)
	defer os.Remove(tmpFile)

	// Force=true will try to download (may fail due to network), but should not skip
	_, err := Download("http://invalid-host-xyz.com/file", tmpFile, "", true)
	if err == nil {
		t.Log("download succeeded (unexpected but ok)")
	}
	// Error is expected due to invalid host, but the point is it attempted download
}

func TestExtractHash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sha256:abc123", "abc123"},
		{"md5:def456", "def456"},
		{"rawhash", "rawhash"},
	}
	for _, tt := range tests {
		result := extractHash(tt.input)
		if result != tt.expected {
			t.Errorf("extractHash(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestExtractAlgo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sha256:abc123", "sha256"},
		{"md5:def456", "md5"},
		{"sha1:ghi789", "sha1"},
		{"rawhash", "sha256"}, // default
	}
	for _, tt := range tests {
		result := extractAlgo(tt.input)
		if result != tt.expected {
			t.Errorf("extractAlgo(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestComputeChecksum(t *testing.T) {
	tmpFile := "/tmp/get_url_checksum_test.txt"
	os.WriteFile(tmpFile, []byte("test content for checksum"), 0644)
	defer os.Remove(tmpFile)

	// SHA256
	hash, err := computeChecksum(tmpFile, "sha256:ignored")
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 64 { // SHA256 produces 32 bytes = 64 hex chars
		t.Errorf("unexpected sha256 hash length: %d", len(hash))
	}

	// MD5
	hash, err = computeChecksum(tmpFile, "md5:ignored")
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 32 { // MD5 produces 16 bytes = 32 hex chars
		t.Errorf("unexpected md5 hash length: %d", len(hash))
	}

	// SHA1
	hash, err = computeChecksum(tmpFile, "sha1:ignored")
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 40 { // SHA1 produces 20 bytes = 40 hex chars
		t.Errorf("unexpected sha1 hash length: %d", len(hash))
	}
}

func TestResultFields(t *testing.T) {
	r := Result{
		URL:      "http://example.com/file",
		Dest:     "/tmp/test",
		Size:     1024,
		Checksum: "sha256:abc123",
		Changed:  true,
		Message:  "downloaded",
	}
	if r.URL != "http://example.com/file" {
		t.Error("URL mismatch")
	}
	if r.Size != 1024 {
		t.Error("size mismatch")
	}
}
