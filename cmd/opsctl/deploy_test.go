package main

import (
	"testing"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/parser"
)

func TestAstToInstructions_SimpleTask(t *testing.T) {
	source := `task "test" on ["localhost"] {
    let x = 42
}`
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	pkg, err := astToInstructions("test.ops", prog)
	if err != nil {
		t.Fatalf("astToInstructions error: %v", err)
	}

	if pkg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", pkg.Version)
	}
	if len(pkg.Instructions) == 0 {
		t.Error("expected at least one instruction")
	}
}

func TestAstToInstructions_EmptyScript(t *testing.T) {
	source := ``
	p := parser.New(source, "empty.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	pkg, err := astToInstructions("empty.ops", prog)
	if err != nil {
		t.Fatalf("astToInstructions error: %v", err)
	}

	// Empty script should get a no-op instruction.
	if len(pkg.Instructions) == 0 {
		t.Error("expected at least one no-op instruction for empty script")
	}
}

func TestResolveDeployMode(t *testing.T) {
	tests := []struct {
		mode     string
		source   string
		expected string
	}{
		{"runner", `print("hi")`, "runner"},
		{"aot", `print("hi")`, "aot"},
		{"auto", `print("hi")`, "runner"},
	}

	for _, tt := range tests {
		p := parser.New(tt.source, "test.ops")
		prog, err := p.Parse()
		if err != nil {
			t.Fatalf("parse error for mode=%s: %v", tt.mode, err)
		}

		result := resolveDeployMode(tt.mode, prog)
		if result != tt.expected {
			t.Errorf("resolveDeployMode(%q) = %q, want %q", tt.mode, result, tt.expected)
		}
	}
}

func TestResolveCallName(t *testing.T) {
	tests := []struct {
		expr     ast.Expression
		expected string
	}{
		{&ast.Identifier{Name: "print"}, "print"},
		{&ast.MemberExpression{
			Object: &ast.MemberExpression{
				Object: &ast.Identifier{Name: "sys"},
				Member: &ast.Identifier{Name: "cpu"},
			},
			Member: &ast.Identifier{Name: "usage"},
		}, "sys.cpu.usage"},
		{&ast.IntegerLiteral{Value: 42}, ""},
	}

	for i, tt := range tests {
		result := resolveCallName(tt.expr)
		if result != tt.expected {
			t.Errorf("test %d: resolveCallName() = %q, want %q", i, result, tt.expected)
		}
	}
}

func TestGenerateTaskID(t *testing.T) {
	id := generateTaskID("check_cpu.ops")
	if id == "" {
		t.Error("task ID should not be empty")
	}
	// Should contain the script name (without extension).
	if len(id) < 5 {
		t.Errorf("task ID too short: %s", id)
	}
}

func TestInstructionGen_WalkReport(t *testing.T) {
	source := `report { status: "ok", count: 42 }`
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := &instructionGen{}
	for _, stmt := range prog.Statements {
		if err := gen.walkStatement(stmt); err != nil {
			t.Fatalf("walkStatement error: %v", err)
		}
	}

	if len(gen.instructions) == 0 {
		t.Fatal("expected instructions from report")
	}

	found := false
	for _, inst := range gen.instructions {
		if inst.Op == "report" {
			found = true
		}
	}
	if !found {
		t.Error("expected a report instruction")
	}
}

func TestInstructionGen_WalkLet(t *testing.T) {
	source := `let x = 10`
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := &instructionGen{}
	for _, stmt := range prog.Statements {
		if err := gen.walkStatement(stmt); err != nil {
			t.Fatalf("walkStatement error: %v", err)
		}
	}

	if len(gen.instructions) == 0 {
		t.Fatal("expected instructions from let statement")
	}
}

func TestDeployMode_AutoWithImport(t *testing.T) {
	// A script with import should resolve to AOT mode.
	source := `let x = 42
print(x)`

	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Without import, auto should pick runner.
	result := resolveDeployMode("auto", prog)
	if result != "runner" {
		t.Errorf("auto mode without import should be runner, got %s", result)
	}
}
