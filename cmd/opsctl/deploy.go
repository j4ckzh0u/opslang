// deploy command for opsctl - remote orchestration with script compilation
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/j4ckzh0u/opslang/internal/ast"
	"github.com/j4ckzh0u/opslang/internal/compiler"
	opsexec "github.com/j4ckzh0u/opslang/internal/exec"
	"github.com/j4ckzh0u/opslang/internal/inventory"
	"github.com/j4ckzh0u/opslang/internal/modules"
	"github.com/j4ckzh0u/opslang/internal/parser"
	"github.com/j4ckzh0u/opslang/internal/runner"
	"github.com/j4ckzh0u/opslang/internal/security"
	"github.com/spf13/cobra"
)

var (
	deployTargets         string
	deployInventory       string
	deployParallel        int
	deployDryRun          bool
	deployMode            string
	deployUser            string
	deployKey             string
	deployPassword        string
	deployOutput          string
	deployInsecureHostKey bool
	deployAutoApprove     bool
	deployLimitCPU        int
	deployLimitMemMB      int64
	deploySignKey         string
	deployVerifyKey       string
	deployVulnDBURL       string
	deployVulnDBToken     string
	deployVulnDBVersion   string
	deployVulnDBSHA256    string
	deployVulnDBCA        string
	deployVulnDBTimeout   time.Duration
)

// deploySignKeyBytes holds the Ed25519 private key loaded once per deploy
// run; nil means packages are sent unsigned.
var deploySignKeyBytes []byte

// Injection points so tests can drive the approval gate without a TTY.
var (
	deployConfirmFn          = promptConfirm
	deployStdinIsInteractive = stdinIsInteractive
)

