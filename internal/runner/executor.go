package runner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/security"
)

// varPrefix marks an argument value as a variable reference. Only strings
// starting with "$" are resolved against executor variables; literal
// strings are never rewritten (the previous scheme replaced ANY string
// equal to a variable name, silently corrupting literal values).
const varPrefix = "$"

// Executor processes instructions using a registry, maintaining a variable
// context for assignment and reference resolution.
type Executor struct {
	registry *Registry
	vars     map[string]interface{}
	dryRun   bool
	warnings []string
	// privilege is the package-declared script privilege used as the
	// runner-side second check: mutating operations that contradict the
	// declaration fail that instruction with a structured error. The zero
	// value means the package did not declare a privilege (legacy format)
	// and no runner-side check applies — controller-side enforcement
	// (interpreter, compiler, instruction generator) covers those.
	privilege ast.PrivilegeLevel
}

// NewExecutor creates a new executor with the given registry and dry-run setting.
func NewExecutor(registry *Registry, dryRun bool) *Executor {
	return &Executor{
		registry: registry,
		vars:     make(map[string]interface{}),
		dryRun:   dryRun,
		warnings: []string{},
	}
}

// SetPrivilege sets the script privilege level enforced on mutating ops.
func (e *Executor) SetPrivilege(level ast.PrivilegeLevel) {
	e.privilege = level
}

// Execute runs a single instruction with variable resolution.
func (e *Executor) Execute(inst *Instruction) (interface{}, error) {
	// Look up the operation first so unknown ops fail even in dry-run.
	fn, ok := e.registry.Get(inst.Op)
	if !ok {
		return nil, fmt.Errorf("unknown operation: %q", inst.Op)
	}

	// Privilege second check: when the package declares a privilege level,
	// a mutating instruction that contradicts it fails here (structured
	// error, never a panic) — defense in depth behind the controller-side
	// enforcement that already rejects these at generation time.
	if e.privilege != "" {
		if err := security.CheckFuncPrivilege(e.privilege, inst.Op); err != nil {
			return nil, err
		}
	}

	// Resolve variable references in args.
	resolvedArgs := e.resolveArgs(inst.Args)

	// Handle dry-run mode.
	if e.dryRun {
		return map[string]interface{}{
			"dry_run":   true,
			"operation": inst.Op,
			"args":      resolvedArgs,
		}, nil
	}

	// Handle built-in special operations.
	switch inst.Op {
	case "report":
		return e.executeReport(resolvedArgs)
	case "log":
		return e.executeLog(resolvedArgs)
	}

	// Execute the operation.
	return fn(resolvedArgs)
}

// resolveArgs replaces "$name" string values with the variable's value.
// "$$" escapes a literal leading dollar sign.
func (e *Executor) resolveArgs(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return make(map[string]interface{})
	}
	resolved := make(map[string]interface{}, len(args))
	for key, value := range args {
		resolved[key] = e.resolveValue(value)
	}
	return resolved
}

// resolveValue resolves one argument value, recursing into lists and maps.
func (e *Executor) resolveValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, varPrefix+varPrefix) {
			return strings.TrimPrefix(v, varPrefix)
		}
		if strings.HasPrefix(v, varPrefix) {
			return e.resolveRef(v[1:])
		}
		return v
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, elem := range v {
			out[i] = e.resolveValue(elem)
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, elem := range v {
			out[k] = e.resolveValue(elem)
		}
		return out
	default:
		return value
	}
}

// resolveRef resolves "$name" and "$name.field.path" references against
// executor variables, indexing into nested maps for each dotted segment.
// SDK results are typed structs; they are converted to generic maps via a
// JSON round-trip on first member access (mirroring the interpreter's
// structToMap), so "$cpu.percent" works against any result type.
func (e *Executor) resolveRef(ref string) interface{} {
	parts := strings.Split(ref, ".")
	current, ok := e.vars[parts[0]]
	if !ok {
		return nil // unresolved reference: yields nil, downstream ops validate
	}
	for _, field := range parts[1:] {
		m, ok := asStringMap(current)
		if !ok {
			return nil
		}
		current, ok = m[field]
		if !ok {
			return nil
		}
	}
	return current
}

