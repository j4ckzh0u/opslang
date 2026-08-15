// Package compiler implements the AOT compilation pipeline for OpsLang.
// It translates an AST into Go source code and compiles it into a static binary.
package compiler

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/opslang/opslang/internal/ast"
)

// sdkMapping maps OpsLang dotted function names to Go SDK calls.
var sdkMapping = map[string]sdkFunc{
	// sys
	"sys.cpu.usage":       {pkg: "sys", goName: "GetCPUUsage"},
	"sys.cpu.count":       {pkg: "sys", goName: "GetCPUCount"},
	"sys.cpu.info":        {pkg: "sys", goName: "GetCPUInfo"},
	"sys.memory.info":     {pkg: "sys", goName: "GetMemoryInfo"},
	"sys.disk.usage":      {pkg: "sys", goName: "GetDiskUsage", args: true},
	"sys.disk.partitions": {pkg: "sys", goName: "GetDiskPartitions"},
	"sys.host.info":       {pkg: "sys", goName: "GetHostInfo"},
	"sys.hostname":        {pkg: "sys", goName: "Hostname"},
	"sys.load":            {pkg: "sys", goName: "GetLoadAvg"},
	"sys.net.interfaces":  {pkg: "sys", goName: "GetNetInterfaces"},
	"sys.users":           {pkg: "sys", goName: "Users"},
	"sys.uptime":          {pkg: "sys", goName: "Uptime"},

	// file
	"file.read":     {pkg: "file", goName: "Read", args: true},
	"file.write":    {pkg: "file", goName: "Write", args: true},
	"file.exists":   {pkg: "file", goName: "Exists", args: true},
	"file.copy":     {pkg: "file", goName: "Copy", args: true},
	"file.move":     {pkg: "file", goName: "Move", args: true},
	"file.delete":   {pkg: "file", goName: "Delete", args: true},
	"file.info":     {pkg: "file", goName: "Stat", args: true},
	"file.list":     {pkg: "file", goName: "List", args: true},
	"file.mkdir":    {pkg: "file", goName: "Mkdir", args: true},
	"file.checksum": {pkg: "file", goName: "Checksum", args: true},

	// net
	"net.http.get":    {pkg: "net", goName: "HTTPGet", args: true},
	"net.http.post":   {pkg: "net", goName: "HTTPPost", args: true},
	"net.tcp.ping":    {pkg: "net", goName: "TCPConnect", args: true},
	"net.dns.resolve": {pkg: "net", goName: "DNSLookup", args: true},
	"net.interfaces":  {pkg: "net", goName: "Interfaces"},

	// process
	"process.list":         {pkg: "process", goName: "List"},
	"process.find.by_name": {pkg: "process", goName: "FindByName", args: true},
	"process.find.by_port": {pkg: "process", goName: "FindByPort", args: true},
	"process.exec":         {pkg: "process", goName: "Exec", args: true},

	// service
	"service.status":  {pkg: "service", goName: "Status", args: true},
	"service.start":   {pkg: "service", goName: "Start", args: true},
	"service.stop":    {pkg: "service", goName: "Stop", args: true},
	"service.restart": {pkg: "service", goName: "Restart", args: true},
	"service.enable":  {pkg: "service", goName: "Enable", args: true},
	"service.disable": {pkg: "service", goName: "Disable", args: true},

	// pkg
	"pkg.install": {pkg: "pkg", goName: "Install", args: true},
	"pkg.remove":  {pkg: "pkg", goName: "Remove", args: true},
	"pkg.info":    {pkg: "pkg", goName: "Info", args: true},
	"pkg.list":    {pkg: "pkg", goName: "List"},

	// time
	"time.now":    {pkg: "time", goName: "Now"},
	"time.format": {pkg: "time", goName: "Format", args: true},
	"time.parse":  {pkg: "time", goName: "Parse", args: true},
	"time.since":  {pkg: "time", goName: "Since", args: true},
	"time.sleep":  {pkg: "time", goName: "Sleep", args: true},
	"time.diff":   {pkg: "time", goName: "Diff", args: true},

	// json / yaml
	"json.encode": {pkg: "json", goName: "Encode", args: true},
	"json.decode": {pkg: "json", goName: "Decode", args: true},
	"yaml.encode": {pkg: "yaml", goName: "Encode", args: true},
	"yaml.decode": {pkg: "yaml", goName: "Decode", args: true},
}

