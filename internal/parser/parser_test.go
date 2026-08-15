package parser

import (
	"testing"

	"github.com/opslang/opslang/internal/ast"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustParse(t *testing.T, src string) *ast.Program {
	t.Helper()
	p := New(src, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", src, err)
	}
	return prog
}

func expectParseError(t *testing.T, src string) {
	t.Helper()
	p := New(src, "test.ops")
	_, err := p.Parse()
	if err == nil {
		t.Fatalf("Parse(%q) expected error, got nil", src)
	}
}

// ---------------------------------------------------------------------------
// Let statements
// ---------------------------------------------------------------------------

func TestLetStatement(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantName string
		wantVal  string // String() of the value expression
	}{
		{"int", "let x = 42", "x", "42"},
		{"string", `let s = "hello"`, "s", `"hello"`},
		{"float", "let pi = 3.14", "pi", "3.14"},
		{"bool", "let flag = true", "flag", "true"},
		{"nil", "let n = nil", "n", "nil"},
		{"binary", "let sum = 1 + 2", "sum", "(1 + 2)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParse(t, tt.src)
			if len(prog.Statements) != 1 {
				t.Fatalf("want 1 statement, got %d", len(prog.Statements))
			}
			let, ok := prog.Statements[0].(*ast.LetStatement)
			if !ok {
				t.Fatalf("expected *ast.LetStatement, got %T", prog.Statements[0])
			}
			if let.Name.Name != tt.wantName {
				t.Errorf("name = %q, want %q", let.Name.Name, tt.wantName)
			}
			got := let.Value.String()
			if got != tt.wantVal {
				t.Errorf("value = %q, want %q", got, tt.wantVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Function statements
// ---------------------------------------------------------------------------

func TestFnStatement(t *testing.T) {
	src := `fn add(a, b) { return a + b }`
	prog := mustParse(t, src)
	if len(prog.Statements) != 1 {
		t.Fatalf("want 1 statement, got %d", len(prog.Statements))
	}
	fn, ok := prog.Statements[0].(*ast.FnStatement)
	if !ok {
		t.Fatalf("expected *ast.FnStatement, got %T", prog.Statements[0])
	}
	if fn.Name.Name != "add" {
		t.Errorf("name = %q, want %q", fn.Name.Name, "add")
	}
	if len(fn.Params) != 2 {
		t.Fatalf("want 2 params, got %d", len(fn.Params))
	}
	if fn.Params[0].Name.Name != "a" || fn.Params[1].Name.Name != "b" {
		t.Errorf("params = [%s, %s], want [a, b]",
			fn.Params[0].Name.Name, fn.Params[1].Name.Name)
	}
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("want 1 body stmt, got %d", len(fn.Body.Statements))
	}
	ret, ok := fn.Body.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected *ast.ReturnStatement, got %T", fn.Body.Statements[0])
	}
	if ret.Value.String() != "(a + b)" {
		t.Errorf("return value = %q, want %q", ret.Value.String(), "(a + b)")
	}
}

func TestFnWithDefaults(t *testing.T) {
	src := `fn greet(name, greeting = "hello") { return greeting }`
	prog := mustParse(t, src)
	fn := prog.Statements[0].(*ast.FnStatement)
	if len(fn.Params) != 2 {
		t.Fatalf("want 2 params, got %d", len(fn.Params))
	}
	if fn.Params[0].Default != nil {
		t.Error("first param should have no default")
	}
	if fn.Params[1].Default == nil {
		t.Fatal("second param should have a default")
	}
	if fn.Params[1].Default.String() != `"hello"` {
		t.Errorf("default = %q, want %q", fn.Params[1].Default.String(), `"hello"`)
	}
}

// ---------------------------------------------------------------------------
// If / else
// ---------------------------------------------------------------------------

func TestIfStatement(t *testing.T) {
	src := `if x > 0 { return x } else { return 0 }`
	prog := mustParse(t, src)
	ifStmt, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected *ast.IfStatement, got %T", prog.Statements[0])
	}
	if ifStmt.Condition.String() != "(x > 0)" {
		t.Errorf("condition = %q", ifStmt.Condition.String())
	}
	elseBlock, ok := ifStmt.ElseClause.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected else *ast.BlockStatement, got %T", ifStmt.ElseClause)
	}
	if len(elseBlock.Statements) != 1 {
		t.Fatalf("want 1 else stmt, got %d", len(elseBlock.Statements))
	}
}

func TestIfElseIf(t *testing.T) {
	src := `if x > 0 { return 1 } else if x < 0 { return -1 } else { return 0 }`
	prog := mustParse(t, src)
	ifStmt := prog.Statements[0].(*ast.IfStatement)

	elseIf, ok := ifStmt.ElseClause.(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected else-if *ast.IfStatement, got %T", ifStmt.ElseClause)
	}
	finalElse, ok := elseIf.ElseClause.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("expected final else *ast.BlockStatement, got %T", elseIf.ElseClause)
	}
	if len(finalElse.Statements) != 1 {
		t.Fatalf("want 1 final-else stmt, got %d", len(finalElse.Statements))
	}
}

