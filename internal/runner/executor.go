package runner

import (
	"encoding/json"
	"fmt"
)

// Executor processes instructions using a registry, maintaining a variable
// context for assignment and reference resolution.
type Executor struct {
	registry *Registry
	vars     map[string]interface{}
	dryRun   bool
	warnings []string
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

// Execute runs a single instruction with variable resolution.
// String argument values that match a previously assigned variable name
// are replaced with that variable's value.
func (e *Executor) Execute(inst *Instruction) (interface{}, error) {
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

	// Look up the operation.
	fn, ok := e.registry.Get(inst.Op)
	if !ok {
		return nil, fmt.Errorf("unknown operation: %q", inst.Op)
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

// resolveArgs replaces string values that match variable names with their values.
func (e *Executor) resolveArgs(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return make(map[string]interface{})
	}
	resolved := make(map[string]interface{}, len(args))
	for key, value := range args {
		if s, ok := value.(string); ok {
			if varVal, exists := e.vars[s]; exists {
				resolved[key] = varVal
				continue
			}
		}
		resolved[key] = value
	}
	return resolved
}

// executeReport collects named variables into the output data map.
// Each arg value is treated as a variable name; the variable's value is
// included in the report under the arg's key.
func (e *Executor) executeReport(args map[string]interface{}) (interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range args {
		// If the value is a string that matches a variable name, use the variable's value.
		if varName, ok := value.(string); ok {
			if varVal, exists := e.vars[varName]; exists {
				result[key] = varVal
				continue
			}
		}
		// Otherwise, use the value as-is.
		result[key] = value
	}
	return result, nil
}

// executeLog outputs a message to warnings and returns the message.
func (e *Executor) executeLog(args map[string]interface{}) (interface{}, error) {
	msg := ""
	if v, ok := args["message"]; ok {
		if s, ok := v.(string); ok {
			msg = s
		}
	}
	if msg == "" {
		if v, ok := args["msg"]; ok {
			if s, ok := v.(string); ok {
				msg = s
			}
		}
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
// Errors in individual instructions do not stop execution — they are
// collected in the output's Errors array.
func Run(pkg *InstructionPackage, registry *Registry) *Output {
	output := &Output{
		Status:   "ok",
		Data:     make(map[string]interface{}),
		Errors:   []string{},
		Warnings: []string{},
	}

	executor := NewExecutor(registry, pkg.DryRun)
	var lastReport map[string]interface{}

	for i, inst := range pkg.Instructions {
		result, err := executor.Execute(&inst)
		if err != nil {
			output.Errors = append(output.Errors,
				fmt.Sprintf("instruction %d (%s): %v", i, inst.Op, err))
			continue
		}

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

	// Determine final status.
	if len(output.Errors) > 0 {
		output.Status = "partial"
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
