package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead_Success(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "hello.txt")
	want := "hello world"
	if err := os.WriteFile(p, []byte(want), 0644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	got, err := Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Path != p {
		t.Errorf("Path = %q, want %q", got.Path, p)
	}
	if got.Content != want {
		t.Errorf("Content = %q, want %q", got.Content, want)
	}
	if got.Size != int64(len(want)) {
		t.Errorf("Size = %d, want %d", got.Size, len(want))
	}
}

func TestRead_Empty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(p, []byte{}, 0644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	got, err := Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Content != "" {
		t.Errorf("Content = %q, want empty", got.Content)
	}
	if got.Size != 0 {
		t.Errorf("Size = %d, want 0", got.Size)
	}
}

func TestRead_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nonexistent.txt")
	_, err := Read(p)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestWrite_Success(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.txt")
	content := "some content"

	got, err := Write(p, content)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got.Path != p {
		t.Errorf("Path = %q, want %q", got.Path, p)
	}
	if got.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", got.Size, len(content))
	}

	// Verify actual file content
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content = %q, want %q", string(data), content)
	}

	// Verify permissions
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0644 {
		t.Errorf("permissions = %04o, want 0644", perm)
	}
}

func TestWrite_Overwrite(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "overwrite.txt")

	if _, err := Write(p, "first"); err != nil {
		t.Fatalf("Write first: %v", err)
	}
	got, err := Write(p, "second")
	if err != nil {
		t.Fatalf("Write second: %v", err)
	}
	if got.Size != 6 { // len("second")
		t.Errorf("Size = %d, want 6", got.Size)
	}

	data, _ := os.ReadFile(p)
	if string(data) != "second" {
		t.Errorf("file content = %q, want %q", string(data), "second")
	}
}

func TestWrite_InvalidPath(t *testing.T) {
	// Writing to a directory that doesn't exist should fail
	_, err := Write("/nonexistent_dir_xyz/file.txt", "data")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestCopy_Success(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	content := "copy this"

	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatalf("setup src: %v", err)
	}

	got, err := Copy(src, dst)
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if got.Src != src {
		t.Errorf("Src = %q, want %q", got.Src, src)
	}
	if got.Dst != dst {
		t.Errorf("Dst = %q, want %q", got.Dst, dst)
	}
	if got.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", got.Size, len(content))
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != content {
		t.Errorf("dst content = %q, want %q", string(data), content)
	}
}

func TestCopy_PreservesMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "exec.txt")
	dst := filepath.Join(dir, "exec_copy.txt")

	if err := os.WriteFile(src, []byte("data"), 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := Copy(src, dst); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0755 {
		t.Errorf("permissions = %04o, want 0755", perm)
	}
}

func TestCopy_SrcNotFound(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "missing.txt")
	dst := filepath.Join(dir, "dst.txt")
	_, err := Copy(src, dst)
	if err == nil {
		t.Fatal("expected error for missing src")
	}
}

func TestMove_Success(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "orig.txt")
	dst := filepath.Join(dir, "moved.txt")

	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Move(src, dst)
	if err != nil {
		t.Fatalf("Move: %v", err)
	}
	if got.Src != src {
		t.Errorf("Src = %q, want %q", got.Src, src)
	}
	if got.Dst != dst {
		t.Errorf("Dst = %q, want %q", got.Dst, dst)
	}

	// Source should be gone
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source still exists after move")
	}

	// Destination should exist
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("dst content = %q, want %q", string(data), "data")
	}
}

func TestMove_SrcNotFound(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "missing.txt")
	dst := filepath.Join(dir, "dst.txt")
	_, err := Move(src, dst)
	if err == nil {
		t.Fatal("expected error for missing src")
	}
}

func TestDelete_Existing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "to_delete.txt")
	if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Delete(p)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.Path != p {
		t.Errorf("Path = %q, want %q", got.Path, p)
	}
	if !got.Existed {
		t.Error("Existed = false, want true")
	}

	// File should be gone
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Error("file still exists after delete")
	}
}

func TestDelete_NonExisting(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "never_existed.txt")

	got, err := Delete(p)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.Existed {
		t.Error("Existed = true, want false")
	}
}

func TestDelete_Directory(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Delete(subdir)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !got.Existed {
		t.Error("Existed = false, want true for directory")
	}
}

