package compiler

import (
	"fmt"

	"github.com/opslang/opslang/internal/ast"
)

// ExecutionMode represents how a script should be executed.
type ExecutionMode string

const (
	// ModeAuto selects the execution mode automatically based on script characteristics.
	ModeAuto ExecutionMode = "auto"
	// ModeRunner uses the universal runner with JSON instruction packages.
	ModeRunner ExecutionMode = "runner"
	// ModeAOT compiles the script to a static binary ahead of time.
	ModeAOT ExecutionMode = "aot"
)

// RequiresAOT reports whether the program contains statements the linear
// runner instruction VM cannot express exactly (control flow, functions,
// ensure, parallel). Such scripts must be compiled to preserve semantics.
func RequiresAOT(prog *ast.Program) bool {
	if prog == nil {
		return false
	}
	var requires bool
	walk := func(stmts []ast.Statement) {}
	var visit func(stmt ast.Statement) bool // returns true -> stop
	visit = func(stmt ast.Statement) bool {
		switch s := stmt.(type) {
		case *ast.IfStatement, *ast.ForStatement, *ast.WhileStatement,
			*ast.FnStatement, *ast.EnsureStatement, *ast.ParallelStatement,
			*ast.ReturnStatement:
			requires = true
			return true
		case *ast.TaskStatement:
			for _, inner := range s.Body.Statements {
				if visit(inner) {
					return true
				}
			}
		}
		return false
	}
	walk = func(stmts []ast.Statement) {
		for _, stmt := range stmts {
			if visit(stmt) {
				return
			}
		}
	}
	walk(prog.Statements)
	return requires
}

// SelectMode determines the execution mode based on script characteristics.
// The runner mode only supports linear scripts, so anything requiring AOT
// semantics is routed there.
func SelectMode(prog *ast.Program, source string, mode ExecutionMode) ExecutionMode {
	// Explicit mode override.
	if mode != ModeAuto && mode != "" {
		return mode
	}

	if RequiresAOT(prog) {
		return ModeAOT
	}
	return ModeRunner
}

// ParseMode parses a mode string into an ExecutionMode.
func ParseMode(s string) (ExecutionMode, error) {
	switch s {
	case "auto", "Auto", "AUTO", "":
		return ModeAuto, nil
	case "runner", "Runner", "RUNNER":
		return ModeRunner, nil
	case "aot", "Aot", "AOT":
		return ModeAOT, nil
	default:
		return "", fmt.Errorf("unknown execution mode: %q (valid: auto, runner, aot)", s)
	}
}
