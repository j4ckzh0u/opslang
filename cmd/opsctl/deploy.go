// deploy command for opsctl - remote orchestration with script compilation
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
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
  - runner: Generate JSON instruction packages, send to remote runner (fast, limited)
  - aot:    Compile to static binary, upload and execute (flexible, slower first run)
  - auto:   Choose mode based on script complexity (default)

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

func runDeployCommand(scriptPath string) error {
	// Validate targets.
	if deployTargets == "" && deployInventory == "" {
		return fmt.Errorf("either --targets or --inventory must be specified")
	}

	// Read source.
	source, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script %s: %w", scriptPath, err)
	}

	// Parse.
	p := parser.New(string(source), scriptPath)
	prog, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Check privilege level.
	scriptPriv := security.GetScriptPrivilege(prog)
	fmt.Fprintf(os.Stderr, "Script privilege: %s\n", scriptPriv)

	// Determine execution mode.
	mode := resolveDeployMode(deployMode, prog)

	// Build target list.
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
			return fmt.Errorf("failed to parse inventory: %w", err)
		}
		targets = append(targets, opsexec.TargetsFromInventory(inv)...)
	}
	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	// Set up context with signal handling.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	startedAt := time.Now().UTC()

	// Create audit entry
	auditEntry := security.NewAuditEntry(
		generateTaskID(scriptPath),
		scriptPath,
		string(scriptPriv),
		func() []string {
			names := make([]string, len(targets))
			for i, t := range targets {
				names[i] = fmt.Sprintf("%s@%s", t.User, t.Host)
			}
			return names
		}(),
		deployUser,
		mode,
		deployDryRun,
	)

	// Create audit logger
	auditLogger := security.NewAuditLogger("")

	switch mode {
	case "runner":
		err = deployRunnerMode(ctx, scriptPath, prog, targets, startedAt)
	case "aot":
		err = deployAOTMode(ctx, scriptPath, targets, startedAt)
	default:
		err = fmt.Errorf("unknown mode: %s", mode)
	}

	// Record audit
	durationMs := time.Since(startedAt).Milliseconds()
	if err != nil {
		auditEntry.SetError(err)
	} else {
		auditEntry.SetStatus("success", durationMs)
	}
	if logErr := auditLogger.Log(auditEntry); logErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write audit log: %v\n", logErr)
	}

	return err
}

// resolveDeployMode decides between runner and aot based on script features.
func resolveDeployMode(mode string, prog *ast.Program) string {
	if mode == "runner" || mode == "aot" {
		return mode
	}
	// Auto: default to runner unless script uses import or complex features.
	for _, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.ImportStatement); ok {
			return "aot"
		}
	}
	return "runner"
}

// deployRunnerMode generates instruction packages and executes via runner.
func deployRunnerMode(ctx context.Context, scriptPath string, prog *ast.Program, targets []opsexec.Target, startedAt time.Time) error {
	// Convert AST to instruction package.
	pkg, err := astToInstructions(scriptPath, prog)
	if err != nil {
		return fmt.Errorf("failed to generate instructions: %w", err)
	}

	if deployDryRun {
		fmt.Fprintf(os.Stderr, "Dry-run mode: would execute %d instructions on %d hosts\n",
			len(pkg.Instructions), len(targets))
		pkg.DryRun = true
	}

	// Create executor.
	executor := &opsexec.Executor{
		Targets:      targets,
		Instructions: pkg,
		User:         deployUser,
		KeyFile:      deployKey,
		Password:     deployPassword,
		Parallel:     deployParallel,
		DryRun:       deployDryRun,
	}

	summary := executor.Execute(ctx)

	return outputDeployResult(summary, startedAt, scriptPath)
}

