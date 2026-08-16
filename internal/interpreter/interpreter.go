// Package interpreter implements AST traversal execution for OpsLang.
package interpreter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/security"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// BuiltinFunc is a Go-implemented function callable from OpsLang.
type BuiltinFunc func(args ...interface{}) (interface{}, error)

// OutputEntry represents a single output (report, alert, or print).
type OutputEntry struct {
	Type string      // "report", "alert", "print"
	Data interface{} // the data payload
}

// Result holds the execution result of a program.
type Result struct {
	Output      []OutputEntry
	ReturnValue interface{}
	Variables   map[string]interface{}
}

// RuntimeError represents a runtime error with position.
type RuntimeError struct {
	Pos ast.Position
	Msg string
}

func (e *RuntimeError) Error() string {
	if e.Pos.File != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Msg)
	}
	return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Column, e.Msg)
}

// FunctionValue represents a user-defined function.
type FunctionValue struct {
	Params  []ast.Parameter
	Body    *ast.BlockStatement
	Closure *Environment
	Name    string
}

func (f *FunctionValue) String() string {
	return fmt.Sprintf("<fn %s>", f.Name)
}

// returnSignal is used to propagate return values through the call stack.
type returnSignal struct {
	Value interface{}
}

func (r *returnSignal) Error() string { return "return" }

// ---------------------------------------------------------------------------
// Environment (scope chain)
// ---------------------------------------------------------------------------

// Environment is a variable scope.
type Environment struct {
	vars   map[string]interface{}
	parent *Environment
	// barrier: when true, update() writes into this scope instead of
	// walking up to the parent. Parallel-block goroutines use barrier
	// environments so concurrent statements never write shared maps;
	// execParallel merges the captured variables back serially afterwards.
	barrier bool
}

func newEnv(parent *Environment) *Environment {
	return &Environment{
		vars:   make(map[string]interface{}),
		parent: parent,
	}
}

func (e *Environment) get(name string) (interface{}, bool) {
	v, ok := e.vars[name]
	if !ok && e.parent != nil {
		return e.parent.get(name)
	}
	return v, ok
}

func (e *Environment) set(name string, value interface{}) {
	e.vars[name] = value
}

// update walks the scope chain to find and update an existing variable.
// Returns an error if the variable is not found. With barrier set, writes
// are captured in the current scope instead.
func (e *Environment) update(name string, value interface{}) error {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = value
		return nil
	}
	if e.barrier {
		e.vars[name] = value
		return nil
	}
	if e.parent != nil {
		if _, ok := e.parent.lookup(name); ok {
			return e.parent.update(name, value)
		}
	}
	return fmt.Errorf("undefined variable: %s", name)
}

