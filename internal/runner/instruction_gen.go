// Package runner implements the OpsLang universal runner that executes
// JSON instruction packages and outputs structured JSON results.
package runner

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/opslang/opslang/internal/ast"
)

// InstructionGenerator converts AST statements into instruction packages
// that the Runner can execute sequentially.
type InstructionGenerator struct {
	instructions []Instruction
	varCounter   int
}

// Generate creates an instruction package from a task statement's body.
func (g *InstructionGenerator) Generate(task *ast.TaskStatement, dryRun bool) (*InstructionPackage, error) {
	g.instructions = nil
	g.varCounter = 0

	if task.Body == nil {
		return &InstructionPackage{
			Version:      "1.0",
			TaskID:       generateTaskID(),
			DryRun:       dryRun,
			Instructions: []Instruction{},
		}, nil
	}

	if err := g.genBlock(task.Body); err != nil {
		return nil, err
	}

	return &InstructionPackage{
		Version:      "1.0",
		TaskID:       generateTaskID(),
		DryRun:       dryRun,
		Instructions: g.instructions,
	}, nil
}

// GenerateFromStatements creates an instruction package from a list of statements.
func (g *InstructionGenerator) GenerateFromStatements(stmts []ast.Statement, dryRun bool) (*InstructionPackage, error) {
	g.instructions = nil
	g.varCounter = 0

	for _, stmt := range stmts {
		if err := g.genStatement(stmt); err != nil {
			return nil, err
		}
	}

	return &InstructionPackage{
		Version:      "1.0",
		TaskID:       generateTaskID(),
		DryRun:       dryRun,
		Instructions: g.instructions,
	}, nil
}

// genBlock iterates over all statements in a block and generates instructions.
func (g *InstructionGenerator) genBlock(block *ast.BlockStatement) error {
	if block == nil {
		return nil
	}
	for _, stmt := range block.Statements {
		if err := g.genStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

// genStatement dispatches statement generation by concrete type.
func (g *InstructionGenerator) genStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return g.genLetStatement(s)
	case *ast.ExpressionStatement:
		return g.genExpressionStatement(s)
	case *ast.ReportStatement:
		return g.genReportStatement(s)
	case *ast.AlertStatement:
		return g.genAlertStatement(s)
	case *ast.IfStatement:
		return g.genIfStatement(s)
	case *ast.AssignStatement:
		return g.genAssignStatement(s)
	case *ast.ForStatement:
		g.addWarning("for loops are not supported in runner mode; use AOT compilation")
		return nil
	case *ast.WhileStatement:
		g.addWarning("while loops are not supported in runner mode; use AOT compilation")
		return nil
	case *ast.ReturnStatement:
		// Return statements are no-ops in runner mode.
		return nil
	case *ast.FnStatement:
		g.addWarning("function definitions are not supported in runner mode; use AOT compilation")
		return nil
	case *ast.TaskStatement:
		return g.genBlock(s.Body)
	default:
		return fmt.Errorf("unsupported statement type in runner mode: %T", stmt)
	}
}

// genLetStatement handles: let <name> = <expr>
// When the value is a function call, the call becomes an instruction with Assign.
func (g *InstructionGenerator) genLetStatement(s *ast.LetStatement) error {
	if call, ok := s.Value.(*ast.CallExpression); ok {
		return g.genCallAsTarget(call, s.Name.Name)
	}

	// For non-call values, evaluate the expression and assign directly.
	value, err := g.genExpression(s.Value)
	if err != nil {
		return fmt.Errorf("let %s: %w", s.Name.Name, err)
	}

	g.emit(Instruction{
		Op:     "set",
		Args:   map[string]interface{}{"value": value},
		Assign: s.Name.Name,
	})
	return nil
}

// genExpressionStatement handles standalone expression statements.
// Only call expressions produce instructions; other expressions are no-ops.
func (g *InstructionGenerator) genExpressionStatement(s *ast.ExpressionStatement) error {
	if call, ok := s.Expr.(*ast.CallExpression); ok {
		return g.genCallAsInstruction(call)
	}
	// Non-call expressions used as statements are silently ignored.
	return nil
}

