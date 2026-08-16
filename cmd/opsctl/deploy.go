// deploy command for opsctl - remote orchestration with script compilation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/compiler"
	opsexec "github.com/opslang/opslang/internal/exec"
	"github.com/opslang/opslang/internal/inventory"
	"github.com/opslang/opslang/internal/parser"
	"github.com/opslang/opslang/internal/runner"
	"github.com/opslang/opslang/internal/security"
	"github.com/spf13/cobra"
)

var (
	deployTargets   string
	deployInventory string
	deployParallel  int
	deployDryRun    bool
	deployMode      string
	deployUser      string
	deployKey       string
	deployPassword  string
	deployOutput    string
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
		return runDeployCommand(args[0])
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
}

// deployStep is one instruction package to run on one subset of targets.
// Statements outside any task form the "all targets" step; each task
// statement routes its body to the targets its on-clause selects.
type deployStep struct {
	name    string
	targets []opsexec.Target
	pkg     *runner.InstructionPackage
}

func runDeployCommand(scriptPath string) error {
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

	mode := resolveDeployMode(deployMode, prog)
	fmt.Fprintf(os.Stderr, "Deploy mode: %s\n", mode)

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

	var result *deployAggregate
	switch mode {
	case "runner":
		result, err = deployRunnerMode(ctx, scriptPath, prog, targets, taskID)
	case "aot":
		result, err = deployAOTMode(ctx, scriptPath, prog, targets, taskID)
	default:
		err = fmt.Errorf("unknown mode: %s", mode)
	}

	recordDeployAudit(auditParams{
		taskID:     taskID,
		scriptPath: scriptPath,
		privilege:  string(scriptPriv),
		targets:    targets,
		user:       deployUser,
		mode:       mode,
		dryRun:     deployDryRun,
		startedAt:  startedAt,
		runErr:     err,
	})

	if err != nil {
		return err
	}

	return outputDeployResult(result, startedAt, scriptPath)
}

// resolveDeployMode picks the execution mode. Runner mode can only express
// linear scripts, so anything with control flow goes to AOT.
func resolveDeployMode(mode string, prog *ast.Program) string {
	if mode == "runner" || mode == "aot" {
		return mode
	}
	if compiler.RequiresAOT(prog) {
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
	TaskID  string
	Status  string
	Targets []string
	Results map[string]*opsexec.HostResult
}

func (a *deployAggregate) add(summary *opsexec.Summary) {
	if a.Results == nil {
		a.Results = make(map[string]*opsexec.HostResult)
	}
	for name, r := range summary.Results {
		// Earlier steps win the slot; later failures still downgrade status.
		if existing, ok := a.Results[name]; ok && existing.Status == "failed" {
			continue
		}
		a.Results[name] = r
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

func deployRunnerMode(ctx context.Context, scriptPath string, prog *ast.Program, targets []opsexec.Target, taskID string) (*deployAggregate, error) {
	steps, err := buildDeploySteps(prog, targets, taskID)
	if err != nil {
		return nil, err
	}

	agg := &deployAggregate{TaskID: taskID, Targets: targetNames(targets)}
	for _, step := range steps {
		fmt.Fprintf(os.Stderr, "Step %q: %d instruction(s) on %d host(s)\n",
			step.name, len(step.pkg.Instructions), len(step.targets))

		if deployDryRun {
			step.pkg.DryRun = true
		}

		executor := &opsexec.Executor{
			Targets:      step.targets,
			Instructions: step.pkg,
			User:         deployUser,
			KeyFile:      deployKey,
			Password:     deployPassword,
			Parallel:     deployParallel,
			DryRun:       deployDryRun,
			TaskID:       taskID + "-" + step.name,
		}

		summary := executor.Execute(ctx)
		agg.add(summary)
	}

	return agg, nil
}

// buildDeploySteps converts the program into per-task instruction packages
// with their target subsets. Statements outside tasks run on all targets.
func buildDeploySteps(prog *ast.Program, targets []opsexec.Target, taskID string) ([]deployStep, error) {
	var steps []deployStep

	var prelude []ast.Statement
	for _, stmt := range prog.Statements {
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

		gen := &runner.InstructionGenerator{}
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
			pkg:     pkg,
		})
	}

	if len(prelude) > 0 {
		gen := &runner.InstructionGenerator{}
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
			pkg:     pkg,
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

func deployAOTMode(ctx context.Context, scriptPath string, prog *ast.Program, targets []opsexec.Target, taskID string) (*deployAggregate, error) {
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
	// placeholder to the per-host remote path after upload.
	pkg := &runner.InstructionPackage{
		Version: "1.0",
		TaskID:  taskID,
		DryRun:  deployDryRun,
		Instructions: []runner.Instruction{
			{
				Op:   "binary.exec",
				Args: map[string]interface{}{"path": opsexec.AppBinaryPlaceholder},
			},
		},
	}

	executor := &opsexec.Executor{
		Targets:      targets,
		Instructions: pkg,
		User:         deployUser,
		KeyFile:      deployKey,
		Password:     deployPassword,
		Parallel:     deployParallel,
		DryRun:       deployDryRun,
		TaskID:       taskID,
		AppBinary:    appBinary,
	}

	summary := executor.Execute(ctx)

	agg := &deployAggregate{TaskID: taskID, Targets: targetNames(targets)}
	agg.add(summary)
	return agg, nil
}

// ============================================================
// Result output and audit
// ============================================================

// auditParams carries everything the audit entry needs.
type auditParams struct {
	taskID     string
	scriptPath string
	privilege  string
	targets    []opsexec.Target
	user       string
	mode       string
	dryRun     bool
	startedAt  time.Time
	runErr     error
}

// recordDeployAudit writes an honest audit entry: only a nil runErr with a
// successful aggregate is a success. Partial deployments used to be audited
// as "success".
func recordDeployAudit(p auditParams) {
	entry := security.NewAuditEntry(
		p.taskID,
		p.scriptPath,
		p.privilege,
		targetAddressList(p.targets),
		p.user,
		p.mode,
		p.dryRun,
	)
	durationMs := time.Since(p.startedAt).Milliseconds()
	if p.runErr != nil {
		entry.SetError(p.runErr)
	} else {
		entry.SetStatus("success", durationMs)
	}
	logger := security.NewAuditLogger("")
	if err := logger.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write audit log: %v\n", err)
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
		"task_id":     agg.TaskID,
		"script":      scriptPath,
		"started_at":  startedAt.Format(time.RFC3339),
		"finished_at": time.Now().UTC().Format(time.RFC3339),
		"status":      agg.Status,
		"targets":     agg.Targets,
		"results":     agg.Results,
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