// lookup finds which scope contains the variable.
func (e *Environment) lookup(name string) (*Environment, bool) {
	if _, ok := e.vars[name]; ok {
		return e, true
	}
	if e.parent != nil {
		return e.parent.lookup(name)
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Interpreter
// ---------------------------------------------------------------------------

// Interpreter executes an OpsLang AST.
type Interpreter struct {
	builtins  map[string]BuiltinFunc
	outputMu  sync.Mutex // guards output: parallel blocks append concurrently
	output    []OutputEntry
	globalEnv *Environment
	// DryRun prints the ensure apply steps instead of executing them.
	dryRun bool
	// scriptPriv is the privilege level of the program being executed,
	// derived from its `privilege:` statement (read_only when undeclared).
	// It is set at the start of Execute and read-only afterwards, so
	// parallel-block goroutines can read it without synchronization.
	scriptPriv ast.PrivilegeLevel
}

// SetDryRun enables dry-run mode: ensure bodies are reported as planned
// actions and not executed. Conditions are still evaluated (reads happen);
// only mutations inside ensure blocks are suppressed.
func (interp *Interpreter) SetDryRun(v bool) { interp.dryRun = v }

// appendOutput records one output entry. Safe for concurrent use: print
// builtins run inside parallel blocks on separate goroutines.
func (interp *Interpreter) appendOutput(entry OutputEntry) {
	interp.outputMu.Lock()
	interp.output = append(interp.output, entry)
	interp.outputMu.Unlock()
}

// New creates a new interpreter with optional extra built-in functions.
func New(builtins map[string]BuiltinFunc) *Interpreter {
	interp := &Interpreter{
		builtins:  make(map[string]BuiltinFunc),
		globalEnv: newEnv(nil),
	}

	// Register default builtins
	interp.registerDefaults()

	// Merge user-provided builtins
	for k, v := range builtins {
		interp.builtins[k] = v
	}

	return interp
}

func (interp *Interpreter) registerDefaults() {
	interp.builtins["print"] = func(args ...interface{}) (interface{}, error) {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = formatValue(a)
		}
		interp.appendOutput(OutputEntry{
			Type: "print",
			Data: strings.Join(parts, " "),
		})
		return nil, nil
	}

	interp.builtins["len"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("len() takes exactly 1 argument, got %d", len(args))
		}
		switch v := args[0].(type) {
		case string:
			return int64(len(v)), nil
		case []interface{}:
			return int64(len(v)), nil
		case map[string]interface{}:
			return int64(len(v)), nil
		default:
			return nil, fmt.Errorf("len() unsupported type: %T", args[0])
		}
	}

	interp.builtins["str"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("str() takes exactly 1 argument")
		}
		return formatValue(args[0]), nil
	}

	interp.builtins["int"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("int() takes exactly 1 argument")
		}
		return toInt(args[0])
	}

	interp.builtins["float"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("float() takes exactly 1 argument")
		}
		return toFloat(args[0])
	}

	interp.builtins["type"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("type() takes exactly 1 argument")
		}
		return typeName(args[0]), nil
	}

	interp.builtins["log"] = func(args ...interface{}) (interface{}, error) {
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = formatValue(a)
		}
		interp.appendOutput(OutputEntry{
			Type: "log",
			Data: strings.Join(parts, " "),
		})
		return nil, nil
	}

	interp.builtins["metric"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("metric() requires at least 2 arguments (name, value)")
		}
		entry := map[string]interface{}{
			"name":  formatValue(args[0]),
			"value": formatValue(args[1]),
		}
		if len(args) >= 3 {
			entry["labels"] = args[2]
		}
		interp.appendOutput(OutputEntry{
			Type: "metric",
			Data: entry,
		})
		return nil, nil
	}
}

// RegisterBuiltin adds a single built-in function by name.
func (interp *Interpreter) RegisterBuiltin(name string, fn BuiltinFunc) {
	interp.builtins[name] = fn
}

// Execute runs a program and returns the result.
func (interp *Interpreter) Execute(prog *ast.Program) (*Result, error) {
	interp.output = nil
	// Privilege enforcement: derive the script's declared level once per
	// run (undeclared scripts default to read_only) and check every
	// builtin call against it in evalCall. This is what makes
	// `privilege: read_only` actually deny mutating functions.
	interp.scriptPriv = security.GetScriptPrivilege(prog)
	var retVal interface{}

	for _, stmt := range prog.Statements {
		_, err := interp.execStatement(stmt, interp.globalEnv)
		if err != nil {
			if ret, ok := err.(*returnSignal); ok {
				retVal = ret.Value
				break // stop execution after return
			}
			return nil, err
		}
	}

	return &Result{
		Output:      interp.output,
		ReturnValue: retVal,
		Variables:   interp.globalEnv.vars,
	}, nil
}

// ---------------------------------------------------------------------------
// Statement execution
// ---------------------------------------------------------------------------

func (interp *Interpreter) execStatement(stmt ast.Statement, env *Environment) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return interp.execLet(s, env)
	case *ast.FnStatement:
		return interp.execFn(s, env)
	case *ast.IfStatement:
		return interp.execIf(s, env)
	case *ast.ForStatement:
		return interp.execFor(s, env)
	case *ast.WhileStatement:
		return interp.execWhile(s, env)
	case *ast.ReturnStatement:
		return interp.execReturn(s, env)
	case *ast.TaskStatement:
		return interp.execTask(s, env)
	case *ast.ImportStatement:
		// SDK builtins are globally registered; imports are declarative
		// only. Third-party Go imports are rejected at parse/run entry.
		if strings.HasPrefix(s.Path, "go ") || strings.HasPrefix(s.Path, "go:") {
			return nil, &RuntimeError{
				Pos: s.Pos(),
				Msg: fmt.Sprintf("import %q: third-party Go imports are not supported yet", s.Path),
			}
		}
		return nil, nil
	case *ast.PrivilegeStatement:
		// Declares the script's required privilege level. Execute() derives
		// the level from it once and evalCall enforces it on every builtin
		// call; the statement itself has no runtime effect.
		return nil, nil
	case *ast.ExpressionStatement:
		return interp.evalExpression(s.Expr, env)
	case *ast.AssignStatement:
		return interp.execAssign(s, env)
	case *ast.ReportStatement:
		return interp.execReport(s, env)
	case *ast.AlertStatement:
		return interp.execAlert(s, env)
	case *ast.EnsureStatement:
		return interp.execEnsure(s, env)
	case *ast.MetricStatement:
		return interp.execMetric(s, env)
	case *ast.LogStatement:
		return interp.execLog(s, env)
	case *ast.ParallelStatement:
		return interp.execParallel(s, env)
	case *ast.BlockStatement:
		blockEnv := newEnv(env)
		return interp.execBlock(s, blockEnv)
	default:
		return nil, &RuntimeError{Pos: stmt.Pos(), Msg: fmt.Sprintf("unknown statement type: %T", stmt)}
	}
}

