// Package runner implements the OpsLang universal runner that executes
// JSON instruction packages and outputs structured JSON results.
package runner

// InstructionPackage represents the JSON input — a set of instructions
// to execute sequentially with variable assignment support.
type InstructionPackage struct {
	Version      string        `json:"version"`
	TaskID       string        `json:"task_id"`
	DryRun       bool          `json:"dry_run"`
	Instructions []Instruction `json:"instructions"`
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
