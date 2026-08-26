package modules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/parser"
)

// parseProgram is the test helper: parse source as a file at path.
func parseProgram(t *testing.T, source, path string) *ast.Program {
	t.Helper()
	p := parser.New(source, path)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return prog
}

func writeModule(t *testing.T, dir, rel, content string) string {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fnNames(prog *ast.Program) []string {
	var names []string
	for _, s := range prog.Statements {
		if fn, ok := s.(*ast.FnStatement); ok {
			names = append(names, fn.Name.Name)
		}
	}
	return names
}

// TestLinkSplicesModuleDeclarations: the core contract - imported fns and
// lets appear in the linked program, the import statement disappears.
func TestLinkSplicesModuleDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "lib/helpers.ops", `
let max_retries = 3
fn double(n) { return n * 2 }
`)
	main := parseProgram(t, `
import "./lib/helpers.ops"
let start = double(max_retries)
`, filepath.Join(dir, "main.ops"))

	linked, err := Link(main, filepath.Join(dir, "main.ops"))
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	got := fnNames(linked)
	if len(got) != 1 || got[0] != "double" {
		t.Errorf("linked fn names = %v, want [double]", got)
	}
	for _, s := range linked.Statements {
		if imp, ok := s.(*ast.ImportStatement); ok && isFileModule(imp.Path) {
			t.Errorf("file import survived linking: %q", imp.Path)
		}
	}
	letCount := 0
	for _, s := range linked.Statements {
		if _, ok := s.(*ast.LetStatement); ok {
			letCount++
		}
	}
	if letCount != 2 {
		t.Errorf("linked let count = %d, want 2 (module constant + main)", letCount)
	}
}

// TestLinkNestedImports: a module importing another module resolves
// relative to the IMPORTING file.
func TestLinkNestedImports(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "lib/base.ops", "fn ping() { return 1 }\n")
	writeModule(t, dir, "lib/mid.ops", `
import "./base.ops"
fn pong() { return ping() + 1 }
`)
	main := parseProgram(t, `import "./lib/mid.ops"`, filepath.Join(dir, "main.ops"))
	linked, err := Link(main, filepath.Join(dir, "main.ops"))
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	names := map[string]bool{}
	for _, n := range fnNames(linked) {
		names[n] = true
	}
	if !names["ping"] || !names["pong"] {
		t.Errorf("nested module declarations missing: %v", names)
	}
}

// TestLinkCycleDetected with an explicit chain message.
func TestLinkCycleDetected(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "a.ops", `import "./b.ops"`)
	writeModule(t, dir, "b.ops", `import "./a.ops"`)
	main := parseProgram(t, `import "./a.ops"`, filepath.Join(dir, "main.ops"))
	_, err := Link(main, filepath.Join(dir, "main.ops"))
	if err == nil {
		t.Fatal("circular import must error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error should mention circularity: %v", err)
	}
}

// TestLinkRejectsTaskInModule: tasks are entry-point constructs.
func TestLinkRejectsTaskInModule(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "bad.ops", `
task "deploy" on "host1" {
	print("nope")
}
`)
	main := parseProgram(t, `import "./bad.ops"`, filepath.Join(dir, "main.ops"))
	_, err := Link(main, filepath.Join(dir, "main.ops"))
	if err == nil {
		t.Fatal("task in module must error")
	}
	if !strings.Contains(err.Error(), "bad.ops") || !strings.Contains(err.Error(), "entry-point") {
		t.Errorf("error should name file and reason: %v", err)
	}
}

// TestLinkDuplicateNamesAcrossModules: no silent shadowing between modules.
func TestLinkDuplicateNamesAcrossModules(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "a.ops", "fn same() { return 1 }")
	writeModule(t, dir, "b.ops", "fn same() { return 2 }")
	main := parseProgram(t, `
import "./a.ops"
import "./b.ops"
`, filepath.Join(dir, "main.ops"))
	_, err := Link(main, filepath.Join(dir, "main.ops"))
	if err == nil {
		t.Fatal("duplicate top-level names across modules must error")
	}
	if !strings.Contains(err.Error(), `"same"`) {
		t.Errorf("error should name the colliding symbol: %v", err)
	}
}

// TestLinkDedupesRepeatedImport: importing the same file twice merges once.
func TestLinkDedupesRepeatedImport(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "lib.ops", "fn once() { return 1 }")
	main := parseProgram(t, `
import "./lib.ops"
import "./lib.ops"
`, filepath.Join(dir, "main.ops"))
	linked, err := Link(main, filepath.Join(dir, "main.ops"))
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	if got := len(fnNames(linked)); got != 1 {
		t.Errorf("deduped fn count = %d, want 1", got)
	}
}

// TestLinkKeepsDeclarativeImports: standard-library imports stay for the
// engines to treat as before.
func TestLinkKeepsDeclarativeImports(t *testing.T) {
	main := parseProgram(t, `import "sys"`, filepath.Join(t.TempDir(), "main.ops"))
	linked, err := Link(main, filepath.Join(t.TempDir(), "main.ops"))
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	if len(linked.Statements) != 1 {
		t.Fatalf("declarative import was touched: %d statements", len(linked.Statements))
	}
	if _, ok := linked.Statements[0].(*ast.ImportStatement); !ok {
		t.Error("declarative import should survive unchanged")
	}
}

// TestNoFileImportsReturnsSameProgram: zero-cost fast path; callers may
// compare pointer identity to skip work.
func TestNoFileImportsReturnsSameProgram(t *testing.T) {
	main := parseProgram(t, `let x = 1`, filepath.Join(t.TempDir(), "main.ops"))
	linked, err := Link(main, "whatever.ops")
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}
	if linked != main {
		t.Error("program without file imports should return the identical pointer")
	}
}
