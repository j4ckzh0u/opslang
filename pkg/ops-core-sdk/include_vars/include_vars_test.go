package include_vars

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	r := Load("")
	if r.Error == "" {
		t.Error("expected error for empty path")
	}
}

func TestLoadYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vars.yml")
	os.WriteFile(f, []byte("name: test\nport: 8080\n"), 0644)

	r := Load(f)
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Count == 0 {
		t.Error("expected variables loaded")
	}
	if r.Changed != true {
		t.Error("expected changed")
	}
}

func TestLoadJSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vars.json")
	os.WriteFile(f, []byte(`{"name":"test","port":8080}`), 0644)

	r := Load(f)
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Count < 1 {
		t.Error("expected variables loaded")
	}
}

func TestLoadNonExistent(t *testing.T) {
	r := Load("/nonexistent/file.yml")
	if r.Error == "" {
		t.Error("expected error for non-existent file")
	}
}

func TestGet(t *testing.T) {
	mu.Lock()
	vars = map[string]string{}
	mu.Unlock()

	dir := t.TempDir()
	f := filepath.Join(dir, "vars.yml")
	os.WriteFile(f, []byte("mykey: myvalue\n"), 0644)
	Load(f)

	v, ok := Get("mykey")
	if !ok || v != "myvalue" {
		t.Errorf("unexpected: %v %v", v, ok)
	}
}
