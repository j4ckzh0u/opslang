// Package ast defines all AST node types for the OpsLang language.
package ast

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Position
// ---------------------------------------------------------------------------

// Position represents a source location in an OpsLang script.
type Position struct {
	Line   int
	Column int
	File   string
}

// ---------------------------------------------------------------------------
// Core interfaces
// ---------------------------------------------------------------------------

// Node is the base interface implemented by every AST node.
type Node interface {
	Pos() Position
	String() string
}

// Statement is a node that can appear in a statement list.
type Statement interface {
	Node
	statementNode()
}

// Expression is a node that produces a value.
type Expression interface {
	Node
	expressionNode()
}

// ---------------------------------------------------------------------------
// Program — top-level container
// ---------------------------------------------------------------------------

// Program is the root node of every OpsLang script.
type Program struct {
	Position   Position
	Statements []Statement
}

func (p *Program) Pos() Position { return p.Position }
func (p *Program) String() string {
	var b strings.Builder
	for i, s := range p.Statements {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("%s", s))
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Statement types
// ---------------------------------------------------------------------------

// LetStatement represents: let <Name> = <Value>
type LetStatement struct {
	Position Position
	Name     *Identifier
	Value    Expression
}

func (s *LetStatement) Pos() Position  { return s.Position }
func (s *LetStatement) statementNode() {}
func (s *LetStatement) String() string {
	return fmt.Sprintf("let %s = %s", s.Name, s.Value)
}

// FnStatement represents: fn <Name>(<Params>) { <Body> }
type FnStatement struct {
	Position Position
	Name     *Identifier
	Params   []Parameter
	Body     *BlockStatement
}

func (s *FnStatement) Pos() Position  { return s.Position }
func (s *FnStatement) statementNode() {}
func (s *FnStatement) String() string {
	params := make([]string, len(s.Params))
	for i, p := range s.Params {
		params[i] = p.String()
	}
	return fmt.Sprintf("fn %s(%s) { ... }", s.Name, strings.Join(params, ", "))
}

// IfStatement represents: if <Condition> { <Body> } [ else <ElseClause> ]
// ElseClause may be *BlockStatement or *IfStatement (for else-if chains).
type IfStatement struct {
	Position   Position
	Condition  Expression
	Body       *BlockStatement
	ElseClause Node // *BlockStatement or *IfStatement; nil when no else
}

func (s *IfStatement) Pos() Position  { return s.Position }
func (s *IfStatement) statementNode() {}
func (s *IfStatement) String() string {
	out := fmt.Sprintf("if %s { ... }", s.Condition)
	if s.ElseClause != nil {
		switch e := s.ElseClause.(type) {
		case *BlockStatement:
			out += " else { ... }"
		case *IfStatement:
			out += " else " + e.String()
		}
	}
	return out
}

// ForStatement represents: for <Init>; <Condition>; <Post> { <Body> }
type ForStatement struct {
	Position  Position
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (s *ForStatement) Pos() Position  { return s.Position }
func (s *ForStatement) statementNode() {}
func (s *ForStatement) String() string {
	return fmt.Sprintf("for %s; %s; %s { ... }", s.Init, s.Condition, s.Post)
}

// WhileStatement represents: while <Condition> { <Body> }
type WhileStatement struct {
	Position  Position
	Condition Expression
	Body      *BlockStatement
}

func (s *WhileStatement) Pos() Position  { return s.Position }
func (s *WhileStatement) statementNode() {}
func (s *WhileStatement) String() string {
	return fmt.Sprintf("while %s { ... }", s.Condition)
}

// ReturnStatement represents: return [<Value>]
type ReturnStatement struct {
	Position Position
	Value    Expression // nil when returning no value
}

func (s *ReturnStatement) Pos() Position  { return s.Position }
func (s *ReturnStatement) statementNode() {}
func (s *ReturnStatement) String() string {
	if s.Value == nil {
		return "return"
	}
	return fmt.Sprintf("return %s", s.Value)
}

// TaskStatement represents: task "<Name>" on <Targets> { <Body> }
type TaskStatement struct {
	Position Position
	Name     string
	Targets  *TargetClause
	Body     *BlockStatement
}

func (s *TaskStatement) Pos() Position  { return s.Position }
func (s *TaskStatement) statementNode() {}
func (s *TaskStatement) String() string {
	return fmt.Sprintf("task %q on %s { ... }", s.Name, s.Targets)
}

// ImportStatement represents: import "<Path>"
type ImportStatement struct {
	Position Position
	Path     string
}

func (s *ImportStatement) Pos() Position  { return s.Position }
func (s *ImportStatement) statementNode() {}
func (s *ImportStatement) String() string {
	return fmt.Sprintf("import %q", s.Path)
}

// ExpressionStatement wraps an expression used as a statement.
type ExpressionStatement struct {
	Position Position
	Expr     Expression
}

func (s *ExpressionStatement) Pos() Position  { return s.Position }
func (s *ExpressionStatement) statementNode() {}
func (s *ExpressionStatement) String() string {
	return fmt.Sprintf("%s", s.Expr)
}

// ReportField is a single key-value pair inside a report { } block.
type ReportField struct {
	Key   string
	Value Expression
}

// ReportStatement represents: report { key: value, ... }
type ReportStatement struct {
	Position Position
	Fields   []ReportField
}

func (s *ReportStatement) Pos() Position  { return s.Position }
func (s *ReportStatement) statementNode() {}
func (s *ReportStatement) String() string {
	pairs := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		pairs[i] = fmt.Sprintf("%s: %s", f.Key, f.Value)
	}
	return fmt.Sprintf("report { %s }", strings.Join(pairs, ", "))
}

// AlertStatement represents: alert(<Message>)
type AlertStatement struct {
	Position Position
	Message  Expression
}

func (s *AlertStatement) Pos() Position  { return s.Position }
func (s *AlertStatement) statementNode() {}
func (s *AlertStatement) String() string {
	return fmt.Sprintf("alert(%s)", s.Message)
}

// EnsureStatement represents: ensure <condition> { <Body> }
// Implements check → apply → verify semantics.
type EnsureStatement struct {
	Position  Position
	Condition Expression      // the condition to ensure
	Body      *BlockStatement // actions to take if condition is false
	Notify    Expression      // optional notification expression (nil if not set)
}

func (s *EnsureStatement) Pos() Position  { return s.Position }
func (s *EnsureStatement) statementNode() {}
func (s *EnsureStatement) String() string {
	out := fmt.Sprintf("ensure %s { ... }", s.Condition)
	if s.Notify != nil {
		out += " notify " + s.Notify.String()
	}
	return out
}

// MetricStatement represents: metric(name, value, labels)
type MetricStatement struct {
	Position Position
	Name     Expression
	Value    Expression
	Labels   Expression // dict expression or nil
}

func (s *MetricStatement) Pos() Position  { return s.Position }
func (s *MetricStatement) statementNode() {}
func (s *MetricStatement) String() string {
	if s.Labels != nil {
		return fmt.Sprintf("metric(%s, %s, %s)", s.Name, s.Value, s.Labels)
	}
	return fmt.Sprintf("metric(%s, %s)", s.Name, s.Value)
}

// LogStatement represents: log(msg)
type LogStatement struct {
	Position Position
	Message  Expression
}

func (s *LogStatement) Pos() Position  { return s.Position }
func (s *LogStatement) statementNode() {}
func (s *LogStatement) String() string {
	return fmt.Sprintf("log(%s)", s.Message)
}

// BlockStatement is a sequence of statements enclosed in braces.
type BlockStatement struct {
	Position   Position
	Statements []Statement
}

func (s *BlockStatement) Pos() Position  { return s.Position }
func (s *BlockStatement) statementNode() {}
func (s *BlockStatement) String() string {
	parts := make([]string, len(s.Statements))
	for i, st := range s.Statements {
		parts[i] = fmt.Sprintf("%s", st)
	}
	return "{ " + strings.Join(parts, "; ") + " }"
}

// AssignStatement represents: <Target> = <Value>
// Target is typically *Identifier, *IndexExpression, or *MemberExpression.
type AssignStatement struct {
	Position Position
	Target   Expression
	Value    Expression
}

func (s *AssignStatement) Pos() Position  { return s.Position }
func (s *AssignStatement) statementNode() {}
func (s *AssignStatement) String() string {
	return fmt.Sprintf("%s = %s", s.Target, s.Value)
}

// ---------------------------------------------------------------------------
// Expression types
// ---------------------------------------------------------------------------

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Position Position
	Value    int64
}

