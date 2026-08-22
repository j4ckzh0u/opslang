package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureDirectoryCreatesThenNoOps(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "data", "opslang")

	first, err := Ensure(dir, "directory", "0755")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !first.Changed || first.Type != "directory" {
		t.Errorf("create must report changed=true type=directory, got %+v", first)
	}
	if len(first.Actions) != 1 || first.Actions[0] != "mkdir" {
		t.Errorf("actions must be [mkdir], got %v", first.Actions)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("directory must really exist: %v", err)
	}
	if got := info.Mode().Perm(); got != os.FileMode(0755) {
		t.Errorf("mode must be 0755, got %04o", got)
	}

	second, err := Ensure(dir, "directory", "0755")
	if err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if second.Changed || len(second.Actions) != 0 {
		t.Errorf("converged re-run must be a no-op, got %+v", second)
	}
	if second.Message != "directory already up to date" {
		t.Errorf("unexpected message %q", second.Message)
	}
}

func TestEnsureDirectoryConvergesMode(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "shared")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}

	res, err := Ensure(dir, "directory", "0775")
	if err != nil {
		t.Fatalf("mode converge: %v", err)
	}
	if !res.Changed || res.Mode != "0775" {
		t.Errorf("mode drift must converge, got %+v", res)
	}
	info, _ := os.Stat(dir)
	if got := info.Mode().Perm(); got != os.FileMode(0775) {
		t.Errorf("real mode must be 0775, got %04o", got)
	}

	again, err := Ensure(dir, "directory", "0775")
	if err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if again.Changed {
		t.Error("second mode converge must be changed=false")
	}
}

func TestEnsureDirectoryRefusesFileConflict(t *testing.T) {
	base := t.TempDir()
	p := filepath.Join(base, "conflict")
	if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Ensure(p, "directory", "")
	if err == nil {
		t.Fatal("must refuse to replace a file with a directory")
	}
	if _, statErr := os.Stat(p); statErr != nil {
		t.Error("the conflicting file must survive")
	}
}

func TestEnsureFileRequiresExisting(t *testing.T) {
	base := t.TempDir()

	if _, err := Ensure(filepath.Join(base, "missing"), "file", ""); err == nil {
		t.Fatal("state=file must not create (Ansible semantics)")
	}
}

func TestEnsureFileConvergesModeOnly(t *testing.T) {
	base := t.TempDir()
	p := filepath.Join(base, "app.conf")
	if err := os.WriteFile(p, []byte("k=v"), 0600); err != nil {
		t.Fatal(err)
	}

	res, err := Ensure(p, "file", "0644")
	if err != nil {
		t.Fatalf("file mode converge: %v", err)
	}
	if !res.Changed || res.Mode != "0644" {
		t.Errorf("expected mode convergence, got %+v", res)
	}
	again, err := Ensure(p, "file", "0644")
	if err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if again.Changed {
		t.Error("second run must be changed=false")
	}
	if again.Message != "file already up to date" {
		t.Errorf("unexpected message %q", again.Message)
	}
}

func TestEnsureTouchCreatesEmptyFile(t *testing.T) {
	base := t.TempDir()
	p := filepath.Join(base, "flag")

	res, err := Ensure(p, "touch", "0640")
	if err != nil {
		t.Fatalf("touch: %v", err)
	}
	if !res.Changed || res.Type != "file" {
		t.Errorf("touch must create, got %+v", res)
	}
	data, err := os.ReadFile(p)
	if err != nil || len(data) != 0 {
		t.Errorf("created file must exist and be empty: %v", err)
	}
	info, _ := os.Stat(p)
	if got := info.Mode().Perm(); got != os.FileMode(0640) {
		t.Errorf("mode must be 0640, got %04o", got)
	}

	// Deliberate divergence from Ansible documented at Ensure: an existing
	// file is left untouched so convergence runs stay idempotent.
	again, err := Ensure(p, "touch", "0640")
	if err != nil {
		t.Fatalf("touch re-run: %v", err)
	}
	if again.Changed {
		t.Error("touch on existing file must be changed=false")
	}
}

func TestEnsureAbsentRemovesFileAndDirectoryTree(t *testing.T) {
	base := t.TempDir()
	tree := filepath.Join(base, "tree")
	if err := os.MkdirAll(filepath.Join(tree, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tree, "nested", "f.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Ensure(tree, "absent", "")
	if err != nil {
		t.Fatalf("absent: %v", err)
	}
	if !res.Changed || res.Type != "" {
		t.Errorf("absent must remove, got %+v", res)
	}
	if _, statErr := os.Stat(tree); !os.IsNotExist(statErr) {
		t.Error("tree must really be gone")
	}

	again, err := Ensure(tree, "absent", "")
	if err != nil {
		t.Fatalf("absent re-run: %v", err)
	}
	if again.Changed || again.Message != "path already absent" {
		t.Errorf("second absent must be a no-op, got %+v", again)
	}
}

func TestEnsureRejectsInvalidInput(t *testing.T) {
	base := t.TempDir()
	cases := []struct {
		path, state, mode string
		wantErr           string
	}{
		{filepath.Join(base, "x"), "", "", "state is required"},
		{filepath.Join(base, "x"), "symlink", "", "invalid state"},
		{filepath.Join(base, "x"), "file", "0999", "invalid mode"},
		{filepath.Join(base, "x"), "file", "rw-r--r--", "invalid mode"},
		{filepath.Join(base, "x"), "file", "1777", "invalid mode"},
	}
	for _, tc := range cases {
		res, err := Ensure(tc.path, tc.state, tc.mode)
		if err == nil {
			t.Errorf("Ensure(%q,%q,%q) must fail", tc.path, tc.state, tc.mode)
			continue
		}
		if !strings.Contains(res.Error, tc.wantErr) && !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("error must mention %q, got %q / %q", tc.wantErr, res.Error, err)
		}
	}
}

func TestEnsureResultJSONContract(t *testing.T) {
	res := EnsureResult{Path: "/x", State: "directory", Type: "directory", Mode: "0755", Changed: false, Actions: []string{}, Message: "m"}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"path"`, `"state"`, `"type"`, `"mode"`, `"changed"`, `"actions"`, `"message"`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON must contain %s, got %s", key, b)
		}
	}
}