// ---------------------------------------------------------------------------
// For loop
// ---------------------------------------------------------------------------

func TestForStatement(t *testing.T) {
	src := `for let i = 0; i < 10; i = i + 1 { print(i) }`
	prog := mustParse(t, src)
	forStmt, ok := prog.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("expected *ast.ForStatement, got %T", prog.Statements[0])
	}

	// Init should be a LetStatement.
	if _, ok := forStmt.Init.(*ast.LetStatement); !ok {
		t.Errorf("init = %T, want *ast.LetStatement", forStmt.Init)
	}
	// Condition.
	if forStmt.Condition.String() != "(i < 10)" {
		t.Errorf("condition = %q", forStmt.Condition.String())
	}
	// Post should be an AssignStatement.
	if _, ok := forStmt.Post.(*ast.AssignStatement); !ok {
		t.Errorf("post = %T, want *ast.AssignStatement", forStmt.Post)
	}
	// Body.
	if len(forStmt.Body.Statements) != 1 {
		t.Fatalf("want 1 body stmt, got %d", len(forStmt.Body.Statements))
	}
}

// ---------------------------------------------------------------------------
// While loop
// ---------------------------------------------------------------------------

func TestWhileStatement(t *testing.T) {
	src := `while x > 0 { x = x - 1 }`
	prog := mustParse(t, src)
	w, ok := prog.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected *ast.WhileStatement, got %T", prog.Statements[0])
	}
	if w.Condition.String() != "(x > 0)" {
		t.Errorf("condition = %q", w.Condition.String())
	}
}