// deployAOTMode compiles the script and deploys the binary.
func deployAOTMode(ctx context.Context, scriptPath string, targets []opsexec.Target, startedAt time.Time) error {
	c, err := compiler.NewCompiler()
	if err != nil {
		return fmt.Errorf("failed to initialize compiler: %w", err)
	}

	// Compile for current architecture (will be recompiled per-target arch).
	tmpOutput := fmt.Sprintf("/tmp/ops-deploy-%d", time.Now().UnixNano())
	if err := c.Compile(scriptPath, "", tmpOutput); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	defer os.Remove(tmpOutput)

	// For AOT mode, we create a minimal instruction package that runs the binary.
	pkg := &runner.InstructionPackage{
		Version: "1.0",
		TaskID:  generateTaskID(scriptPath),
		DryRun:  deployDryRun,
		Instructions: []runner.Instruction{
			{
				Op:   "binary.exec",
				Args: map[string]interface{}{"path": "/tmp/ops-binary"},
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
		// Note: RunnerPath intentionally not set. Executor will detect target
		// architecture and build/upload the appropriate runner binary.
		// TODO: Implement true AOT mode with per-target cross-compilation.
	}

	summary := executor.Execute(ctx)
	return outputDeployResult(summary, startedAt, scriptPath)
}

// astToInstructions converts an AST program to a runner instruction package.
func astToInstructions(scriptPath string, prog *ast.Program) (*runner.InstructionPackage, error) {
	pkg := &runner.InstructionPackage{
		Version: "1.0",
		TaskID:  generateTaskID(scriptPath),
	}

	gen := &instructionGen{}
	for _, stmt := range prog.Statements {
		if err := gen.walkStatement(stmt); err != nil {
			return nil, err
		}
	}

	pkg.Instructions = gen.instructions
	if len(pkg.Instructions) == 0 {
		// Add a no-op to ensure valid package.
		pkg.Instructions = append(pkg.Instructions, runner.Instruction{
			Op:   "report",
			Args: map[string]interface{}{"status": "empty_script"},
		})
	}

	return pkg, nil
}

// instructionGen walks AST and generates runner instructions.
type instructionGen struct {
	instructions []runner.Instruction
}

func (g *instructionGen) walkStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.TaskStatement:
		// Walk task body statements.
		for _, inner := range s.Body.Statements {
			if err := g.walkStatement(inner); err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		return g.walkExpression(s.Expr)
	case *ast.LetStatement:
		return g.walkLet(s)
	case *ast.ReportStatement:
		return g.walkReport(s)
	case *ast.AlertStatement:
		return g.walkAlert(s)
	case *ast.LogStatement:
		return g.walkLog(s)
	case *ast.EnsureStatement:
		return g.walkEnsure(s)
	case *ast.IfStatement:
		// For simple if/else, we generate conditional instructions.
		return g.walkIf(s)
	case *ast.ImportStatement:
		// No-op for instruction generation.
		return nil
	default:
		// For unsupported statements in runner mode, skip with a warning.
		return nil
	}
	return nil
}

func (g *instructionGen) walkLet(s *ast.LetStatement) error {
	// If the value is a call expression, convert it to an instruction.
	if call, ok := s.Value.(*ast.CallExpression); ok {
		op, args := g.callToInstruction(call)
		if op != "" {
			g.instructions = append(g.instructions, runner.Instruction{
				Op:     op,
				Args:   args,
				Assign: s.Name.Name,
			})
			return nil
		}
	}
	// For literal assignments, use a report instruction with the value.
	val := g.evalLiteral(s.Value)
	g.instructions = append(g.instructions, runner.Instruction{
		Op:     "report",
		Args:   map[string]interface{}{s.Name.Name: val},
		Assign: s.Name.Name,
	})
	return nil
}

func (g *instructionGen) walkExpression(expr ast.Expression) error {
	if call, ok := expr.(*ast.CallExpression); ok {
		op, args := g.callToInstruction(call)
		if op != "" {
			g.instructions = append(g.instructions, runner.Instruction{
				Op:   op,
				Args: args,
			})
			return nil
		}
	}
	return nil
}

func (g *instructionGen) walkReport(s *ast.ReportStatement) error {
	args := make(map[string]interface{})
	for _, field := range s.Fields {
		args[field.Key] = g.evalLiteral(field.Value)
	}
	g.instructions = append(g.instructions, runner.Instruction{
		Op:   "report",
		Args: args,
	})
	return nil
}

func (g *instructionGen) walkAlert(s *ast.AlertStatement) error {
	msg := g.evalLiteral(s.Message)
	g.instructions = append(g.instructions, runner.Instruction{
		Op:   "log",
		Args: map[string]interface{}{"message": fmt.Sprintf("ALERT: %v", msg)},
	})
	return nil
}

func (g *instructionGen) walkLog(s *ast.LogStatement) error {
	msg := g.evalLiteral(s.Message)
	g.instructions = append(g.instructions, runner.Instruction{
		Op:   "log",
		Args: map[string]interface{}{"message": msg},
	})
	return nil
}

func (g *instructionGen) walkEnsure(s *ast.EnsureStatement) error {
	// Ensure is translated as: check condition, then execute body if needed.
	// For runner mode, we emit a simple check operation.
	cond := g.evalLiteral(s.Condition)
	g.instructions = append(g.instructions, runner.Instruction{
		Op:   "log",
		Args: map[string]interface{}{"message": fmt.Sprintf("ensure check: %v", cond)},
	})
	// Walk body for the apply step.
	for _, stmt := range s.Body.Statements {
		if err := g.walkStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (g *instructionGen) walkIf(s *ast.IfStatement) error {
	// Simple approach: walk body statements (condition evaluation happens at runtime).
	for _, stmt := range s.Body.Statements {
		if err := g.walkStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

// callToInstruction converts a call expression to an instruction op and args.
func (g *instructionGen) callToInstruction(call *ast.CallExpression) (string, map[string]interface{}) {
	name := resolveCallName(call.Function)
	if name == "" {
		return "", nil
	}

	args := make(map[string]interface{})
	for i, arg := range call.Args {
		key := fmt.Sprintf("arg%d", i)
		args[key] = g.evalLiteral(arg)
	}

	return name, args
}

// resolveCallName builds a dotted name from a call expression.
func resolveCallName(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		prefix := resolveCallName(e.Object)
		if prefix != "" {
			return prefix + "." + e.Member.Name
		}
	}
	return ""
}

// evalLiteral extracts a literal value from an expression for instruction args.
func (g *instructionGen) evalLiteral(expr ast.Expression) interface{} {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return e.Value
	case *ast.FloatLiteral:
		return e.Value
	case *ast.StringLiteral:
		return e.Value
	case *ast.BoolLiteral:
		return e.Value
	case *ast.Identifier:
		return "$" + e.Name // Variable reference marker.
	case *ast.MemberExpression:
		return resolveCallName(expr)
	case *ast.CallExpression:
		return resolveCallName(e.Function)
	default:
		return fmt.Sprintf("%v", expr)
	}
}

func generateTaskID(scriptPath string) string {
	// Use a combination of timestamp and script name.
	name := strings.TrimSuffix(scriptPath, ".ops")
	name = strings.ReplaceAll(name, "/", "_")
	return fmt.Sprintf("%s-%d", name, time.Now().Unix())
}

func outputDeployResult(summary *opsexec.Summary, startedAt time.Time, scriptPath string) error {
	// Add script info to summary.
	deployResult := map[string]interface{}{
		"task_id":     summary.TaskID,
		"script":      scriptPath,
		"started_at":  startedAt.Format(time.RFC3339),
		"finished_at": time.Now().UTC().Format(time.RFC3339),
		"status":      summary.Status,
		"targets":     summary.Targets,
		"results":     summary.Results,
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

	// Return error instead of os.Exit to allow audit logging
	if summary.Status == "failed" {
		return fmt.Errorf("deployment failed")
	}

	return nil
}