func (interp *Interpreter) execLet(s *ast.LetStatement, env *Environment) (interface{}, error) {
	val, err := interp.evalExpression(s.Value, env)
	if err != nil {
		return nil, err
	}
	if _, exists := env.vars[s.Name.Name]; exists {
		return nil, &RuntimeError{
			Pos: s.Pos(),
			Msg: fmt.Sprintf("variable %q is already declared in this scope (use assignment without let to update it)", s.Name.Name),
		}
	}
	env.set(s.Name.Name, val)
	return nil, nil
}

func (interp *Interpreter) execFn(s *ast.FnStatement, env *Environment) (interface{}, error) {
	fn := &FunctionValue{
		Params:  s.Params,
		Body:    s.Body,
		Closure: env,
		Name:    s.Name.Name,
	}
	env.set(s.Name.Name, fn)
	return nil, nil
}

func (interp *Interpreter) execIf(s *ast.IfStatement, env *Environment) (interface{}, error) {
	cond, err := interp.evalExpression(s.Condition, env)
	if err != nil {
		return nil, err
	}

	if isTruthy(cond) {
		blockEnv := newEnv(env)
		return interp.execBlock(s.Body, blockEnv)
	} else if s.ElseClause != nil {
		switch e := s.ElseClause.(type) {
		case *ast.BlockStatement:
			blockEnv := newEnv(env)
			return interp.execBlock(e, blockEnv)
		case *ast.IfStatement:
			return interp.execIf(e, env)
		}
	}

	return nil, nil
}