// sdkFunc describes how an OpsLang SDK call maps to Go.
type sdkFunc struct {
	pkg    string // short package key (e.g. "sys", "net")
	goName string // Go function name without package prefix
	args   bool   // whether the function takes arguments
}

// pkgImportAlias maps our short package key to the import alias used in generated Go code.
var pkgImportAlias = map[string]string{
	"sys":     "sys",
	"file":    "file",
	"net":     "opsnet",
	"process": "process",
	"service": "service",
	"pkg":     "opspkg",
	"time":    "opstime",
	"json":    "opsjson",
	"yaml":    "opsyaml",
}

// pkgImportPath maps our short package key to the full import path.
var pkgImportPath = map[string]string{
	"sys":     "github.com/opslang/opslang/pkg/ops-core-sdk/sys",
	"file":    "github.com/opslang/opslang/pkg/ops-core-sdk/file",
	"net":     "github.com/opslang/opslang/pkg/ops-core-sdk/net",
	"process": "github.com/opslang/opslang/pkg/ops-core-sdk/process",
	"service": "github.com/opslang/opslang/pkg/ops-core-sdk/service",
	"pkg":     "github.com/opslang/opslang/pkg/ops-core-sdk/pkg",
	"time":    "github.com/opslang/opslang/pkg/ops-core-sdk/time",
	"json":    "github.com/opslang/opslang/pkg/ops-core-sdk/json",
	"yaml":    "github.com/opslang/opslang/pkg/ops-core-sdk/yaml",
}

// CodeGenerator translates an AST Program into Go source code.
type CodeGenerator struct {
	indent    int
	buf       strings.Builder
	usedSDK   map[string]bool // tracks which SDK packages are used
	useOS     bool            // whether "os" is needed
	userFuncs []userFunc      // user-defined functions collected during generation
}

type userFunc struct {
	name   string
	params []ast.Parameter
	body   *ast.BlockStatement
}

// Generate takes an AST Program and returns a complete Go source string.
func (g *CodeGenerator) Generate(prog *ast.Program) (string, error) {
	g.usedSDK = make(map[string]bool)
	g.userFuncs = nil
	g.useOS = false

	// First pass: collect user-defined functions
	for _, stmt := range prog.Statements {
		if fn, ok := stmt.(*ast.FnStatement); ok {
			g.userFuncs = append(g.userFuncs, userFunc{
				name:   fn.Name.Name,
				params: fn.Params,
				body:   fn.Body,
			})
		}
	}

	// Second pass: generate main body into a temp buffer
	var mainBody strings.Builder
	savedBuf := g.buf
	savedIndent := g.indent
	g.buf = mainBody
	g.indent = 1

	for _, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.FnStatement); ok {
			continue // already collected
		}
		if err := g.genStatement(stmt); err != nil {
			return "", err
		}
	}

	mainCode := g.buf.String()
	g.buf = savedBuf
	g.indent = savedIndent

	// Assemble the full file
	return g.assemble(mainCode)
}