// ---------------------------------------------------------------------------
// Return
// ---------------------------------------------------------------------------

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantNil bool
		wantVal string
	}{
		{"bare", "return", true, ""},
		{"with value", "return x + 1", false, "(x + 1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParse(t, tt.src)
			ret := prog.Statements[0].(*ast.ReturnStatement)
			if tt.wantNil && ret.Value != nil {
				t.Errorf("expected nil value, got %s", ret.Value)
			}
			if !tt.wantNil && ret.Value == nil {
				t.Fatal("expected non-nil value")
			}
			if !tt.wantNil && ret.Value.String() != tt.wantVal {
				t.Errorf("value = %q, want %q", ret.Value.String(), tt.wantVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Task statement
// ---------------------------------------------------------------------------

func TestTaskStatement(t *testing.T) {
	src := `task "deploy" on hosts { print("ok") }`
	prog := mustParse(t, src)
	task, ok := prog.Statements[0].(*ast.TaskStatement)
	if !ok {
		t.Fatalf("expected *ast.TaskStatement, got %T", prog.Statements[0])
	}
	if task.Name != "deploy" {
		t.Errorf("name = %q, want %q", task.Name, "deploy")
	}
	if task.Targets.Var == nil || task.Targets.Var.Name != "hosts" {
		t.Errorf("targets.Var = %v, want hosts", task.Targets.Var)
	}
}

func TestTaskWithHostList(t *testing.T) {
	src := `task "check" on "h1", "h2" { print("ok") }`
	prog := mustParse(t, src)
	task := prog.Statements[0].(*ast.TaskStatement)
	if len(task.Targets.Hosts) != 2 {
		t.Fatalf("want 2 hosts, got %d", len(task.Targets.Hosts))
	}
	if task.Targets.Hosts[0].String() != `"h1"` {
		t.Errorf("host[0] = %s", task.Targets.Hosts[0])
	}
}

// ---------------------------------------------------------------------------
// Import
// ---------------------------------------------------------------------------

func TestImportStatement(t *testing.T) {
	src := `import "sys"`
	prog := mustParse(t, src)
	imp, ok := prog.Statements[0].(*ast.ImportStatement)
	if !ok {
		t.Fatalf("expected *ast.ImportStatement, got %T", prog.Statements[0])
	}
	if imp.Path != "sys" {
		t.Errorf("path = %q, want %q", imp.Path, "sys")
	}
}

// ---------------------------------------------------------------------------
// Report
// ---------------------------------------------------------------------------

func TestReportStatement(t *testing.T) {
	src := `report { cpu: x, mem: y }`
	prog := mustParse(t, src)
	rpt, ok := prog.Statements[0].(*ast.ReportStatement)
	if !ok {
		t.Fatalf("expected *ast.ReportStatement, got %T", prog.Statements[0])
	}
	if len(rpt.Fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(rpt.Fields))
	}
	if rpt.Fields[0].Key != "cpu" || rpt.Fields[0].Value.String() != "x" {
		t.Errorf("field[0] = %s: %s", rpt.Fields[0].Key, rpt.Fields[0].Value)
	}
	if rpt.Fields[1].Key != "mem" || rpt.Fields[1].Value.String() != "y" {
		t.Errorf("field[1] = %s: %s", rpt.Fields[1].Key, rpt.Fields[1].Value)
	}
}

// ---------------------------------------------------------------------------
// Alert
// ---------------------------------------------------------------------------

func TestAlertStatement(t *testing.T) {
	src := `alert("high cpu")`
	prog := mustParse(t, src)
	alert, ok := prog.Statements[0].(*ast.AlertStatement)
	if !ok {
		t.Fatalf("expected *ast.AlertStatement, got %T", prog.Statements[0])
	}
	if alert.Message.String() != `"high cpu"` {
		t.Errorf("message = %s", alert.Message)
	}
}

// ---------------------------------------------------------------------------
// Binary expressions / precedence
// ---------------------------------------------------------------------------

func TestBinaryPrecedence(t *testing.T) {
	// 1 + 2 * 3 should parse as 1 + (2 * 3)
	src := `let r = 1 + 2 * 3`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	bin, ok := let.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression, got %T", let.Value)
	}
	if bin.Op != "+" {
		t.Errorf("top op = %q, want +", bin.Op)
	}
	right, ok := bin.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("right = %T, want *ast.BinaryExpression", bin.Right)
	}
	if right.Op != "*" {
		t.Errorf("right op = %q, want *", right.Op)
	}
}

func TestAssociativity(t *testing.T) {
	// 1 - 2 - 3 should parse as (1 - 2) - 3 (left-associative)
	src := `let r = 1 - 2 - 3`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	top, ok := let.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected *ast.BinaryExpression, got %T", let.Value)
	}
	if top.Op != "-" {
		t.Errorf("top op = %q, want -", top.Op)
	}
	left, ok := top.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("left = %T, want *ast.BinaryExpression", top.Left)
	}
	if left.Op != "-" {
		t.Errorf("left op = %q, want -", left.Op)
	}
}