func (e *IntegerLiteral) Pos() Position   { return e.Position }
func (e *IntegerLiteral) expressionNode() {}
func (e *IntegerLiteral) String() string  { return fmt.Sprintf("%d", e.Value) }

// FloatLiteral represents a floating-point value.
type FloatLiteral struct {
	Position Position
	Value    float64
}

func (e *FloatLiteral) Pos() Position   { return e.Position }
func (e *FloatLiteral) expressionNode() {}
func (e *FloatLiteral) String() string  { return fmt.Sprintf("%g", e.Value) }

// StringLiteral represents a quoted string value.
type StringLiteral struct {
	Position Position
	Value    string
}

func (e *StringLiteral) Pos() Position   { return e.Position }
func (e *StringLiteral) expressionNode() {}
func (e *StringLiteral) String() string  { return fmt.Sprintf("%q", e.Value) }

// BoolLiteral represents true or false.
type BoolLiteral struct {
	Position Position
	Value    bool
}

func (e *BoolLiteral) Pos() Position   { return e.Position }
func (e *BoolLiteral) expressionNode() {}
func (e *BoolLiteral) String() string {
	if e.Value {
		return "true"
	}
	return "false"
}

// NilLiteral represents the nil value.
type NilLiteral struct {
	Position Position
}

func (e *NilLiteral) Pos() Position   { return e.Position }
func (e *NilLiteral) expressionNode() {}
func (e *NilLiteral) String() string  { return "nil" }