// assemble builds the complete Go source file with imports, helpers, and main.
func (g *CodeGenerator) assemble(mainCode string) (string, error) {
	var b strings.Builder

	b.WriteString("// Code generated by OpsLang AOT compiler. DO NOT EDIT.\n")
	b.WriteString("package main\n\n")

	// Collect all imports
	b.WriteString("import (\n")
	// Standard library imports (always needed by helpers)
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"strings\"\n")
	if g.useOS {
		b.WriteString("\t\"os\"\n")
	}

	// SDK imports
	sdkOrder := []string{"sys", "file", "net", "process", "service", "pkg", "time", "json", "yaml"}
	for _, pkg := range sdkOrder {
		if g.usedSDK[pkg] {
			alias := pkgImportAlias[pkg]
			path := pkgImportPath[pkg]
			b.WriteString(fmt.Sprintf("\t%s %q\n", alias, path))
		}
	}
	b.WriteString(")\n\n")

	// Suppress unused import warnings for always-imported stdlib packages
	b.WriteString("// Ensure standard library imports are used.\n")
	b.WriteString("var (\n")
	b.WriteString("\t_ = fmt.Println\n")
	b.WriteString("\t_ = json.Marshal\n")
	b.WriteString("\t_ = strings.Join\n")
	if g.useOS {
		b.WriteString("\t_ = os.Stderr\n")
	}
	b.WriteString(")\n\n")

	// Runtime helpers
	g.writeHelpers(&b)

	// User-defined functions
	for _, fn := range g.userFuncs {
		if err := g.writeUserFunc(&b, fn); err != nil {
			return "", err
		}
	}

	// Main function
	b.WriteString("func main() {\n")
	b.WriteString("\t_output := make(map[string]interface{})\n")
	b.WriteString("\t_ = _output\n")
	b.WriteString(mainCode)
	b.WriteString("\n\t// Print final output as JSON\n")
	b.WriteString("\tif len(_output) > 0 {\n")
	b.WriteString("\t\tdata, _ := json.MarshalIndent(_output, \"\", \"  \")\n")
	b.WriteString("\t\tfmt.Println(string(data))\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")

	return b.String(), nil
}

// writeHelpers emits runtime helper functions used by generated code.
func (g *CodeGenerator) writeHelpers(b *strings.Builder) {
	b.WriteString(`func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case bool:
		if val {
			return 1.0
		}
		return 0.0
	case nil:
		return 0.0
	default:
		return 0.0
	}
}

func toInt(v interface{}) int64 {
	return int64(toFloat(v))
}

func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case string:
		return val
	case []interface{}:
		parts := make([]string, len(val))
		for i, elem := range val {
			parts[i] = formatValue(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		data, _ := json.Marshal(val)
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case map[string]interface{}:
		return len(val) > 0
	default:
		return true
	}
}

`)
}

// writeUserFunc generates a Go function for a user-defined OpsLang function.
func (g *CodeGenerator) writeUserFunc(b *strings.Builder, fn userFunc) error {
	b.WriteString(fmt.Sprintf("func %s(", fn.name))
	for i, param := range fn.params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s interface{}", sanitizeName(param.Name.Name)))
	}
	b.WriteString(") interface{} {\n")

	for _, stmt := range fn.body.Statements {
		if err := g.genStatementTo(b, stmt, 1); err != nil {
			return err
		}
	}

	b.WriteString("\treturn nil\n")
	b.WriteString("}\n\n")
	return nil
}

// genStatement generates Go code for a statement, appending to g.buf.
func (g *CodeGenerator) genStatement(stmt ast.Statement) error {
	return g.genStatementTo(&g.buf, stmt, g.indent)
}

func (g *CodeGenerator) genStatementTo(b *strings.Builder, stmt ast.Statement, indent int) error {
	prefix := strings.Repeat("\t", indent)

	switch s := stmt.(type) {
	case *ast.LetStatement:
		expr, err := g.genExpr(s.Value)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s%s := %s\n", prefix, sanitizeName(s.Name.Name), expr))

	case *ast.AssignStatement:
		target, err := g.genExpr(s.Target)
		if err != nil {
			return err
		}
		val, err := g.genExpr(s.Value)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s%s = %s\n", prefix, target, val))

	case *ast.ExpressionStatement:
		expr, err := g.genExpr(s.Expr)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s_ = %s\n", prefix, expr))

	case *ast.IfStatement:
		return g.genIf(b, s, indent)

	case *ast.ForStatement:
		return g.genFor(b, s, indent)

	case *ast.WhileStatement:
		return g.genWhile(b, s, indent)

	case *ast.ReturnStatement:
		if s.Value != nil {
			expr, err := g.genExpr(s.Value)
			if err != nil {
				return err
			}
			b.WriteString(fmt.Sprintf("%sreturn %s\n", prefix, expr))
		} else {
			b.WriteString(fmt.Sprintf("%sreturn nil\n", prefix))
		}

	case *ast.TaskStatement:
		// In AOT mode, execute task body directly
		for _, inner := range s.Body.Statements {
			if err := g.genStatementTo(b, inner, indent); err != nil {
				return err
			}
		}

	case *ast.ReportStatement:
		return g.genReport(b, s, indent)

	case *ast.AlertStatement:
		g.useOS = true
		msg, err := g.genExpr(s.Message)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s_output[\"alert\"] = %s\n", prefix, msg))
		b.WriteString(fmt.Sprintf("%sfmt.Fprintf(os.Stderr, \"ALERT: %%s\\n\", formatValue(%s))\n", prefix, msg))

	case *ast.ImportStatement:
		// No-op in compiled mode

	case *ast.FnStatement:
		// Already collected

	case *ast.BlockStatement:
		for _, inner := range s.Statements {
			if err := g.genStatementTo(b, inner, indent); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}

	return nil
}

