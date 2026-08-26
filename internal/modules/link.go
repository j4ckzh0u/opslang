// Package modules implements the user-module system: scripts can pull in
// reusable declarations from other .ops files via `import "./lib/x.ops"`.
//
// Linking happens on the AST before any engine runs: imported modules'
// top-level declarations are spliced into the importing program at the
// import site, so the interpreter, the runner generator and the AOT
// codegen all see one flat program and need no module awareness of their
// own.
//
// Module rules (deliberately strict, for predictable fleets):
//   - a module may declare only fn and let, plus further imports
//   - tasks/ensure/parallel stay in entry-point scripts; a task inside a
//     module is an error naming the file and line
//   - duplicate top-level names across modules are an error, not silent
//     shadowing
//   - cycles are detected and reported with the import chain
package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/parser"
)

type linker struct {
	// declared maps each top-level name to the file that declared it,
	// for collision errors that name both sides.
	declared map[string]string
	// loaded holds absolute paths already merged (import dedup).
	loaded map[string]bool
	// stack is the active import chain, innermost last, for cycle errors.
	stack []string
}

// Link resolves file-module imports in prog. scriptPath is the entry
// script's path (used to resolve relative imports); empty means stdin /
// REPL, resolving imports against the working directory. Returns prog
// unchanged when it contains no file imports.
func Link(prog *ast.Program, scriptPath string) (*ast.Program, error) {
	l := &linker{
		declared: map[string]string{},
		loaded:   map[string]bool{},
	}
	if !hasFileImports(prog.Statements) {
		return prog, nil
	}
	base := "."
	if scriptPath != "" {
		base = filepath.Dir(scriptPath)
	}
	stmts, err := l.link(prog.Statements, base)
	if err != nil {
		return nil, err
	}
	return &ast.Program{Position: prog.Position, Statements: stmts}, nil
}

// isFileModule reports whether an import path refers to a local .ops
// module rather than a declarative standard-library import.
func isFileModule(path string) bool {
	return strings.HasSuffix(path, ".ops")
}

func hasFileImports(stmts []ast.Statement) bool {
	for _, s := range stmts {
		if imp, ok := s.(*ast.ImportStatement); ok && isFileModule(imp.Path) {
			return true
		}
	}
	return false
}

// link splices module declarations into stmts at each import site.
// currentDir resolves the next relative hop.
func (l *linker) link(stmts []ast.Statement, currentDir string) ([]ast.Statement, error) {
	out := make([]ast.Statement, 0, len(stmts))
	for _, s := range stmts {
		switch node := s.(type) {
		case *ast.ImportStatement:
			if !isFileModule(node.Path) {
				out = append(out, s) // declarative import: engines handle it
				continue
			}
			merged, err := l.importModule(node, currentDir)
			if err != nil {
				return nil, err
			}
			out = append(out, merged...)
		case *ast.FnStatement:
			if err := l.declare(node.Name.Name, describe(node.Pos())); err != nil {
				return nil, err
			}
			out = append(out, s)
		case *ast.LetStatement:
			if err := l.declare(node.Name.Name, describe(node.Pos())); err != nil {
				return nil, err
			}
			out = append(out, s)
		default:
			out = append(out, s)
		}
	}
	return out, nil
}

func (l *linker) importModule(imp *ast.ImportStatement, currentDir string) ([]ast.Statement, error) {
	path := imp.Path
	if !filepath.IsAbs(path) {
		path = filepath.Join(currentDir, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("import %q: %w", imp.Path, err)
	}
	for _, active := range l.stack {
		if active == abs {
			chain := append(append([]string{}, l.stack...), abs)
			return nil, fmt.Errorf("import %q: circular import chain: %s",
				imp.Path, strings.Join(chain, " -> "))
		}
	}
	if l.loaded[abs] {
		return nil, nil // already merged once; dedupe
	}

	source, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("import %q: %w", imp.Path, err)
	}
	p := parser.New(string(source), abs)
	mod, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("import %q: parse error: %w", imp.Path, err)
	}
	for _, s := range mod.Statements {
		if err := validateModuleStatement(s, abs); err != nil {
			return nil, fmt.Errorf("import %q: %w", imp.Path, err)
		}
	}

	l.stack = append(l.stack, abs)
	merged, err := l.link(mod.Statements, filepath.Dir(abs))
	l.stack = l.stack[:len(l.stack)-1]
	if err != nil {
		return nil, err
	}
	l.loaded[abs] = true
	return merged, nil
}

// validateModuleStatement rejects entry-point-only constructs inside
// modules, with file:line context.
func validateModuleStatement(s ast.Statement, file string) error {
	switch node := s.(type) {
	case *ast.FnStatement, *ast.LetStatement, *ast.ImportStatement:
		return nil
	case *ast.TaskStatement:
		return posErr(file, node.Pos(), "tasks belong in entry-point scripts, not modules")
	case *ast.EnsureStatement:
		return posErr(file, node.Pos(), "ensure blocks belong in entry-point scripts, not modules")
	default:
		return posErr(file, node.Pos(), "only fn, let and import are allowed at module top level (got %T)", s)
	}
}

func (l *linker) declare(name, where string) error {
	if prev, exists := l.declared[name]; exists {
		return fmt.Errorf("duplicate top-level name %q defined in %s and %s", name, prev, where)
	}
	l.declared[name] = where
	return nil
}

func describe(pos ast.Position) string {
	return fmt.Sprintf("line %d", pos.Line)
}

func posErr(file string, pos ast.Position, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s:%d:%d: %s", filepath.Base(file), pos.Line, pos.Column, msg)
}
