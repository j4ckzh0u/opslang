package compiler

import (
	"testing"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/parser"
)

func parseOrDie(t *testing.T, source string) *ast.Program {
	t.Helper()
	p := parser.New(source, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return prog
}

func TestSelectMode(t *testing.T) {
	tests := []struct {
		name   string
		source string
		mode   ExecutionMode
		want   ExecutionMode
	}{
		{
			name:   "explicit runner stays runner",
			source: `if true { print("x") }`,
			mode:   ModeRunner,
			want:   ModeRunner,
		},
		{
			name:   "explicit aot stays aot",
			source: `print("linear")`,
			mode:   ModeAOT,
			want:   ModeAOT,
		},
		{
			name:   "linear script selects runner",
			source: "let cpu = sys.cpu.usage()\nreport { cpu: cpu }",
			mode:   ModeAuto,
			want:   ModeRunner,
		},
		{
			name:   "if statement requires aot",
			source: "if true { print(\"x\") }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "for loop requires aot",
			source: "for let i = 0; i < 3; i = i + 1 { print(i) }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "while loop requires aot",
			source: "while false { print(\"x\") }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "fn definition requires aot",
			source: "fn helper() { print(\"hi\") }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "ensure requires aot",
			source: "ensure file.exists(\"/tmp/x\").exists { file.mkdir(\"/tmp/x\") }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "parallel requires aot",
			source: "parallel { sys.cpu.usage() }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "control flow inside task requires aot",
			source: "task \"t\" on \"h\" { if true { print(\"x\") } }",
			mode:   ModeAuto,
			want:   ModeAOT,
		},
		{
			name:   "linear task body selects runner",
			source: "task \"t\" on \"h\" { sys.cpu.usage() }",
			mode:   ModeAuto,
			want:   ModeRunner,
		},
		{
			// Long linear scripts are fine in runner mode; the old
			// line-count heuristic sent them to AOT for no reason.
			name:   "long linear script stays runner",
			source: "print(\"a\")\nprint(\"b\")\n" + repeatLines("print(\"filler\")", 200),
			mode:   ModeAuto,
			want:   ModeRunner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := parseOrDie(t, tt.source)
			got := SelectMode(prog, tt.source, tt.mode)
			if got != tt.want {
				t.Errorf("SelectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func repeatLines(line string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += line + "\n"
	}
	return out
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		input string
		want  ExecutionMode
	}{
		{"", ModeAuto},
		{"auto", ModeAuto},
		{"runner", ModeRunner},
		{"aot", ModeAOT},
		{"AOT", ModeAOT},
	}
	for _, tt := range tests {
		got, err := ParseMode(tt.input)
		if err != nil {
			t.Errorf("ParseMode(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseMode(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseModeErrorMessage(t *testing.T) {
	_, err := ParseMode("turbo")
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}