func (g *CodeGenerator) genIf(b *strings.Builder, s *ast.IfStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	cond, err := g.genExpr(s.Condition)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("%sif isTruthy(%s) {\n", prefix, cond))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}

	if s.ElseClause != nil {
		switch e := s.ElseClause.(type) {
		case *ast.BlockStatement:
			b.WriteString(fmt.Sprintf("%s} else {\n", prefix))
			for _, stmt := range e.Statements {
				if err := g.genStatementTo(b, stmt, indent+1); err != nil {
					return err
				}
			}
			b.WriteString(fmt.Sprintf("%s}\n", prefix))
		case *ast.IfStatement:
			b.WriteString(fmt.Sprintf("%s} else ", prefix))
			return g.genIf(b, e, indent)
		}
	} else {
		b.WriteString(fmt.Sprintf("%s}\n", prefix))
	}
	return nil
}

func (g *CodeGenerator) genFor(b *strings.Builder, s *ast.ForStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)

	// C-style for loop
	initStr := ""
	if s.Init != nil {
		var tmp strings.Builder
		if err := g.genStatementTo(&tmp, s.Init, 0); err != nil {
			return err
		}
		initStr = strings.TrimSpace(tmp.String())
	}

	condStr := "true"
	if s.Condition != nil {
		cond, err := g.genExpr(s.Condition)
		if err != nil {
			return err
		}
		condStr = fmt.Sprintf("isTruthy(%s)", cond)
	}

	postStr := ""
	if s.Post != nil {
		var tmp strings.Builder
		if err := g.genStatementTo(&tmp, s.Post, 0); err != nil {
			return err
		}
		postStr = strings.TrimSpace(tmp.String())
	}

	b.WriteString(fmt.Sprintf("%sfor %s; %s; %s {\n", prefix, initStr, condStr, postStr))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

func (g *CodeGenerator) genWhile(b *strings.Builder, s *ast.WhileStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	cond, err := g.genExpr(s.Condition)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("%sfor isTruthy(%s) {\n", prefix, cond))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

func (g *CodeGenerator) genReport(b *strings.Builder, s *ast.ReportStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	for _, field := range s.Fields {
		val, err := g.genExpr(field.Value)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s_output[%q] = %s\n", prefix, field.Key, val))
	}
	return nil
}

