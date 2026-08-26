package htpasswd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetMissingArgs(t *testing.T) {
	r := Set("", "", "", false)
	if r.Status != "failed" {
		t.Error("expected failure for empty path")
	}
	r2 := Set("/tmp/test", "", "", false)
	if r2.Status != "failed" {
		t.Error("expected failure for empty username")
	}
}

func TestRemoveMissingArgs(t *testing.T) {
	r := Remove("", "")
	if r.Status != "failed" {
		t.Error("expected failure for missing args")
	}
}

func TestInfoMissingPath(t *testing.T) {
	r := Info("")
	if r.Status != "failed" {
		t.Error("expected failure for empty path")
	}
}

func TestSetAndRemove(t *testing.T) {
	// The Set/Remove implementation shells out to the apache htpasswd
	// binary. GitHub's fat runner image ships it, minimal hosts do not -
	// skip honestly instead of failing on the missing tool.
	if _, err := exec.LookPath("htpasswd"); err != nil {
		t.Skip("requires the htpasswd binary (apache2-utils), not installed")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, ".htpasswd")

	// Write initial content
	os.WriteFile(path, []byte(""), 0644)

	r := Set(path, "testuser", "testpass", false)
	if r.Status != "success" {
		t.Errorf("expected success, got %s: %s", r.Status, r.Error)
	}

	info := Info(path)
	if info.Status != "success" || len(info.Users) != 1 || info.Users[0] != "testuser" {
		t.Errorf("expected 1 user, got %v", info.Users)
	}

	// Remove
	rr := Remove(path, "testuser")
	if rr.Status != "success" || !rr.Changed {
		t.Errorf("expected changed, got status=%s changed=%v", rr.Status, rr.Changed)
	}

	info2 := Info(path)
	if len(info2.Users) != 0 {
		t.Errorf("expected 0 users after remove, got %v", info2.Users)
	}
}

func TestHashSHA1(t *testing.T) {
	h := HashSHA1("password")
	if !strings.HasPrefix(h, "{SHA}") {
		t.Error("expected {SHA} prefix")
	}
}