func TestExists_RegularFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Exists(p)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !got.Exists {
		t.Error("Exists = false, want true")
	}
	if got.IsDir {
		t.Error("IsDir = true, want false")
	}
}

func TestExists_Directory(t *testing.T) {
	dir := t.TempDir()
	got, err := Exists(dir)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !got.Exists {
		t.Error("Exists = false, want true")
	}
	if !got.IsDir {
		t.Error("IsDir = false, want true")
	}
}

func TestExists_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nope.txt")

	got, err := Exists(p)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if got.Exists {
		t.Error("Exists = true, want false")
	}
	if got.IsDir {
		t.Error("IsDir = true, want false")
	}
}

func TestStat_File(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "stat_me.txt")
	content := "stat content"
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Stat(p)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got.Path != p {
		t.Errorf("Path = %q, want %q", got.Path, p)
	}
	if got.Name != "stat_me.txt" {
		t.Errorf("Name = %q, want %q", got.Name, "stat_me.txt")
	}
	if got.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", got.Size, len(content))
	}
	if got.Mode != "0644" {
		t.Errorf("Mode = %q, want %q", got.Mode, "0644")
	}
	if got.ModTime == 0 {
		t.Error("ModTime = 0, expected non-zero")
	}
	if got.IsDir {
		t.Error("IsDir = true, want false")
	}
}

func TestStat_Directory(t *testing.T) {
	dir := t.TempDir()

	got, err := Stat(dir)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !got.IsDir {
		t.Error("IsDir = false, want true")
	}
	if got.Size == 0 {
		t.Error("Size = 0 for directory")
	}
}

func TestStat_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "missing.txt")
	_, err := Stat(p)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestChmod_Success(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "chmod_me.txt")
	if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Chmod(p, 0755)
	if err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	if got.Path != p {
		t.Errorf("Path = %q, want %q", got.Path, p)
	}
	if got.Mode != "0755" {
		t.Errorf("Mode = %q, want %q", got.Mode, "0755")
	}

	// Verify actual permissions
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0755 {
		t.Errorf("actual permissions = %04o, want 0755", perm)
	}
}

func TestChmod_NotFound(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nope.txt")
	_, err := Chmod(p, 0644)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestChmod_Executable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\necho hi"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got, err := Chmod(p, 0700)
	if err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	if got.Mode != "0700" {
		t.Errorf("Mode = %q, want %q", got.Mode, "0700")
	}
}

// TestIntegrationChain tests a sequence of operations in a realistic workflow.
func TestIntegrationChain(t *testing.T) {
	dir := t.TempDir()

	// Write a file
	p := filepath.Join(dir, "chain.txt")
	wr, err := Write(p, "initial")
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if wr.Size != 7 {
		t.Errorf("Write Size = %d, want 7", wr.Size)
	}

	// Read it back
	rc, err := Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if rc.Content != "initial" {
		t.Errorf("Read Content = %q, want %q", rc.Content, "initial")
	}

	// Copy it
	cp := filepath.Join(dir, "chain_copy.txt")
	cr, err := Copy(p, cp)
	if err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if cr.Size != 7 {
		t.Errorf("Copy Size = %d, want 7", cr.Size)
	}

	// Stat the copy
	si, err := Stat(cp)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if si.IsDir {
		t.Error("copy should not be a directory")
	}

	// Chmod the copy
	cm, err := Chmod(cp, 0600)
	if err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	if cm.Mode != "0600" {
		t.Errorf("Chmod Mode = %q, want %q", cm.Mode, "0600")
	}

	// Move the copy
	mv := filepath.Join(dir, "chain_moved.txt")
	mr, err := Move(cp, mv)
	if err != nil {
		t.Fatalf("Move: %v", err)
	}
	if mr.Dst != mv {
		t.Errorf("Move Dst = %q, want %q", mr.Dst, mv)
	}

	// Exists should find the moved file
	ex, err := Exists(mv)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ex.Exists {
		t.Error("moved file should exist")
	}

	// Delete the original
	dr, err := Delete(p)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !dr.Existed {
		t.Error("original should have existed")
	}

	// Delete again — should report not existed
	dr2, err := Delete(p)
	if err != nil {
		t.Fatalf("Delete again: %v", err)
	}
	if dr2.Existed {
		t.Error("should not exist on second delete")
	}
}