// genExpr generates a Go expression string from an AST expression.
func (g *CodeGenerator) genExpr(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("int64(%d)", e.Value), nil

	case *ast.FloatLiteral:
		return fmt.Sprintf("float64(%g)", e.Value), nil

	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value), nil

	case *ast.BoolLiteral:
		if e.Value {
			return "true", nil
		}
		return "false", nil

	case *ast.NilLiteral:
		return "nil", nil

	case *ast.Identifier:
		return sanitizeName(e.Name), nil

	case *ast.BinaryExpression:
		return g.genBinary(e)

	case *ast.UnaryExpression:
		right, err := g.genExpr(e.Right)
		if err != nil {
			return "", err
		}
		if e.Op == "!" {
			return fmt.Sprintf("!isTruthy(%s)", right), nil
		}
		return fmt.Sprintf("(%s%s)", e.Op, right), nil

	case *ast.CallExpression:
		return g.genCall(e)

	case *ast.IndexExpression:
		left, err := g.genExpr(e.Left)
		if err != nil {
			return "", err
		}
		idx, err := g.genExpr(e.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { l := %s; i := %s; switch v := l.(type) { case []interface{}: idx := int(toFloat(i)); if idx >= 0 && idx < len(v) { return v[idx] }; return nil; case map[string]interface{}: return v[fmt.Sprintf(\"%%v\", i)]; default: return nil } }()", left, idx), nil

	case *ast.MemberExpression:
		obj, err := g.genExpr(e.Object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { m, ok := %s.(map[string]interface{}); if !ok { return nil }; return m[%q] }()", obj, e.Member.Name), nil

	case *ast.ListLiteral:
		return g.genList(e)

	case *ast.DictLiteral:
		return g.genDict(e)

	case *ast.IfExpression:
		return g.genIfExpr(e)

	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (g *CodeGenerator) genBinary(e *ast.BinaryExpression) (string, error) {
	left, err := g.genExpr(e.Left)
	if err != nil {
		return "", err
	}
	right, err := g.genExpr(e.Right)
	if err != nil {
		return "", err
	}

	switch e.Op {
	case "+":
		return fmt.Sprintf("func() interface{} { l, r := %s, %s; if ls, ok := l.(string); ok { return ls + formatValue(r) }; if _, ok := r.(string); ok { return formatValue(l) + formatValue(r) }; return toFloat(l) + toFloat(r) }()", left, right), nil
	case "-":
		return fmt.Sprintf("func() interface{} { l, r := toFloat(%s), toFloat(%s); if l == float64(int64(l)) && r == float64(int64(r)) { return int64(l) - int64(r) }; return l - r }()", left, right), nil
	case "*":
		return fmt.Sprintf("func() interface{} { l, r := toFloat(%s), toFloat(%s); if l == float64(int64(l)) && r == float64(int64(r)) { return int64(l) * int64(r) }; return l * r }()", left, right), nil
	case "/":
		return fmt.Sprintf("func() interface{} { l, r := toFloat(%s), toFloat(%s); if r == 0 { return nil }; if l == float64(int64(l)) && r == float64(int64(r)) { return int64(l) / int64(r) }; return l / r }()", left, right), nil
	case "%":
		return fmt.Sprintf("func() interface{} { l, r := toInt(%s), toInt(%s); if r == 0 { return nil }; return l %% r }()", left, right), nil
	case "==":
		return fmt.Sprintf("func() interface{} { l, r := %s, %s; return fmt.Sprintf(\"%%v\", l) == fmt.Sprintf(\"%%v\", r) }()", left, right), nil
	case "!=":
		return fmt.Sprintf("func() interface{} { l, r := %s, %s; return fmt.Sprintf(\"%%v\", l) != fmt.Sprintf(\"%%v\", r) }()", left, right), nil
	case "<":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) < toFloat(%s) }()", left, right), nil
	case ">":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) > toFloat(%s) }()", left, right), nil
	case "<=":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) <= toFloat(%s) }()", left, right), nil
	case ">=":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) >= toFloat(%s) }()", left, right), nil
	case "&&":
		return fmt.Sprintf("func() interface{} { if !isTruthy(%s) { return false }; return isTruthy(%s) }()", left, right), nil
	case "||":
		return fmt.Sprintf("func() interface{} { if isTruthy(%s) { return true }; return isTruthy(%s) }()", left, right), nil
	default:
		return "", fmt.Errorf("unsupported binary operator: %s", e.Op)
	}
}

