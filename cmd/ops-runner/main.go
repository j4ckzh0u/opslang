// ops-runner is the universal runner that executes JSON instruction packages
// from stdin and outputs JSON results to stdout.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/opslang/opslang/internal/runner"
)

var (
	dryRun  bool
	version bool
)

func main() {
	flag.BoolVar(&dryRun, "dry-run", false, "Execute in dry-run mode (no actual changes)")
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.Parse()

	if version {
		fmt.Println("ops-runner v0.1.0")
		os.Exit(0)
	}

	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(in io.Reader, out io.Writer) error {
	// Read instruction package from stdin.
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Parse instruction package.
	var pkg runner.InstructionPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to parse instruction package: %w", err)
	}

	// Validate version.
	if pkg.Version != "1.0" {
		return fmt.Errorf("unsupported protocol version: %s (expected 1.0)", pkg.Version)
	}

	// Apply dry-run flag.
	if dryRun {
		pkg.DryRun = true
	}

	// Create registry with all standard operations.
	registry := runner.NewRegistry()

	// Execute instructions.
	output := runner.Run(&pkg, registry)

	// Marshal output.
	result, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	// Write output.
	if _, err := out.Write(result); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	if _, err := out.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}
