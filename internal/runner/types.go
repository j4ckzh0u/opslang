// Package runner implements the OpsLang universal runner that executes
// JSON instruction packages and outputs structured JSON results.
package runner

import "github.com/opslang/opslang/internal/ast"

// InstructionPackage represents the JSON input — a set of instructions
// to execute sequentially with variable assignment support.
type InstructionPackage struct {
	Version      string        `json:"version"`
	TaskID       string        `json:"task_id"`
	DryRun       bool          `json:"dry_run"`
	Instructions []Instruction `json:"instructions"`
	// Privilege carries the script's declared privilege level
	// (read_only | admin | root) so the runner can refuse mutating
	// instructions that contradict it. Empty means the field predates the
	// declaration (legacy packages): the runner then skips this
	// second-line check and relies on controller-side enforcement.
	Privilege string `json:"privilege,omitempty"`
}

// PrivilegeLevel returns the package's declared privilege as a typed
// level. Empty (unset) is reported as the zero level, which callers must
// treat as "not declared" rather than read_only.
func (p *InstructionPackage) PrivilegeLevel() ast.PrivilegeLevel {
	return ast.PrivilegeLevel(p.Privilege)
}

// Instruction represents a single operation to execute.
// Args values can be literal values or variable names — if a string arg
// matches a previously assigned variable name, the variable's value is used.
type Instruction struct {
	Op     string                 `json:"op"`
	Args   map[string]interface{} `json:"args"`
	Assign string                 `json:"assign,omitempty"`
}

// Output represents the JSON output to stdout.
type Output struct {
	Status   string                 `json:"status"`
	Data     map[string]interface{} `json:"data"`
	Errors   []string               `json:"errors"`
	Warnings []string               `json:"warnings"`
}

// OperationFunc is a function that executes an operation with the given arguments.
// It receives the resolved arguments (variable references already substituted)
// and returns the result value and any error.
type OperationFunc func(args map[string]interface{}) (interface{}, error)
