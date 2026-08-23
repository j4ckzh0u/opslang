// repl command for opsctl - interactive OpsLang environment
package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/opslang/opslang/internal/interpreter"
	"github.com/opslang/opslang/internal/parser"
	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive OpsLang REPL",
	Long: `Start a Read-Eval-Print Loop for interactive OpsLang scripting.

Supports multi-line input (end blocks with empty line).
Press Ctrl+C to cancel current input, Ctrl+D to exit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runREPL()
	},
}

func runREPL() error {
	// Handle SIGINT gracefully for the non-TTY path (raw mode handles
	// Ctrl+C itself inside the line editor): cancel the line, keep running.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for range sigCh {
			fmt.Println()
			printPrompt()
		}
	}()

	interp := interpreter.New(nil)
	interpreter.RegisterSDKBuiltins(interp)

	hist := &history{}
	reader := newLineReader("ops> ", hist)
	if err := reader.init(); err != nil {
		return err
	}
	defer reader.close()

	reader.emit("OpsLang REPL v0.1.0\n")
	reader.emit("Type OpsLang expressions. Ctrl+D to exit, Ctrl+C to cancel line.\n")
	reader.emit("Up/Down browse history, Left/Home/End edit the line.\n\n")
	reader.emit("SDK builtins loaded: sys.*, file.*, net.*, process.*, service.*, pkg.*, time.*, json.*, yaml.*\n\n")

	var multiLine strings.Builder

	for {
		line, err := reader.read()
		if err != nil {
			if err == errInterrupted {
				// Line discarded, fresh prompt on the next loop turn.
				if multiLine.Len() > 0 {
					multiLine.Reset()
				}
				continue
			}
			if err == io.EOF {
				reader.emit("\nBye!\n")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		trimmed := strings.TrimSpace(line)

		// Empty line: if we have accumulated multi-line input, execute it.
		if trimmed == "" {
			if multiLine.Len() > 0 {
				submitREPLInput(interp, hist, reader, multiLine.String())
				multiLine.Reset()
			}
			continue
		}

		// Special commands.
		if trimmed == "exit" || trimmed == "quit" {
			reader.emit("Bye!\n")
			return nil
		}

		if trimmed == "help" {
			printREPLHelp(reader)
			continue
		}

		// Check if this line opens a block (ends with {) for multi-line.
		multiLine.WriteString(line)
		multiLine.WriteString("\n")
		if isOpenBlock(trimmed) {
			// Continue reading more lines.
			continue
		}

		// Execute accumulated input; the whole submission (including
		// multi-line blocks) becomes one history entry, so Up recalls
		// the complete block.
		submitREPLInput(interp, hist, reader, multiLine.String())
		multiLine.Reset()
	}
}

// submitREPLInput records the submission in history and executes it.
func submitREPLInput(interp *interpreter.Interpreter, hist *history, reader *lineReader, source string) {
	hist.add(source)
	executeREPLInput(interp, reader, source)
}

func executeREPLInput(interp *interpreter.Interpreter, reader *lineReader, source string) {
	p := parser.New(source, "<repl>")
	prog, err := p.Parse()
	if err != nil {
		reader.emit("  parse error: %v\n", err)
		return
	}

	result, err := interp.Execute(prog)
	if err != nil {
		reader.emit("  error: %v\n", err)
		return
	}

	// Print outputs.
	for _, entry := range result.Output {
		switch entry.Type {
		case "print", "log":
			if s, ok := entry.Data.(string); ok {
				reader.emit("  %s\n", s)
			} else {
				reader.emit("  %v\n", entry.Data)
			}
		case "report", "metric":
			reader.emit("  %v\n", entry.Data)
		case "alert":
			reader.emit("  ALERT: %v\n", entry.Data)
		}
	}

	// Print return value if any.
	if result.ReturnValue != nil {
		reader.emit("  => %v\n", result.ReturnValue)
	}
}

// isOpenBlock returns true if the line ends with { indicating more input needed.
func isOpenBlock(line string) bool {
	return strings.HasSuffix(line, "{")
}

func printPrompt() {
	fmt.Print("ops> ")
}

var _ = printPrompt // kept: referenced by the non-TTY SIGINT path

func printREPLHelp(reader *lineReader) {
	reader.emit("OpsLang REPL Commands:\n")
	reader.emit("  exit, quit  - Exit the REPL\n")
	reader.emit("  help        - Show this help\n")
	reader.emit("\n")
	reader.emit("Line editing: Up/Down browse history (multi-line blocks recalled\n")
	reader.emit("as one entry), Left/Right/Home/End move, Backspace/Delete edit,\n")
	reader.emit("Ctrl+A/E line edges, Ctrl+U/K delete to edge, Ctrl+L clear screen.\n")
	reader.emit("\n")
	reader.emit("Multi-line input: lines ending with { continue to next line.\n")
	reader.emit("An empty line executes accumulated input.\n")
	reader.emit("\n")
	reader.emit("Examples:\n")
	reader.emit("  let x = 42\n")
	reader.emit("  print(x)\n")
	reader.emit("  fn add(a, b) { return a + b }\n")
	reader.emit("  add(1, 2)\n")
}