func TestLogicalPrecedence(t *testing.T) {
	// a || b && c should parse as a || (b && c)
	src := `let r = a || b && c`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	top, ok := let.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression, got %T", let.Value)
	}
	if top.Op != "||" {
		t.Errorf("top op = %q, want ||", top.Op)
	}
	right, ok := top.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("right = %T, want BinaryExpression", top.Right)
	}
	if right.Op != "&&" {
		t.Errorf("right op = %q, want &&", right.Op)
	}
}

// ---------------------------------------------------------------------------
// Unary expressions
// ---------------------------------------------------------------------------

func TestUnaryExpressions(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"let r = -x", "(-x)"},
		{"let r = !flag", "(!flag)"},
		{"let r = -1 + 2", "((-1) + 2)"},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			prog := mustParse(t, tt.src)
			let := prog.Statements[0].(*ast.LetStatement)
			got := let.Value.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Call expressions
// ---------------------------------------------------------------------------

func TestCallExpression(t *testing.T) {
	src := `print(x)`
	prog := mustParse(t, src)
	exprStmt := prog.Statements[0].(*ast.ExpressionStatement)
	call, ok := exprStmt.Expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", exprStmt.Expr)
	}
	if call.Function.String() != "print" {
		t.Errorf("function = %s", call.Function)
	}
	if len(call.Args) != 1 || call.Args[0].String() != "x" {
		t.Errorf("args = %v", call.Args)
	}
}

func TestChainedCall(t *testing.T) {
	src := `sys.cpu.usage()`
	prog := mustParse(t, src)
	exprStmt := prog.Statements[0].(*ast.ExpressionStatement)
	call, ok := exprStmt.Expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected *ast.CallExpression, got %T", exprStmt.Expr)
	}
	// Function should be MemberExpression: sys.cpu.usage
	member, ok := call.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("function = %T, want *ast.MemberExpression", call.Function)
	}
	if member.Member.Name != "usage" {
		t.Errorf("member = %s, want usage", member.Member.Name)
	}
}

// ---------------------------------------------------------------------------
// Index expression
// ---------------------------------------------------------------------------

func TestIndexExpression(t *testing.T) {
	src := `let r = arr[0]`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	idx, ok := let.Value.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("expected *ast.IndexExpression, got %T", let.Value)
	}
	if idx.Left.String() != "arr" {
		t.Errorf("left = %s", idx.Left)
	}
	if idx.Index.String() != "0" {
		t.Errorf("index = %s", idx.Index)
	}
}

// ---------------------------------------------------------------------------
// Member expression
// ---------------------------------------------------------------------------

func TestMemberExpression(t *testing.T) {
	src := `let r = obj.field`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	mem, ok := let.Value.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("expected *ast.MemberExpression, got %T", let.Value)
	}
	if mem.Object.String() != "obj" {
		t.Errorf("object = %s", mem.Object)
	}
	if mem.Member.Name != "field" {
		t.Errorf("member = %s", mem.Member.Name)
	}
}

// ---------------------------------------------------------------------------
// List literal
// ---------------------------------------------------------------------------

func TestListLiteral(t *testing.T) {
	src := `let r = [1, 2, 3]`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	list, ok := let.Value.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected *ast.ListLiteral, got %T", let.Value)
	}
	if len(list.Elements) != 3 {
		t.Fatalf("want 3 elements, got %d", len(list.Elements))
	}
	if list.Elements[0].String() != "1" {
		t.Errorf("elem[0] = %s", list.Elements[0])
	}
}

// ---------------------------------------------------------------------------
// Dict literal
// ---------------------------------------------------------------------------

func TestDictLiteral(t *testing.T) {
	src := `let r = {"a": 1, "b": 2}`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	dict, ok := let.Value.(*ast.DictLiteral)
	if !ok {
		t.Fatalf("expected *ast.DictLiteral, got %T", let.Value)
	}
	if len(dict.Keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(dict.Keys))
	}
	if dict.Keys[0].String() != `"a"` || dict.Values[0].String() != "1" {
		t.Errorf("pair[0] = %s: %s", dict.Keys[0], dict.Values[0])
	}
}

