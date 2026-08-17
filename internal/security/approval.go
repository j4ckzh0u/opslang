// Approval flow: privileged deploys (privilege admin or root) that target
// production hosts require explicit human approval before any host is
// contacted. The decision logic is pure and testable; the interactive half
// is injected via ConfirmFunc so tests never need a TTY.
package security

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/opsspec"
)

// EnvAutoApprove is the environment variable that pre-approves gated runs
// in non-interactive contexts (CI). --auto-approve takes precedence.
const EnvAutoApprove = "OPSCTL_AUTO_APPROVE"

// TargetTags carries the identity and environment metadata of one deploy
// target. Inventory entries contribute their tags; inline hosts carry none
// and therefore never classify as production — production classification
// is a property of inventory metadata, not of the connection string.
type TargetTags struct {
	Name string
	Tags map[string]string
}

// IsProduction reports whether a target's tags mark it as a production
// host: tag key "env" with value "prod" or "production" (case-insensitive),
// matching the inventory conventions already used in this repo.
func IsProduction(tags map[string]string) bool {
	env := strings.ToLower(strings.TrimSpace(tags["env"]))
	return env == "prod" || env == "production"
}

// ApprovalDecision is the outcome of the pure trigger evaluation: whether
// approval is required plus the summary data shown to the approver.
type ApprovalDecision struct {
	Required     bool
	Privilege    ast.PrivilegeLevel
	MutatingOps  []string // sorted, deduplicated mutating calls in the script
	ProdTargets  []string // names of production-classified targets
	TotalTargets int
}

// EvaluateApproval decides whether a run needs approval: the script must
// declare privilege admin or root AND at least one target must carry a
// production tag. read_only scripts and non-production target sets are
// never gated; undeclared privilege ("", legacy instruction packages)
// behaves like read_only here — its mutating surface is already refused by
// the privilege enforcement layers.
func EvaluateApproval(scriptPriv ast.PrivilegeLevel, mutatingOps []string, targets []TargetTags) ApprovalDecision {
	d := ApprovalDecision{Privilege: scriptPriv, TotalTargets: len(targets)}
	for _, t := range targets {
		if IsProduction(t.Tags) {
			d.ProdTargets = append(d.ProdTargets, t.Name)
		}
	}
	if scriptPriv != ast.PrivilegeAdmin && scriptPriv != ast.PrivilegeRoot {
		return d
	}
	if len(d.ProdTargets) > 0 {
		d.Required = true
		d.MutatingOps = sortedUnique(mutatingOps)
	}
	return d
}