// genReportStatement handles: report { key: value, ... }
// Sub-calls in field values are extracted into temporary variables first.
func (g *InstructionGenerator) genReportStatement(s *ast.ReportStatement) error {
	args := make(map[string]interface{}, len(s.Fields))

	for _, field := range s.Fields {
		val, err := g.genFieldValue(field.Value)
		if err != nil {
			return fmt.Errorf("report field %q: %w", field.Key, err)
		}
		args[field.Key] = val
	}

	g.emit(Instruction{
		Op:   "report",
		Args: args,
	})
	return nil
}

// genAlertStatement handles: alert(<message>)
func (g *InstructionGenerator) genAlertStatement(s *ast.AlertStatement) error {
	msg, err := g.genExpression(s.Message)
	if err != nil {
		return fmt.Errorf("alert: %w", err)
	}

	g.emit(Instruction{
		Op:   "alert",
		Args: map[string]interface{}{"message": msg},
	})
	return nil
}

// genIfStatement handles if statements. Complex control flow is not directly
// translatable to linear instructions, so the condition is evaluated for
// side-effect calls and a warning is emitted.
func (g *InstructionGenerator) genIfStatement(s *ast.IfStatement) error {
	// Evaluate the condition expression for any sub-calls (e.g. sys.cpu.usage()).
	_, err := g.genExpression(s.Condition)
	if err != nil {
		return fmt.Errorf("if condition: %w", err)
	}

	g.addWarning("if statements are not fully supported in runner mode; use AOT compilation")
	return nil
}

// genAssignStatement handles: <target> = <value>
func (g *InstructionGenerator) genAssignStatement(s *ast.AssignStatement) error {
	targetName := expressionToString(s.Target)
	if targetName == "" {
		return fmt.Errorf("unsupported assignment target: %T", s.Target)
	}

	if call, ok := s.Value.(*ast.CallExpression); ok {
		return g.genCallAsTarget(call, targetName)
	}

	value, err := g.genExpression(s.Value)
	if err != nil {
		return fmt.Errorf("assign %s: %w", targetName, err)
	}

	g.emit(Instruction{
		Op:     "set",
		Args:   map[string]interface{}{"value": value},
		Assign: targetName,
	})
	return nil
}

// ---------------------------------------------------------------------------
// Expression evaluation
// ---------------------------------------------------------------------------