// ListLiteral represents [elem1, elem2, ...]
type ListLiteral struct {
	Position Position
	Elements []Expression
}

func (e *ListLiteral) Pos() Position   { return e.Position }
func (e *ListLiteral) expressionNode() {}
func (e *ListLiteral) String() string {
	elems := make([]string, len(e.Elements))
	for i, el := range e.Elements {
		elems[i] = fmt.Sprintf("%s", el)
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

// DictLiteral represents { key: value, ... }
// Keys and Values are parallel slices of equal length.
type DictLiteral struct {
	Position Position
	Keys     []Expression
	Values   []Expression
}

func (e *DictLiteral) Pos() Position   { return e.Position }
func (e *DictLiteral) expressionNode() {}
func (e *DictLiteral) String() string {
	pairs := make([]string, len(e.Keys))
	for i := range e.Keys {
		pairs[i] = fmt.Sprintf("%s: %s", e.Keys[i], e.Values[i])
	}
	return fmt.Sprintf("{ %s }", strings.Join(pairs, ", "))
}

// Identifier represents a variable or field name.
type Identifier struct {
	Position Position
	Name     string
}

func (e *Identifier) Pos() Position   { return e.Position }
func (e *Identifier) expressionNode() {}
func (e *Identifier) String() string  { return e.Name }

// CallExpression represents a function call: <Function>(<Args>...)
// Function can be *Identifier (e.g. print(x)) or *MemberExpression
// (e.g. sys.cpu.usage()) to support dotted calls.
type CallExpression struct {
	Position Position
	Function Expression
	Args     []Expression
}

func (e *CallExpression) Pos() Position   { return e.Position }
func (e *CallExpression) expressionNode() {}
func (e *CallExpression) String() string {
	args := make([]string, len(e.Args))
	for i, a := range e.Args {
		args[i] = fmt.Sprintf("%s", a)
	}
	return fmt.Sprintf("%s(%s)", e.Function, strings.Join(args, ", "))
}

// BinaryExpression represents: <Left> <Op> <Right>
type BinaryExpression struct {
	Position Position
	Left     Expression
	Op       string
	Right    Expression
}

func (e *BinaryExpression) Pos() Position   { return e.Position }
func (e *BinaryExpression) expressionNode() {}
func (e *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Op, e.Right)
}

// UnaryExpression represents a prefix unary operator: <Op><Right>
type UnaryExpression struct {
	Position Position
	Op       string
	Right    Expression
}

func (e *UnaryExpression) Pos() Position   { return e.Position }
func (e *UnaryExpression) expressionNode() {}
func (e *UnaryExpression) String() string {
	return fmt.Sprintf("(%s%s)", e.Op, e.Right)
}

// IndexExpression represents: <Left>[<Index>]
type IndexExpression struct {
	Position Position
	Left     Expression
	Index    Expression
}

func (e *IndexExpression) Pos() Position   { return e.Position }
func (e *IndexExpression) expressionNode() {}
func (e *IndexExpression) String() string {
	return fmt.Sprintf("%s[%s]", e.Left, e.Index)
}

// MemberExpression represents: <Object>.<Member>
// Used for dotted access such as sys.cpu or result.value.
type MemberExpression struct {
	Position Position
	Object   Expression
	Member   *Identifier
}

func (e *MemberExpression) Pos() Position   { return e.Position }
func (e *MemberExpression) expressionNode() {}
func (e *MemberExpression) String() string {
	return fmt.Sprintf("%s.%s", e.Object, e.Member)
}

// IfExpression represents a ternary-style conditional expression:
//
//	if <Condition> { <Then> } else { <Else> }
//
// When used as a short expression (e.g. inline), Condition may be nil and
// only Then/Else carry the values.
type IfExpression struct {
	Position  Position
	Condition Expression
	Then      Expression
	Else      Expression
}

func (e *IfExpression) Pos() Position   { return e.Position }
func (e *IfExpression) expressionNode() {}
func (e *IfExpression) String() string {
	if e.Condition == nil {
		return fmt.Sprintf("if { %s } else { %s }", e.Then, e.Else)
	}
	return fmt.Sprintf("if %s { %s } else { %s }", e.Condition, e.Then, e.Else)
}

// ---------------------------------------------------------------------------
// Helper types
// ---------------------------------------------------------------------------

// Parameter describes a single function parameter.
// Default is nil when the parameter has no default value.
type Parameter struct {
	Name    *Identifier
	Default Expression
}

// String returns the parameter in a debug-friendly form.
func (p Parameter) String() string {
	if p.Default != nil {
		return fmt.Sprintf("%s = %s", p.Name, p.Default)
	}
	return p.Name.String()
}

// TargetClause represents the `on <targets>` part of a task statement.
// It can hold a list of host expressions (string literals or call
// expressions like group("role=web")), or a variable reference.
type TargetClause struct {
	Position Position
	Hosts    []Expression // string literals or CallExpression for group()
	Var      *Identifier  // non-nil when referencing a variable
}

func (t *TargetClause) Pos() Position { return t.Position }
func (t *TargetClause) String() string {
	if t.Var != nil {
		return fmt.Sprintf("%s", t.Var)
	}
	parts := make([]string, len(t.Hosts))
	for i, h := range t.Hosts {
		parts[i] = fmt.Sprintf("%s", h)
	}
	return strings.Join(parts, ", ")
}