// Summary renders the operator-facing prompt for the decision.
func (d ApprovalDecision) Summary(script string) string {
	var b strings.Builder
	b.WriteString("\n=== OpsLang approval required ===\n")
	fmt.Fprintf(&b, "Script:         %s\n", script)
	fmt.Fprintf(&b, "Privilege:      %s\n", d.Privilege)
	fmt.Fprintf(&b, "Targets:        %d total, %d production\n", d.TotalTargets, len(d.ProdTargets))
	if len(d.ProdTargets) > 0 {
		sample := d.ProdTargets
		if len(sample) > 5 {
			sample = sample[:5]
		}
		fmt.Fprintf(&b, "Prod targets:   %s", strings.Join(sample, ", "))
		if len(d.ProdTargets) > len(sample) {
			fmt.Fprintf(&b, " (+%d more)", len(d.ProdTargets)-len(sample))
		}
		b.WriteString("\n")
	}
	if len(d.MutatingOps) > 0 {
		ops := d.MutatingOps
		if len(ops) > 8 {
			ops = ops[:8]
		}
		fmt.Fprintf(&b, "Mutating ops:   %s", strings.Join(ops, ", "))
		if len(d.MutatingOps) > len(ops) {
			fmt.Fprintf(&b, " (+%d more)", len(d.MutatingOps)-len(ops))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ============================================================
// Mutating-op collection
// ============================================================

// CollectMutatingOps walks the program and returns the sorted, deduplicated
// names of every mutating operation it calls (per the canonical opsspec
// table). It feeds the approval summary; enforcement itself lives in the
// privilege layers.
func CollectMutatingOps(prog *ast.Program) []string {
	seen := make(map[string]bool)
	collectMutatingStatements(prog.Statements, seen)
	ops := make([]string, 0, len(seen))
	for op := range seen {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	return ops
}

func collectMutatingStatements(stmts []ast.Statement, seen map[string]bool) {
	for _, stmt := range stmts {
		collectMutatingStatement(stmt, seen)
	}
}

func collectMutatingStatement(stmt ast.Statement, seen map[string]bool) {
	switch s := stmt.(type) {
	case nil:
	case *ast.LetStatement:
		collectMutatingExpr(s.Value, seen)
	case *ast.FnStatement:
		for _, p := range s.Params {
			collectMutatingExpr(p.Default, seen)
		}
		collectMutatingBlock(s.Body, seen)
	case *ast.IfStatement:
		collectMutatingExpr(s.Condition, seen)
		collectMutatingBlock(s.Body, seen)
		switch e := s.ElseClause.(type) {
		case *ast.BlockStatement:
			collectMutatingBlock(e, seen)
		case *ast.IfStatement:
			collectMutatingStatement(e, seen)
		}
	case *ast.ForStatement:
		collectMutatingStatement(s.Init, seen)
		collectMutatingExpr(s.Condition, seen)
		collectMutatingStatement(s.Post, seen)
		collectMutatingBlock(s.Body, seen)
	case *ast.WhileStatement:
		collectMutatingExpr(s.Condition, seen)
		collectMutatingBlock(s.Body, seen)
	case *ast.ReturnStatement:
		collectMutatingExpr(s.Value, seen)
	case *ast.TaskStatement:
		collectMutatingBlock(s.Body, seen)
	case *ast.ExpressionStatement:
		collectMutatingExpr(s.Expr, seen)
	case *ast.AssignStatement:
		collectMutatingExpr(s.Target, seen)
		collectMutatingExpr(s.Value, seen)
	case *ast.ReportStatement:
		for _, f := range s.Fields {
			collectMutatingExpr(f.Value, seen)
		}
	case *ast.AlertStatement:
		collectMutatingExpr(s.Message, seen)
	case *ast.EnsureStatement:
		collectMutatingExpr(s.Condition, seen)
		collectMutatingBlock(s.Body, seen)
		collectMutatingExpr(s.Notify, seen)
	case *ast.MetricStatement:
		collectMutatingExpr(s.Name, seen)
		collectMutatingExpr(s.Value, seen)
		collectMutatingExpr(s.Labels, seen)
	case *ast.LogStatement:
		collectMutatingExpr(s.Message, seen)
	case *ast.ParallelStatement:
		collectMutatingBlock(s.Body, seen)
	case *ast.BlockStatement:
		collectMutatingBlock(s, seen)
	}
}

func collectMutatingBlock(block *ast.BlockStatement, seen map[string]bool) {
	if block == nil {
		return
	}
	collectMutatingStatements(block.Statements, seen)
}

func collectMutatingExpr(expr ast.Expression, seen map[string]bool) {
	switch e := expr.(type) {
	case nil:
	case *ast.CallExpression:
		if name := resolveCallName(e.Function); name != "" {
			if mutating, known := opsspec.Mutating(name); known && mutating {
				seen[name] = true
			}
		}
		for _, arg := range e.Args {
			collectMutatingExpr(arg, seen)
		}
	case *ast.ListLiteral:
		for _, elem := range e.Elements {
			collectMutatingExpr(elem, seen)
		}
	case *ast.DictLiteral:
		for i := range e.Keys {
			collectMutatingExpr(e.Keys[i], seen)
			collectMutatingExpr(e.Values[i], seen)
		}
	case *ast.BinaryExpression:
		collectMutatingExpr(e.Left, seen)
		collectMutatingExpr(e.Right, seen)
	case *ast.UnaryExpression:
		collectMutatingExpr(e.Right, seen)
	case *ast.IndexExpression:
		collectMutatingExpr(e.Left, seen)
		collectMutatingExpr(e.Index, seen)
	case *ast.MemberExpression:
		collectMutatingExpr(e.Object, seen)
	case *ast.IfExpression:
		collectMutatingExpr(e.Condition, seen)
		collectMutatingExpr(e.Then, seen)
		collectMutatingExpr(e.Else, seen)
	}
}

// resolveCallName builds the dotted name of a call's function expression
// (e.g. sys.cpu.usage) when it is a static identifier/member chain. Empty
// string for anything dynamic.
func resolveCallName(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		if prefix := resolveCallName(e.Object); prefix != "" {
			return prefix + "." + e.Member.Name
		}
	}
	return ""
}

// ============================================================
// Gate: decision + environment → outcome
// ============================================================

// ApprovalSource records how the approval decision was made, for the audit
// trail.
type ApprovalSource string

const (
	ApprovalInteractive    ApprovalSource = "interactive"
	ApprovalAutoFlag       ApprovalSource = "auto-approve-flag"
	ApprovalAutoEnv        ApprovalSource = "auto-approve-env"
	ApprovalNonInteractive ApprovalSource = "non-interactive"
)

// ApprovalRecord is the auditable outcome of an approval decision.
type ApprovalRecord struct {
	Required     bool      `json:"required"`
	Decision     string    `json:"decision"` // "approved" | "denied"
	Source       string    `json:"source"`
	User         string    `json:"user,omitempty"`
	Privilege    string    `json:"privilege"`
	MutatingOps  []string  `json:"mutating_ops,omitempty"`
	ProdTargets  []string  `json:"prod_targets"`
	TotalTargets int       `json:"total_targets"`
	DecidedAt    time.Time `json:"decided_at"`
}

// ConfirmFunc renders the summary, asks the operator, and reports whether
// the run may proceed. Injected by tests.
type ConfirmFunc func(summary string) bool

// ApprovalGate couples a decision with the environment it runs in.
type ApprovalGate struct {
	Decision ApprovalDecision
	// ScriptName is shown in the prompt and record context.
	ScriptName string
	// AutoApprove pre-approves the run (CI); AutoSource records whether it
	// came from the flag or the environment variable.
	AutoApprove bool
	AutoSource  ApprovalSource
	// Interactive reports whether a human can be asked (stdin is a TTY).
	Interactive bool
	// Confirm is the prompt implementation; nil means PromptConfirm.
	Confirm ConfirmFunc
	// User overrides the recorded approver identity (defaults to $USER).
	User string
	// Now overrides the decision timestamp clock.
	Now func() time.Time
}

// Check enforces the gate. It returns the ApprovalRecord for the audit
// trail (nil when no approval was required) and a non-nil error when the
// run must be aborted before contacting any host.
func (g *ApprovalGate) Check() (*ApprovalRecord, error) {
	if !g.Decision.Required {
		return nil, nil
	}

	user := g.User
	if user == "" {
		user = currentUser()
	}
	now := time.Now().UTC()
	if g.Now != nil {
		now = g.Now()
	}
	rec := &ApprovalRecord{
		Required:     true,
		Privilege:    string(g.Decision.Privilege),
		MutatingOps:  g.Decision.MutatingOps,
		ProdTargets:  g.Decision.ProdTargets,
		TotalTargets: g.Decision.TotalTargets,
		User:         user,
		DecidedAt:    now,
	}

	switch {
	case g.AutoApprove:
		rec.Decision = "approved"
		rec.Source = string(g.AutoSource)
		return rec, nil

	case g.Interactive:
		confirm := g.Confirm
		if confirm == nil {
			confirm = PromptConfirm
		}
		if confirm(g.Decision.Summary(g.ScriptName)) {
			rec.Decision = "approved"
			rec.Source = string(ApprovalInteractive)
			return rec, nil
		}
		rec.Decision = "denied"
		rec.Source = string(ApprovalInteractive)
		// The escape hatch is repeated here: a stdin that looks like a
		// terminal but hits EOF immediately (e.g. `< /dev/null`, which is
		// also a char device) lands on this branch, and the operator still
		// needs to know how to proceed in unattended contexts.
		return rec, fmt.Errorf("approval denied by %s: aborted before contacting any host (unattended? use --auto-approve or %s=1)",
			user, EnvAutoApprove)

	default:
		rec.Decision = "denied"
		rec.Source = string(ApprovalNonInteractive)
		return rec, fmt.Errorf(
			"approval required (privilege %s, %d production target(s)) but stdin is not a terminal; "+
				"non-interactive runs are refused by default — rerun with --auto-approve or set %s=1",
			g.Decision.Privilege, len(g.Decision.ProdTargets), EnvAutoApprove)
	}
}

// ResolveAutoApprove applies flag-over-environment precedence. flagSet
// reports whether --auto-approve was explicitly given; an explicit (even
// false) flag wins over OPSCTL_AUTO_APPROVE.
func ResolveAutoApprove(flagSet, flagValue bool) (auto bool, source ApprovalSource) {
	if flagSet {
		return flagValue, ApprovalAutoFlag
	}
	switch strings.ToLower(os.Getenv(EnvAutoApprove)) {
	case "1", "true":
		return true, ApprovalAutoEnv
	}
	return false, ""
}

// StdinIsInteractive reports whether stdin is attached to a terminal.
func StdinIsInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// PromptConfirm shows the summary on stderr and asks a y/N question on
// stdin. Anything other than y/yes denies.
func PromptConfirm(summary string) bool {
	fmt.Fprint(os.Stderr, summary)
	fmt.Fprint(os.Stderr, "Approve execution on these targets? [y/N]: ")

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	}
	return false
}

func currentUser() string {
	for _, key := range []string{"USER", "LOGNAME"} {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return "unknown"
}

func sortedUnique(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
