package security

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/parser"
)

// ============================================================
// Production classification
// ============================================================

func TestIsProduction(t *testing.T) {
	cases := []struct {
		name string
		tags map[string]string
		want bool
	}{
		{"env=prod", map[string]string{"env": "prod"}, true},
		{"env=production", map[string]string{"env": "production"}, true},
		{"env=PROD case-insensitive", map[string]string{"env": "PROD"}, true},
		{"env with spaces", map[string]string{"env": " prod "}, true},
		{"env=dev", map[string]string{"env": "dev"}, false},
		{"env=staging", map[string]string{"env": "staging"}, false},
		{"no tags", nil, false},
		{"empty tags", map[string]string{}, false},
		{"wrong key", map[string]string{"tier": "prod"}, false},
		{"empty env", map[string]string{"env": ""}, false},
	}
	for _, tc := range cases {
		if got := IsProduction(tc.tags); got != tc.want {
			t.Errorf("%s: IsProduction(%v) = %v, want %v", tc.name, tc.tags, got, tc.want)
		}
	}
}

// ============================================================
// Decision evaluation: all trigger branches
// ============================================================

func approvalFixtureTargets() []TargetTags {
	return []TargetTags{
		{Name: "web-01", Tags: map[string]string{"env": "prod"}},
		{Name: "web-02", Tags: map[string]string{"env": "prod"}},
		{Name: "dev-01", Tags: map[string]string{"env": "dev"}},
		{Name: "bare-01"}, // inline host: no tags
	}
}

func TestEvaluateApproval_Branches(t *testing.T) {
	cases := []struct {
		name  string
		priv  ast.PrivilegeLevel
		want  bool
		prods int
	}{
		{"admin + prod targets triggers", ast.PrivilegeAdmin, true, 2},
		{"root + prod targets triggers", ast.PrivilegeRoot, true, 2},
		{"admin + non-prod targets does not trigger", ast.PrivilegeAdmin, false, 0},
		{"root + non-prod targets does not trigger", ast.PrivilegeRoot, false, 0},
		{"read_only + prod targets does not trigger", ast.PrivilegeReadOnly, false, 2},
		{"undeclared privilege + prod targets does not trigger", ast.PrivilegeLevel(""), false, 2},
	}
	for _, tc := range cases {
		var targets []TargetTags
		if tc.prods > 0 {
			targets = approvalFixtureTargets()
		} else {
			targets = []TargetTags{
				{Name: "dev-01", Tags: map[string]string{"env": "dev"}},
				{Name: "bare-01"},
			}
		}

		d := EvaluateApproval(tc.priv, []string{"file.write"}, targets)
		if d.Required != tc.want {
			t.Errorf("%s: Required = %v, want %v", tc.name, d.Required, tc.want)
		}
		if len(d.ProdTargets) != tc.prods {
			t.Errorf("%s: %d prod targets, want %d (%v)", tc.name, len(d.ProdTargets), tc.prods, d.ProdTargets)
		}
		if tc.want {
			if d.Privilege != tc.priv {
				t.Errorf("%s: decision privilege = %q, want %q", tc.name, d.Privilege, tc.priv)
			}
			if len(d.MutatingOps) != 1 || d.MutatingOps[0] != "file.write" {
				t.Errorf("%s: MutatingOps = %v, want [file.write]", tc.name, d.MutatingOps)
			}
		} else {
			// Not required: the op list is only populated for gated runs.
			if len(d.MutatingOps) != 0 {
				t.Errorf("%s: MutatingOps should be empty when not required, got %v", tc.name, d.MutatingOps)
			}
		}
		if d.TotalTargets != len(targets) {
			t.Errorf("%s: TotalTargets = %d, want %d", tc.name, d.TotalTargets, len(targets))
		}
	}
}

func TestEvaluateApproval_ProdTargetNames(t *testing.T) {
	d := EvaluateApproval(ast.PrivilegeAdmin, nil, approvalFixtureTargets())
	if d.ProdTargets[0] != "web-01" || d.ProdTargets[1] != "web-02" {
		t.Errorf("prod target names = %v, want [web-01 web-02]", d.ProdTargets)
	}
}

func TestEvaluateApproval_MutatingOpsDedupSorted(t *testing.T) {
	d := EvaluateApproval(ast.PrivilegeRoot,
		[]string{"pkg.install", "file.write", "pkg.install", "service.stop"},
		[]TargetTags{{Name: "p1", Tags: map[string]string{"env": "prod"}}})
	want := []string{"file.write", "pkg.install", "service.stop"}
	if strings.Join(d.MutatingOps, ",") != strings.Join(want, ",") {
		t.Errorf("MutatingOps = %v, want %v", d.MutatingOps, want)
	}
}

