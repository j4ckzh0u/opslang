package unarchive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestUnarchiveMissing(t *testing.T) {
	r := Unarchive("", "", "", "", "", "")
	if r.Error == "" {
		t.Error("expected error for missing src/dest")
	}
}

func TestUnarchiveUnsupported(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.xyz")
	os.WriteFile(f, []byte("data"), 0644)
	r := Unarchive(f, dir, "", "", "", "")
	if r.Error == "" {
		t.Error("expected error for unsupported format")
	}
}

func TestUnarchiveTarGz(t *testing.T) {
	dir := t.TempDir()
	tgz := filepath.Join(dir, "test.tar.gz")

	// create tar.gz
	f, err := os.Create(tgz)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	content := []byte("hello world")
	hdr := &tar.Header{Name: "test.txt", Mode: 0644, Size: int64(len(content))}
	tw.WriteHeader(hdr)
	tw.Write(content)
	tw.Close()
	gz.Close()
	f.Close()

	dest := filepath.Join(dir, "out")
	r := Unarchive(tgz, dest, "", "", "", "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}
	if _, err := os.Stat(filepath.Join(dest, "test.txt")); err != nil {
		t.Error("expected test.txt to exist")
	}
}

func TestUnarchiveZip(t *testing.T) {
	dir := t.TempDir()
	zipf := filepath.Join(dir, "test.zip")

	f, err := os.Create(zipf)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("hello.txt")
	w.Write([]byte("world"))
	zw.Close()
	f.Close()

	dest := filepath.Join(dir, "out")
	r := Unarchive(zipf, dest, "", "", "", "")
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if !r.Changed {
		t.Error("expected changed")
	}
}

func TestUnarchiveCreates(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	os.WriteFile(marker, []byte("exists"), 0644)
	r := Unarchive("dummy.tar", dir, "", "", "", marker)
	if r.Changed {
		t.Error("creates file exists, should not change")
	}
}