func (g *CodeGenerator) genCall(e *ast.CallExpression) (string, error) {
	fnName := g.resolveFuncName(e.Function)

	// Check builtins
	switch fnName {
	case "print":
		args, err := g.genArgs(e.Args)
		if err != nil {
			return "", err
		}
		if len(args) == 0 {
			return "func() interface{} { fmt.Println(); return nil }()", nil
		}
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = fmt.Sprintf("formatValue(%s)", a)
		}
		return fmt.Sprintf("func() interface{} { fmt.Println(%s); return nil }()", strings.Join(parts, ", \" \", ")), nil

	case "len":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("len() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { v := %s; switch c := v.(type) { case string: return int64(len(c)); case []interface{}: return int64(len(c)); case map[string]interface{}: return int64(len(c)); default: return int64(0) } }()", arg), nil

	case "str":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("str() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("formatValue(%s)", arg), nil

	case "int":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("int() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("toInt(%s)", arg), nil

	case "float":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("float() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("toFloat(%s)", arg), nil

	case "type":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("type() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("fmt.Sprintf(\"%%T\", %s)", arg), nil
	}

	// Check user-defined functions
	for _, fn := range g.userFuncs {
		if fn.name == fnName {
			args, err := g.genArgs(e.Args)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s(%s)", fn.name, strings.Join(args, ", ")), nil
		}
	}

	// Check SDK mapping
	if sdk, ok := sdkMapping[fnName]; ok {
		g.usedSDK[sdk.pkg] = true
		alias := pkgImportAlias[sdk.pkg]
		args, err := g.genArgs(e.Args)
		if err != nil {
			return "", err
		}
		call := fmt.Sprintf("%s.%s(%s)", alias, sdk.goName, strings.Join(args, ", "))
		return fmt.Sprintf("func() interface{} { v, err := %s; if err != nil { return fmt.Sprintf(\"error: %%v\", err) }; return v }()", call), nil
	}

	// Unknown function - try to evaluate the function expression
	funcExpr, err := g.genExpr(e.Function)
	if err != nil {
		return "", err
	}
	args, err := g.genArgs(e.Args)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s)", funcExpr, strings.Join(args, ", ")), nil
}

func (g *CodeGenerator) genList(e *ast.ListLiteral) (string, error) {
	if len(e.Elements) == 0 {
		return "[]interface{}{}", nil
	}
	elems, err := g.genArgs(e.Elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elems, ", ")), nil
}

func (g *CodeGenerator) genDict(e *ast.DictLiteral) (string, error) {
	if len(e.Keys) == 0 {
		return "map[string]interface{}{}", nil
	}
	var pairs []string
	for i := range e.Keys {
		key, err := g.genExpr(e.Keys[i])
		if err != nil {
			return "", err
		}
		val, err := g.genExpr(e.Values[i])
		if err != nil {
			return "", err
		}
		pairs = append(pairs, fmt.Sprintf("fmt.Sprintf(\"%%v\", %s): %s", key, val))
	}
	return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(pairs, ", ")), nil
}

func (g *CodeGenerator) genIfExpr(e *ast.IfExpression) (string, error) {
	thenExpr, err := g.genExpr(e.Then)
	if err != nil {
		return "", err
	}
	elseExpr, err := g.genExpr(e.Else)
	if err != nil {
		return "", err
	}
	if e.Condition != nil {
		cond, err := g.genExpr(e.Condition)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { if isTruthy(%s) { return %s }; return %s }()", cond, thenExpr, elseExpr), nil
	}
	return thenExpr, nil
}

func (g *CodeGenerator) genArgs(args []ast.Expression) ([]string, error) {
	result := make([]string, len(args))
	for i, arg := range args {
		s, err := g.genExpr(arg)
		if err != nil {
			return nil, err
		}
		result[i] = s
	}
	return result, nil
}

// resolveFuncName builds a dotted name from a call's function expression.
func (g *CodeGenerator) resolveFuncName(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		prefix := g.resolveFuncName(e.Object)
		if prefix != "" {
			return prefix + "." + e.Member.Name
		}
	}
	return ""
}

// sanitizeName escapes Go reserved words and invalid identifier characters
// that might appear in OpsLang variable names.
func sanitizeName(name string) string {
	goReserved := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	if goReserved[name] {
		return "_" + name
	}
	var result strings.Builder
	for i, ch := range name {
		if i == 0 && !unicode.IsLetter(ch) && ch != '_' {
			result.WriteRune('_')
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			result.WriteRune('_')
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