// ============================================================
// Mutating-op collection from the AST
// ============================================================

func TestCollectMutatingOps(t *testing.T) {
	src := `
privilege: admin
import "std"
let ok = sys.cpu.usage()
file.write("/tmp/a", "x")
if true {
	file.write("/tmp/a", "y")            // duplicate
	pkg.install("nginx")
}
for let i = 0; i < 3; i = i + 1 {
	service.restart("app")
}
fn helper() {
	file.delete("/tmp/b")
}
task "t" on "web-01" {
	process.kill(42, "TERM")
	net.http_get("http://x")             // not mutating
}
ensure file.exists("/tmp/c").exists {
	file.append("/tmp/c", "line")
}
parallel {
	pkg.remove("old")
}
`
	p := parser.New(src, "test.ops")
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := CollectMutatingOps(prog)
	want := []string{"file.append", "file.delete", "file.write", "pkg.install", "pkg.remove", "process.kill", "service.restart"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("CollectMutatingOps = %v, want %v", got, want)
	}
}

func TestCollectMutatingOps_Empty(t *testing.T) {
	if got := CollectMutatingOps(&ast.Program{}); len(got) != 0 {
		t.Errorf("empty program should have no mutating ops, got %v", got)
	}
}

// ============================================================
// Gate behavior
// ============================================================

func requiredDecision() ApprovalDecision {
	return EvaluateApproval(ast.PrivilegeAdmin, []string{"file.write"}, approvalFixtureTargets())
}

func TestApprovalGate_NotRequiredPasses(t *testing.T) {
	g := &ApprovalGate{
		Decision: EvaluateApproval(ast.PrivilegeReadOnly, nil, approvalFixtureTargets()),
	}
	rec, err := g.Check()
	if err != nil {
		t.Fatalf("read_only run must not be gated: %v", err)
	}
	if rec != nil {
		t.Errorf("no approval record expected when not required, got %+v", rec)
	}
}

func TestApprovalGate_AutoApprove(t *testing.T) {
	for _, src := range []ApprovalSource{ApprovalAutoFlag, ApprovalAutoEnv} {
		g := &ApprovalGate{
			Decision:    requiredDecision(),
			AutoApprove: true,
			AutoSource:  src,
			Interactive: true, // auto-approve must win even with a TTY
			Confirm: func(string) bool {
				t.Error("auto-approve must not prompt")
				return false
			},
			User: "ci-bot",
			Now:  func() time.Time { return time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC) },
		}
		rec, err := g.Check()
		if err != nil {
			t.Fatalf("source %s: auto-approve must pass: %v", src, err)
		}
		if rec.Decision != "approved" || rec.Source != string(src) {
			t.Errorf("source %s: record = %+v", src, rec)
		}
		if rec.User != "ci-bot" || rec.Privilege != "admin" {
			t.Errorf("source %s: record fields wrong: %+v", src, rec)
		}
		if rec.DecidedAt.IsZero() {
			t.Errorf("source %s: DecidedAt must be set", src)
		}
		if !rec.Required || len(rec.ProdTargets) != 2 {
			t.Errorf("source %s: record must carry prod targets: %+v", src, rec)
		}
	}
}

func TestApprovalGate_InteractiveApprove(t *testing.T) {
	var shown string
	g := &ApprovalGate{
		Decision:    requiredDecision(),
		ScriptName:  "deploy.ops",
		Interactive: true,
		Confirm: func(summary string) bool {
			shown = summary
			return true
		},
	}
	rec, err := g.Check()
	if err != nil {
		t.Fatalf("interactive approval failed: %v", err)
	}
	if rec.Decision != "approved" || rec.Source != string(ApprovalInteractive) {
		t.Errorf("record = %+v", rec)
	}
	if !strings.Contains(shown, "deploy.ops") || !strings.Contains(shown, "admin") ||
		!strings.Contains(shown, "web-01") || !strings.Contains(shown, "file.write") {
		t.Errorf("prompt summary must show script, privilege, prod targets and mutating ops:\n%s", shown)
	}
}

func TestApprovalGate_InteractiveDeny(t *testing.T) {
	g := &ApprovalGate{
		Decision:    requiredDecision(),
		Interactive: true,
		Confirm:     func(string) bool { return false },
		User:        "alice",
	}
	rec, err := g.Check()
	if err == nil {
		t.Fatal("denial must abort the run")
	}
	if rec.Decision != "denied" || rec.Source != string(ApprovalInteractive) {
		t.Errorf("record = %+v", rec)
	}
	if !strings.Contains(err.Error(), "denied by alice") {
		t.Errorf("error should name the denier: %v", err)
	}
}