var deployCmd = &cobra.Command{
	Use:   "deploy [script.ops]",
	Short: "Deploy an OpsLang script to remote hosts",
	Long: `Parse an OpsLang script, compile or interpret it, and deploy to remote hosts.

Supports two execution modes:
  - runner: Generate JSON instruction packages, send to remote runner.
            Fast, zero compile on the target, supports linear scripts
            (calls, let, report, alert, log). Control flow is rejected
            with an explicit error rather than mistranslated.
  - aot:    Compile the script to a static binary for each target
            architecture, upload it, and execute it. Supports the full
            language (if/for/while/fn/ensure/parallel).
  - auto:   Choose runner unless the script uses control flow (default).

Results are aggregated and output as JSON.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve --auto-approve (flag over env) here: reading the flag set
		// from helpers would create an init cycle via deployCmd.
		auto, src := resolveAutoApprove(cmd.Flags().Changed("auto-approve"), deployAutoApprove)
		return runDeployCommand(args[0], auto, src)
	},
}

func init() {
	deployCmd.Flags().StringVar(&deployTargets, "targets", "", "Target hosts (comma-separated user@host)")
	deployCmd.Flags().StringVar(&deployInventory, "inventory", "", "Inventory file path (YAML)")
	deployCmd.Flags().IntVar(&deployParallel, "parallel", 10, "Maximum concurrent hosts")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "Execute in dry-run mode")
	deployCmd.Flags().StringVar(&deployMode, "mode", "auto", "Execution mode: auto, runner, aot")
	deployCmd.Flags().StringVarP(&deployUser, "user", "u", "root", "Default SSH user")
	deployCmd.Flags().StringVarP(&deployKey, "key", "i", "", "SSH private key path")
	deployCmd.Flags().StringVarP(&deployPassword, "password", "p", "", "SSH password")
	deployCmd.Flags().StringVarP(&deployOutput, "output", "o", "", "Output file path (default: stdout)")
	deployCmd.Flags().BoolVar(&deployInsecureHostKey, "insecure-host-key", false, "Skip SSH host key verification (TOFU still applies by default; lab use only)")
	deployCmd.Flags().BoolVar(&deployAutoApprove, "auto-approve", false, "Pre-approve gated deploys (admin/root scripts on production targets); non-interactive runs are refused without it")
	deployCmd.Flags().IntVar(&deployLimitCPU, "limit-cpu", 0, "Cap remote runner CPU usage (percent, requires systemd-run on targets; 0 = off)")
	deployCmd.Flags().Int64Var(&deployLimitMemMB, "limit-mem", 0, "Cap remote runner memory (MB, requires systemd-run on targets; 0 = off)")
	deployCmd.Flags().StringVar(&deploySignKey, "sign-key", "", "Ed25519 private key (from opsctl keygen) used to sign instruction packages")
	deployCmd.Flags().StringVar(&deployVerifyKey, "verify-key", "", "REMOTE path of the trusted public key; runners refuse unsigned/tampered packages")
	deployCmd.Flags().StringVar(&deployVulnDBURL, "vulndb-url", "", "Controller vulnerability HTTPS endpoint")
	deployCmd.Flags().StringVar(&deployVulnDBToken, "vulndb-token", "", "Task-scoped vulnerability service bearer token")
	deployCmd.Flags().StringVar(&deployVulnDBVersion, "vulndb-rule-version", "", "Expected vulnerability rule version")
	deployCmd.Flags().StringVar(&deployVulnDBSHA256, "vulndb-rule-sha256", "", "Expected vulnerability rule SHA-256")
	deployCmd.Flags().StringVar(&deployVulnDBCA, "vulndb-ca", "", "PEM CA bundle for the vulnerability service")
	deployCmd.Flags().DurationVar(&deployVulnDBTimeout, "vulndb-timeout", 30*time.Second, "Vulnerability service request timeout")
}

// deployStep is one task with independently signed execution phases.
// Statements outside any task form the "all targets" step; each task
// statement routes its body to the targets its on-clause selects.
type deployStep struct {
	name    string
	targets []opsexec.Target
	main    *runner.InstructionPackage
	rescue  *runner.InstructionPackage
	always  *runner.InstructionPackage
}

func runDeployCommand(scriptPath string, autoApprove bool, autoSource approvalSource) error {
	if deployTargets == "" && deployInventory == "" {
		return fmt.Errorf("either --targets or --inventory must be specified")
	}

	source, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script %s: %w", scriptPath, err)
	}

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

	// Third-party Go imports are not implemented in any engine; fail now
	// instead of generating code that cannot compile.
	for _, stmt := range prog.Statements {
		if imp, ok := stmt.(*ast.ImportStatement); ok {
			if strings.HasPrefix(imp.Path, "go ") || strings.HasPrefix(imp.Path, "go:") {
				return fmt.Errorf("import %q: third-party Go imports are not supported yet", imp.Path)
			}
		}
	}

	scriptPriv := security.GetScriptPrivilege(prog)
	fmt.Fprintf(os.Stderr, "Script privilege: %s\n", scriptPriv)

	// Privilege gate before anything touches a host: a read_only script
	// with mutating calls is refused here in every mode (runner generation
	// and AOT compilation both re-check independently).
	if err := compiler.CheckPrivileges(prog); err != nil {
		return fmt.Errorf("privilege check failed: %w", err)
	}

	mode := resolveDeployMode(deployMode, prog)
	fmt.Fprintf(os.Stderr, "Deploy mode: %s\n", mode)
	if mode == "aot" && hasRemoteVulnerabilitySettings(deployVulnDBURL, deployVulnDBToken, deployVulnDBVersion, deployVulnDBSHA256, deployVulnDBCA) {
		return fmt.Errorf("vulnerability database options require runner mode")
	}

	// Load the signing key before any host is contacted: a bad key path
	// must fail the deploy here, not mid-flight after some hosts ran.
	if deploySignKey != "" {
		key, err := loadSignKey(deploySignKey)
		if err != nil {
			return fmt.Errorf("failed to load signing key: %w", err)
		}
		deploySignKeyBytes = key
	}

	targets := buildDeployTargets()
	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	startedAt := time.Now().UTC()
	taskID := generateTaskID(scriptPath)

	// Approval gate: privileged scripts (admin/root) aimed at production
	// targets need explicit approval before any host is contacted.
	approvalRec, err := enforceDeployApprovalGate(scriptPath, prog, targets, autoApprove, autoSource)
	if err != nil {
		writeAudit(auditFacts{
			taskID: taskID, script: scriptPath, privilege: string(scriptPriv),
			targets: targetAddressList(targets), user: deployUser, mode: mode, dryRun: deployDryRun,
			durationMs: time.Since(startedAt).Milliseconds(), runErr: err, approval: approvalRec,
		})
		return err
	}

	var result *deployAggregate
	switch mode {
	case "runner":
		result, err = deployRunnerMode(ctx, scriptPath, prog, targets, taskID, scriptPriv)
	case "aot":
		result, err = deployAOTMode(ctx, scriptPath, prog, targets, taskID, scriptPriv)
	default:
		err = fmt.Errorf("unknown mode: %s", mode)
	}
	if err != nil {
		writeAudit(auditFacts{
			taskID: taskID, script: scriptPath, privilege: string(scriptPriv),
			targets: targetAddressList(targets), user: deployUser, mode: mode, dryRun: deployDryRun,
			durationMs: time.Since(startedAt).Milliseconds(), runErr: err, approval: approvalRec,
		})
		return err
	}

	// A partial deployment is a failure from the operator's point of view;
	// the audit entry must reflect that, not record "success".
	statusErr := error(nil)
	switch result.Status {
	case "failed":
		statusErr = fmt.Errorf("deployment failed")
	case "partial":
		statusErr = fmt.Errorf("deployment partially failed: some hosts did not complete successfully")
	}

	writeAudit(auditFacts{
		taskID:    taskID,
		script:    scriptPath,
		privilege: string(scriptPriv),
		targets:   targetAddressList(targets),
		user:      deployUser,
		mode:      mode,
		dryRun:    deployDryRun,
		status:    result.Status,
		runErr:    statusErr,
		approval:  approvalRec,
		results: map[string]interface{}{
			"tasks":    result.TaskResults,
			"packages": result.TaskPackages,
		},
	})

	return outputDeployResult(result, startedAt, scriptPath)
}

// resolveDeployMode picks the execution mode. Runner mode can only express
// linear scripts (no control flow, no computed expressions), so auto mode
// generates a trial instruction package: whatever the real generator
// rejects goes to AOT. This keeps the mode decision and the generator in
// lockstep - feature-detection here used to drift from the generator.
func resolveDeployMode(mode string, prog *ast.Program) string {
	if mode == "runner" || mode == "aot" {
		return mode
	}
	if compiler.RequiresAOT(prog) {
		return "aot"
	}
	gen := &runner.InstructionGenerator{}
	if _, err := gen.GenerateFromStatements(prog.Statements, false); err != nil {
		return "aot"
	}
	return "runner"
}

// buildDeployTargets assembles the target list from flags and inventory.
func buildDeployTargets() []opsexec.Target {
	var targets []opsexec.Target
	if deployTargets != "" {
		hosts := strings.Split(deployTargets, ",")
		for i := range hosts {
			hosts[i] = strings.TrimSpace(hosts[i])
		}
		targets = append(targets, opsexec.ParseTargets(hosts, deployUser)...)
	}
	if deployInventory != "" {
		inv, err := inventory.ParseFile(deployInventory)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to parse inventory: %v\n", err)
			return targets
		}
		targets = append(targets, opsexec.TargetsFromInventory(inv)...)
	}
	return targets
}

// ============================================================
// Runner mode
// ============================================================

// deployAggregate merges the per-step execution summaries.
type deployAggregate struct {
	TaskID       string
	Status       string
	Targets      []string
	Results      map[string]*opsexec.HostResult
	TaskResults  map[string]map[string]*deployHostPhases
	TaskPackages map[string]map[string]string
}

type deployHostPhases struct {
	Status  string              `json:"status"`
	Trigger string              `json:"trigger,omitempty"`
	Main    *opsexec.HostResult `json:"main"`
	Rescue  *opsexec.HostResult `json:"rescue,omitempty"`
	Always  *opsexec.HostResult `json:"always,omitempty"`
}

func (a *deployAggregate) add(summary *opsexec.Summary) {
	if a.Results == nil {
		a.Results = make(map[string]*opsexec.HostResult)
	}
	for name, r := range summary.Results {
		if existing, ok := a.Results[name]; ok {
			// A playbook runs several steps per host. Later steps add
			// their reports and errors to the host record instead of
			// silently overwriting earlier ones: the final JSON must
			// show every step's outcome, not just the last.
			if len(r.Data) > 0 {
				if existing.Data == nil {
					existing.Data = make(map[string]interface{})
				}
				for k, v := range r.Data {
					existing.Data[k] = v
				}
			}
			if r.Status == "failed" || r.Status == "partial" {
				existing.Status = r.Status
			}
			existing.Errors = append(existing.Errors, r.Errors...)
			if len(existing.Errors) > 0 && existing.Error == "" {
				existing.Error = existing.Errors[0]
			}
			continue
		}
		cp := *r
		a.Results[name] = &cp
	}
	switch summary.Status {
	case "failed":
		a.Status = "failed"
	case "partial":
		if a.Status != "failed" {
			a.Status = "partial"
		}
	case "success":
		if a.Status == "" {
			a.Status = "success"
		}
	}
}

func (a *deployAggregate) recordPhase(taskName, phase string, summary *opsexec.Summary) {
	if a.TaskResults == nil {
		a.TaskResults = make(map[string]map[string]*deployHostPhases)
	}
	if a.TaskResults[taskName] == nil {
		a.TaskResults[taskName] = make(map[string]*deployHostPhases)
	}
	for host, result := range summary.Results {
		phases := a.TaskResults[taskName][host]
		if phases == nil {
			phases = &deployHostPhases{}
			a.TaskResults[taskName][host] = phases
		}
		switch phase {
		case "main":
			phases.Main = result
		case "rescue":
			phases.Rescue = result
		case "always":
			phases.Always = result
		}
	}
	a.add(summary)
}

func (a *deployAggregate) finalizeTask(taskName string) bool {
	mainFailed := false
	for host, phases := range a.TaskResults[taskName] {
		phases.Status = phaseStatus(phases)
		if phases.Main == nil || phases.Main.Status != "success" {
			mainFailed = true
			if phases.Main != nil {
				phases.Trigger = phases.Main.Error
				if phases.Trigger == "" && len(phases.Main.Errors) > 0 {
					phases.Trigger = phases.Main.Errors[0]
				}
			}
		}
		if result := a.Results[host]; result != nil {
			result.Status = phases.Status
		}
	}
	if mainFailed {
		a.Status = "failed"
	}
	return mainFailed
}

func phaseStatus(phases *deployHostPhases) string {
	if phases == nil {
		return "failed"
	}
	if phases.Always != nil && phases.Always.Status != "success" {
		return "cleanup_failed"
	}
	if phases.Main != nil && phases.Main.Status == "success" {
		return "success"
	}
	if phases.Rescue == nil {
		return "failed"
	}
	if phases.Rescue.Status == "success" {
		return "rolled_back"
	}
	return "rollback_failed"
}

// signPkg signs an instruction package when --sign-key was provided; it is
// a no-op otherwise. Called after DryRun is applied so the dry-run bit is
// covered by the signature.
func signPkg(pkg *runner.InstructionPackage) error {
	if deploySignKeyBytes == nil {
		return nil
	}
	return runner.SignPackage(pkg, ed25519.PrivateKey(deploySignKeyBytes))
}

func deployRunnerMode(ctx context.Context, scriptPath string, prog *ast.Program, targets []opsexec.Target, taskID string, scriptPriv ast.PrivilegeLevel) (*deployAggregate, error) {
	remoteConfig, err := loadRemoteVulnerabilityConfig(deployVulnDBURL, deployVulnDBToken, deployVulnDBVersion, deployVulnDBSHA256, deployVulnDBCA, deployVulnDBTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid vulnerability database configuration: %w", err)
	}
	steps, err := buildDeploySteps(prog, targets, taskID, scriptPriv)
	if err != nil {
		return nil, err
	}

	agg := &deployAggregate{TaskID: taskID, Targets: targetNames(targets)}
	// One architecture cache per deploy run: task steps share detections
	// instead of re-reading the disk file per executor.
	archCache := archCacheForRun()
	limit := resourceLimitFromFlags(deployLimitCPU, deployLimitMemMB)
	connectionPool, err := opsexec.NewConnectionPool(deployParallel)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH connection pool: %w", err)
	}
	executor := &opsexec.Executor{
		User:                      deployUser,
		KeyFile:                   deployKey,
		Password:                  deployPassword,
		Parallel:                  deployParallel,
		DryRun:                    deployDryRun,
		InsecureSkipHostKeyVerify: deployInsecureHostKey,
		ArchCache:                 archCache,
		ResourceLimit:             limit,
		RunnerVerifyKeyPath:       deployVerifyKey,
		ConnectionPool:            connectionPool,
	}
	defer func() {
		if err := executor.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close SSH connection pool: %v\n", err)
		}
	}()
	for _, step := range steps {
		fmt.Fprintf(os.Stderr, "Step %q: %d instruction(s) on %d host(s)\n",
			step.name, len(step.main.Instructions), len(step.targets))

		for _, phasePkg := range []struct {
			name string
			pkg  *runner.InstructionPackage
		}{
			{name: "main", pkg: step.main},
			{name: "rescue", pkg: step.rescue},
			{name: "always", pkg: step.always},
		} {
			phase, pkg := phasePkg.name, phasePkg.pkg
			if pkg == nil {
				continue
			}
			pkg.DryRun = deployDryRun
			if err := runner.InjectRemoteVulnerabilityConfig(pkg, remoteConfig); err != nil {
				return nil, fmt.Errorf("configure vulnerability query for step %q %s: %w", step.name, phase, err)
			}
			if err := signPkg(pkg); err != nil {
				return nil, fmt.Errorf("failed to sign package for step %q %s: %w", step.name, phase, err)
			}
		}
		if agg.TaskPackages == nil {
			agg.TaskPackages = make(map[string]map[string]string)
		}
		agg.TaskPackages[step.name] = phasePackageIDs(step)

		executePhase := func(phase string, pkg *runner.InstructionPackage, phaseTargets []opsexec.Target) *opsexec.Summary {
			executor.Targets = phaseTargets
			executor.Instructions = pkg
			executor.TaskID = taskID + "-" + step.name + "-" + phase
			return executor.Execute(ctx)
		}
		if runDeployStep(step, deployDryRun, executePhase, agg) {
			break
		}
	}

	return agg, nil
}

func phasePackageIDs(step deployStep) map[string]string {
	ids := map[string]string{"main": step.main.TaskID}
	if step.rescue != nil {
		ids["rescue"] = step.rescue.TaskID
	}
	if step.always != nil {
		ids["always"] = step.always.TaskID
	}
	return ids
}

type deployPhaseExecutor func(string, *runner.InstructionPackage, []opsexec.Target) *opsexec.Summary

func runDeployStep(step deployStep, dryRun bool, execute deployPhaseExecutor, agg *deployAggregate) bool {
	mainSummary := execute("main", step.main, step.targets)
	agg.recordPhase(step.name, "main", mainSummary)
	failedTargets := targetsWithFailedResults(step.targets, mainSummary)
	if step.rescue != nil && (len(failedTargets) > 0 || dryRun) {
		rescueTargets := failedTargets
		if dryRun {
			rescueTargets = step.targets
		}
		agg.recordPhase(step.name, "rescue", execute("rescue", step.rescue, rescueTargets))
	}
	if step.always != nil {
		agg.recordPhase(step.name, "always", execute("always", step.always, step.targets))
	}
	return agg.finalizeTask(step.name)
}

func targetsWithFailedResults(targets []opsexec.Target, summary *opsexec.Summary) []opsexec.Target {
	failed := make([]opsexec.Target, 0)
	if summary == nil {
		return append(failed, targets...)
	}
	for _, target := range targets {
		result, ok := summary.Results[target.Name]
		if !ok || result.Status != "success" {
			failed = append(failed, target)
		}
	}
	return failed
}

// buildDeploySteps converts the program into per-task instruction packages
// with their target subsets. Statements outside tasks run on all targets.
// scriptPriv governs generation-time privilege enforcement and travels with
// every package so the remote runner can re-check it.
func buildDeploySteps(prog *ast.Program, targets []opsexec.Target, taskID string, scriptPriv ast.PrivilegeLevel) ([]deployStep, error) {
	var steps []deployStep

	var prelude []ast.Statement
	for _, stmt := range prog.Statements {
		// Privilege declarations are metadata (checked by the gate above
		// and carried on every package); a script whose only non-task
		// statement is `privilege:` has no runnable prelude.
		if _, isPriv := stmt.(*ast.PrivilegeStatement); isPriv {
			continue
		}
		task, ok := stmt.(*ast.TaskStatement)
		if !ok {
			prelude = append(prelude, stmt)
			continue
		}

		subset, err := selectTaskTargets(task, targets)
		if err != nil {
			return nil, fmt.Errorf("task %q: %w", task.Name, err)
		}
		if len(subset) == 0 {
			return nil, fmt.Errorf("task %q: its on-clause selects none of the deploy targets %v",
				task.Name, targetNames(targets))
		}

		gen := &runner.InstructionGenerator{Privilege: scriptPriv}
		pkg, err := gen.Generate(task, deployDryRun)
		if err != nil {
			return nil, fmt.Errorf("task %q: %w", task.Name, err)
		}
		pkg.TaskID = taskID + "-" + sanitizeStepName(task.Name)
		if err := runner.ValidatePackage(pkg); err != nil {
			return nil, fmt.Errorf("task %q: invalid instruction package: %w", task.Name, err)
		}
		steps = append(steps, deployStep{
			name:    sanitizeStepName(task.Name),
			targets: subset,
			main:    pkg,
		})
		step := &steps[len(steps)-1]
		for _, phase := range []struct {
			name  string
			block *ast.BlockStatement
			dest  **runner.InstructionPackage
		}{
			{name: "rescue", block: task.Rescue, dest: &step.rescue},
			{name: "always", block: task.Always, dest: &step.always},
		} {
			if phase.block == nil {
				continue
			}
			phaseGen := &runner.InstructionGenerator{Privilege: scriptPriv}
			phasePkg, phaseErr := phaseGen.GenerateBlock(phase.block, deployDryRun)
			if phaseErr != nil {
				return nil, fmt.Errorf("task %q %s: %w", task.Name, phase.name, phaseErr)
			}
			phasePkg.TaskID = taskID + "-" + sanitizeStepName(task.Name) + "-" + phase.name
			if phaseErr := runner.ValidatePackage(phasePkg); phaseErr != nil {
				return nil, fmt.Errorf("task %q %s: invalid instruction package: %w", task.Name, phase.name, phaseErr)
			}
			*phase.dest = phasePkg
		}
	}

	if len(prelude) > 0 {
		gen := &runner.InstructionGenerator{Privilege: scriptPriv}
		pkg, err := gen.GenerateFromStatements(prelude, deployDryRun)
		if err != nil {
			return nil, err
		}
		pkg.TaskID = taskID + "-main"
		if err := runner.ValidatePackage(pkg); err != nil {
			return nil, fmt.Errorf("invalid instruction package: %w", err)
		}
		// Prelude runs first, on every target.
		steps = append([]deployStep{{
			name:    "main",
			targets: targets,
			main:    pkg,
		}}, steps...)
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("script contains no runnable statements")
	}
	return steps, nil
}

// selectTaskTargets resolves a task's on-clause against the deploy targets.
// Accepted selectors: exact target name, glob (path.Match syntax), or an
// exact host address. Anything dynamic fails loudly.
func selectTaskTargets(task *ast.TaskStatement, targets []opsexec.Target) ([]opsexec.Target, error) {
	if task.Targets == nil {
		return targets, nil
	}
	if task.Targets.Var != nil {
		return nil, fmt.Errorf("variable target %q cannot be resolved at deploy time; use literal host selectors", task.Targets.Var.Name)
	}

	var selected []opsexec.Target
	for _, expr := range task.Targets.Hosts {
		switch e := expr.(type) {
		case *ast.StringLiteral:
			for _, t := range targets {
				if targetMatchesSelector(t, e.Value) && !containsTarget(selected, t) {
					selected = append(selected, t)
				}
			}
		case *ast.CallExpression:
			return nil, fmt.Errorf("dynamic selectors like %s are not supported in deploy yet; list hosts literally", expr.String())
		default:
			return nil, fmt.Errorf("unsupported target selector: %s", expr.String())
		}
	}
	return selected, nil
}

// targetMatchesSelector matches a selector against the target's name,
// host address, or user@host form. Globbing via path.Match is allowed.
func targetMatchesSelector(t opsexec.Target, selector string) bool {
	candidates := []string{t.Name, t.Host, t.User + "@" + t.Host}
	if t.Group != "" {
		// Inventory groups route like Ansible's hosts: field.
		candidates = append(candidates, t.Group)
	}
	for _, c := range candidates {
		if c == selector {
			return true
		}
		if ok, err := path.Match(selector, c); err == nil && ok {
			return true
		}
	}
	return false
}

func containsTarget(list []opsexec.Target, t opsexec.Target) bool {
	for _, x := range list {
		if x.Name == t.Name && x.Host == t.Host {
			return true
		}
	}
	return false
}

func targetNames(targets []opsexec.Target) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = t.Name
	}
	return names
}

func sanitizeStepName(name string) string {
	r := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ":", "_", "\"", "")
	return r.Replace(name)
}

// ============================================================
// AOT mode
// ============================================================

func deployAOTMode(ctx context.Context, scriptPath string, prog *ast.Program, targets []opsexec.Target, taskID string, scriptPriv ast.PrivilegeLevel) (*deployAggregate, error) {
	// Task routing happens inside the compiled binary on each host — but a
	// self-contained binary cannot know which host it lands on, so routed
	// tasks would silently run on EVERY target. Reject instead of misroute.
	for _, stmt := range prog.Statements {
		if task, ok := stmt.(*ast.TaskStatement); ok && task.Targets != nil {
			return nil, fmt.Errorf("task %q: task-level \"on\" routing requires runner mode (linear task bodies); AOT runs the whole script on every target", task.Name)
		}
	}

	c, err := compiler.NewCompiler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize compiler: %w", err)
	}

	// One shared output directory; one binary per target architecture.
	outDir, err := os.MkdirTemp("", "ops-deploy-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(outDir)

	appBinary := func(goos, goarch string) (string, error) {
		binPath := filepath.Join(outDir, "ops-app-"+goos+"-"+goarch)
		if _, err := os.Stat(binPath); err == nil {
			return binPath, nil
		}
		if err := c.Compile(scriptPath, goos+"/"+goarch, binPath); err != nil {
			return "", fmt.Errorf("compilation for %s/%s failed: %w", goos, goarch, err)
		}
		return binPath, nil
	}

	// The runner executes the uploaded binary; the executor rewrites the
	// placeholder to the per-host remote path after upload. The binary's
	// parsed report is assigned to "app" so it lands in the deploy output.
	pkg := newAOTInstructionPackage(taskID, deployDryRun, scriptPriv)

	if err := signPkg(pkg); err != nil {
		return nil, fmt.Errorf("failed to sign AOT instruction package: %w", err)
	}
	connectionPool, err := opsexec.NewConnectionPool(deployParallel)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH connection pool: %w", err)
	}

	executor := &opsexec.Executor{
		Targets:                   targets,
		Instructions:              pkg,
		User:                      deployUser,
		KeyFile:                   deployKey,
		Password:                  deployPassword,
		Parallel:                  deployParallel,
		DryRun:                    deployDryRun,
		TaskID:                    taskID,
		AppBinary:                 appBinary,
		InsecureSkipHostKeyVerify: deployInsecureHostKey,
		ArchCache:                 archCacheForRun(),
		ResourceLimit:             resourceLimitFromFlags(deployLimitCPU, deployLimitMemMB),
		RunnerVerifyKeyPath:       deployVerifyKey,
		ConnectionPool:            connectionPool,
	}

	summary := executor.Execute(ctx)
	if err := executor.Close(); err != nil {
		return nil, fmt.Errorf("failed to close SSH connection pool: %w", err)
	}

	agg := &deployAggregate{TaskID: taskID, Targets: targetNames(targets)}
	agg.add(summary)
	return agg, nil
}

// newAOTInstructionPackage wraps a statically checked compiled program in a
// transport package. The wrapper itself launches a binary, so it carries
// admin transport privilege. The original script privilege is checked before
// compilation by compiler.CheckPrivileges and by the deploy approval gate.
func newAOTInstructionPackage(taskID string, dryRun bool, _ ast.PrivilegeLevel) *runner.InstructionPackage {
	return &runner.InstructionPackage{
		Version:   "1.0",
		TaskID:    taskID,
		DryRun:    dryRun,
		Privilege: string(ast.PrivilegeAdmin),
		Instructions: []runner.Instruction{
			{
				Op:     "binary.exec",
				Args:   map[string]interface{}{"path": opsexec.AppBinaryPlaceholder},
				Assign: "app",
			},
		},
	}
}

func targetAddressList(targets []opsexec.Target) []string {
	names := make([]string, len(targets))
	for i, t := range targets {
		names[i] = fmt.Sprintf("%s@%s", t.User, t.Host)
	}
	return names
}

func generateTaskID(scriptPath string) string {
	name := strings.TrimSuffix(scriptPath, ".ops")
	name = strings.ReplaceAll(name, "/", "_")
	return fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
}

func outputDeployResult(agg *deployAggregate, startedAt time.Time, scriptPath string) error {
	deployResult := map[string]interface{}{
		"task_id":       agg.TaskID,
		"script":        scriptPath,
		"started_at":    startedAt.Format(time.RFC3339),
		"finished_at":   time.Now().UTC().Format(time.RFC3339),
		"status":        agg.Status,
		"targets":       agg.Targets,
		"results":       agg.Results,
		"task_results":  agg.TaskResults,
		"task_packages": agg.TaskPackages,
	}

	result, err := json.MarshalIndent(deployResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if deployOutput != "" {
		if err := os.WriteFile(deployOutput, append(result, '\n'), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Results written to %s\n", deployOutput)
	} else {
		fmt.Println(string(result))
	}

	// partial is a failure from the operator's point of view: some hosts
	// did not reach the desired state.
	switch agg.Status {
	case "failed":
		return fmt.Errorf("deployment failed")
	case "partial":
		return fmt.Errorf("deployment partially failed: some hosts did not complete successfully")
	}
	return nil
}
