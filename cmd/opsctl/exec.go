// exec command for opsctl - remote execution via SSH
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/inventory"
	"github.com/j4ckzh0u/opslang/internal/opsspec"
	"github.com/j4ckzh0u/opslang/internal/runner"
	"github.com/j4ckzh0u/opslang/internal/security"
	"github.com/spf13/cobra"
)

var (
	execHosts           []string
	execUser            string
	execKey             string
	execPassword        string
	execInventory       string
	execInstructions    string
	execParallel        int
	execDryRun          bool
	execRunnerPath      string
	execOutputFile      string
	execInsecureHostKey bool
	execAutoApprove     bool
	execLimitCPU        int
	execLimitMemMB      int64
	execSignKey         string
	execVerifyKey       string
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
		// Resolve --auto-approve (flag over env) here: reading the flag set
		// from helpers would create an init cycle via execCmd.
		auto, src := security.ResolveAutoApprove(cmd.Flags().Changed("auto-approve"), execAutoApprove)
		return runExecCommand(auto, src)
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
	execCmd.Flags().BoolVar(&execInsecureHostKey, "insecure-host-key", false, "Skip SSH host key verification (lab use only)")
	execCmd.Flags().StringVarP(&execOutputFile, "output", "o", "", "Output file path (default: stdout)")
	execCmd.Flags().BoolVar(&execAutoApprove, "auto-approve", false, "Pre-approve gated executions (privileged packages on production targets); non-interactive runs are refused without it")
	execCmd.Flags().IntVar(&execLimitCPU, "limit-cpu", 0, "Cap remote runner CPU usage (percent, requires systemd-run on targets; 0 = off)")
	execCmd.Flags().Int64Var(&execLimitMemMB, "limit-mem", 0, "Cap remote runner memory (MB, requires systemd-run on targets; 0 = off)")
	execCmd.Flags().StringVar(&execSignKey, "sign-key", "", "Ed25519 private key (from opsctl keygen) used to sign the instruction package")
	execCmd.Flags().StringVar(&execVerifyKey, "verify-key", "", "REMOTE path of the trusted public key; runners refuse unsigned/tampered packages")
}

// runExecCommand is the main execution logic for the exec subcommand.
func runExecCommand(autoApprove bool, autoSource security.ApprovalSource) error {
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

	// Sign the package before any host is contacted; a bad key path fails
	// the run here rather than mid-flight.
	if execSignKey != "" {
		key, err := security.LoadPrivateKey(execSignKey)
		if err != nil {
			return fmt.Errorf("failed to load signing key: %w", err)
		}
		if err := runner.SignPackage(pkg, key); err != nil {
			return fmt.Errorf("failed to sign instruction package: %w", err)
		}
	}

	// Build target list from --hosts and/or --inventory.
	targets, err := buildExecTargets()
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	// Approval gate: privileged packages aimed at production targets need
	// explicit approval before any host is contacted.
	taskID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	approvalRec, err := enforceExecApproval(execInstructions, pkg, targets, autoApprove, autoSource)
	if err != nil {
		entry := security.NewAuditEntry(
			taskID,
			execInstructions,
			pkg.Privilege,
			targetAddressList(targets),
			execUser,
			"instruction-exec",
			execDryRun,
		)
		entry.SetError(err)
		entry.Approval = approvalRec
		if lerr := security.NewAuditLogger("").Log(entry); lerr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write audit log: %v\n", lerr)
		}
		return err
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
		Targets:                   targets,
		Instructions:              pkg,
		User:                      execUser,
		KeyFile:                   execKey,
		Password:                  execPassword,
		Parallel:                  execParallel,
		DryRun:                    execDryRun,
		RunnerPath:                execRunnerPath,
		InsecureSkipHostKeyVerify: execInsecureHostKey,
		TaskID:                    taskID,
		ArchCache:                 archCacheForRun(),
		ResourceLimit:             resourceLimitFromFlags(execLimitCPU, execLimitMemMB),
		RunnerVerifyKeyPath:       execVerifyKey,
	}

	summary := executor.Execute(ctx)

	// Audit every remote execution: operator, targets, outcome.
	auditEntry := security.NewAuditEntry(
		summary.TaskID,
		execInstructions,
		"system",
		append([]string(nil), execHosts...),
		execUser,
		"instruction-exec",
		execDryRun,
	)
	auditEntry.Approval = approvalRec
	if summary.Status == "success" {
		auditEntry.SetStatus("success", summary.FinishedAt.Sub(summary.StartedAt).Milliseconds())
	} else {
		auditEntry.SetError(fmt.Errorf("execution status: %s", summary.Status))
	}
	if err := security.NewAuditLogger("").Log(auditEntry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write audit log: %v\n", err)
	}

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
	switch summary.Status {
	case "failed":
		os.Exit(2)
	case "partial":
		// Some hosts succeeded, some did not: the operator must notice.
		os.Exit(1)
	}

	return nil
}

// buildExecTargets assembles the target list from --hosts and --inventory.
func buildExecTargets() ([]opsexec.Target, error) {
	var targets []opsexec.Target
	if len(execHosts) > 0 {
		targets = append(targets, opsexec.ParseTargets(execHosts, execUser)...)
	}
	if execInventory != "" {
		inv, err := inventory.ParseFile(execInventory)
		if err != nil {
			return targets, fmt.Errorf("failed to parse inventory: %w", err)
		}
		targets = append(targets, opsexec.TargetsFromInventory(inv)...)
	}
	return targets, nil
}

// enforceExecApproval applies the approval gate to a raw instruction-package
// execution. The package's declared privilege drives the decision (empty —
// legacy packages — never gates, matching the runner's treatment); the
// mutating operations come straight from the instructions.
func enforceExecApproval(instrPath string, pkg *runner.InstructionPackage, targets []opsexec.Target, autoApprove bool, autoSource security.ApprovalSource) (*security.ApprovalRecord, error) {
	var mutatingOps []string
	for _, in := range pkg.Instructions {
		if mutating, known := opsspec.Mutating(in.Op); known && mutating {
			mutatingOps = append(mutatingOps, in.Op)
		}
	}
	gate := &security.ApprovalGate{
		Decision: security.EvaluateApproval(
			pkg.PrivilegeLevel(),
			mutatingOps,
			targetTags(targets),
		),
		ScriptName:  instrPath,
		AutoApprove: autoApprove,
		AutoSource:  autoSource,
		Interactive: deployStdinIsInteractive(),
		Confirm:     deployConfirmFn,
	}
	rec, err := gate.Check()
	if err != nil {
		return rec, fmt.Errorf("execution blocked: %w", err)
	}
	if rec != nil {
		fmt.Fprintf(os.Stderr, "Approved via %s by %s\n", rec.Source, rec.User)
	}
	return rec, nil
}
