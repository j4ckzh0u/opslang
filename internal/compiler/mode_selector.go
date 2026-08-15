package compiler

import (
	"fmt"
	"strings"

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

// SelectMode determines the execution mode based on script characteristics.
// Rules:
//   - If the mode is explicitly set (not "auto" or ""), use that mode.
//   - If source has "import go" or "import go:" statements -> AOT.
//   - If source has fewer than 100 lines -> Runner.
//   - Otherwise -> AOT.
func SelectMode(prog *ast.Program, source string, mode ExecutionMode) ExecutionMode {
	// Explicit mode override.
	if mode != ModeAuto && mode != "" {
		return mode
	}

	// Check for Go imports that require AOT compilation.
	if prog != nil {
		for _, stmt := range prog.Statements {
			if imp, ok := stmt.(*ast.ImportStatement); ok {
				if strings.HasPrefix(imp.Path, "go ") || strings.HasPrefix(imp.Path, "go:") {
					return ModeAOT
				}
			}
		}
	}

	// Count source lines.
	lineCount := strings.Count(source, "\n") + 1
	if lineCount < 100 {
		return ModeRunner
	}

	return ModeAOT
}

// ParseMode parses a mode string into an ExecutionMode.
func ParseMode(s string) (ExecutionMode, error) {
	switch strings.ToLower(s) {
	case "auto", "":
		return ModeAuto, nil
	case "runner":
		return ModeRunner, nil
	case "aot":
		return ModeAOT, nil
	default:
		return "", fmt.Errorf("unknown execution mode: %q (valid: auto, runner, aot)", s)
	}
}