func TestApprovalGate_NonInteractiveDeniesByDefault(t *testing.T) {
	g := &ApprovalGate{Decision: requiredDecision(), Interactive: false}
	rec, err := g.Check()
	if err == nil {
		t.Fatal("non-interactive runs without --auto-approve must be refused")
	}
	if rec.Decision != "denied" || rec.Source != string(ApprovalNonInteractive) {
		t.Errorf("record = %+v", rec)
	}
	if !strings.Contains(err.Error(), "--auto-approve") || !strings.Contains(err.Error(), EnvAutoApprove) {
		t.Errorf("error must tell the operator how to proceed: %v", err)
	}
}

func TestApprovalGate_DefaultUserFromEnv(t *testing.T) {
	t.Setenv("USER", "env-user")
	g := &ApprovalGate{Decision: requiredDecision(), AutoApprove: true, AutoSource: ApprovalAutoFlag}
	rec, err := g.Check()
	if err != nil {
		t.Fatalf("auto-approve: %v", err)
	}
	if rec.User != "env-user" {
		t.Errorf("record user = %q, want env-user", rec.User)
	}
}

// ============================================================
// --auto-approve / OPSCTL_AUTO_APPROVE resolution
// ============================================================

func TestResolveAutoApprove(t *testing.T) {
	cases := []struct {
		name       string
		flagSet    bool
		flagValue  bool
		env        string
		wantAuto   bool
		wantSource ApprovalSource
	}{
		{"flag true", true, true, "", true, ApprovalAutoFlag},
		{"flag false wins over env=1", true, false, "1", false, ApprovalAutoFlag},
		{"env=1", false, false, "1", true, ApprovalAutoEnv},
		{"env=true", false, false, "true", true, ApprovalAutoEnv},
		{"env=TRUE case-insensitive", false, false, "TRUE", true, ApprovalAutoEnv},
		{"env=0", false, false, "0", false, ""},
		{"env=garbage", false, false, "yes-please", false, ""},
		{"nothing set", false, false, "", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvAutoApprove, tc.env)
			auto, src := ResolveAutoApprove(tc.flagSet, tc.flagValue)
			if auto != tc.wantAuto || src != tc.wantSource {
				t.Errorf("ResolveAutoApprove(%v, %v) with env=%q = (%v, %q), want (%v, %q)",
					tc.flagSet, tc.flagValue, tc.env, auto, src, tc.wantAuto, tc.wantSource)
			}
		})
	}
}

// ============================================================
// Audit integration
// ============================================================

func TestAuditEntryCarriesApprovalRecord(t *testing.T) {
	entry := NewAuditEntry("t1", "s.ops", "admin", []string{"h1"}, "u", "runner", false)
	entry.SetError(errString("approval denied"))
	entry.Approval = &ApprovalRecord{
		Required:    true,
		Decision:    "denied",
		Source:      string(ApprovalNonInteractive),
		User:        "bob",
		Privilege:   "admin",
		ProdTargets: []string{"h1"},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back struct {
		Approval *ApprovalRecord `json:"approval"`
	}
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Approval == nil || back.Approval.Decision != "denied" ||
		back.Approval.Source != string(ApprovalNonInteractive) || back.Approval.User != "bob" {
		t.Errorf("approval record did not survive the audit round-trip: %+v", back.Approval)
	}
}

func TestAuditEntryOmitsApprovalWhenNone(t *testing.T) {
	entry := NewAuditEntry("t2", "s.ops", "read_only", nil, "u", "runner", false)
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(data), "approval") {
		t.Errorf("approval field should be omitted when no approval was required: %s", data)
	}
}

// errString is a tiny error helper to avoid importing errors in two places.
type errString string

func (e errString) Error() string { return string(e) }

// ============================================================
// Summary rendering
// ============================================================

func TestApprovalDecisionSummary(t *testing.T) {
	d := ApprovalDecision{
		Required:     true,
		Privilege:    ast.PrivilegeRoot,
		ProdTargets:  []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7"},
		TotalTargets: 10,
		MutatingOps:  []string{"file.write", "pkg.install"},
	}
	s := d.Summary("reboot.ops")
	for _, want := range []string{"reboot.ops", "root", "10 total, 7 production", "p1", "(+2 more)", "file.write"} {
		if !strings.Contains(s, want) {
			t.Errorf("summary missing %q:\n%s", want, s)
		}
	}
}

// StdinIsInteractive must at least not crash when stdin is a pipe (as under go test).
func TestStdinIsInteractiveUnderTest(t *testing.T) {
	if os.Stdin == nil {
		t.Skip("no stdin")
	}
	_ = StdinIsInteractive() // just exercising; the value depends on the runner
}