func (interp *Interpreter) execFor(s *ast.ForStatement, env *Environment) (interface{}, error) {
	loopEnv := newEnv(env)

	// Execute init
	if s.Init != nil {
		if _, err := interp.execStatement(s.Init, loopEnv); err != nil {
			return nil, err
		}
	}

	for {
		// Check condition
		cond, err := interp.evalExpression(s.Condition, loopEnv)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			break
		}

		// Execute body
		bodyEnv := newEnv(loopEnv)
		_, err = interp.execBlock(s.Body, bodyEnv)
		if err != nil {
			if _, ok := err.(*returnSignal); ok {
				return nil, err
			}
			return nil, err
		}

		// Execute post
		if s.Post != nil {
			if _, err := interp.execStatement(s.Post, loopEnv); err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

func (interp *Interpreter) execWhile(s *ast.WhileStatement, env *Environment) (interface{}, error) {
	for {
		cond, err := interp.evalExpression(s.Condition, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			break
		}

		bodyEnv := newEnv(env)
		_, err = interp.execBlock(s.Body, bodyEnv)
		if err != nil {
			if _, ok := err.(*returnSignal); ok {
				return nil, err
			}
			return nil, err
		}
	}
	return nil, nil
}

func (interp *Interpreter) execReturn(s *ast.ReturnStatement, env *Environment) (interface{}, error) {
	var val interface{}
	if s.Value != nil {
		var err error
		val, err = interp.evalExpression(s.Value, env)
		if err != nil {
			return nil, err
		}
	}
	return nil, &returnSignal{Value: val}
}

func (interp *Interpreter) execTask(s *ast.TaskStatement, env *Environment) (interface{}, error) {
	// A task with an on-clause targets remote hosts. Executing its body on
	// the local machine would silently do the wrong thing; `opsctl run` has
	// no SSH context, so route it to `opsctl deploy` instead.
	if s.Targets != nil {
		return nil, &RuntimeError{
			Pos: s.Pos(),
			Msg: fmt.Sprintf("task %q targets remote hosts; use `opsctl deploy` to run it (opsctl run executes locally)", s.Name),
		}
	}
	blockEnv := newEnv(env)
	return interp.execBlock(s.Body, blockEnv)
}

func (interp *Interpreter) execAssign(s *ast.AssignStatement, env *Environment) (interface{}, error) {
	val, err := interp.evalExpression(s.Value, env)
	if err != nil {
		return nil, err
	}

	switch target := s.Target.(type) {
	case *ast.Identifier:
		if err := env.update(target.Name, val); err != nil {
			return nil, &RuntimeError{Pos: s.Pos(), Msg: err.Error()}
		}
	case *ast.IndexExpression:
		return nil, interp.assignIndex(target, val, env)
	case *ast.MemberExpression:
		return nil, interp.assignMember(target, val, env)
	default:
		return nil, &RuntimeError{Pos: s.Pos(), Msg: "invalid assignment target"}
	}

	return nil, nil
}

func (interp *Interpreter) assignIndex(idx *ast.IndexExpression, val interface{}, env *Environment) error {
	left, err := interp.evalExpression(idx.Left, env)
	if err != nil {
		return err
	}
	index, err := interp.evalExpression(idx.Index, env)
	if err != nil {
		return err
	}

	switch container := left.(type) {
	case []interface{}:
		i, ok := toInt64(index)
		if !ok {
			return &RuntimeError{Pos: idx.Pos(), Msg: "list index must be an integer"}
		}
		if i < 0 || i >= int64(len(container)) {
			return &RuntimeError{Pos: idx.Pos(), Msg: "list index out of range"}
		}
		container[i] = val
	case map[string]interface{}:
		key, ok := index.(string)
		if !ok {
			return &RuntimeError{Pos: idx.Pos(), Msg: "dict key must be a string"}
		}
		container[key] = val
	default:
		return &RuntimeError{Pos: idx.Pos(), Msg: "cannot index this type"}
	}
	return nil
}

func (interp *Interpreter) assignMember(mem *ast.MemberExpression, val interface{}, env *Environment) error {
	obj, err := interp.evalExpression(mem.Object, env)
	if err != nil {
		return err
	}
	dict, ok := obj.(map[string]interface{})
	if !ok {
		return &RuntimeError{Pos: mem.Pos(), Msg: "member assignment requires a dict"}
	}
	dict[mem.Member.Name] = val
	return nil
}

func (interp *Interpreter) execReport(s *ast.ReportStatement, env *Environment) (interface{}, error) {
	data := make(map[string]interface{})
	for _, field := range s.Fields {
		val, err := interp.evalExpression(field.Value, env)
		if err != nil {
			return nil, err
		}
		data[field.Key] = val
	}
	interp.appendOutput(OutputEntry{Type: "report", Data: data})
	return nil, nil
}

func (interp *Interpreter) execAlert(s *ast.AlertStatement, env *Environment) (interface{}, error) {
	msg, err := interp.evalExpression(s.Message, env)
	if err != nil {
		return nil, err
	}
	interp.appendOutput(OutputEntry{Type: "alert", Data: formatValue(msg)})
	return nil, nil
}

func (interp *Interpreter) execEnsure(s *ast.EnsureStatement, env *Environment) (interface{}, error) {
	// Step 1: CHECK - evaluate the condition
	cond, err := interp.evalExpression(s.Condition, env)
	if err != nil {
		return nil, err
	}

	// If condition is true, nothing to do (state is already desired)
	if isTruthy(cond) {
		return nil, nil
	}

	// Dry-run: report the planned apply steps without executing them.
	if interp.dryRun {
		interp.appendOutput(OutputEntry{
			Type: "dry-run",
			Data: fmt.Sprintf("ensure: would apply %d action(s) (condition is false)", len(s.Body.Statements)),
		})
		return nil, nil
	}

	// Step 2: APPLY - execute the body to reach desired state
	blockEnv := newEnv(env)
	_, err = interp.execBlock(s.Body, blockEnv)
	if err != nil {
		return nil, err
	}

	// Step 3: VERIFY - re-check the condition
	cond2, err := interp.evalExpression(s.Condition, env)
	if err != nil {
		return nil, err
	}
	if !isTruthy(cond2) {
		return nil, &RuntimeError{
			Pos: s.Pos(),
			Msg: "ensure: condition still false after applying actions",
		}
	}

	// Step 4: NOTIFY - optional expression evaluated after a change was
	// applied (typically alert(...)).
	if s.Notify != nil {
		if _, err := interp.evalExpression(s.Notify, env); err != nil {
			return nil, &RuntimeError{
				Pos: s.Pos(),
				Msg: fmt.Sprintf("ensure notify: %v", err),
			}
		}
	}

	return nil, nil
}

func (interp *Interpreter) execMetric(s *ast.MetricStatement, env *Environment) (interface{}, error) {
	name, err := interp.evalExpression(s.Name, env)
	if err != nil {
		return nil, err
	}
	value, err := interp.evalExpression(s.Value, env)
	if err != nil {
		return nil, err
	}

	entry := map[string]interface{}{
		"name":  formatValue(name),
		"value": formatValue(value),
	}

	if s.Labels != nil {
		labels, err := interp.evalExpression(s.Labels, env)
		if err != nil {
			return nil, err
		}
		entry["labels"] = labels
	}

	interp.appendOutput(OutputEntry{
		Type: "metric",
		Data: entry,
	})
	return nil, nil
}

func (interp *Interpreter) execLog(s *ast.LogStatement, env *Environment) (interface{}, error) {
	msg, err := interp.evalExpression(s.Message, env)
	if err != nil {
		return nil, err
	}
	interp.appendOutput(OutputEntry{
		Type: "log",
		Data: formatValue(msg),
	})
	return nil, nil
}

func (interp *Interpreter) execBlock(block *ast.BlockStatement, env *Environment) (interface{}, error) {
	var result interface{}
	for _, stmt := range block.Statements {
		val, err := interp.execStatement(stmt, env)
		if err != nil {
			return nil, err
		}
		result = val
	}
	return result, nil
}

// execParallel executes a parallel block: each statement runs in its own
// goroutine with an isolated child environment (concurrent writes to shared
// variables would be a data race). After all statements finish, their
// declared variables are merged back into the enclosing scope in source
// order, so results ARE visible after the block - deterministically: when
// two statements declare the same name, the later statement wins.
func (interp *Interpreter) execParallel(s *ast.ParallelStatement, env *Environment) (interface{}, error) {
	if s.Body == nil || len(s.Body.Statements) == 0 {
		return nil, nil
	}

	type outcome struct {
		err error
		env *Environment
	}
	outcomes := make([]outcome, len(s.Body.Statements))
	var wg sync.WaitGroup

	for i, stmt := range s.Body.Statements {
		wg.Add(1)
		go func(idx int, st ast.Statement) {
			defer wg.Done()
			// Isolated barrier env per statement: assignments are captured
			// locally, never written to the shared parent map concurrently.
			childEnv := newEnv(env)
			childEnv.barrier = true
			_, err := interp.execStatement(st, childEnv)
			outcomes[idx] = outcome{err: err, env: childEnv}
		}(i, stmt)
	}
	wg.Wait()

	// Deterministic serial merge in source order.
	for _, o := range outcomes {
		if o.err != nil {
			if _, ok := o.err.(*returnSignal); ok {
				continue // return inside parallel is meaningless; ignore
			}
			return nil, o.err
		}
	}
	for _, o := range outcomes {
		if o.env == nil {
			continue
		}
		for name, val := range o.env.vars {
			env.set(name, val)
		}
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Expression evaluation
// ---------------------------------------------------------------------------

func (interp *Interpreter) evalExpression(expr ast.Expression, env *Environment) (interface{}, error) {
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
	case *ast.ListLiteral:
		return interp.evalList(e, env)
	case *ast.DictLiteral:
		return interp.evalDict(e, env)
	case *ast.Identifier:
		return interp.evalIdent(e, env)
	case *ast.BinaryExpression:
		return interp.evalBinary(e, env)
	case *ast.UnaryExpression:
		return interp.evalUnary(e, env)
	case *ast.CallExpression:
		return interp.evalCall(e, env)
	case *ast.IndexExpression:
		return interp.evalIndex(e, env)
	case *ast.MemberExpression:
		return interp.evalMember(e, env)
	case *ast.IfExpression:
		return interp.evalIfExpression(e, env)
	default:
		return nil, &RuntimeError{Pos: expr.Pos(), Msg: fmt.Sprintf("unknown expression type: %T", expr)}
	}
}

func (interp *Interpreter) evalIfExpression(e *ast.IfExpression, env *Environment) (interface{}, error) {
	// If Condition is nil, evaluate Then directly (short form)
	if e.Condition == nil {
		return interp.evalExpression(e.Then, env)
	}

	cond, err := interp.evalExpression(e.Condition, env)
	if err != nil {
		return nil, err
	}

	if isTruthy(cond) {
		return interp.evalExpression(e.Then, env)
	} else if e.Else != nil {
		return interp.evalExpression(e.Else, env)
	}
	return nil, nil
}

func (interp *Interpreter) evalList(e *ast.ListLiteral, env *Environment) (interface{}, error) {
	list := make([]interface{}, len(e.Elements))
	for i, elem := range e.Elements {
		val, err := interp.evalExpression(elem, env)
		if err != nil {
			return nil, err
		}
		list[i] = val
	}
	return list, nil
}

func (interp *Interpreter) evalDict(e *ast.DictLiteral, env *Environment) (interface{}, error) {
	dict := make(map[string]interface{})
	for i, key := range e.Keys {
		k, err := interp.evalExpression(key, env)
		if err != nil {
			return nil, err
		}
		keyStr, ok := k.(string)
		if !ok {
			keyStr = formatValue(k)
		}
		val, err := interp.evalExpression(e.Values[i], env)
		if err != nil {
			return nil, err
		}
		dict[keyStr] = val
	}
	return dict, nil
}

func (interp *Interpreter) evalIdent(e *ast.Identifier, env *Environment) (interface{}, error) {
	val, ok := env.get(e.Name)
	if !ok {
		return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("undefined variable: %s", e.Name)}
	}
	return val, nil
}

func (interp *Interpreter) evalBinary(e *ast.BinaryExpression, env *Environment) (interface{}, error) {
	left, err := interp.evalExpression(e.Left, env)
	if err != nil {
		return nil, err
	}

	// Short-circuit for logical operators
	switch e.Op {
	case "&&":
		if !isTruthy(left) {
			return false, nil
		}
		right, err := interp.evalExpression(e.Right, env)
		if err != nil {
			return nil, err
		}
		return isTruthy(right), nil
	case "||":
		if isTruthy(left) {
			return true, nil
		}
		right, err := interp.evalExpression(e.Right, env)
		if err != nil {
			return nil, err
		}
		return isTruthy(right), nil
	}

	right, err := interp.evalExpression(e.Right, env)
	if err != nil {
		return nil, err
	}

	return applyBinary(e.Op, left, right, e.Pos())
}

func applyBinary(op string, left, right interface{}, pos ast.Position) (interface{}, error) {
	switch op {
	case "+":
		// List concatenation: list + list or list + item
		if ll, ok := left.([]interface{}); ok {
			if rl, ok := right.([]interface{}); ok {
				result := make([]interface{}, len(ll)+len(rl))
				copy(result, ll)
				copy(result[len(ll):], rl)
				return result, nil
			}
			// list + single item → append
			result := make([]interface{}, len(ll)+1)
			copy(result, ll)
			result[len(ll)] = right
			return result, nil
		}
		if _, ok := right.([]interface{}); ok {
			// single item + list → prepend
			result := make([]interface{}, 0, 1+len(right.([]interface{})))
			result = append(result, left)
			result = append(result, right.([]interface{})...)
			return result, nil
		}
		// String concatenation
		if ls, ok := left.(string); ok {
			return ls + formatValue(right), nil
		}
		if _, ok := right.(string); ok {
			return formatValue(left) + formatValue(right), nil
		}
		return applyArithmetic(op, left, right, pos)
	case "-", "*", "/", "%":
		return applyArithmetic(op, left, right, pos)
	case "==":
		return isEqual(left, right), nil
	case "!=":
		return !isEqual(left, right), nil
	case "<":
		return compareOrdered(left, right, pos, func(a, b float64) bool { return a < b })
	case ">":
		return compareOrdered(left, right, pos, func(a, b float64) bool { return a > b })
	case "<=":
		return compareOrdered(left, right, pos, func(a, b float64) bool { return a <= b })
	case ">=":
		return compareOrdered(left, right, pos, func(a, b float64) bool { return a >= b })
	default:
		return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("unknown operator: %s", op)}
	}
}

func applyArithmetic(op string, left, right interface{}, pos ast.Position) (interface{}, error) {
	// If either is float, promote to float
	if isFloat(left) || isFloat(right) {
		l, err := toFloat(left)
		if err != nil {
			return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("cannot use %s with %T", op, left)}
		}
		r, err := toFloat(right)
		if err != nil {
			return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("cannot use %s with %T", op, right)}
		}
		switch op {
		case "+":
			return l + r, nil
		case "-":
			return l - r, nil
		case "*":
			return l * r, nil
		case "/":
			if r == 0 {
				return nil, &RuntimeError{Pos: pos, Msg: "division by zero"}
			}
			return l / r, nil
		case "%":
			if r == 0 {
				return nil, &RuntimeError{Pos: pos, Msg: "division by zero"}
			}
			return float64(int64(l) % int64(r)), nil
		}
	}

	// Integer arithmetic
	l, ok := toInt64(left)
	if !ok {
		return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("cannot use %s with %T", op, left)}
	}
	r, ok := toInt64(right)
	if !ok {
		return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("cannot use %s with %T", op, right)}
	}

	switch op {
	case "+":
		return l + r, nil
	case "-":
		return l - r, nil
	case "*":
		return l * r, nil
	case "/":
		if r == 0 {
			return nil, &RuntimeError{Pos: pos, Msg: "division by zero"}
		}
		return l / r, nil
	case "%":
		if r == 0 {
			return nil, &RuntimeError{Pos: pos, Msg: "division by zero"}
		}
		return l % r, nil
	}
	return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("unknown operator: %s", op)}
}

