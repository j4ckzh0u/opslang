//go:build !opssec

package main

// Default build: the side-effect half of internal/security is compiled out
// (stability-first decision). Approval gates are skipped and audit logging
// is disabled. Rebuild with -tags opssec to restore the full feature set.

import (
	"fmt"
	"os"

	"github.com/j4ckzh0u/opslang/internal/ast"
	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/runner"
)

func resolveAutoApprove(flagSet, flagValue bool) (bool, approvalSource) {
	// No gate exists in this build, so approval resolution is moot; keep
	// flag-over-env ordering so behavior stays predictable if opssec is
	// re-enabled later.
	if flagSet {
		return flagValue, approveAutoFlag
	}
	if flagValue {
		return true, approveAutoEnv
	}
	return false, approveNonInteractive
}

func promptConfirm(summary string) bool {
	fmt.Fprintln(os.Stderr, "approval prompt unavailable without opssec build; denying")
	return false
}

func stdinIsInteractive() bool { return false }

func loadSignKey(string) ([]byte, error) {
	return nil, fmt.Errorf("--sign-key requires an opssec build (-tags opssec)")
}

// Approval always passes in the default build.
func enforceDeployApprovalGate(string, *ast.Program, []opsexec.Target, bool, approvalSource) (*approvalRecord, error) {
	return nil, nil
}

func enforceExecApprovalGate(string, *runner.InstructionPackage, []opsexec.Target, bool, approvalSource) (*approvalRecord, error) {
	return nil, nil
}

// auditFacts mirrors the opssec shape so call sites are tag-agnostic.
type auditFacts struct {
	taskID     string
	script     string
	privilege  string
	targets    []string
	user       string
	mode       string
	dryRun     bool
	runErr     error
	status     string
	durationMs int64
	approval   *approvalRecord
}

// writeAudit is a no-op: audit logging needs the opssec build.
func writeAudit(auditFacts) {}
