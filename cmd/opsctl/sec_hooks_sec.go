package main

// The full security feature set (approval gate, audit log, Ed25519 signing)
// is enabled in the default build and backed by internal/security. The
// historical opssec build tag remains source-compatible.

import (
	"fmt"
	"os"

	"github.com/j4ckzh0u/opslang/internal/ast"
	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/opsspec"
	"github.com/j4ckzh0u/opslang/internal/runner"
	"github.com/j4ckzh0u/opslang/internal/security"
)

func resolveAutoApprove(flagSet, flagValue bool) (bool, approvalSource) {
	auto, src := security.ResolveAutoApprove(flagSet, flagValue)
	return auto, approvalSource(src)
}

func promptConfirm(summary string) bool {
	return security.PromptConfirm(summary)
}

func stdinIsInteractive() bool {
	return security.StdinIsInteractive()
}

func loadSignKey(path string) ([]byte, error) {
	return security.LoadPrivateKey(path)
}

// targetTags converts deploy targets into the tag view the approval
// decision consumes (name + env metadata; inline targets carry no tags).
func targetTags(targets []opsexec.Target) []security.TargetTags {
	out := make([]security.TargetTags, len(targets))
	for i, t := range targets {
		out[i] = security.TargetTags{Name: t.Name, Tags: t.Tags}
	}
	return out
}

func enforceDeployApprovalGate(scriptPath string, prog *ast.Program, targets []opsexec.Target, autoApprove bool, autoSrc approvalSource) (*approvalRecord, error) {
	gate := &security.ApprovalGate{
		Decision: security.EvaluateApproval(
			security.GetScriptPrivilege(prog),
			security.CollectMutatingOps(prog),
			targetTags(targets),
		),
		ScriptName:  scriptPath,
		AutoApprove: autoApprove,
		AutoSource:  security.ApprovalSource(autoSrc),
		Interactive: deployStdinIsInteractive(),
		Confirm:     deployConfirmFn,
	}
	rec, err := gate.Check()
	if err != nil {
		return recordFrom(rec), fmt.Errorf("deploy blocked: %w", err)
	}
	if rec != nil {
		fmt.Fprintf(os.Stderr, "Approved via %s by %s\n", rec.Source, rec.User)
	}
	return recordFrom(rec), nil
}

func enforceExecApprovalGate(instrPath string, pkg *runner.InstructionPackage, targets []opsexec.Target, autoApprove bool, autoSrc approvalSource) (*approvalRecord, error) {
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
		AutoSource:  security.ApprovalSource(autoSrc),
		Interactive: deployStdinIsInteractive(),
		Confirm:     deployConfirmFn,
	}
	rec, err := gate.Check()
	if err != nil {
		return recordFrom(rec), fmt.Errorf("execution blocked: %w", err)
	}
	if rec != nil {
		fmt.Fprintf(os.Stderr, "Approved via %s by %s\n", rec.Source, rec.User)
	}
	return recordFrom(rec), nil
}

func recordFrom(rec *security.ApprovalRecord) *approvalRecord {
	if rec == nil {
		return nil
	}
	return &approvalRecord{
		Required:     rec.Required,
		Decision:     rec.Decision,
		Source:       rec.Source,
		User:         rec.User,
		Privilege:    rec.Privilege,
		MutatingOps:  rec.MutatingOps,
		ProdTargets:  rec.ProdTargets,
		TotalTargets: rec.TotalTargets,
		DecidedAt:    rec.DecidedAt,
	}
}

// auditFacts carries everything an audit entry needs.
type auditFacts struct {
	taskID     string
	script     string
	privilege  string
	targets    []string
	user       string
	mode       string
	dryRun     bool
	runErr     error
	status     string // empty or "success" with nil runErr records a success
	durationMs int64
	approval   *approvalRecord
}

func writeAudit(f auditFacts) {
	entry := security.NewAuditEntry(
		f.taskID,
		f.script,
		f.privilege,
		f.targets,
		f.user,
		f.mode,
		f.dryRun,
	)
	switch {
	case f.runErr != nil:
		entry.SetError(f.runErr)
	case f.status == "" || f.status == "success":
		entry.SetStatus("success", f.durationMs)
	default:
		entry.SetError(fmt.Errorf("execution status: %s", f.status))
	}
	if f.approval != nil {
		entry.Approval = &security.ApprovalRecord{
			Required:     f.approval.Required,
			Decision:     f.approval.Decision,
			Source:       f.approval.Source,
			User:         f.approval.User,
			Privilege:    f.approval.Privilege,
			MutatingOps:  f.approval.MutatingOps,
			ProdTargets:  f.approval.ProdTargets,
			TotalTargets: f.approval.TotalTargets,
			DecidedAt:    f.approval.DecidedAt,
		}
	}
	logger := security.NewAuditLogger("")
	if err := logger.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write audit log: %v\n", err)
	}
}