func (interp *Interpreter) evalUnary(e *ast.UnaryExpression, env *Environment) (interface{}, error) {
	val, err := interp.evalExpression(e.Right, env)
	if err != nil {
		return nil, err
	}

	switch e.Op {
	case "-":
		switch v := val.(type) {
		case int64:
			return -v, nil
		case float64:
			return -v, nil
		default:
			return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("cannot negate %T", val)}
		}
	case "!":
		return !isTruthy(val), nil
	default:
		return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("unknown unary operator: %s", e.Op)}
	}
}

func (interp *Interpreter) evalCall(e *ast.CallExpression, env *Environment) (interface{}, error) {
	// Evaluate arguments
	args := make([]interface{}, len(e.Args))
	for i, arg := range e.Args {
		val, err := interp.evalExpression(arg, env)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	// Try to resolve function name
	fnName := interp.resolveFuncName(e.Function)
	if fnName != "" {
		if builtin, ok := interp.builtins[fnName]; ok {
			// Privilege enforcement: a read_only script may not call
			// mutating builtins. Unknown names (custom builtins) are not
			// restricted — the opsspec table defines what mutates.
			if err := security.CheckFuncPrivilege(interp.scriptPriv, fnName); err != nil {
				return nil, &RuntimeError{Pos: e.Pos(), Msg: err.Error()}
			}
			result, err := builtin(args...)
			if err != nil {
				return nil, &RuntimeError{Pos: e.Pos(), Msg: err.Error()}
			}
			return result, nil
		}
		// If we have a function name but it's not a builtin,
		// check if it exists as a user-defined function before trying to evaluate
		fnVal, lookupErr := interp.evalExpression(e.Function, env)
		if lookupErr != nil {
			// Function name not found in environment
			return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("unknown function: %s", fnName)}
		}
		if fn, ok := fnVal.(*FunctionValue); ok {
			return interp.callFunction(fn, args, e.Pos())
		}
		return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("not a function: %s", fnName)}
	}

	// Anonymous function call (e.g., (fn() { return 1 })())
	fnVal, err := interp.evalExpression(e.Function, env)
	if err != nil {
		return nil, err
	}

	fn, ok := fnVal.(*FunctionValue)
	if !ok {
		return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("not a function: %T", fnVal)}
	}

	return interp.callFunction(fn, args, e.Pos())
}

