package compiler

import (
	"strings"
	"testing"

	"github.com/opslang/opslang/internal/ast"
)

// ---------------------------------------------------------------------------
// SelectMode
// ---------------------------------------------------------------------------

func TestSelectMode(t *testing.T) {
	tests := []struct {
		name   string
		prog   *ast.Program
		source string
		mode   ExecutionMode
		want   ExecutionMode
	}{
		// Explicit overrides
		{
			name: "explicit runner override",
			prog: &ast.Program{},
			mode: ModeRunner,
			want: ModeRunner,
		},
		{
			name: "explicit aot override",
			prog: &ast.Program{},
			mode: ModeAOT,
			want: ModeAOT,
		},
		{
			name: "explicit aot overrides go import detection",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ImportStatement{Path: "go something"},
				},
			},
			mode: ModeRunner,
			want: ModeRunner,
		},
		{
			name:   "explicit override ignores line count",
			source: strings.Repeat("x\n", 200),
			mode:   ModeRunner,
			want:   ModeRunner,
		},

		// Auto mode: Go import detection
		{
			name: "auto with go-space import triggers aot",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ImportStatement{Path: "go something"},
				},
			},
			mode: ModeAuto,
			want: ModeAOT,
		},
		{
			name: "auto with go-colon import triggers aot",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ImportStatement{Path: "go:something"},
				},
			},
			mode: ModeAuto,
			want: ModeAOT,
		},
		{
			name: "auto with non-go import does not trigger aot",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ImportStatement{Path: "python os"},
				},
			},
			source: "let x = 1",
			mode:   ModeAuto,
			want:   ModeRunner,
		},
		{
			name: "auto with mixed imports triggers aot on go import",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.ImportStatement{Path: "python os"},
					&ast.ImportStatement{Path: "go fmt"},
				},
			},
			mode: ModeAuto,
			want: ModeAOT,
		},
		{
			name: "auto with non-import statements falls to line count",
			prog: &ast.Program{
				Statements: []ast.Statement{
					&ast.LetStatement{Name: &ast.Identifier{Name: "x"}, Value: &ast.IntegerLiteral{Value: 1}},
				},
			},
			source: "let x = 1",
			mode:   ModeAuto,
			want:   ModeRunner,
		},

		// Auto mode: line count threshold
		{
			name:   "auto with empty mode and short source returns runner",
			source: "let x = 1",
			mode:   "",
			want:   ModeRunner,
		},
		{
			name:   "auto with exactly 99 lines returns runner",
			source: strings.Repeat("x\n", 98) + "x", // 99 lines
			mode:   ModeAuto,
			want:   ModeRunner,
		},
		{
			name:   "auto with exactly 100 lines returns aot",
			source: strings.Repeat("x\n", 99) + "x", // 100 lines
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "auto with 101 lines returns aot",
			source: strings.Repeat("x\n", 100) + "x", // 101 lines
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "auto with empty source is 1 line and returns runner",
			source: "",
			mode:   ModeAuto,
			want:   ModeRunner,
		},

		// Nil prog
		{
			name:   "nil prog with short source returns runner",
			prog:   nil,
			source: "short",
			mode:   ModeAuto,
			want:   ModeRunner,
		},
		{
			name:   "nil prog with long source returns aot",
			prog:   nil,
			source: strings.Repeat("x\n", 100),
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name: "nil prog skips go import check",
			prog: nil,
			// Even though source is short, if prog were non-nil with go imports
			// it would be AOT. With nil prog, we go straight to line count.
			source: "short",
			mode:   ModeAuto,
			want:   ModeRunner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectMode(tt.prog, tt.source, tt.mode)
			if got != tt.want {
				t.Errorf("SelectMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseMode
// ---------------------------------------------------------------------------

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ExecutionMode
		wantErr bool
	}{
		{"auto lowercase", "auto", ModeAuto, false},
		{"auto uppercase", "AUTO", ModeAuto, false},
		{"auto mixed case", "Auto", ModeAuto, false},
		{"empty string defaults to auto", "", ModeAuto, false},
		{"runner lowercase", "runner", ModeRunner, false},
		{"runner uppercase", "RUNNER", ModeRunner, false},
		{"runner mixed case", "Runner", ModeRunner, false},
		{"aot lowercase", "aot", ModeAOT, false},
		{"aot uppercase", "AOT", ModeAOT, false},
		{"aot mixed case", "Aot", ModeAOT, false},
		{"unknown mode returns error", "invalid", "", true},
		{"unknown mode with spaces", " auto ", "", true},
		{"numeric is unknown", "123", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseMode(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMode(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseModeErrorMessage(t *testing.T) {
	_, err := ParseMode("badmode")
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "unknown execution mode") {
		t.Errorf("error message should contain 'unknown execution mode', got: %s", msg)
	}
	if !strings.Contains(msg, "badmode") {
		t.Errorf("error message should contain the bad input, got: %s", msg)
	}
	if !strings.Contains(msg, "valid: auto, runner, aot") {
		t.Errorf("error message should list valid modes, got: %s", msg)
	}
}