// asStringMap converts maps (and SDK structs via JSON) to map[string]interface{}.
func asStringMap(v interface{}) (map[string]interface{}, bool) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	return m, true
}

// executeReport collects named variables into the output data map.
func (e *Executor) executeReport(args map[string]interface{}) (interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range args {
		result[key] = e.resolveValue(value)
	}
	return result, nil
}

// executeLog outputs a message to warnings and returns the message.
func (e *Executor) executeLog(args map[string]interface{}) (interface{}, error) {
	msg := getStringArg(args, "message", "")
	if msg == "" {
		msg = getStringArg(args, "msg", "")
	}
	if msg != "" {
		e.warnings = append(e.warnings, msg)
	}
	return msg, nil
}

// SetVar sets a variable in the executor's context.
func (e *Executor) SetVar(name string, value interface{}) {
	e.vars[name] = value
}

// GetVar returns a variable from the executor's context.
func (e *Executor) GetVar(name string) (interface{}, bool) {
	v, ok := e.vars[name]
	return v, ok
}

// Warnings returns all accumulated warnings.
func (e *Executor) Warnings() []string {
	return e.warnings
}

// Run processes an instruction package and returns the output.
//
// An instruction error does not abort the remaining instructions (later
// steps may still collect useful state), but it is recorded and affects
// the final status:
//
//	ok      - every instruction succeeded
//	partial - some succeeded, some failed
//	failed  - every instruction failed (or none produced a result)
func Run(pkg *InstructionPackage, registry *Registry) *Output {
	output := &Output{
		Status:   "ok",
		Data:     make(map[string]interface{}),
		Errors:   []string{},
		Warnings: []string{},
	}

	executor := NewExecutor(registry, pkg.DryRun)
	if pkg.Privilege != "" {
		switch lvl := pkg.PrivilegeLevel(); lvl {
		case ast.PrivilegeReadOnly, ast.PrivilegeAdmin, ast.PrivilegeRoot:
			executor.SetPrivilege(lvl)
		default:
			// A malformed privilege value must not silently pass the
			// second check; fail the whole run with a structured error.
			output.Errors = append(output.Errors,
				fmt.Sprintf("invalid package privilege %q (expected read_only, admin, or root)", pkg.Privilege))
			output.Status = "failed"
			return output
		}
	}
	var lastReport map[string]interface{}
	succeeded, failed := 0, 0

	for i, inst := range pkg.Instructions {
		result, err := executor.Execute(&inst)
		if err != nil {
			failed++
			output.Errors = append(output.Errors,
				fmt.Sprintf("instruction %d (%s): %v", i, inst.Op, err))
			continue
		}
		succeeded++

		// Assign result to variable if specified.
		if inst.Assign != "" {
			executor.SetVar(inst.Assign, result)
			output.Data[inst.Assign] = result
		}

		// Track report results for final output override.
		if inst.Op == "report" {
			if report, ok := result.(map[string]interface{}); ok {
				lastReport = report
			}
		}
	}

	// Collect warnings from executor (from log operations).
	output.Warnings = append(output.Warnings, executor.Warnings()...)

	// Determine final status honestly.
	switch {
	case failed == 0:
		output.Status = "ok"
	case succeeded > 0:
		output.Status = "partial"
	default:
		output.Status = "failed"
	}

	// If a report was generated, use it as the output data.
	if lastReport != nil {
		output.Data = lastReport
	}

	return output
}

// OutputToJSON marshals an Output to JSON bytes.
func OutputToJSON(output *Output) ([]byte, error) {
	return json.MarshalIndent(output, "", "  ")
}