// resolveFuncName builds a dotted name from a call's function expression.
func (interp *Interpreter) resolveFuncName(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		prefix := interp.resolveFuncName(e.Object)
		if prefix != "" {
			return prefix + "." + e.Member.Name
		}
	}
	return ""
}

func (interp *Interpreter) callFunction(fn *FunctionValue, args []interface{}, pos ast.Position) (interface{}, error) {
	fnEnv := newEnv(fn.Closure)

	// Bind parameters
	for i, param := range fn.Params {
		if i < len(args) {
			fnEnv.set(param.Name.Name, args[i])
		} else if param.Default != nil {
			val, err := interp.evalExpression(param.Default, fnEnv)
			if err != nil {
				return nil, err
			}
			fnEnv.set(param.Name.Name, val)
		} else {
			return nil, &RuntimeError{
				Pos: pos,
				Msg: fmt.Sprintf("missing argument %q (parameter %d)", param.Name.Name, i+1),
			}
		}
	}

	_, err := interp.execBlock(fn.Body, fnEnv)
	if err != nil {
		if ret, ok := err.(*returnSignal); ok {
			return ret.Value, nil
		}
		return nil, err
	}

	return nil, nil
}

func (interp *Interpreter) evalIndex(e *ast.IndexExpression, env *Environment) (interface{}, error) {
	left, err := interp.evalExpression(e.Left, env)
	if err != nil {
		return nil, err
	}
	index, err := interp.evalExpression(e.Index, env)
	if err != nil {
		return nil, err
	}

	switch container := left.(type) {
	case []interface{}:
		i, ok := toInt64(index)
		if !ok {
			return nil, &RuntimeError{Pos: e.Pos(), Msg: "list index must be an integer"}
		}
		if i < 0 || i >= int64(len(container)) {
			return nil, &RuntimeError{Pos: e.Pos(), Msg: "list index out of range"}
		}
		return container[i], nil
	case map[string]interface{}:
		key, ok := index.(string)
		if !ok {
			return nil, &RuntimeError{Pos: e.Pos(), Msg: "dict key must be a string"}
		}
		val, ok := container[key]
		if !ok {
			return nil, &RuntimeError{
				Pos: e.Pos(),
				Msg: fmt.Sprintf("map has no key %q (available: %v)", key, mapKeys(container)),
			}
		}
		return val, nil
	default:
		return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("cannot index %T", left)}
	}
}