// genExpression converts an AST expression into a value suitable for use
// as an instruction argument. Literals are converted directly; calls are
// extracted into temporary variables; variable references become strings.
func (g *InstructionGenerator) genExpression(expr ast.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value, nil
	case *ast.FloatLiteral:
		return e.Value, nil
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.BoolLiteral:
		return e.Value, nil
	case *ast.NilLiteral:
		return nil, nil
	case *ast.Identifier:
		return e.Name, nil
	case *ast.CallExpression:
		return g.evaluateCallExpression(e)
	case *ast.ListLiteral:
		return g.evaluateListLiteral(e)
	case *ast.DictLiteral:
		return g.evaluateDictLiteral(e)
	case *ast.BinaryExpression:
		return g.evaluateBinaryExpression(e)
	case *ast.MemberExpression:
		return resolveMemberPath(e), nil
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// evaluateCallExpression handles a call expression used as a value.
// The call is extracted into a temporary variable and the variable name
// is returned so the caller can reference it.
func (g *InstructionGenerator) evaluateCallExpression(call *ast.CallExpression) (interface{}, error) {
	tmp := g.newTemp()
	if err := g.genCallAsTarget(call, tmp); err != nil {
		return nil, err
	}
	return tmp, nil
}

// evaluateListLiteral converts a list literal into []interface{}.
func (g *InstructionGenerator) evaluateListLiteral(list *ast.ListLiteral) (interface{}, error) {
	result := make([]interface{}, len(list.Elements))
	for i, elem := range list.Elements {
		val, err := g.genExpression(elem)
		if err != nil {
			return nil, fmt.Errorf("list element %d: %w", i, err)
		}
		result[i] = val
	}
	return result, nil
}

// evaluateDictLiteral converts a dict literal into map[string]interface{}.
func (g *InstructionGenerator) evaluateDictLiteral(dict *ast.DictLiteral) (interface{}, error) {
	result := make(map[string]interface{}, len(dict.Keys))
	for i := range dict.Keys {
		key := expressionToString(dict.Keys[i])
		if key == "" {
			return nil, fmt.Errorf("dict key %d is not a simple expression", i)
		}
		val, err := g.genExpression(dict.Values[i])
		if err != nil {
			return nil, fmt.Errorf("dict value for key %q: %w", key, err)
		}
		result[key] = val
	}
	return result, nil
}

// evaluateBinaryExpression handles binary expressions. Numeric literals are
// folded at generation time; otherwise sub-calls are extracted and a string
// representation is returned.
func (g *InstructionGenerator) evaluateBinaryExpression(bin *ast.BinaryExpression) (interface{}, error) {
	// Try constant folding for numeric literals.
	leftLit, leftIsLit := bin.Left.(*ast.IntegerLiteral)
	rightLit, rightIsLit := bin.Right.(*ast.IntegerLiteral)
	if leftIsLit && rightIsLit {
		switch bin.Op {
		case "+":
			return leftLit.Value + rightLit.Value, nil
		case "-":
			return leftLit.Value - rightLit.Value, nil
		case "*":
			return leftLit.Value * rightLit.Value, nil
		}
	}

	// Evaluate both sides for sub-calls (they may produce temp variables).
	left, err := g.genExpression(bin.Left)
	if err != nil {
		return nil, fmt.Errorf("binary left: %w", err)
	}
	right, err := g.genExpression(bin.Right)
	if err != nil {
		return nil, fmt.Errorf("binary right: %w", err)
	}

	return fmt.Sprintf("%v %s %v", left, bin.Op, right), nil
}

// genFieldValue generates the value for a report field. Sub-calls are
// extracted into temporary variables.
func (g *InstructionGenerator) genFieldValue(expr ast.Expression) (interface{}, error) {
	if call, ok := expr.(*ast.CallExpression); ok {
		tmp := g.newTemp()
		if err := g.genCallAsTarget(call, tmp); err != nil {
			return nil, err
		}
		return tmp, nil
	}
	return g.genExpression(expr)
}

// ---------------------------------------------------------------------------
// Call instruction generation
// ---------------------------------------------------------------------------

// genCallAsInstruction generates an instruction for a standalone call expression.
// The result is NOT assigned to a variable.
func (g *InstructionGenerator) genCallAsInstruction(call *ast.CallExpression) error {
	opName := resolveFunctionName(call.Function)
	if opName == "" {
		return fmt.Errorf("unsupported function call: %s", call.Function)
	}

	args, err := g.buildArgs(call, opName)
	if err != nil {
		return err
	}

	g.emit(Instruction{Op: opName, Args: args})
	return nil
}

// genCallAsTarget generates an instruction for a call expression and assigns
// the result to the given variable name.
func (g *InstructionGenerator) genCallAsTarget(call *ast.CallExpression, assign string) error {
	opName, err := resolveOpName(call.Function)
	if err != nil {
		return err
	}

	args, err := g.buildArgs(call, opName)
	if err != nil {
		return err
	}

	g.emit(Instruction{Op: opName, Args: args, Assign: assign})
	return nil
}

// buildArgs extracts argument values from a call expression's argument list
// and maps them to named parameters using the known operation signatures.
func (g *InstructionGenerator) buildArgs(call *ast.CallExpression, opName string) (map[string]interface{}, error) {
	paramNames := argNamesForOp(opName)
	args := make(map[string]interface{}, len(call.Args))

	for i, argExpr := range call.Args {
		val, err := g.genExpression(argExpr)
		if err != nil {
			return nil, fmt.Errorf("arg %d: %w", i, err)
		}
		paramName := paramNames[i]
		args[paramName] = val
	}
	return args, nil
}

// ---------------------------------------------------------------------------
// Function name resolution
// ---------------------------------------------------------------------------

// resolveOpName converts a function expression into an operation name.
// It handles dotted member expressions (sys.cpu.usage), built-in aliases
// (print -> log), and bare identifiers.
func resolveOpName(fn ast.Expression) (string, error) {
	switch f := fn.(type) {
	case *ast.Identifier:
		switch f.Name {
		case "print":
			return "log", nil
		case "log":
			return "log", nil
		default:
			return f.Name, nil
		}
	case *ast.MemberExpression:
		return resolveMemberPath(f), nil
	default:
		return "", fmt.Errorf("unsupported function expression type: %T", fn)
	}
}

// resolveFunctionName is like resolveOpName but returns empty string instead of error.
func resolveFunctionName(fn ast.Expression) string {
	name, _ := resolveOpName(fn)
	return name
}

// resolveMemberPath flattens a chain of MemberExpression nodes into a
// dotted path string. For example, sys.cpu.usage -> "sys.cpu.usage".
func resolveMemberPath(member *ast.MemberExpression) string {
	var parts []string
	current := ast.Expression(member)

	for {
		switch node := current.(type) {
		case *ast.MemberExpression:
			parts = append([]string{node.Member.Name}, parts...)
			current = node.Object
		case *ast.Identifier:
			parts = append([]string{node.Name}, parts...)
			return strings.Join(parts, ".")
		default:
			return strings.Join(parts, ".")
		}
	}
}

// ---------------------------------------------------------------------------
// Argument name mapping
// ---------------------------------------------------------------------------

// argNamesForOp returns the positional parameter names for a known operation.
// Unknown operations get generic names (arg0, arg1, ...).
func argNamesForOp(op string) []string {
	switch op {
	case "sys.disk.usage":
		return []string{"path"}
	case "file.read":
		return []string{"path"}
	case "file.write":
		return []string{"path", "content"}
	case "file.append":
		return []string{"path", "content"}
	case "file.copy":
		return []string{"src", "dst"}
	case "file.move":
		return []string{"src", "dst"}
	case "file.delete":
		return []string{"path"}
	case "file.exists":
		return []string{"path"}
	case "file.info":
		return []string{"path"}
	case "file.list":
		return []string{"dir"}
	case "file.mkdir":
		return []string{"path"}
	case "file.checksum":
		return []string{"path", "algo"}
	case "net.http.get":
		return []string{"url"}
	case "net.http.post":
		return []string{"url", "body"}
	case "net.tcp.ping":
		return []string{"host", "port"}
	case "net.dns.resolve":
		return []string{"host"}
	case "process.find.by_name":
		return []string{"name"}
	case "process.find.by_port":
		return []string{"port"}
	case "process.exec":
		return []string{"command", "args"}
	case "service.status", "service.start", "service.stop",
		"service.restart", "service.enable", "service.disable":
		return []string{"name"}
	case "pkg.install":
		return []string{"name"}
	case "pkg.remove":
		return []string{"name"}
	case "pkg.search":
		return []string{"name"}
	case "time.format":
		return []string{"layout", "ts"}
	case "time.parse":
		return []string{"layout", "value"}
	case "time.diff":
		return []string{"t1", "t2"}
	case "time.sleep":
		return []string{"ms"}
	case "log":
		return []string{"message"}
	case "alert":
		return []string{"message"}
	default:
		return nil // caller falls back to arg0, arg1, ...
	}
}

// getParamName returns the parameter name for position i of the given operation.
func getParamName(op string, i int) string {
	names := argNamesForOp(op)
	if i < len(names) {
		return names[i]
	}
	return fmt.Sprintf("arg%d", i)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// emit appends an instruction to the generator's instruction list.
func (g *InstructionGenerator) emit(inst Instruction) {
	if inst.Args == nil {
		inst.Args = make(map[string]interface{})
	}
	g.instructions = append(g.instructions, inst)
}

// addWarning emits a log instruction with a warning message.
func (g *InstructionGenerator) addWarning(msg string) {
	g.emit(Instruction{
		Op:   "log",
		Args: map[string]interface{}{"message": "[runner-mode warning] " + msg},
	})
}

// newTemp returns a fresh temporary variable name.
func (g *InstructionGenerator) newTemp() string {
	g.varCounter++
	return fmt.Sprintf("__tmp_%d", g.varCounter-1)
}

// generateTaskID creates a random 12-character hex task identifier.
func generateTaskID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// expressionToString converts simple expressions (identifiers, member
// expressions, string literals) to a string representation. Returns
// empty string for unsupported types.
func expressionToString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		return resolveMemberPath(e)
	case *ast.StringLiteral:
		return e.Value
	default:
		return ""
	}
}
