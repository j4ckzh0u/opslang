package interpreter

import (
	"strings"
	"testing"

	"github.com/opslang/opslang/internal/ast"
)

// ---------------------------------------------------------------------------
// Helper constructors
// ---------------------------------------------------------------------------

func pos() ast.Position { return ast.Position{Line: 1, Column: 1} }

func ident(name string) *ast.Identifier {
	return &ast.Identifier{Position: pos(), Name: name}
}

func intLit(v int64) *ast.IntegerLiteral {
	return &ast.IntegerLiteral{Position: pos(), Value: v}
}

func floatLit(v float64) *ast.FloatLiteral {
	return &ast.FloatLiteral{Position: pos(), Value: v}
}

func strLit(v string) *ast.StringLiteral {
	return &ast.StringLiteral{Position: pos(), Value: v}
}

func boolLit(v bool) *ast.BoolLiteral {
	return &ast.BoolLiteral{Position: pos(), Value: v}
}

func nilLit() *ast.NilLiteral {
	return &ast.NilLiteral{Position: pos()}
}

func prog(stmts ...ast.Statement) *ast.Program {
	return &ast.Program{Position: pos(), Statements: stmts}
}

func let(name string, val ast.Expression) *ast.LetStatement {
	return &ast.LetStatement{Position: pos(), Name: ident(name), Value: val}
}

func binOp(left ast.Expression, op string, right ast.Expression) *ast.BinaryExpression {
	return &ast.BinaryExpression{Position: pos(), Left: left, Op: op, Right: right}
}

func unaryOp(op string, right ast.Expression) *ast.UnaryExpression {
	return &ast.UnaryExpression{Position: pos(), Op: op, Right: right}
}

func call(fn ast.Expression, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{Position: pos(), Function: fn, Args: args}
}

func member(obj ast.Expression, name string) *ast.MemberExpression {
	return &ast.MemberExpression{Position: pos(), Object: obj, Member: ident(name)}
}

func index(left ast.Expression, idx ast.Expression) *ast.IndexExpression {
	return &ast.IndexExpression{Position: pos(), Left: left, Index: idx}
}

func block(stmts ...ast.Statement) *ast.BlockStatement {
	return &ast.BlockStatement{Position: pos(), Statements: stmts}
}

func exprStmt(expr ast.Expression) *ast.ExpressionStatement {
	return &ast.ExpressionStatement{Position: pos(), Expr: expr}
}

func assign(target ast.Expression, val ast.Expression) *ast.AssignStatement {
	return &ast.AssignStatement{Position: pos(), Target: target, Value: val}
}

func newInterp() *Interpreter {
	return New(nil)
}

// ---------------------------------------------------------------------------
// Variable declaration and lookup
// ---------------------------------------------------------------------------