// ---------------------------------------------------------------------------
// Assignment statement
// ---------------------------------------------------------------------------

func TestAssignStatement(t *testing.T) {
	src := `x = 42`
	prog := mustParse(t, src)
	assign, ok := prog.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("expected *ast.AssignStatement, got %T", prog.Statements[0])
	}
	if assign.Target.String() != "x" {
		t.Errorf("target = %s", assign.Target)
	}
	if assign.Value.String() != "42" {
		t.Errorf("value = %s", assign.Value)
	}
}

func TestAssignMemberTarget(t *testing.T) {
	src := `obj.field = 1`
	prog := mustParse(t, src)
	assign := prog.Statements[0].(*ast.AssignStatement)
	mem, ok := assign.Target.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("target = %T, want *ast.MemberExpression", assign.Target)
	}
	if mem.Member.Name != "field" {
		t.Errorf("member = %s", mem.Member.Name)
	}
}

// ---------------------------------------------------------------------------
// Multiple statements
// ---------------------------------------------------------------------------

func TestMultipleStatements(t *testing.T) {
	src := `let x = 1
let y = 2
let z = x + y`
	prog := mustParse(t, src)
	if len(prog.Statements) != 3 {
		t.Fatalf("want 3 statements, got %d", len(prog.Statements))
	}
}

func TestSemicolonSeparated(t *testing.T) {
	src := `let x = 1; let y = 2; let z = 3`
	prog := mustParse(t, src)
	if len(prog.Statements) != 3 {
		t.Fatalf("want 3 statements, got %d", len(prog.Statements))
	}
}

// ---------------------------------------------------------------------------
// If-expression
// ---------------------------------------------------------------------------

func TestIfExpression(t *testing.T) {
	src := `let r = if x > 0 { x } else { 0 }`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	ifExpr, ok := let.Value.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected *ast.IfExpression, got %T", let.Value)
	}
	if ifExpr.Condition.String() != "(x > 0)" {
		t.Errorf("condition = %s", ifExpr.Condition)
	}
	if ifExpr.Then.String() != "x" {
		t.Errorf("then = %s", ifExpr.Then)
	}
	if ifExpr.Else.String() != "0" {
		t.Errorf("else = %s", ifExpr.Else)
	}
}

// ---------------------------------------------------------------------------
// Empty program
// ---------------------------------------------------------------------------

func TestEmptyProgram(t *testing.T) {
	prog := mustParse(t, "")
	if len(prog.Statements) != 0 {
		t.Errorf("want 0 statements, got %d", len(prog.Statements))
	}
}

