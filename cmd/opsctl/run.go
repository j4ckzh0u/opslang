// run command for opsctl - local script interpretation
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/j4ckzh0u/opslang/internal/interpreter"
	"github.com/j4ckzh0u/opslang/internal/modules"
	"github.com/j4ckzh0u/opslang/internal/parser"
	"github.com/spf13/cobra"
)

var (
	runOutputJSON bool
	runVerbose    bool
	runDryRun     bool
)

var runCmd = &cobra.Command{
	Use:   "run [script.ops]",
	Short: "Execute an OpsLang script locally via interpreter",
	Long: `Parse and interpret an OpsLang script file locally.

The script goes through lexer → parser → interpreter pipeline.
Output is printed to stdout; use --json for machine-readable output.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRunCommand(args[0])
	},
}

func init() {
	runCmd.Flags().BoolVar(&runOutputJSON, "json", false, "Output results as JSON")
	runCmd.Flags().BoolVarP(&runVerbose, "verbose", "v", false, "Print execution details")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Print ensure apply actions without executing them")
}

func runRunCommand(scriptPath string) error {
	// Read source file.
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script %s: %w", scriptPath, err)
	}

	// Parse.
	p := parser.New(string(source), scriptPath)
	prog, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Link file-module imports into one flat program.
	prog, err = modules.Link(prog, scriptPath)
	if err != nil {
		return fmt.Errorf("module error: %w", err)
	}

	if runVerbose {
		fmt.Fprintf(os.Stderr, "Parsed %d statements from %s\n", len(prog.Statements), scriptPath)
	}

	// Interpret.
	interp := interpreter.New(nil)
	interpreter.RegisterSDKBuiltins(interp)
	interp.SetDryRun(runDryRun)
	result, err := interp.Execute(prog)
	if err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}

	// Output.
	if runOutputJSON {
		return printJSONResult(result)
	}

	return printTextResult(result)
}

func printJSONResult(result *interpreter.Result) error {
	output := map[string]interface{}{
		"output": result.Output,
	}
	if result.ReturnValue != nil {
		output["return"] = result.ReturnValue
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printTextResult(result *interpreter.Result) error {
	for _, entry := range result.Output {
		switch entry.Type {
		case "print", "log":
			fmt.Println(formatOutputData(entry.Data))
		case "report":
			data, err := json.MarshalIndent(entry.Data, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format report: %w", err)
			}
			fmt.Println(string(data))
		case "alert":
			fmt.Fprintf(os.Stderr, "ALERT: %s\n", formatOutputData(entry.Data))
		case "metric":
			data, err := json.Marshal(entry.Data)
			if err != nil {
				return fmt.Errorf("failed to format metric: %w", err)
			}
			fmt.Printf("METRIC: %s\n", string(data))
		}
	}
	return nil
}

func formatOutputData(data interface{}) string {
	if s, ok := data.(string); ok {
		return s
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(b)
}