func TestLetAndIdentifier(t *testing.T) {
	p := prog(
		let("x", intLit(42)),
		let("y", strLit("hello")),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != int64(42) {
		t.Errorf("x = %v, want 42", r.Variables["x"])
	}
	if r.Variables["y"] != "hello" {
		t.Errorf("y = %v, want hello", r.Variables["y"])
	}
}

// ---------------------------------------------------------------------------
// Arithmetic
// ---------------------------------------------------------------------------

func TestArithmeticInt(t *testing.T) {
	p := prog(
		let("a", binOp(intLit(10), "+", intLit(5))),
		let("b", binOp(intLit(10), "-", intLit(3))),
		let("c", binOp(intLit(6), "*", intLit(7))),
		let("d", binOp(intLit(15), "/", intLit(4))),
		let("e", binOp(intLit(17), "%", intLit(5))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]int64{"a": 15, "b": 7, "c": 42, "d": 3, "e": 2}
	for k, want := range cases {
		if got, ok := r.Variables[k].(int64); !ok || got != want {
			t.Errorf("%s = %v (%T), want %d", k, r.Variables[k], r.Variables[k], want)
		}
	}
}

func TestArithmeticFloat(t *testing.T) {
	p := prog(
		let("a", binOp(floatLit(1.5), "+", floatLit(2.5))),
		let("b", binOp(intLit(3), "+", floatLit(0.14))),
		let("c", binOp(floatLit(10.0), "/", floatLit(4.0))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := r.Variables["a"].(float64); !ok || v != 4.0 {
		t.Errorf("a = %v, want 4.0", r.Variables["a"])
	}
	if v, ok := r.Variables["b"].(float64); !ok || v != 3.14 {
		t.Errorf("b = %v, want 3.14", r.Variables["b"])
	}
	if v, ok := r.Variables["c"].(float64); !ok || v != 2.5 {
		t.Errorf("c = %v, want 2.5", r.Variables["c"])
	}
}

func TestStringConcat(t *testing.T) {
	p := prog(let("s", binOp(strLit("hello"), "+", strLit(" world"))))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["s"] != "hello world" {
		t.Errorf("s = %v, want 'hello world'", r.Variables["s"])
	}
}

// ---------------------------------------------------------------------------
// Comparison and boolean logic
// ---------------------------------------------------------------------------

func TestComparisons(t *testing.T) {
	p := prog(
		let("a", binOp(intLit(3), "<", intLit(5))),
		let("b", binOp(intLit(5), ">", intLit(3))),
		let("c", binOp(intLit(3), "==", intLit(3))),
		let("d", binOp(intLit(3), "!=", intLit(4))),
		let("e", binOp(intLit(3), "<=", intLit(3))),
		let("f", binOp(intLit(5), ">=", intLit(6))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{"a": true, "b": true, "c": true, "d": true, "e": true, "f": false}
	for k, want := range cases {
		if got, ok := r.Variables[k].(bool); !ok || got != want {
			t.Errorf("%s = %v, want %v", k, r.Variables[k], want)
		}
	}
}

func TestLogicalShortCircuit(t *testing.T) {
	// && short-circuit: left is false so right (which would error) is never evaluated.
	p := prog(
		let("x", binOp(boolLit(false), "&&",
			// This would be an undefined variable, but should not be evaluated.
			ident("undefined_var_xyz"))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != false {
		t.Errorf("x = %v, want false", r.Variables["x"])
	}

	// || short-circuit: left is true.
	p2 := prog(
		let("y", binOp(boolLit(true), "||", ident("undefined_var_xyz"))),
	)
	r2, err := newInterp().Execute(p2)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Variables["y"] != true {
		t.Errorf("y = %v, want true", r2.Variables["y"])
	}
}

func TestLogicalOperators(t *testing.T) {
	p := prog(
		let("a", binOp(boolLit(true), "&&", boolLit(true))),
		let("b", binOp(boolLit(true), "&&", boolLit(false))),
		let("c", binOp(boolLit(false), "||", boolLit(true))),
		let("d", binOp(boolLit(false), "||", boolLit(false))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]bool{"a": true, "b": false, "c": true, "d": false}
	for k, want := range cases {
		if got, ok := r.Variables[k].(bool); !ok || got != want {
			t.Errorf("%s = %v, want %v", k, r.Variables[k], want)
		}
	}
}

// ---------------------------------------------------------------------------
// If / else statements
// ---------------------------------------------------------------------------

func TestIfStatement(t *testing.T) {
	p := prog(
		let("x", intLit(10)),
		&ast.IfStatement{
			Position:  pos(),
			Condition: binOp(ident("x"), ">", intLit(5)),
			Body:      block(assign(ident("x"), intLit(100))),
			ElseClause: block(assign(ident("x"), intLit(0))),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != int64(100) {
		t.Errorf("x = %v, want 100", r.Variables["x"])
	}
}

func TestIfElseStatement(t *testing.T) {
	p := prog(
		let("x", intLit(2)),
		&ast.IfStatement{
			Position:  pos(),
			Condition: binOp(ident("x"), ">", intLit(5)),
			Body:      block(assign(ident("x"), intLit(100))),
			ElseClause: block(assign(ident("x"), intLit(0))),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != int64(0) {
		t.Errorf("x = %v, want 0", r.Variables["x"])
	}
}

// ---------------------------------------------------------------------------
// For loops
// ---------------------------------------------------------------------------

func TestForLoop(t *testing.T) {
	// sum = 0; for (let i = 0; i < 5; i = i + 1) { sum = sum + i }
	p := prog(
		let("sum", intLit(0)),
		&ast.ForStatement{
			Position:  pos(),
			Init:      let("i", intLit(0)),
			Condition: binOp(ident("i"), "<", intLit(5)),
			Post:      assign(ident("i"), binOp(ident("i"), "+", intLit(1))),
			Body: block(
				assign(ident("sum"), binOp(ident("sum"), "+", ident("i"))),
			),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["sum"] != int64(10) {
		t.Errorf("sum = %v, want 10", r.Variables["sum"])
	}
}

// ---------------------------------------------------------------------------
// While loops
// ---------------------------------------------------------------------------

func TestWhileLoop(t *testing.T) {
	p := prog(
		let("n", intLit(5)),
		let("fact", intLit(1)),
		&ast.WhileStatement{
			Position:  pos(),
			Condition: binOp(ident("n"), ">", intLit(0)),
			Body: block(
				assign(ident("fact"), binOp(ident("fact"), "*", ident("n"))),
				assign(ident("n"), binOp(ident("n"), "-", intLit(1))),
			),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["fact"] != int64(120) {
		t.Errorf("fact = %v, want 120", r.Variables["fact"])
	}
}

// ---------------------------------------------------------------------------
// Function definition and calling
// ---------------------------------------------------------------------------

func TestFunctionDefAndCall(t *testing.T) {
	p := prog(
		&ast.FnStatement{
			Position: pos(),
			Name:     ident("add"),
			Params:   []ast.Parameter{{Name: ident("a")}, {Name: ident("b")}},
			Body: block(
				&ast.ReturnStatement{Position: pos(), Value: binOp(ident("a"), "+", ident("b"))},
			),
		},
		let("result", call(ident("add"), intLit(3), intLit(4))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["result"] != int64(7) {
		t.Errorf("result = %v, want 7", r.Variables["result"])
	}
}

func TestFunctionDefaultParams(t *testing.T) {
	p := prog(
		&ast.FnStatement{
			Position: pos(),
			Name:     ident("greet"),
			Params: []ast.Parameter{
				{Name: ident("name")},
				{Name: ident("greeting"), Default: strLit("Hello")},
			},
			Body: block(
				&ast.ReturnStatement{
					Position: pos(),
					Value:    binOp(ident("greeting"), "+", binOp(strLit(", "), "+", ident("name"))),
				},
			),
		},
		let("a", call(ident("greet"), strLit("Alice"))),
		let("b", call(ident("greet"), strLit("Bob"), strLit("Hi"))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["a"] != "Hello, Alice" {
		t.Errorf("a = %v, want 'Hello, Alice'", r.Variables["a"])
	}
	if r.Variables["b"] != "Hi, Bob" {
		t.Errorf("b = %v, want 'Hi, Bob'", r.Variables["b"])
	}
}

func TestRecursiveFactorial(t *testing.T) {
	p := prog(
		&ast.FnStatement{
			Position: pos(),
			Name:     ident("fact"),
			Params:   []ast.Parameter{{Name: ident("n")}},
			Body: block(
				&ast.IfStatement{
					Position:  pos(),
					Condition: binOp(ident("n"), "<=", intLit(1)),
					Body: block(
						&ast.ReturnStatement{Position: pos(), Value: intLit(1)},
					),
				},
				&ast.ReturnStatement{
					Position: pos(),
					Value:    binOp(ident("n"), "*", call(ident("fact"), binOp(ident("n"), "-", intLit(1)))),
				},
			),
		},
		let("r", call(ident("fact"), intLit(6))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["r"] != int64(720) {
		t.Errorf("r = %v, want 720", r.Variables["r"])
	}
}

// ---------------------------------------------------------------------------
// List and dict operations
// ---------------------------------------------------------------------------

func TestListLiteral(t *testing.T) {
	p := prog(let("xs", &ast.ListLiteral{
		Position: pos(),
		Elements: []ast.Expression{intLit(1), intLit(2), intLit(3)},
	}))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	xs, ok := r.Variables["xs"].([]interface{})
	if !ok || len(xs) != 3 {
		t.Fatalf("xs = %v, want [1 2 3]", r.Variables["xs"])
	}
	if xs[0] != int64(1) || xs[1] != int64(2) || xs[2] != int64(3) {
		t.Errorf("xs = %v, want [1 2 3]", xs)
	}
}

func TestDictLiteral(t *testing.T) {
	p := prog(let("d", &ast.DictLiteral{
		Position: pos(),
		Keys:     []ast.Expression{strLit("name"), strLit("age")},
		Values:   []ast.Expression{strLit("Alice"), intLit(30)},
	}))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	d, ok := r.Variables["d"].(map[string]interface{})
	if !ok {
		t.Fatalf("d is not a dict: %T", r.Variables["d"])
	}
	if d["name"] != "Alice" || d["age"] != int64(30) {
		t.Errorf("d = %v", d)
	}
}

// ---------------------------------------------------------------------------
// Index access
// ---------------------------------------------------------------------------

func TestIndexAccess(t *testing.T) {
	p := prog(
		let("xs", &ast.ListLiteral{Position: pos(), Elements: []ast.Expression{intLit(10), intLit(20), intLit(30)}}),
		let("d", &ast.DictLiteral{Position: pos(), Keys: []ast.Expression{strLit("k")}, Values: []ast.Expression{intLit(99)}}),
		let("a", index(ident("xs"), intLit(1))),
		let("b", index(ident("d"), strLit("k"))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["a"] != int64(20) {
		t.Errorf("a = %v, want 20", r.Variables["a"])
	}
	if r.Variables["b"] != int64(99) {
		t.Errorf("b = %v, want 99", r.Variables["b"])
	}
}

// ---------------------------------------------------------------------------
// Member access
// ---------------------------------------------------------------------------

func TestMemberAccess(t *testing.T) {
	p := prog(
		let("d", &ast.DictLiteral{
			Position: pos(),
			Keys:     []ast.Expression{strLit("x"), strLit("y")},
			Values:   []ast.Expression{intLit(1), intLit(2)},
		}),
		let("a", member(ident("d"), "x")),
		let("b", member(ident("d"), "y")),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["a"] != int64(1) {
		t.Errorf("a = %v, want 1", r.Variables["a"])
	}
	if r.Variables["b"] != int64(2) {
		t.Errorf("b = %v, want 2", r.Variables["b"])
	}
}

// ---------------------------------------------------------------------------
// Built-in functions
// ---------------------------------------------------------------------------

func TestBuiltinLen(t *testing.T) {
	p := prog(
		let("a", call(ident("len"), strLit("hello"))),
		let("b", call(ident("len"), &ast.ListLiteral{Position: pos(), Elements: []ast.Expression{intLit(1), intLit(2)}})),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["a"] != int64(5) {
		t.Errorf("len(hello) = %v, want 5", r.Variables["a"])
	}
	if r.Variables["b"] != int64(2) {
		t.Errorf("len([1,2]) = %v, want 2", r.Variables["b"])
	}
}

func TestBuiltinStrIntFloat(t *testing.T) {
	p := prog(
		let("a", call(ident("str"), intLit(42))),
		let("b", call(ident("int"), floatLit(3.9))),
		let("c", call(ident("float"), intLit(7))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["a"] != "42" {
		t.Errorf("str(42) = %v, want '42'", r.Variables["a"])
	}
	if r.Variables["b"] != int64(3) {
		t.Errorf("int(3.9) = %v, want 3", r.Variables["b"])
	}
	if r.Variables["c"] != float64(7) {
		t.Errorf("float(7) = %v, want 7.0", r.Variables["c"])
	}
}

func TestBuiltinType(t *testing.T) {
	p := prog(
		let("a", call(ident("type"), intLit(1))),
		let("b", call(ident("type"), strLit("hi"))),
		let("c", call(ident("type"), boolLit(true))),
		let("d", call(ident("type"), nilLit())),
		let("e", call(ident("type"), &ast.ListLiteral{Position: pos()})),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{"a": "int", "b": "string", "c": "bool", "d": "nil", "e": "list"}
	for k, want := range cases {
		if got, ok := r.Variables[k].(string); !ok || got != want {
			t.Errorf("type(%s) = %v, want %s", k, r.Variables[k], want)
		}
	}
}

// ---------------------------------------------------------------------------
// Report and alert statements
// ---------------------------------------------------------------------------

func TestReportStatement(t *testing.T) {
	p := prog(
		&ast.ReportStatement{
			Position: pos(),
			Fields: []ast.ReportField{
				{Key: "host", Value: strLit("server1")},
				{Key: "cpu", Value: floatLit(12.5)},
			},
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Type != "report" {
		t.Errorf("type = %s, want report", r.Output[0].Type)
	}
	data, ok := r.Output[0].Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map: %T", r.Output[0].Data)
	}
	if data["host"] != "server1" {
		t.Errorf("host = %v, want server1", data["host"])
	}
	if data["cpu"] != float64(12.5) {
		t.Errorf("cpu = %v, want 12.5", data["cpu"])
	}
}

func TestAlertStatement(t *testing.T) {
	p := prog(
		&ast.AlertStatement{
			Position: pos(),
			Message:  strLit("CPU high!"),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Type != "alert" {
		t.Errorf("type = %s, want alert", r.Output[0].Type)
	}
	if r.Output[0].Data != "CPU high!" {
		t.Errorf("data = %v, want 'CPU high!'", r.Output[0].Data)
	}
}

// ---------------------------------------------------------------------------
// Nested scopes
// ---------------------------------------------------------------------------

func TestNestedScopes(t *testing.T) {
	// Outer x=10, inner block creates x=20. After block, x should still be 10.
	p := prog(
		let("x", intLit(10)),
		block(let("x", intLit(20))),
		let("y", ident("x")),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["y"] != int64(10) {
		t.Errorf("y = %v, want 10 (outer scope preserved)", r.Variables["y"])
	}
}

func TestScopeChainLookup(t *testing.T) {
	// Function accesses variable from enclosing scope via closure.
	p := prog(
		let("base", intLit(100)),
		&ast.FnStatement{
			Position: pos(),
			Name:     ident("addToBase"),
			Params:   []ast.Parameter{{Name: ident("n")}},
			Body: block(
				&ast.ReturnStatement{Position: pos(), Value: binOp(ident("base"), "+", ident("n"))},
			),
		},
		let("r", call(ident("addToBase"), intLit(5))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["r"] != int64(105) {
		t.Errorf("r = %v, want 105", r.Variables["r"])
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestUndefinedVariable(t *testing.T) {
	p := prog(let("x", ident("nope")))
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for undefined variable")
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T: %v", err, err)
	}
	if !strings.Contains(re.Msg, "undefined variable") {
		t.Errorf("error msg = %q, want it to mention 'undefined variable'", re.Msg)
	}
}

func TestTypeMismatch(t *testing.T) {
	p := prog(let("x", binOp(strLit("hi"), "-", intLit(1))))
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
}

func TestDivisionByZero(t *testing.T) {
	p := prog(let("x", binOp(intLit(10), "/", intLit(0))))
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for division by zero")
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T: %v", err, err)
	}
	if !strings.Contains(re.Msg, "division by zero") {
		t.Errorf("error msg = %q, want it to mention 'division by zero'", re.Msg)
	}
}

func TestUnknownFunction(t *testing.T) {
	p := prog(let("x", call(ident("nonexistent_func"))))
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for unknown function")
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T: %v", err, err)
	}
	if !strings.Contains(re.Msg, "unknown function") {
		t.Errorf("error msg = %q, want it to mention 'unknown function'", re.Msg)
	}
}

// ---------------------------------------------------------------------------
// Return statements
// ---------------------------------------------------------------------------

func TestReturnValue(t *testing.T) {
	p := prog(&ast.ReturnStatement{Position: pos(), Value: intLit(99)})
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.ReturnValue != int64(99) {
		t.Errorf("ReturnValue = %v, want 99", r.ReturnValue)
	}
}

func TestReturnStopsExecution(t *testing.T) {
	// After return, subsequent statements should not execute.
	p := prog(
		let("x", intLit(1)),
		&ast.ReturnStatement{Position: pos(), Value: intLit(42)},
		let("x", intLit(999)), // should not run
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.ReturnValue != int64(42) {
		t.Errorf("ReturnValue = %v, want 42", r.ReturnValue)
	}
	if r.Variables["x"] != int64(1) {
		t.Errorf("x = %v, want 1 (not overwritten)", r.Variables["x"])
	}
}

// ---------------------------------------------------------------------------
// CallExpression with dotted names (builtin lookup)
// ---------------------------------------------------------------------------

func TestDottedCallBuiltin(t *testing.T) {
	called := false
	builtins := map[string]BuiltinFunc{
		"sys.cpu.usage": func(args ...interface{}) (interface{}, error) {
			called = true
			return map[string]interface{}{"percent": 42.0}, nil
		},
	}
	interp := New(builtins)

	// sys.cpu.usage() → MemberExpression(MemberExpression(Identifier("sys"), "cpu"), "usage")
	callExpr := call(
		member(member(ident("sys"), "cpu"), "usage"),
	)
	p := prog(let("result", callExpr))

	r, err := interp.Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("builtin sys.cpu.usage was not called")
	}
	d, ok := r.Variables["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a dict: %T", r.Variables["result"])
	}
	if d["percent"] != float64(42) {
		t.Errorf("percent = %v, want 42", d["percent"])
	}
}

// ---------------------------------------------------------------------------
// AssignStatement (identifier, index, member)
// ---------------------------------------------------------------------------

func TestAssignIdentifier(t *testing.T) {
	p := prog(
		let("x", intLit(1)),
		assign(ident("x"), intLit(42)),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != int64(42) {
		t.Errorf("x = %v, want 42", r.Variables["x"])
	}
}

func TestAssignIndex(t *testing.T) {
	p := prog(
		let("xs", &ast.ListLiteral{Position: pos(), Elements: []ast.Expression{intLit(1), intLit(2), intLit(3)}}),
		assign(index(ident("xs"), intLit(1)), intLit(99)),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	xs := r.Variables["xs"].([]interface{})
	if xs[1] != int64(99) {
		t.Errorf("xs[1] = %v, want 99", xs[1])
	}
}

func TestAssignMember(t *testing.T) {
	p := prog(
		let("d", &ast.DictLiteral{
			Position: pos(),
			Keys:     []ast.Expression{strLit("a")},
			Values:   []ast.Expression{intLit(1)},
		}),
		assign(member(ident("d"), "b"), intLit(2)),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	d := r.Variables["d"].(map[string]interface{})
	if d["b"] != int64(2) {
		t.Errorf("d[b] = %v, want 2", d["b"])
	}
}

// ---------------------------------------------------------------------------
// Unary operators
// ---------------------------------------------------------------------------

func TestUnaryNegate(t *testing.T) {
	p := prog(
		let("a", unaryOp("-", intLit(5))),
		let("b", unaryOp("-", floatLit(3.14))),
		let("c", unaryOp("!", boolLit(false))),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["a"] != int64(-5) {
		t.Errorf("a = %v, want -5", r.Variables["a"])
	}
	if r.Variables["b"] != float64(-3.14) {
		t.Errorf("b = %v, want -3.14", r.Variables["b"])
	}
	if r.Variables["c"] != true {
		t.Errorf("c = %v, want true", r.Variables["c"])
	}
}

// ---------------------------------------------------------------------------
// IfExpression (ternary)
// ---------------------------------------------------------------------------

func TestIfExpression(t *testing.T) {
	p := prog(
		let("x", intLit(10)),
		let("r", &ast.IfExpression{
			Position:  pos(),
			Condition: binOp(ident("x"), ">", intLit(5)),
			Then:      strLit("big"),
			Else:      strLit("small"),
		}),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["r"] != "big" {
		t.Errorf("r = %v, want 'big'", r.Variables["r"])
	}
}

// ---------------------------------------------------------------------------
// TaskStatement (executes body directly)
// ---------------------------------------------------------------------------

func TestTaskStatement(t *testing.T) {
	p := prog(
		&ast.TaskStatement{
			Position: pos(),
			Name:     "test_task",
			Body: block(
				let("inside", intLit(42)),
			),
		},
		// "inside" is in the task's block scope, not visible here.
		let("outside", intLit(1)),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["outside"] != int64(1) {
		t.Errorf("outside = %v, want 1", r.Variables["outside"])
	}
}

// ---------------------------------------------------------------------------
// ImportStatement (no-op)
// ---------------------------------------------------------------------------

func TestImportStatement(t *testing.T) {
	p := prog(
		&ast.ImportStatement{Position: pos(), Path: "some/path"},
		let("x", intLit(1)),
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != int64(1) {
		t.Errorf("x = %v, want 1", r.Variables["x"])
	}
}

// ---------------------------------------------------------------------------
// Nil literal
// ---------------------------------------------------------------------------

func TestNilLiteral(t *testing.T) {
	p := prog(let("x", nilLit()))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Variables["x"] != nil {
		t.Errorf("x = %v, want nil", r.Variables["x"])
	}
}

// ---------------------------------------------------------------------------
// RuntimeError.Error() formatting
// ---------------------------------------------------------------------------

func TestRuntimeErrorFormat(t *testing.T) {
	e := &RuntimeError{
		Pos: ast.Position{Line: 5, Column: 10, File: "test.ops"},
		Msg: "something broke",
	}
	want := "test.ops:5:10: something broke"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	e2 := &RuntimeError{Pos: ast.Position{Line: 3, Column: 1}, Msg: "oops"}
	want2 := "3:1: oops"
	if got := e2.Error(); got != want2 {
		t.Errorf("Error() = %q, want %q", got, want2)
	}
}

// ---------------------------------------------------------------------------
// Assign to undefined variable
// ---------------------------------------------------------------------------

func TestAssignUndefinedVariable(t *testing.T) {
	p := prog(assign(ident("ghost"), intLit(1)))
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for assigning undefined variable")
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T", err)
	}
	if !strings.Contains(re.Msg, "undefined variable") {
		t.Errorf("error msg = %q, want 'undefined variable'", re.Msg)
	}
}

// ---------------------------------------------------------------------------
// Index out of range
// ---------------------------------------------------------------------------

func TestIndexOutOfRange(t *testing.T) {
	p := prog(
		let("xs", &ast.ListLiteral{Position: pos(), Elements: []ast.Expression{intLit(1)}}),
		let("x", index(ident("xs"), intLit(5))),
	)
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for index out of range")
	}
}

// ---------------------------------------------------------------------------
// Expression statement (side-effect only)
// ---------------------------------------------------------------------------

func TestExpressionStatement(t *testing.T) {
	// A call expression used as a statement. We use the print builtin.
	p := prog(exprStmt(call(ident("len"), strLit("test"))))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	// No variables, no crash.
	if len(r.Variables) != 0 {
		t.Errorf("expected no variables, got %v", r.Variables)
	}
}

// ---------------------------------------------------------------------------
// EnsureStatement
// ---------------------------------------------------------------------------

func TestEnsureConditionTrue(t *testing.T) {
	// Condition is true => body should NOT execute
	p := prog(
		let("x", intLit(10)),
		&ast.EnsureStatement{
			Position:  pos(),
			Condition: binOp(ident("x"), ">", intLit(5)),
			Body: block(
				assign(ident("x"), intLit(999)),
			),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	// x should still be 10, body should not have run
	if r.Variables["x"] != int64(10) {
		t.Errorf("x = %v, want 10 (condition was true, body should not run)", r.Variables["x"])
	}
}

func TestEnsureConditionFalseApplySuccess(t *testing.T) {
	// Condition is false initially, body fixes it, re-check passes
	p := prog(
		let("x", intLit(0)),
		&ast.EnsureStatement{
			Position:  pos(),
			Condition: binOp(ident("x"), ">", intLit(0)),
			Body: block(
				assign(ident("x"), intLit(5)),
			),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	// x should be 5 now (body set it)
	if r.Variables["x"] != int64(5) {
		t.Errorf("x = %v, want 5", r.Variables["x"])
	}
}

func TestEnsureConditionFalseApplyFail(t *testing.T) {
	// Condition is false, body does NOT fix it => error
	p := prog(
		let("x", intLit(0)),
		&ast.EnsureStatement{
			Position:  pos(),
			Condition: binOp(ident("x"), ">", intLit(0)),
			Body: block(
				// body does nothing useful
				assign(ident("x"), intLit(0)),
			),
		},
	)
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error when ensure condition still false after apply")
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T: %v", err, err)
	}
	if !strings.Contains(re.Msg, "ensure") {
		t.Errorf("error msg = %q, want it to mention 'ensure'", re.Msg)
	}
}

// ---------------------------------------------------------------------------
// log() builtin
// ---------------------------------------------------------------------------

func TestBuiltinLog(t *testing.T) {
	p := prog(exprStmt(call(ident("log"), strLit("server started"))))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Type != "log" {
		t.Errorf("type = %s, want log", r.Output[0].Type)
	}
	if r.Output[0].Data != "server started" {
		t.Errorf("data = %v, want 'server started'", r.Output[0].Data)
	}
}

func TestBuiltinLogMultipleArgs(t *testing.T) {
	p := prog(exprStmt(call(ident("log"), strLit("count:"), intLit(42))))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Data != "count: 42" {
		t.Errorf("data = %v, want 'count: 42'", r.Output[0].Data)
	}
}

// ---------------------------------------------------------------------------
// metric() builtin
// ---------------------------------------------------------------------------

func TestBuiltinMetric(t *testing.T) {
	p := prog(exprStmt(call(ident("metric"), strLit("cpu_usage"), floatLit(42.5))))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Type != "metric" {
		t.Errorf("type = %s, want metric", r.Output[0].Type)
	}
	data, ok := r.Output[0].Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map: %T", r.Output[0].Data)
	}
	if data["name"] != "cpu_usage" {
		t.Errorf("name = %v, want cpu_usage", data["name"])
	}
	if data["value"] != "42.5" {
		t.Errorf("value = %v, want 42.5", data["value"])
	}
	if _, hasLabels := data["labels"]; hasLabels {
		t.Errorf("should not have labels key when no labels arg")
	}
}

func TestBuiltinMetricWithLabels(t *testing.T) {
	labels := &ast.DictLiteral{
		Position: pos(),
		Keys:     []ast.Expression{strLit("host")},
		Values:   []ast.Expression{strLit("server1")},
	}
	p := prog(exprStmt(call(ident("metric"), strLit("mem_used"), intLit(1024), labels)))
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	data := r.Output[0].Data.(map[string]interface{})
	if data["name"] != "mem_used" {
		t.Errorf("name = %v, want mem_used", data["name"])
	}
	if data["value"] != "1024" {
		t.Errorf("value = %v, want 1024", data["value"])
	}
	lbls, ok := data["labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("labels is not a map: %T", data["labels"])
	}
	if lbls["host"] != "server1" {
		t.Errorf("labels.host = %v, want server1", lbls["host"])
	}
}

func TestBuiltinMetricTooFewArgs(t *testing.T) {
	p := prog(exprStmt(call(ident("metric"), strLit("only_name"))))
	_, err := newInterp().Execute(p)
	if err == nil {
		t.Fatal("expected error for metric() with too few args")
	}
}

// ---------------------------------------------------------------------------
// LogStatement (AST-direct execution)
// ---------------------------------------------------------------------------

func TestLogStatement(t *testing.T) {
	p := prog(
		&ast.LogStatement{
			Position: pos(),
			Message:  strLit("something happened"),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Type != "log" {
		t.Errorf("type = %s, want log", r.Output[0].Type)
	}
	if r.Output[0].Data != "something happened" {
		t.Errorf("data = %v, want 'something happened'", r.Output[0].Data)
	}
}

// ---------------------------------------------------------------------------
// MetricStatement (AST-direct execution)
// ---------------------------------------------------------------------------

func TestMetricStatement(t *testing.T) {
	p := prog(
		&ast.MetricStatement{
			Position: pos(),
			Name:     strLit("requests"),
			Value:    intLit(100),
		},
	)
	r, err := newInterp().Execute(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Output) != 1 {
		t.Fatalf("expected 1 output entry, got %d", len(r.Output))
	}
	if r.Output[0].Type != "metric" {
		t.Errorf("type = %s, want metric", r.Output[0].Type)
	}
	data := r.Output[0].Data.(map[string]interface{})
	if data["name"] != "requests" {
		t.Errorf("name = %v, want requests", data["name"])
	}
	if data["value"] != "100" {
		t.Errorf("value = %v, want 100", data["value"])
	}
}
