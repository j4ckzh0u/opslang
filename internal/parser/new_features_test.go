package parser

import (
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

func TestForInStatement(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		varName string
	}{
		{"list literal", `for x in [1, 2, 3] { print(x) }`, "x"},
		{"variable", `for item in items { print(item) }`, "item"},
		{"string", `for ch in "hi" { print(ch) }`, "ch"},
		{"function call", `for x in getItems() { print(x) }`, "x"},
		{"dict", `for k in d { print(k) }`, "k"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New(tc.src, "test.ops")
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if len(prog.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
			}
			fi, ok := prog.Statements[0].(*ast.ForInStatement)
			if !ok {
				t.Fatalf("expected *ast.ForInStatement, got %T", prog.Statements[0])
			}
			if fi.Var.Name != tc.varName {
				t.Errorf("expected var %q, got %q", tc.varName, fi.Var.Name)
			}
			if fi.Body == nil || len(fi.Body.Statements) == 0 {
				t.Error("expected non-empty body")
			}
		})
	}
}

func TestForInDoesNotMatchCStyle(t *testing.T) {
	src := `for i = 0; i < 10; i = i + 1 { print(i) }`
	p := New(src, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if _, ok := prog.Statements[0].(*ast.ForStatement); !ok {
		t.Fatalf("expected *ast.ForStatement, got %T", prog.Statements[0])
	}
}

func TestBlockRescueStatement(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		hasBody   bool
		hasRescue bool
		hasAlways bool
	}{
		{"block only", `block { print(1) }`, true, false, false},
		{"block rescue", `block { print(1) } rescue { print(_error) }`, true, true, false},
		{"block always", `block { print(1) } always { print("done") }`, true, false, true},
		{"block rescue always", `block { print(1) } rescue { print(_error) } always { print("done") }`, true, true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New(tc.src, "test.ops")
			prog, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if len(prog.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
			}
			br, ok := prog.Statements[0].(*ast.BlockRescueStatement)
			if !ok {
				t.Fatalf("expected *ast.BlockRescueStatement, got %T", prog.Statements[0])
			}
			if tc.hasBody && br.Body == nil {
				t.Error("expected non-nil Body")
			}
			if tc.hasRescue && br.Rescue == nil {
				t.Error("expected non-nil Rescue")
			}
			if !tc.hasRescue && br.Rescue != nil {
				t.Error("expected nil Rescue")
			}
			if tc.hasAlways && br.Always == nil {
				t.Error("expected non-nil Always")
			}
			if !tc.hasAlways && br.Always != nil {
				t.Error("expected nil Always")
			}
		})
	}
}

func TestBlockRescueMultiline(t *testing.T) {
	src := `block {
    let x = 1
    let y = 2
} rescue {
    print(_error)
} always {
    print("cleanup")
}`
	p := New(src, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	br, ok := prog.Statements[0].(*ast.BlockRescueStatement)
	if !ok {
		t.Fatalf("expected *ast.BlockRescueStatement, got %T", prog.Statements[0])
	}
	if len(br.Body.Statements) != 2 {
		t.Errorf("expected 2 body statements, got %d", len(br.Body.Statements))
	}
	if len(br.Rescue.Statements) != 1 {
		t.Errorf("expected 1 rescue statement, got %d", len(br.Rescue.Statements))
	}
	if len(br.Always.Statements) != 1 {
		t.Errorf("expected 1 always statement, got %d", len(br.Always.Statements))
	}
}
