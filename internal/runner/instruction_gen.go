// Package runner implements the OpsLang universal runner that executes
// JSON instruction packages and outputs structured JSON results.
package runner

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/opsspec"
)

// InstructionGenerator converts AST statements into instruction packages
// that the Runner can execute sequentially.
//
// The runner is a linear instruction VM: it has no expression evaluator.
// The generator therefore REFUSES (with a clear error) any statement whose
// semantics it cannot preserve exactly — control flow, functions, ensure.
// Scripts that need those features must be deployed in AOT mode, where the
// full script is compiled. Silently dropping or mis-translating statements
// is how this component used to lie; that path is closed for good.
type InstructionGenerator struct {
	instructions []Instruction
	varCounter   int
}

// Generate creates an instruction package from a task statement's body.
// The task's on-clause does not affect instruction generation: target
// routing is a deploy-time concern handled by opsctl.
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

// unsupportedErr renders the standard guidance for statements the runner
// VM cannot express.
func unsupportedErr(what string) error {
	return fmt.Errorf("%s is not supported in runner mode; redeploy with --mode aot (the script is compiled to a full binary with exact semantics)", what)
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
	case *ast.LogStatement:
		return g.genLogStatement(s)
	case *ast.AssignStatement:
		return g.genAssignStatement(s)
	case *ast.ImportStatement:
		// SDK functions are globally registered; plain imports carry no
		// behavior. Third-party Go imports are not supported anywhere yet.
		if strings.HasPrefix(s.Path, "go ") || strings.HasPrefix(s.Path, "go:") {
			return fmt.Errorf("import %q: third-party Go imports are not supported yet", s.Path)
		}
		return nil
	case *ast.PrivilegeStatement:
		// Metadata only; enforced by opsctl before deployment.
		return nil
	case *ast.TaskStatement:
		// Target routing is handled by the deploy layer; the body's
		// statements translate like any other block.
		return g.genBlock(s.Body)
	case *ast.IfStatement:
		return unsupportedErr("if statement")
	case *ast.ForStatement:
		return unsupportedErr("for loop")
	case *ast.WhileStatement:
		return unsupportedErr("while loop")
	case *ast.FnStatement:
		return unsupportedErr("function definition")
	case *ast.ReturnStatement:
		return unsupportedErr("return statement")
	case *ast.EnsureStatement:
		return unsupportedErr("ensure statement")
	case *ast.ParallelStatement:
		return unsupportedErr("parallel block")
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
// Only call expressions produce instructions; other expressions are errors:
// silently dropping them hid dead code from the author.
func (g *InstructionGenerator) genExpressionStatement(s *ast.ExpressionStatement) error {
	if call, ok := s.Expr.(*ast.CallExpression); ok {
		return g.genCallAsInstruction(call)
	}
	return fmt.Errorf("expression statement has no effect in runner mode: %s (only calls can be executed)", s.Expr.String())
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

// genLogStatement handles: log(<message>)
func (g *InstructionGenerator) genLogStatement(s *ast.LogStatement) error {
	msg, err := g.genExpression(s.Message)
	if err != nil {
		return fmt.Errorf("log: %w", err)
	}

	g.emit(Instruction{
		Op:   "log",
		Args: map[string]interface{}{"message": msg},
	})
	return nil
}

// genAssignStatement handles: <target> = <value>
func (g *InstructionGenerator) genAssignStatement(s *ast.AssignStatement) error {
	targetName := expressionToString(s.Target)
	if targetName == "" {
		return fmt.Errorf("unsupported assignment target in runner mode: %s", s.Target.String())
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
// as an instruction argument. Calls become temporary variables; variable
// references become "$name" strings resolved by the executor.
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
		return "$" + e.Name, nil
	case *ast.CallExpression:
		return g.evaluateCallExpression(e)
	case *ast.ListLiteral:
		return g.evaluateListLiteral(e)
	case *ast.DictLiteral:
		return g.evaluateDictLiteral(e)
	case *ast.BinaryExpression:
		return nil, fmt.Errorf("runner mode cannot evaluate expression %s at runtime; use AOT mode for computed values", e.String())
	case *ast.UnaryExpression:
		return nil, fmt.Errorf("runner mode cannot evaluate expression %s at runtime; use AOT mode for computed values", e.String())
	case *ast.IfExpression:
		return nil, fmt.Errorf("runner mode cannot evaluate conditional expressions; use AOT mode")
	case *ast.MemberExpression:
		return nil, fmt.Errorf("runner mode cannot dereference %s; assign the call result to a variable first or use AOT mode", e.String())
	case *ast.IndexExpression:
		return nil, fmt.Errorf("runner mode cannot index into %s; use AOT mode", e.Left.String())
	default:
		return nil, fmt.Errorf("unsupported expression type in runner mode: %T", expr)
	}
}

// evaluateCallExpression handles a call expression used as a value.
// The call is extracted into a temporary variable and the reference is
// returned so later instructions can use its result.
func (g *InstructionGenerator) evaluateCallExpression(call *ast.CallExpression) (interface{}, error) {
	tmp := g.newTemp()
	if err := g.genCallAsTarget(call, tmp); err != nil {
		return nil, err
	}
	return "$" + tmp, nil
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
			return nil, fmt.Errorf("dict key %d must be a simple literal in runner mode", i)
		}
		val, err := g.genExpression(dict.Values[i])
		if err != nil {
			return nil, fmt.Errorf("dict value for key %q: %w", key, err)
		}
		result[key] = val
	}
	return result, nil
}