func TestWhitespaceOnly(t *testing.T) {
	prog := mustParse(t, "   \n\n  \n")
	if len(prog.Statements) != 0 {
		t.Errorf("want 0 statements, got %d", len(prog.Statements))
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"unexpected token", "~"},
		{"missing let name", "let = 1"},
		{"missing let value", "let x = "},
		{"missing fn name", "fn() {}"},
		{"missing fn paren", "fn foo {}"},
		{"missing if condition", "if { }"},
		{"missing if body", "if true"},
		{"missing for semicolons", "for let i = 0 { }"},
		{"missing task on", `task "x" { }`},
		{"missing import path", "import"},
		{"unterminated list", "let x = [1, 2"},
		{"unterminated dict", `let x = {"a": 1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectParseError(t, tt.src)
		})
	}
}

// ---------------------------------------------------------------------------
// Nested structures
// ---------------------------------------------------------------------------

func TestNestedBlocks(t *testing.T) {
	src := `fn outer() {
	fn inner() { return 1 }
	return inner()
}`
	prog := mustParse(t, src)
	if len(prog.Statements) != 1 {
		t.Fatalf("want 1 statement, got %d", len(prog.Statements))
	}
	outer := prog.Statements[0].(*ast.FnStatement)
	if len(outer.Body.Statements) != 2 {
		t.Fatalf("want 2 body stmts, got %d", len(outer.Body.Statements))
	}
}

func TestComplexExpression(t *testing.T) {
	// Test that chained member + call works: a.b.c(x, y)
	src := `a.b.c(x, y)`
	prog := mustParse(t, src)
	exprStmt := prog.Statements[0].(*ast.ExpressionStatement)
	call, ok := exprStmt.Expr.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", exprStmt.Expr)
	}
	// call.Function should be a.b.c
	mem, ok := call.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("function = %T, want MemberExpression", call.Function)
	}
	if mem.Member.Name != "c" {
		t.Errorf("outer member = %s", mem.Member.Name)
	}
	// mem.Object should be a.b
	inner, ok := mem.Object.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("object = %T, want MemberExpression", mem.Object)
	}
	if inner.Member.Name != "b" {
		t.Errorf("inner member = %s", inner.Member.Name)
	}
}

func TestParenGrouping(t *testing.T) {
	// (1 + 2) * 3 should override precedence
	src := `let r = (1 + 2) * 3`
	prog := mustParse(t, src)
	let := prog.Statements[0].(*ast.LetStatement)
	top, ok := let.Value.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression, got %T", let.Value)
	}
	if top.Op != "*" {
		t.Errorf("top op = %q, want *", top.Op)
	}
	left, ok := top.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("left = %T, want BinaryExpression", top.Left)
	}
	if left.Op != "+" {
		t.Errorf("left op = %q, want +", left.Op)
	}
}

// ---------------------------------------------------------------------------
// OpsLang example scripts
// ---------------------------------------------------------------------------

func TestCollectInfoScript(t *testing.T) {
	src := `
import "sys"

let cpu = sys.cpu.usage()
let mem = sys.memory.info()

if cpu.percent > 90 {
    alert("CPU high")
}

report {
    cpu: cpu,
    mem: mem
}
`
	prog := mustParse(t, src)
	// import + 2 lets + if + report = 5 statements
	if len(prog.Statements) != 5 {
		t.Errorf("want 5 statements, got %d", len(prog.Statements))
	}
}

// ---------------------------------------------------------------------------
// Ensure statement
// ---------------------------------------------------------------------------

func TestEnsureStatement(t *testing.T) {
	src := `ensure x > 0 { x = 1 }`
	prog := mustParse(t, src)
	if len(prog.Statements) != 1 {
		t.Fatalf("want 1 statement, got %d", len(prog.Statements))
	}
	ens, ok := prog.Statements[0].(*ast.EnsureStatement)
	if !ok {
		t.Fatalf("expected *ast.EnsureStatement, got %T", prog.Statements[0])
	}
	if ens.Condition.String() != "(x > 0)" {
		t.Errorf("condition = %q, want (x > 0)", ens.Condition.String())
	}
	if len(ens.Body.Statements) != 1 {
		t.Fatalf("want 1 body stmt, got %d", len(ens.Body.Statements))
	}
	if ens.Notify != nil {
		t.Errorf("expected nil Notify, got %v", ens.Notify)
	}
}

func TestEnsureStatementWithNotify(t *testing.T) {
	src := `ensure running { start() }
notify "restart needed"`
	prog := mustParse(t, src)
	if len(prog.Statements) != 1 {
		t.Fatalf("want 1 statement, got %d", len(prog.Statements))
	}
	ens, ok := prog.Statements[0].(*ast.EnsureStatement)
	if !ok {
		t.Fatalf("expected *ast.EnsureStatement, got %T", prog.Statements[0])
	}
	if ens.Notify == nil {
		t.Fatal("expected non-nil Notify")
	}
	if ens.Notify.String() != `"restart needed"` {
		t.Errorf("notify = %s, want \"restart needed\"", ens.Notify.String())
	}
}

func TestEnsureStatementString(t *testing.T) {
	src := `ensure ok { fix() }`
	prog := mustParse(t, src)
	ens := prog.Statements[0].(*ast.EnsureStatement)
	got := ens.String()
	want := "ensure ok { ... }"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// log() and metric() as call expressions (parsed normally)
// ---------------------------------------------------------------------------

func TestLogCallParsed(t *testing.T) {
	src := `log("hello world")`
	prog := mustParse(t, src)
	if len(prog.Statements) != 1 {
		t.Fatalf("want 1 statement, got %d", len(prog.Statements))
	}
	logStmt, ok := prog.Statements[0].(*ast.LogStatement)
	if !ok {
		t.Fatalf("expected *ast.LogStatement, got %T", prog.Statements[0])
	}
	if logStmt.Message.String() != `"hello world"` {
		t.Errorf("message = %s, want \"hello world\"", logStmt.Message)
	}
}

func TestMetricCallParsed(t *testing.T) {
	src := `metric("cpu_usage", 42.5)`
	prog := mustParse(t, src)
	metricStmt, ok := prog.Statements[0].(*ast.MetricStatement)
	if !ok {
		t.Fatalf("expected *ast.MetricStatement, got %T", prog.Statements[0])
	}
	if metricStmt.Name.String() != `"cpu_usage"` {
		t.Errorf("name = %s, want \"cpu_usage\"", metricStmt.Name)
	}
	if metricStmt.Value.String() != "42.5" {
		t.Errorf("value = %s, want 42.5", metricStmt.Value)
	}
	if metricStmt.Labels != nil {
		t.Errorf("labels = %s, want nil", metricStmt.Labels)
	}
}

// ---------------------------------------------------------------------------
// Parallel statement
// ---------------------------------------------------------------------------

func TestParallelStatement(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		wantStmts  int
	}{
		{
			name: "basic let statements",
			src: `parallel {
  let x = 1
  let y = 2
  let z = 3
}`,
			wantStmts: 3,
		},
		{
			name:      "empty parallel block",
			src:       `parallel {}`,
			wantStmts: 0,
		},
		{
			name: "function calls",
			src: `parallel {
  print("a")
  print("b")
}`,
			wantStmts: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := mustParse(t, tt.src)
			if len(prog.Statements) != 1 {
				t.Fatalf("want 1 statement, got %d", len(prog.Statements))
			}
			par, ok := prog.Statements[0].(*ast.ParallelStatement)
			if !ok {
				t.Fatalf("expected *ast.ParallelStatement, got %T", prog.Statements[0])
			}
			if len(par.Body.Statements) != tt.wantStmts {
				t.Errorf("body stmts = %d, want %d", len(par.Body.Statements), tt.wantStmts)
			}
		})
	}
}

func TestParallelStatementWithLetValues(t *testing.T) {
	src := `parallel {
  let a = 10
  let b = "hello"
}`
	prog := mustParse(t, src)
	par := prog.Statements[0].(*ast.ParallelStatement)
	if len(par.Body.Statements) != 2 {
		t.Fatalf("want 2 body stmts, got %d", len(par.Body.Statements))
	}

	let0, ok := par.Body.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("stmt[0] = %T, want *ast.LetStatement", par.Body.Statements[0])
	}
	if let0.Name.Name != "a" {
		t.Errorf("stmt[0].Name = %q, want %q", let0.Name.Name, "a")
	}
	if let0.Value.String() != "10" {
		t.Errorf("stmt[0].Value = %s, want 10", let0.Value)
	}

	let1, ok := par.Body.Statements[1].(*ast.LetStatement)
	if !ok {
		t.Fatalf("stmt[1] = %T, want *ast.LetStatement", par.Body.Statements[1])
	}
	if let1.Name.Name != "b" {
		t.Errorf("stmt[1].Name = %q, want %q", let1.Name.Name, "b")
	}
	if let1.Value.String() != `"hello"` {
		t.Errorf("stmt[1].Value = %s, want \"hello\"", let1.Value)
	}
}

func TestParallelStatementString(t *testing.T) {
	src := `parallel { let x = 1 }`
	prog := mustParse(t, src)
	par := prog.Statements[0].(*ast.ParallelStatement)
	got := par.String()
	want := "parallel { ... }"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
