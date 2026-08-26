package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/security"
	"github.com/j4ckzh0u/opslang/internal/sshx"
)

// ============================================================
// Fixtures
// ============================================================

const approvalTestInventory = `
hosts:
  - name: prod-web-01
    host: 10.0.1.1
    tags:
      env: prod
  - name: prod-web-02
    host: 10.0.1.2
    tags:
      env: prod
  - name: dev-box-01
    host: 10.0.2.1
    tags:
      env: dev
`

const approvalTestDevOnlyInventory = `
hosts:
  - name: dev-box-01
    host: 10.0.2.1
    tags:
      env: dev
`

// approvalTestEnv writes script + inventory files and points the deploy
// flag globals at them. The returned restore func must be deferred.
func approvalTestEnv(t *testing.T, script, inventory string) (scriptPath string, restore func()) {
	t.Helper()

	dir := t.TempDir()
	scriptPath = filepath.Join(dir, "script.ops")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	invPath := filepath.Join(dir, "hosts.yaml")
	if err := os.WriteFile(invPath, []byte(inventory), 0644); err != nil {
		t.Fatalf("write inventory: %v", err)
	}

	savedTargets, savedInventory := deployTargets, deployInventory
	savedUser, savedKey, savedPassword := deployUser, deployKey, deployPassword
	savedOutput, savedInsecure := deployOutput, deployInsecureHostKey
	deployTargets, deployInventory = "", invPath
	deployUser, deployKey, deployPassword = "root", "", ""
	deployOutput, deployInsecureHostKey = "", false

	return scriptPath, func() {
		deployTargets, deployInventory = savedTargets, savedInventory
		deployUser, deployKey, deployPassword = savedUser, savedKey, savedPassword
		deployOutput, deployInsecureHostKey = savedOutput, savedInsecure
	}
}

// countingSSHFactory replaces the SSH factory with one that counts calls
// (and refuses connections, so no runner build or upload can follow).
func countingSSHFactory(t *testing.T) (calls *int64, restore func()) {
	t.Helper()
	orig := opsexec.SSHClientFactory
	var n int64
	opsexec.SSHClientFactory = func(cfg *sshx.Config) (*sshx.Client, error) {
		atomic.AddInt64(&n, 1)
		return nil, fmt.Errorf("mock: ssh refused")
	}
	return &n, func() { opsexec.SSHClientFactory = orig }
}

// forceNonInteractive pins the gate's TTY detection to "no terminal" (what
// CI/piped runs see) regardless of how go test's own stdin is wired.
func forceNonInteractive(t *testing.T) {
	t.Helper()
	savedConfirm, savedInteractive := deployConfirmFn, deployStdinIsInteractive
	deployStdinIsInteractive = func() bool { return false }
	deployConfirmFn = func(string) bool {
		t.Error("non-interactive runs must not prompt")
		return false
	}
	t.Cleanup(func() { deployConfirmFn, deployStdinIsInteractive = savedConfirm, savedInteractive })
}

