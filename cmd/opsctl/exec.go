// exec command for opsctl - remote execution via SSH
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	opsexec "github.com/opslang/opslang/internal/exec"
	"github.com/opslang/opslang/internal/inventory"
	"github.com/spf13/cobra"
)

var (
	execHosts        []string
	execUser         string
	execKey          string
	execPassword     string
	execInventory    string
	execInstructions string
	execParallel     int
	execDryRun       bool
	execRunnerPath   string
	execOutputFile   string
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute instructions on remote hosts via SSH",
	Long: `Upload ops-runner to target hosts and execute JSON instruction packages.

Targets can be specified via --hosts (comma-separated user@host entries) or
--inventory (YAML inventory file). Both can be combined.

The runner binary is automatically built for the target architecture on first
use and cached in ~/.cache/opslang/runners/ for subsequent executions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExecCommand()
	},
}

func init() {
	execCmd.Flags().StringSliceVar(&execHosts, "hosts", nil, "Target hosts (user@host or host)")
	execCmd.Flags().StringVarP(&execUser, "user", "u", "root", "Default SSH user")
	execCmd.Flags().StringVarP(&execKey, "key", "i", "", "SSH private key path")
	execCmd.Flags().StringVarP(&execPassword, "password", "p", "", "SSH password")
	execCmd.Flags().StringVar(&execInventory, "inventory", "", "Inventory file path (YAML)")
	execCmd.Flags().StringVar(&execInstructions, "instructions", "", "JSON instructions file path")
	execCmd.Flags().IntVar(&execParallel, "parallel", 10, "Maximum concurrent hosts")
	execCmd.Flags().BoolVar(&execDryRun, "dry-run", false, "Execute in dry-run mode")
	execCmd.Flags().StringVar(&execRunnerPath, "runner-path", "", "Path to pre-built runner binary")
	execCmd.Flags().StringVarP(&execOutputFile, "output", "o", "", "Output file path (default: stdout)")
}

// runExecCommand is the main execution logic for the exec subcommand.
func runExecCommand() error {
	// Validate required flags.
	if execInstructions == "" {
		return fmt.Errorf("--instructions flag is required")
	}
	if len(execHosts) == 0 && execInventory == "" {
		return fmt.Errorf("either --hosts or --inventory must be specified")
	}

	// Load instruction package.
	pkg, err := opsexec.LoadInstructions(execInstructions)
	if err != nil {
		return fmt.Errorf("failed to load instructions: %w", err)
	}

	// Build target list from --hosts and/or --inventory.
	var targets []opsexec.Target

	if len(execHosts) > 0 {
		targets = append(targets, opsexec.ParseTargets(execHosts, execUser)...)
	}

	if execInventory != "" {
		inv, err := inventory.ParseFile(execInventory)
		if err != nil {
			return fmt.Errorf("failed to parse inventory: %w", err)
		}
		targets = append(targets, opsexec.TargetsFromInventory(inv)...)
	}

	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	// Set up context with signal handling for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Create and run executor.
	executor := &opsexec.Executor{
		Targets:      targets,
		Instructions: pkg,
		User:         execUser,
		KeyFile:      execKey,
		Password:     execPassword,
		Parallel:     execParallel,
		DryRun:       execDryRun,
		RunnerPath:   execRunnerPath,
	}

	summary := executor.Execute(ctx)

	// Marshal result to JSON.
	result, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to output file or stdout.
	if execOutputFile != "" {
		if err := os.WriteFile(execOutputFile, append(result, '\n'), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Println(string(result))
	}

	// Exit with non-zero code if any host failed.
	if summary.Status == "failed" {
		os.Exit(1)
	}

	return nil
}