// genFieldValue generates the value for a report field. Sub-calls are
// extracted into temporary variables.
func (g *InstructionGenerator) genFieldValue(expr ast.Expression) (interface{}, error) {
	if call, ok := expr.(*ast.CallExpression); ok {
		tmp := g.newTemp()
		if err := genCallAsTargetSafe(g, call, tmp); err != nil {
			return nil, err
		}
		return "$" + tmp, nil
	}
	return g.genExpression(expr)
}

// ---------------------------------------------------------------------------
// Call instruction generation
// ---------------------------------------------------------------------------

// genCallAsInstruction generates an instruction for a standalone call expression.
func (g *InstructionGenerator) genCallAsInstruction(call *ast.CallExpression) error {
	opName, err := resolveOpName(call.Function)
	if err != nil {
		return err
	}

	args, err := g.buildArgs(call, opName)
	if err != nil {
		return err
	}

	g.emit(Instruction{Op: opName, Args: args})
	return nil
}

// genCallAsTargetSafe is an alias kept for internal readability.
func genCallAsTargetSafe(g *InstructionGenerator, call *ast.CallExpression, assign string) error {
	return g.genCallAsTarget(call, assign)
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
// and maps them to named parameters using the canonical opsspec signature.
// Unknown ops and argument-count mismatches are generation-time errors:
// a typo must fail the deploy, not the remote run.
func (g *InstructionGenerator) buildArgs(call *ast.CallExpression, opName string) (map[string]interface{}, error) {
	// Skip validation for the special alert/log ops that go through the
	// statement path; direct calls of print() map to log.
	checkName := opName
	if checkName == "print" {
		checkName = "log"
	}

	paramNames, known := opsspec.ArgNames(checkName)
	if !known {
		return nil, fmt.Errorf("unknown function %q (not a registered operation)", opName)
	}
	if len(call.Args) > len(paramNames) {
		// process.exec variadic: allow extra args, all named "args".
		if checkName != "process.exec" {
			return nil, fmt.Errorf("%s() takes at most %d argument(s), got %d", opName, len(paramNames), len(call.Args))
		}
	}

	args := make(map[string]interface{}, len(call.Args))
	for i, argExpr := range call.Args {
		val, err := g.genExpression(argExpr)
		if err != nil {
			return nil, fmt.Errorf("%s arg %d: %w", opName, i, err)
		}
		if checkName == "process.exec" && i >= 1 {
			// Variadic tail collapses into the "args" list.
			list, _ := args["args"].([]interface{})
			args["args"] = append(list, val)
			continue
		}
		args[paramNames[i]] = val
	}

	// Required-argument check against the canonical signature.
	for _, p := range requiredArgs(checkName, paramNames) {
		if _, ok := args[p]; !ok {
			return nil, fmt.Errorf("%s() missing required argument %q", opName, p)
		}
	}

	return args, nil
}

// requiredArgs returns the argument names that must be present. For most
// ops every declared argument is required; optional trailing arguments are
// listed here explicitly.
func requiredArgs(op string, paramNames []string) []string {
	optional := map[string]map[string]bool{
		"file.checksum": {"algo": true},
		"process.kill":  {"signal": true},
		"time.format":   {"layout": true},
		"net.http_post": {"body": true}, // an empty POST body is legal
	}
	if opt, ok := optional[op]; ok {
		var req []string
		for _, p := range paramNames {
			if !opt[p] {
				req = append(req, p)
			}
		}
		return req
	}
	return paramNames
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
		return "", fmt.Errorf("unsupported function expression: %s", fn.String())
	}
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
// Helpers
// ---------------------------------------------------------------------------

// emit appends an instruction to the generator's instruction list.
func (g *InstructionGenerator) emit(inst Instruction) {
	if inst.Args == nil {
		inst.Args = make(map[string]interface{})
	}
	g.instructions = append(g.instructions, inst)
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