// readApprovalAudit decodes every audit entry written today that carries an
// approval record.
func readApprovalAudit(t *testing.T) []security.ApprovalRecord {
	t.Helper()
	dir := os.Getenv("OPSLANG_AUDIT_DIR")
	if dir == "" {
		t.Fatal("OPSLANG_AUDIT_DIR not set")
	}
	entries, err := filepath.Glob(filepath.Join(dir, "audit-*.json"))
	if err != nil || len(entries) == 0 {
		t.Fatalf("no audit files in %s (err=%v)", dir, err)
	}
	data, err := os.ReadFile(entries[0])
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	var records []security.ApprovalRecord
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var entry struct {
			Status   string                   `json:"status"`
			Error    string                   `json:"error"`
			Approval *security.ApprovalRecord `json:"approval"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("decode audit line: %v\n%s", err, line)
		}
		if entry.Approval != nil {
			records = append(records, *entry.Approval)
			if entry.Status != "failed" || entry.Error == "" {
				t.Errorf("approval-denied run must audit as failed with an error, got status=%q error=%q", entry.Status, entry.Error)
			}
		}
	}
	return records
}

// ============================================================
// opsctl deploy: end-to-end gate behavior
// ============================================================

// Admin script + production inventory + no TTY + no --auto-approve:
// the deploy is refused and no host is ever contacted.
func TestDeployApproval_NonInteractiveDeniedNoHostContact(t *testing.T) {
	scriptPath, restore := approvalTestEnv(t, `
privilege: admin
let r = file.write("/tmp/approval-check", "data")
`, approvalTestInventory)
	defer restore()

	calls, restoreSSH := countingSSHFactory(t)
	defer restoreSSH()
	forceNonInteractive(t)
	t.Setenv("OPSLANG_AUDIT_DIR", t.TempDir())

	err := runDeployCommand(scriptPath, false, "")
	if err == nil {
		t.Fatal("non-interactive privileged deploy to prod must be refused")
	}
	for _, want := range []string{"approval required", "--auto-approve", security.EnvAutoApprove} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error must mention %q: %v", want, err)
		}
	}
	if n := atomic.LoadInt64(calls); n != 0 {
		t.Fatalf("denied deploy must not contact any host, SSH factory called %d time(s)", n)
	}

	records := readApprovalAudit(t)
	if len(records) != 1 {
		t.Fatalf("expected exactly 1 approval audit record, got %d", len(records))
	}
	rec := records[0]
	if rec.Decision != "denied" || rec.Source != string(security.ApprovalNonInteractive) {
		t.Errorf("record = %+v", rec)
	}
	if rec.Privilege != "admin" || !rec.Required {
		t.Errorf("record privilege fields wrong: %+v", rec)
	}
	if strings.Join(rec.ProdTargets, ",") != "prod-web-01,prod-web-02" {
		t.Errorf("prod targets = %v, want [prod-web-01 prod-web-02]", rec.ProdTargets)
	}
	if len(rec.MutatingOps) != 1 || rec.MutatingOps[0] != "file.write" {
		t.Errorf("mutating ops = %v, want [file.write]", rec.MutatingOps)
	}
}

// With --auto-approve the gate passes and execution proceeds (here until
// the mocked SSH layer refuses).
func TestDeployApproval_AutoApproveProceedsToExecution(t *testing.T) {
	scriptPath, restore := approvalTestEnv(t, `
privilege: admin
let r = file.write("/tmp/approval-check", "data")
`, approvalTestInventory)
	defer restore()

	calls, restoreSSH := countingSSHFactory(t)
	defer restoreSSH()
	forceNonInteractive(t)
	t.Setenv("OPSLANG_AUDIT_DIR", t.TempDir())

	err := runDeployCommand(scriptPath, true, security.ApprovalAutoFlag)
	if err == nil {
		t.Fatal("mocked SSH fails, so the deploy must still report an error")
	}
	if strings.Contains(err.Error(), "approval") {
		t.Errorf("approval must not block an auto-approved run: %v", err)
	}
	if n := atomic.LoadInt64(calls); n == 0 {
		t.Fatal("auto-approved deploy must reach the hosts (SSH factory never called)")
	}
}

// Read-only scripts never trigger the gate, even against prod targets.
func TestDeployApproval_ReadOnlyScriptNotGated(t *testing.T) {
	scriptPath, restore := approvalTestEnv(t, `
let cpu = sys.cpu.usage()
report { cpu: cpu }
`, approvalTestInventory)
	defer restore()

	t.Setenv("OPSLANG_AUDIT_DIR", t.TempDir())
	prog := parseProgram(t, mustRead(t, scriptPath))

	rec, err := enforceDeployApproval(scriptPath, prog, buildDeployTargets(), false, "")
	if err != nil {
		t.Fatalf("read_only script must not be gated: %v", err)
	}
	if rec != nil {
		t.Errorf("no approval record expected: %+v", rec)
	}
}

// Admin scripts on non-production inventories are not gated either.
func TestDeployApproval_NonProdTargetsNotGated(t *testing.T) {
	scriptPath, restore := approvalTestEnv(t, `
privilege: admin
let r = file.write("/tmp/approval-check", "data")
`, approvalTestDevOnlyInventory)
	defer restore()

	prog := parseProgram(t, mustRead(t, scriptPath))
	rec, err := enforceDeployApproval(scriptPath, prog, buildDeployTargets(), false, "")
	if err != nil {
		t.Fatalf("non-prod targets must not be gated: %v", err)
	}
	if rec != nil {
		t.Errorf("no approval record expected: %+v", rec)
	}
}

// Interactive TTY: injected confirm receives the summary; approve passes,
// deny aborts. Both record their outcome.
func TestDeployApproval_InteractiveConfirm(t *testing.T) {
	scriptPath, restore := approvalTestEnv(t, `
privilege: admin
let r = file.write("/tmp/approval-check", "data")
`, approvalTestInventory)
	defer restore()

	savedConfirm, savedInteractive := deployConfirmFn, deployStdinIsInteractive
	defer func() { deployConfirmFn, deployStdinIsInteractive = savedConfirm, savedInteractive }()
	deployStdinIsInteractive = func() bool { return true }
	prog := parseProgram(t, mustRead(t, scriptPath))
	targets := buildDeployTargets()

	var shown string
	deployConfirmFn = func(summary string) bool {
		shown = summary
		return true
	}
	rec, err := enforceDeployApproval(scriptPath, prog, targets, false, "")
	if err != nil {
		t.Fatalf("approved interactively, must proceed: %v", err)
	}
	if rec == nil || rec.Decision != "approved" || rec.Source != string(security.ApprovalInteractive) {
		t.Fatalf("record = %+v", rec)
	}
	for _, want := range []string{scriptPath, "admin", "prod-web-01", "file.write", "3 total, 2 production"} {
		if !strings.Contains(shown, want) {
			t.Errorf("prompt must contain %q:\n%s", want, shown)
		}
	}

	deployConfirmFn = func(string) bool { return false }
	rec, err = enforceDeployApproval(scriptPath, prog, targets, false, "")
	if err == nil {
		t.Fatal("interactive denial must abort")
	}
	if rec == nil || rec.Decision != "denied" || rec.Source != string(security.ApprovalInteractive) {
		t.Errorf("record = %+v", rec)
	}
}

// ============================================================
// opsctl exec: instruction-package gate
// ============================================================

func writeExecInstructions(t *testing.T, privilege string) string {
	t.Helper()
	pkg := map[string]interface{}{
		"version": "1.0",
		"task_id": "approval-exec-test",
		"instructions": []map[string]interface{}{
			{"op": "file.write", "args": map[string]interface{}{"path": "/tmp/x", "content": "y"}},
		},
	}
	if privilege != "" {
		pkg["privilege"] = privilege
	}
	data, err := json.Marshal(pkg)
	if err != nil {
		t.Fatalf("marshal instructions: %v", err)
	}
	path := filepath.Join(t.TempDir(), "instr.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write instructions: %v", err)
	}
	return path
}

func execTestEnv(t *testing.T, instrPath, inventory string) func() {
	t.Helper()
	dir := t.TempDir()
	invPath := filepath.Join(dir, "hosts.yaml")
	if err := os.WriteFile(invPath, []byte(inventory), 0644); err != nil {
		t.Fatalf("write inventory: %v", err)
	}
	savedHosts, savedInv, savedInstr := execHosts, execInventory, execInstructions
	savedUser, savedOutput, savedRunnerPath := execUser, execOutputFile, execRunnerPath
	execHosts, execInventory, execInstructions = nil, invPath, instrPath
	execUser, execOutputFile, execRunnerPath = "root", filepath.Join(dir, "out.json"), ""
	return func() {
		execHosts, execInventory, execInstructions = savedHosts, savedInv, savedInstr
		execUser, execOutputFile, execRunnerPath = savedUser, savedOutput, savedRunnerPath
	}
}

// Privileged instruction package + production inventory + no TTY: refused,
// no host contacted, denial audited.
func TestExecApproval_NonInteractiveDeniedNoHostContact(t *testing.T) {
	instr := writeExecInstructions(t, "admin")
	restore := execTestEnv(t, instr, approvalTestInventory)
	defer restore()

	calls, restoreSSH := countingSSHFactory(t)
	defer restoreSSH()
	forceNonInteractive(t)
	t.Setenv("OPSLANG_AUDIT_DIR", t.TempDir())

	err := runExecCommand(false, "")
	if err == nil {
		t.Fatal("non-interactive privileged exec on prod must be refused")
	}
	if !strings.Contains(err.Error(), "--auto-approve") {
		t.Errorf("error must explain the escape hatch: %v", err)
	}
	if n := atomic.LoadInt64(calls); n != 0 {
		t.Fatalf("denied exec must not contact any host, SSH factory called %d time(s)", n)
	}

	records := readApprovalAudit(t)
	if len(records) != 1 {
		t.Fatalf("expected 1 approval audit record, got %d", len(records))
	}
	if records[0].Decision != "denied" || records[0].Source != string(security.ApprovalNonInteractive) {
		t.Errorf("record = %+v", records[0])
	}
}

// Legacy packages without a privilege field are never gated (consistent
// with the runner's treatment of the empty field).
func TestExecApproval_LegacyPackageNotGated(t *testing.T) {
	instr := writeExecInstructions(t, "")
	restore := execTestEnv(t, instr, approvalTestInventory)
	defer restore()

	pkg, err := opsexec.LoadInstructions(instr)
	if err != nil {
		t.Fatalf("load instructions: %v", err)
	}
	targets, err := buildExecTargets()
	if err != nil {
		t.Fatalf("build targets: %v", err)
	}
	rec, err := enforceExecApproval(instr, pkg, targets, false, "")
	if err != nil {
		t.Fatalf("legacy package must not be gated: %v", err)
	}
	if rec != nil {
		t.Errorf("no approval record expected: %+v", rec)
	}
}

// Auto-approve passes the exec gate without prompting.
func TestExecApproval_AutoApprove(t *testing.T) {
	instr := writeExecInstructions(t, "root")
	restore := execTestEnv(t, instr, approvalTestInventory)
	defer restore()

	forceNonInteractive(t)

	pkg, err := opsexec.LoadInstructions(instr)
	if err != nil {
		t.Fatalf("load instructions: %v", err)
	}
	targets, err := buildExecTargets()
	if err != nil {
		t.Fatalf("build targets: %v", err)
	}
	rec, err := enforceExecApproval(instr, pkg, targets, true, security.ApprovalAutoEnv)
	if err != nil {
		t.Fatalf("auto-approved exec must proceed: %v", err)
	}
	if rec == nil || rec.Decision != "approved" || rec.Source != string(security.ApprovalAutoEnv) {
		t.Errorf("record = %+v", rec)
	}
	if len(rec.MutatingOps) != 1 || rec.MutatingOps[0] != "file.write" {
		t.Errorf("mutating ops = %v, want [file.write]", rec.MutatingOps)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
