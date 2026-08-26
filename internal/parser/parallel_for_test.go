package parser

import (
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

// TestParseParallelFor pins the fan-out form: parallel for <var> in
// <list> { body }. The block form must keep parsing too.
func TestParseParallelFor(t *testing.T) {
	src := `parallel for h in hosts {
	let r = check(h)
	print(r)
}`
	p := New(src, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(prog.Statements))
	}
	pf, ok := prog.Statements[0].(*ast.ParallelForStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ParallelForStatement", prog.Statements[0])
	}
	if pf.Var.Name != "h" {
		t.Errorf("loop var = %q, want h", pf.Var.Name)
	}
	if pf.List == nil || pf.Body == nil || len(pf.Body.Statements) != 2 {
		t.Errorf("list/body not parsed correctly: list=%v body=%d", pf.List, len(pf.Body.Statements))
	}

	// Block form still parses.
	block := `parallel {
	print(1)
}`
	prog2, err := New(block, "b.ops").Parse()
	if err != nil {
		t.Fatalf("block form parse error: %v", err)
	}
	if _, ok := prog2.Statements[0].(*ast.ParallelStatement); !ok {
		t.Errorf("block form broken: %T", prog2.Statements[0])
	}
}
