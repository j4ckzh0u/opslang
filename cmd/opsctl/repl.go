// repl command for opsctl - interactive OpsLang environment
package main

import (
	"bufio"
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
	// Handle SIGINT gracefully (don't exit, just cancel current line).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for range sigCh {
			// Print newline after ^C to keep prompt clean.
			fmt.Println()
			printPrompt()
		}
	}()

	interp := interpreter.New(nil)
	interpreter.RegisterSDKBuiltins(interp)

	fmt.Println("OpsLang REPL v0.1.0")
	fmt.Println("Type OpsLang expressions. Ctrl+D to exit, Ctrl+C to cancel line.")
	fmt.Println()
	fmt.Println("SDK builtins loaded: sys.*, file.*, net.*, process.*, service.*, pkg.*, time.*, json.*, yaml.*")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	var multiLine strings.Builder

	for {
		printPrompt()

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nBye!")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		trimmed := strings.TrimSpace(line)

		// Empty line: if we have accumulated multi-line input, execute it.
		if trimmed == "" {
			if multiLine.Len() > 0 {
				executeREPLInput(interp, multiLine.String())
				multiLine.Reset()
			}
			continue
		}

		// Special commands.
		if trimmed == "exit" || trimmed == "quit" {
			fmt.Println("Bye!")
			return nil
		}

		if trimmed == "help" {
			printREPLHelp()
			continue
		}

		// Check if this line opens a block (ends with {) for multi-line.
		multiLine.WriteString(line)
		if isOpenBlock(trimmed) {
			// Continue reading more lines.
			continue
		}

		// Execute accumulated input.
		executeREPLInput(interp, multiLine.String())
		multiLine.Reset()
	}
}

func executeREPLInput(interp *interpreter.Interpreter, source string) {
	p := parser.New(source, "<repl>")
	prog, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  parse error: %v\n", err)
		return
	}

	result, err := interp.Execute(prog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  error: %v\n", err)
		return
	}

	// Print outputs.
	for _, entry := range result.Output {
		switch entry.Type {
		case "print", "log":
			if s, ok := entry.Data.(string); ok {
				fmt.Println(" ", s)
			} else {
				fmt.Printf(" %v\n", entry.Data)
			}
		case "report", "metric":
			fmt.Printf(" %v\n", entry.Data)
		case "alert":
			fmt.Fprintf(os.Stderr, " ALERT: %v\n", entry.Data)
		}
	}

	// Print return value if any.
	if result.ReturnValue != nil {
		fmt.Printf(" => %v\n", result.ReturnValue)
	}
}

// isOpenBlock returns true if the line ends with { indicating more input needed.
func isOpenBlock(line string) bool {
	return strings.HasSuffix(line, "{")
}

func printPrompt() {
	fmt.Print("ops> ")
}

func printREPLHelp() {
	fmt.Println("OpsLang REPL Commands:")
	fmt.Println("  exit, quit  - Exit the REPL")
	fmt.Println("  help        - Show this help")
	fmt.Println()
	fmt.Println("Multi-line input: lines ending with { continue to next line.")
	fmt.Println("An empty line executes accumulated input.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  let x = 42")
	fmt.Println("  print(x)")
	fmt.Println("  fn add(a, b) { return a + b }")
	fmt.Println("  add(1, 2)")
}