func (interp *Interpreter) evalMember(e *ast.MemberExpression, env *Environment) (interface{}, error) {
	obj, err := interp.evalExpression(e.Object, env)
	if err != nil {
		return nil, err
	}

	switch o := obj.(type) {
	case map[string]interface{}:
		val, ok := o[e.Member.Name]
		if !ok {
			// A missing key is almost always a typo in the field name;
			// failing here points at the real mistake instead of much
			// later where nil eventually breaks an expression.
			return nil, &RuntimeError{
				Pos: e.Pos(),
				Msg: fmt.Sprintf("map has no key %q (available: %v)", e.Member.Name, mapKeys(o)),
			}
		}
		return val, nil
	default:
		return nil, &RuntimeError{Pos: e.Pos(), Msg: fmt.Sprintf("cannot access member on %T", obj)}
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case map[string]interface{}:
		return len(v) > 0
	default:
		return true
	}
}

// mapKeys returns the sorted key list of a dict for error messages.
func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// isEqual: numbers compare numerically, strings as strings, bools as
// bools. Values of different kinds are NOT equal (1 != "1") - cross-type
// string-coincidence matching hid real bugs. Mirrors the AOT compiler's
// opsEqual exactly; the two engines must never disagree.
func isEqual(left, right interface{}) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if lb, ok := left.(bool); ok {
		rb, rok := right.(bool)
		return rok && lb == rb
	}
	lNum, lIsNum := toNumber(left)
	rNum, rIsNum := toNumber(right)
	if lIsNum && rIsNum {
		return lNum == rNum
	}
	if ls, lok := left.(string); lok {
		if rs, rok := right.(string); rok {
			return ls == rs
		}
	}
	return false
}

