// ops-runner is the universal runner that executes JSON instruction packages
// from stdin and outputs JSON results to stdout.
//
// Exit codes:
//
//	0 - all instructions succeeded (status "ok")
//	1 - some instructions failed (status "partial")
//	2 - every instruction failed (status "failed")
//	3 - protocol/usage error (bad input, unsupported version)
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
		os.Exit(3)
	}

	// The run itself may have partially or fully failed; report it via the
	// exit code so the controller (and CI) can detect it without parsing JSON.
	switch status {
	case "partial":
		os.Exit(1)
	case "failed":
		os.Exit(2)
	}
}

// status records the outcome of the last run for exit-code mapping.
var status string

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
	status = output.Status

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
