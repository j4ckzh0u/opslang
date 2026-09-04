package file

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGzipFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	want := strings.Repeat("ops-lang\n", 128)
	if err := os.WriteFile(source, []byte(want), 0600); err != nil {
		t.Fatal(err)
	}
	compressed, err := gzipFile(source)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(compressed)
	destination := filepath.Join(dir, "destination")
	if err := gunzipFile(compressed, destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("round trip mismatch: got %q", string(got))
	}
}

func TestGzipFileEmptyInput(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "empty")
	if err := os.WriteFile(source, nil, 0600); err != nil {
		t.Fatal(err)
	}
	compressed, err := gzipFile(source)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(compressed)
	destination := filepath.Join(dir, "destination")
	if err := gunzipFile(compressed, destination); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 0 {
		t.Fatalf("empty round trip size = %d", info.Size())
	}
}

func TestGzipFileIncompressibleContent(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "random")
	want := make([]byte, 32*1024)
	if _, err := rand.Read(want); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, want, 0600); err != nil {
		t.Fatal(err)
	}
	compressed, err := gzipFile(source)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(compressed)
	destination := filepath.Join(dir, "destination")
	if err := gunzipFile(compressed, destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("incompressible content changed during round trip")
	}
}

func TestGunzipFileRejectsInvalidInput(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "invalid")
	if err := os.WriteFile(source, []byte("plain text"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := gunzipFile(source, filepath.Join(dir, "destination")); err == nil {
		t.Fatal("expected invalid gzip error")
	}
}