func compareOrdered(left, right interface{}, pos ast.Position, cmp func(a, b float64) bool) (interface{}, error) {
	// String comparison
	if ls, ok := left.(string); ok {
		if rs, ok := right.(string); ok {
			return cmp(float64(strings.Compare(ls, rs)), 0) || (ls == rs && cmp(0, 0)), nil
		}
	}

	l, err := toFloat(left)
	if err != nil {
		return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("cannot compare %T", left)}
	}
	r, err := toFloat(right)
	if err != nil {
		return nil, &RuntimeError{Pos: pos, Msg: fmt.Sprintf("cannot compare %T", right)}
	}
	return cmp(l, r), nil
}

func formatValue(val interface{}) string {
	if val == nil {
		return "nil"
	}
	switch v := val.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case string:
		return v
	case []interface{}:
		parts := make([]string, len(v))
		for i, elem := range v {
			parts[i] = formatValue(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		parts := make([]string, 0, len(v))
		for k, val := range v {
			parts = append(parts, fmt.Sprintf("%s: %s", k, formatValue(val)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *FunctionValue:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt(val interface{}) (interface{}, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		// Strict parse: "42abc" and "3.14" are errors, not 42/3. The old
		// Sscanf prefix matching silently corrupted parsed command output.
		i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %q to int", v)
		}
		return i, nil
	case bool:
		if v {
			return int64(1), nil
		}
		return int64(0), nil
	case nil:
		return int64(0), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to int", val)
	}
}

func toFloat(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to float", v)
		}
		return f, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case nil:
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", val)
	}
}

func toInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

func toNumber(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case int64:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func isFloat(val interface{}) bool {
	_, ok := val.(float64)
	return ok
}

func typeName(val interface{}) string {
	if val == nil {
		return "nil"
	}
	switch val.(type) {
	case int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case bool:
		return "bool"
	case []interface{}:
		return "list"
	case map[string]interface{}:
		return "dict"
	case *FunctionValue:
		return "function"
	default:
		return fmt.Sprintf("%T", val)
	}
}
