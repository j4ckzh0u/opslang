package ast

import (
	"strings"
	"testing"
)

func TestPosition(t *testing.T) {
	p := Position{Line: 10, Column: 5, File: "test.ops"}
	if p.Line != 10 || p.Column != 5 || p.File != "test.ops" {
		t.Errorf("Position fields not set correctly: %+v", p)
	}
}

func TestProgramString(t *testing.T) {
	tests := []struct {
		name string
		prog *Program
		want string
	}{
		{
			name: "empty program",
			prog: &Program{Statements: []Statement{}},
			want: "",
		},
		{
			name: "single statement",
			prog: &Program{
				Statements: []Statement{
					&LetStatement{
						Name:  &Identifier{Name: "x"},
						Value: &IntegerLiteral{Value: 42},
					},
				},
			},
			want: "let x = 42",
		},
		{
			name: "multiple statements",
			prog: &Program{
				Statements: []Statement{
					&LetStatement{
						Name:  &Identifier{Name: "x"},
						Value: &IntegerLiteral{Value: 1},
					},
					&LetStatement{
						Name:  &Identifier{Name: "y"},
						Value: &IntegerLiteral{Value: 2},
					},
				},
			},
			want: "let x = 1\nlet y = 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.prog.String()
			if got != tt.want {
				t.Errorf("Program.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProgramPos(t *testing.T) {
	pos := Position{Line: 1, Column: 1}
	prog := &Program{Position: pos}
	if prog.Pos() != pos {
		t.Errorf("Program.Pos() = %v, want %v", prog.Pos(), pos)
	}
}

func TestLetStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *LetStatement
		want string
	}{
		{
			name: "simple let",
			stmt: &LetStatement{
				Name:  &Identifier{Name: "x"},
				Value: &IntegerLiteral{Value: 42},
			},
			want: "let x = 42",
		},
		{
			name: "let with string",
			stmt: &LetStatement{
				Name:  &Identifier{Name: "name"},
				Value: &StringLiteral{Value: "hello"},
			},
			want: `let name = "hello"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("LetStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFnStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *FnStatement
		want string
	}{
		{
			name: "no params",
			stmt: &FnStatement{
				Name:   &Identifier{Name: "greet"},
				Params: []Parameter{},
				Body:   &BlockStatement{},
			},
			want: "fn greet() { ... }",
		},
		{
			name: "with params",
			stmt: &FnStatement{
				Name: &Identifier{Name: "add"},
				Params: []Parameter{
					{Name: &Identifier{Name: "a"}},
					{Name: &Identifier{Name: "b"}},
				},
				Body: &BlockStatement{},
			},
			want: "fn add(a, b) { ... }",
		},
		{
			name: "with default param",
			stmt: &FnStatement{
				Name: &Identifier{Name: "greet"},
				Params: []Parameter{
					{Name: &Identifier{Name: "name"}},
					{Name: &Identifier{Name: "greeting"}, Default: &StringLiteral{Value: "hello"}},
				},
				Body: &BlockStatement{},
			},
			want: `fn greet(name, greeting = "hello") { ... }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("FnStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIfStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *IfStatement
		want string
	}{
		{
			name: "if only",
			stmt: &IfStatement{
				Condition: &BoolLiteral{Value: true},
				Body:      &BlockStatement{},
			},
			want: "if true { ... }",
		},
		{
			name: "if-else",
			stmt: &IfStatement{
				Condition:  &BoolLiteral{Value: true},
				Body:       &BlockStatement{},
				ElseClause: &BlockStatement{},
			},
			want: "if true { ... } else { ... }",
		},
		{
			name: "if-elseif",
			stmt: &IfStatement{
				Condition: &BoolLiteral{Value: true},
				Body:      &BlockStatement{},
				ElseClause: &IfStatement{
					Condition: &BoolLiteral{Value: false},
					Body:      &BlockStatement{},
				},
			},
			want: "if true { ... } else if false { ... }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("IfStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestForStatement(t *testing.T) {
	stmt := &ForStatement{
		Init:      &LetStatement{Name: &Identifier{Name: "i"}, Value: &IntegerLiteral{Value: 0}},
		Condition: &BinaryExpression{Left: &Identifier{Name: "i"}, Op: "<", Right: &IntegerLiteral{Value: 10}},
		Post:      &AssignStatement{Target: &Identifier{Name: "i"}, Value: &BinaryExpression{Left: &Identifier{Name: "i"}, Op: "+", Right: &IntegerLiteral{Value: 1}}},
		Body:      &BlockStatement{},
	}
	got := stmt.String()
	if !strings.Contains(got, "for") || !strings.Contains(got, "{ ... }") {
		t.Errorf("ForStatement.String() = %q, expected 'for ... {{ ... }}'", got)
	}
}

func TestWhileStatement(t *testing.T) {
	stmt := &WhileStatement{
		Condition: &BoolLiteral{Value: true},
		Body:      &BlockStatement{},
	}
	got := stmt.String()
	if got != "while true { ... }" {
		t.Errorf("WhileStatement.String() = %q, want %q", got, "while true { ... }")
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *ReturnStatement
		want string
	}{
		{
			name: "bare return",
			stmt: &ReturnStatement{},
			want: "return",
		},
		{
			name: "return value",
			stmt: &ReturnStatement{Value: &IntegerLiteral{Value: 42}},
			want: "return 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("ReturnStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskStatement(t *testing.T) {
	stmt := &TaskStatement{
		Name: "deploy",
		Targets: &TargetClause{
			Hosts: []Expression{&StringLiteral{Value: "host1"}, &StringLiteral{Value: "host2"}},
		},
		Body: &BlockStatement{},
	}
	got := stmt.String()
	if !strings.Contains(got, `task "deploy"`) || !strings.Contains(got, "host1") {
		t.Errorf("TaskStatement.String() = %q, unexpected format", got)
	}
}

func TestParallelStatement(t *testing.T) {
	stmt := &ParallelStatement{Body: &BlockStatement{}}
	got := stmt.String()
	if got != "parallel { ... }" {
		t.Errorf("ParallelStatement.String() = %q, want %q", got, "parallel { ... }")
	}
}

func TestImportStatement(t *testing.T) {
	stmt := &ImportStatement{Path: "sys"}
	got := stmt.String()
	if got != `import "sys"` {
		t.Errorf("ImportStatement.String() = %q, want %q", got, `import "sys"`)
	}
}

func TestPrivilegeStatement(t *testing.T) {
	tests := []struct {
		name  string
		level PrivilegeLevel
		want  string
	}{
		{"read_only", PrivilegeReadOnly, "privilege: read_only"},
		{"admin", PrivilegeAdmin, "privilege: admin"},
		{"root", PrivilegeRoot, "privilege: root"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt := &PrivilegeStatement{Level: tt.level}
			got := stmt.String()
			if got != tt.want {
				t.Errorf("PrivilegeStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpressionStatement(t *testing.T) {
	stmt := &ExpressionStatement{Expr: &CallExpression{
		Function: &Identifier{Name: "print"},
		Args:     []Expression{&StringLiteral{Value: "hello"}},
	}}
	got := stmt.String()
	if !strings.Contains(got, "print") {
		t.Errorf("ExpressionStatement.String() = %q, expected to contain 'print'", got)
	}
}

func TestReportStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *ReportStatement
		want string
	}{
		{
			name: "empty report",
			stmt: &ReportStatement{Fields: []ReportField{}},
			want: "report {  }",
		},
		{
			name: "with fields",
			stmt: &ReportStatement{
				Fields: []ReportField{
					{Key: "cpu", Value: &Identifier{Name: "cpuVal"}},
					{Key: "mem", Value: &Identifier{Name: "memVal"}},
				},
			},
			want: "report { cpu: cpuVal, mem: memVal }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("ReportStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlertStatement(t *testing.T) {
	stmt := &AlertStatement{Message: &StringLiteral{Value: "high cpu"}}
	got := stmt.String()
	if got != `alert("high cpu")` {
		t.Errorf("AlertStatement.String() = %q, want %q", got, `alert("high cpu")`)
	}
}

func TestEnsureStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *EnsureStatement
		want string
	}{
		{
			name: "without notify",
			stmt: &EnsureStatement{
				Condition: &BoolLiteral{Value: true},
				Body:      &BlockStatement{},
			},
			want: "ensure true { ... }",
		},
		{
			name: "with notify",
			stmt: &EnsureStatement{
				Condition: &BoolLiteral{Value: true},
				Body:      &BlockStatement{},
				Notify:    &StringLiteral{Value: "admin"},
			},
			want: `ensure true { ... } notify "admin"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("EnsureStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetricStatement(t *testing.T) {
	tests := []struct {
		name string
		stmt *MetricStatement
		want string
	}{
		{
			name: "without labels",
			stmt: &MetricStatement{
				Name:  &StringLiteral{Value: "cpu_usage"},
				Value: &FloatLiteral{Value: 45.5},
			},
			want: `metric("cpu_usage", 45.5)`,
		},
		{
			name: "with labels",
			stmt: &MetricStatement{
				Name:   &StringLiteral{Value: "cpu_usage"},
				Value:  &FloatLiteral{Value: 45.5},
				Labels: &DictLiteral{Keys: []Expression{&StringLiteral{Value: "host"}}, Values: []Expression{&StringLiteral{Value: "web1"}}},
			},
			want: `metric("cpu_usage", 45.5, { "host": "web1" })`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.stmt.String()
			if got != tt.want {
				t.Errorf("MetricStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogStatement(t *testing.T) {
	stmt := &LogStatement{Message: &StringLiteral{Value: "starting"}}
	got := stmt.String()
	if got != `log("starting")` {
		t.Errorf("LogStatement.String() = %q, want %q", got, `log("starting")`)
	}
}

func TestBlockStatement(t *testing.T) {
	tests := []struct {
		name string
		blk  *BlockStatement
		want string
	}{
		{
			name: "empty block",
			blk:  &BlockStatement{Statements: []Statement{}},
			want: "{  }",
		},
		{
			name: "single statement",
			blk: &BlockStatement{
				Statements: []Statement{
					&ReturnStatement{Value: &IntegerLiteral{Value: 1}},
				},
			},
			want: "{ return 1 }",
		},
		{
			name: "multiple statements",
			blk: &BlockStatement{
				Statements: []Statement{
					&LetStatement{Name: &Identifier{Name: "x"}, Value: &IntegerLiteral{Value: 1}},
					&ReturnStatement{Value: &Identifier{Name: "x"}},
				},
			},
			want: "{ let x = 1; return x }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.blk.String()
			if got != tt.want {
				t.Errorf("BlockStatement.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAssignStatement(t *testing.T) {
	stmt := &AssignStatement{
		Target: &Identifier{Name: "x"},
		Value:  &IntegerLiteral{Value: 10},
	}
	got := stmt.String()
	if got != "x = 10" {
		t.Errorf("AssignStatement.String() = %q, want %q", got, "x = 10")
	}
}

// --- Expression tests ---

func TestIntegerLiteral(t *testing.T) {
	e := &IntegerLiteral{Value: 42}
	if e.String() != "42" {
		t.Errorf("IntegerLiteral.String() = %q, want %q", e.String(), "42")
	}
}

func TestFloatLiteral(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{3.14, "3.14"},
		{0.0, "0"},
		{1.0, "1"},
	}
	for _, tt := range tests {
		e := &FloatLiteral{Value: tt.value}
		got := e.String()
		if got != tt.want {
			t.Errorf("FloatLiteral(%g).String() = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	e := &StringLiteral{Value: "hello"}
	if e.String() != `"hello"` {
		t.Errorf("StringLiteral.String() = %q, want %q", e.String(), `"hello"`)
	}
}

func TestBoolLiteral(t *testing.T) {
	tests := []struct {
		value bool
		want  string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		e := &BoolLiteral{Value: tt.value}
		if e.String() != tt.want {
			t.Errorf("BoolLiteral(%v).String() = %q, want %q", tt.value, e.String(), tt.want)
		}
	}
}

func TestNilLiteral(t *testing.T) {
	e := &NilLiteral{}
	if e.String() != "nil" {
		t.Errorf("NilLiteral.String() = %q, want %q", e.String(), "nil")
	}
}

func TestListLiteral(t *testing.T) {
	tests := []struct {
		name string
		e    *ListLiteral
		want string
	}{
		{
			name: "empty list",
			e:    &ListLiteral{Elements: []Expression{}},
			want: "[]",
		},
		{
			name: "with elements",
			e:    &ListLiteral{Elements: []Expression{&IntegerLiteral{Value: 1}, &IntegerLiteral{Value: 2}}},
			want: "[1, 2]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("ListLiteral.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDictLiteral(t *testing.T) {
	e := &DictLiteral{
		Keys:   []Expression{&StringLiteral{Value: "a"}, &StringLiteral{Value: "b"}},
		Values: []Expression{&IntegerLiteral{Value: 1}, &IntegerLiteral{Value: 2}},
	}
	got := e.String()
	want := `{ "a": 1, "b": 2 }`
	if got != want {
		t.Errorf("DictLiteral.String() = %q, want %q", got, want)
	}
}

func TestIdentifier(t *testing.T) {
	e := &Identifier{Name: "myVar"}
	if e.String() != "myVar" {
		t.Errorf("Identifier.String() = %q, want %q", e.String(), "myVar")
	}
}

func TestCallExpression(t *testing.T) {
	tests := []struct {
		name string
		e    *CallExpression
		want string
	}{
		{
			name: "no args",
			e: &CallExpression{
				Function: &Identifier{Name: "now"},
				Args:     []Expression{},
			},
			want: "now()",
		},
		{
			name: "with args",
			e: &CallExpression{
				Function: &Identifier{Name: "add"},
				Args:     []Expression{&IntegerLiteral{Value: 1}, &IntegerLiteral{Value: 2}},
			},
			want: "add(1, 2)",
		},
		{
			name: "member call",
			e: &CallExpression{
				Function: &MemberExpression{
					Object: &MemberExpression{
						Object: &Identifier{Name: "sys"},
						Member: &Identifier{Name: "cpu"},
					},
					Member: &Identifier{Name: "usage"},
				},
				Args: []Expression{},
			},
			want: "sys.cpu.usage()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("CallExpression.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBinaryExpression(t *testing.T) {
	e := &BinaryExpression{
		Left:  &IntegerLiteral{Value: 1},
		Op:    "+",
		Right: &IntegerLiteral{Value: 2},
	}
	got := e.String()
	want := "(1 + 2)"
	if got != want {
		t.Errorf("BinaryExpression.String() = %q, want %q", got, want)
	}
}

func TestUnaryExpression(t *testing.T) {
	e := &UnaryExpression{
		Op:    "-",
		Right: &IntegerLiteral{Value: 5},
	}
	got := e.String()
	want := "(-5)"
	if got != want {
		t.Errorf("UnaryExpression.String() = %q, want %q", got, want)
	}
}

func TestIndexExpression(t *testing.T) {
	e := &IndexExpression{
		Left:  &Identifier{Name: "arr"},
		Index: &IntegerLiteral{Value: 0},
	}
	got := e.String()
	want := "arr[0]"
	if got != want {
		t.Errorf("IndexExpression.String() = %q, want %q", got, want)
	}
}

func TestMemberExpression(t *testing.T) {
	e := &MemberExpression{
		Object: &Identifier{Name: "obj"},
		Member: &Identifier{Name: "field"},
	}
	got := e.String()
	want := "obj.field"
	if got != want {
		t.Errorf("MemberExpression.String() = %q, want %q", got, want)
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		name string
		e    *IfExpression
		want string
	}{
		{
			name: "with condition",
			e: &IfExpression{
				Condition: &BoolLiteral{Value: true},
				Then:      &IntegerLiteral{Value: 1},
				Else:      &IntegerLiteral{Value: 2},
			},
			want: "if true { 1 } else { 2 }",
		},
		{
			name: "nil condition",
			e: &IfExpression{
				Then: &IntegerLiteral{Value: 1},
				Else: &IntegerLiteral{Value: 2},
			},
			want: "if { 1 } else { 2 }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("IfExpression.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParameterString(t *testing.T) {
	tests := []struct {
		name string
		p    Parameter
		want string
	}{
		{
			name: "no default",
			p:    Parameter{Name: &Identifier{Name: "x"}},
			want: "x",
		},
		{
			name: "with default",
			p:    Parameter{Name: &Identifier{Name: "x"}, Default: &IntegerLiteral{Value: 10}},
			want: "x = 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("Parameter.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetClauseString(t *testing.T) {
	tests := []struct {
		name string
		tc   *TargetClause
		want string
	}{
		{
			name: "var reference",
			tc:   &TargetClause{Var: &Identifier{Name: "hosts"}},
			want: "hosts",
		},
		{
			name: "host list",
			tc: &TargetClause{
				Hosts: []Expression{
					&StringLiteral{Value: "host1"},
					&StringLiteral{Value: "host2"},
				},
			},
			want: `"host1", "host2"`,
		},
		{
			name: "group call",
			tc: &TargetClause{
				Hosts: []Expression{
					&CallExpression{
						Function: &Identifier{Name: "group"},
						Args:     []Expression{&StringLiteral{Value: "role=web"}},
					},
				},
			},
			want: `group("role=web")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tc.String(); got != tt.want {
				t.Errorf("TargetClause.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetClausePos(t *testing.T) {
	pos := Position{Line: 5, Column: 1}
	tc := &TargetClause{Position: pos, Var: &Identifier{Name: "hosts"}}
	if tc.Pos() != pos {
		t.Errorf("TargetClause.Pos() = %v, want %v", tc.Pos(), pos)
	}
}

// Test interface compliance
func TestInterfaceCompliance(t *testing.T) {
	// Statements
	var _ Statement = &LetStatement{}
	var _ Statement = &FnStatement{}
	var _ Statement = &IfStatement{}
	var _ Statement = &ForStatement{}
	var _ Statement = &WhileStatement{}
	var _ Statement = &ReturnStatement{}
	var _ Statement = &TaskStatement{}
	var _ Statement = &ParallelStatement{}
	var _ Statement = &ImportStatement{}
	var _ Statement = &PrivilegeStatement{}
	var _ Statement = &ExpressionStatement{}
	var _ Statement = &ReportStatement{}
	var _ Statement = &AlertStatement{}
	var _ Statement = &EnsureStatement{}
	var _ Statement = &MetricStatement{}
	var _ Statement = &LogStatement{}
	var _ Statement = &BlockStatement{}
	var _ Statement = &AssignStatement{}

	// Expressions
	var _ Expression = &IntegerLiteral{}
	var _ Expression = &FloatLiteral{}
	var _ Expression = &StringLiteral{}
	var _ Expression = &BoolLiteral{}
	var _ Expression = &NilLiteral{}
	var _ Expression = &ListLiteral{}
	var _ Expression = &DictLiteral{}
	var _ Expression = &Identifier{}
	var _ Expression = &CallExpression{}
	var _ Expression = &BinaryExpression{}
	var _ Expression = &UnaryExpression{}
	var _ Expression = &IndexExpression{}
	var _ Expression = &MemberExpression{}
	var _ Expression = &IfExpression{}

	// Nodes
	var _ Node = &Program{}
	var _ Node = &TargetClause{}
}

// Test Pos() for all statement types
func TestStatementPositions(t *testing.T) {
	pos := Position{Line: 1, Column: 2, File: "test.ops"}

	tests := []struct {
		name string
		node Node
	}{
		{"LetStatement", &LetStatement{Position: pos}},
		{"FnStatement", &FnStatement{Position: pos}},
		{"IfStatement", &IfStatement{Position: pos}},
		{"ForStatement", &ForStatement{Position: pos}},
		{"WhileStatement", &WhileStatement{Position: pos}},
		{"ReturnStatement", &ReturnStatement{Position: pos}},
		{"TaskStatement", &TaskStatement{Position: pos}},
		{"ParallelStatement", &ParallelStatement{Position: pos}},
		{"ImportStatement", &ImportStatement{Position: pos}},
		{"PrivilegeStatement", &PrivilegeStatement{Position: pos}},
		{"ExpressionStatement", &ExpressionStatement{Position: pos}},
		{"ReportStatement", &ReportStatement{Position: pos}},
		{"AlertStatement", &AlertStatement{Position: pos}},
		{"EnsureStatement", &EnsureStatement{Position: pos}},
		{"MetricStatement", &MetricStatement{Position: pos}},
		{"LogStatement", &LogStatement{Position: pos}},
		{"BlockStatement", &BlockStatement{Position: pos}},
		{"AssignStatement", &AssignStatement{Position: pos}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Pos(); got != pos {
				t.Errorf("%s.Pos() = %v, want %v", tt.name, got, pos)
			}
		})
	}
}

// Test Pos() for all expression types
func TestExpressionPositions(t *testing.T) {
	pos := Position{Line: 3, Column: 4, File: "test.ops"}

	tests := []struct {
		name string
		node Node
	}{
		{"IntegerLiteral", &IntegerLiteral{Position: pos}},
		{"FloatLiteral", &FloatLiteral{Position: pos}},
		{"StringLiteral", &StringLiteral{Position: pos}},
		{"BoolLiteral", &BoolLiteral{Position: pos}},
		{"NilLiteral", &NilLiteral{Position: pos}},
		{"ListLiteral", &ListLiteral{Position: pos}},
		{"DictLiteral", &DictLiteral{Position: pos}},
		{"Identifier", &Identifier{Position: pos}},
		{"CallExpression", &CallExpression{Position: pos}},
		{"BinaryExpression", &BinaryExpression{Position: pos}},
		{"UnaryExpression", &UnaryExpression{Position: pos}},
		{"IndexExpression", &IndexExpression{Position: pos}},
		{"MemberExpression", &MemberExpression{Position: pos}},
		{"IfExpression", &IfExpression{Position: pos}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Pos(); got != pos {
				t.Errorf("%s.Pos() = %v, want %v", tt.name, got, pos)
			}
		})
	}
}
