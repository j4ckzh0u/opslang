package compiler

import (
	"strings"
	"testing"

	"github.com/j4ckzh0u/opslang/internal/ast"
)

func TestGenerateLetStatement(t *testing.T) {
	source := `let x = 42`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Top-level lets are hoisted to a package-level declaration and
	// assigned in main, so user functions can reference them.
	if !strings.Contains(code, "var (\n\tx interface{}\n)") {
		t.Errorf("expected hoisted package-level 'x interface{}', got:\n%s", code)
	}
	if !strings.Contains(code, "x = int64(42)") {
		t.Errorf("expected assignment 'x = int64(42)', got:\n%s", code)
	}
}

func TestGenerateStringLiteral(t *testing.T) {
	source := `let name = "hello"`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `"hello"`) {
		t.Errorf("expected string literal, got:\n%s", code)
	}
}

func TestGenerateBoolLiteral(t *testing.T) {
	source := `let flag = true`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "\n\tflag interface{}\n") || !strings.Contains(code, "flag = true") {
		t.Errorf("expected hoisted flag with assignment, got:\n%s", code)
	}
}

func TestGenerateNilLiteral(t *testing.T) {
	source := `let x = nil`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "x = nil") {
		t.Errorf("expected 'x = nil' assignment, got:\n%s", code)
	}
}

func TestGenerateBinaryExpression(t *testing.T) {
	source := `let x = 1 + 2`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "x = ") {
		t.Errorf("expected assignment, got:\n%s", code)
	}
}

func TestGenerateIfStatement(t *testing.T) {
	source := `if true {
  let x = 1
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "if isTruthy(true)") {
		t.Errorf("expected if statement, got:\n%s", code)
	}
}

func TestGenerateIfElse(t *testing.T) {
	source := `if true {
  let x = 1
} else {
  let x = 2
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "} else {") {
		t.Errorf("expected else clause, got:\n%s", code)
	}
}

func TestGenerateWhileStatement(t *testing.T) {
	source := `while true {
  let x = 1
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "for isTruthy(true)") {
		t.Errorf("expected while loop, got:\n%s", code)
	}
}

func TestGenerateForStatement(t *testing.T) {
	source := `for let i = 0; i < 10; i = i + 1 {
  print(i)
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "for") {
		t.Errorf("expected for loop, got:\n%s", code)
	}
}

func TestGenerateReturnStatement(t *testing.T) {
	source := `fn add(a, b) {
  return a
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "func add(") {
		t.Errorf("expected function definition, got:\n%s", code)
	}
	if !strings.Contains(code, "return") {
		t.Errorf("expected return statement, got:\n%s", code)
	}
}

func TestGeneratePrintBuiltin(t *testing.T) {
	source := `print("hello")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "fmt.Println") {
		t.Errorf("expected fmt.Println, got:\n%s", code)
	}
}

func TestGenerateLenBuiltin(t *testing.T) {
	source := `let x = len("hello")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "len(") {
		t.Errorf("expected len() call, got:\n%s", code)
	}
}

func TestGenerateStrBuiltin(t *testing.T) {
	source := `let x = str(42)`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "formatValue(") {
		t.Errorf("expected formatValue() call, got:\n%s", code)
	}
}

func TestGenerateSDKCall(t *testing.T) {
	source := `let cpu = sys.cpu.usage()`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "sys.GetCPUUsage()") {
		t.Errorf("expected sys.GetCPUUsage(), got:\n%s", code)
	}
	if !strings.Contains(code, "sys ") {
		t.Errorf("expected sys import, got:\n%s", code)
	}
}

func TestGenerateSDKCallWithArgs(t *testing.T) {
	source := `let d = sys.disk.usage("/")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `sys.GetDiskUsage(opsStr("/"))`) {
		t.Errorf("expected sys.GetDiskUsage(\"/\"), got:\n%s", code)
	}
}

func TestGenerateNetSDKImport(t *testing.T) {
	source := `let r = net.http_get("http://example.com")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "opsnet") {
		t.Errorf("expected opsnet alias for net SDK, got:\n%s", code)
	}
	if !strings.Contains(code, "opsnet.HTTPGet") {
		t.Errorf("expected opsnet.HTTPGet call, got:\n%s", code)
	}
}

func TestGenerateReportStatement(t *testing.T) {
	source := `report {
  host: "server1",
  cpu: 42
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `setOutput(&_outputMu, _output, "host"`) {
		t.Errorf("expected report output for host, got:\n%s", code)
	}
	if !strings.Contains(code, `setOutput(&_outputMu, _output, "cpu"`) {
		t.Errorf("expected report output for cpu, got:\n%s", code)
	}
}

func TestGenerateAlertStatement(t *testing.T) {
	source := `alert("high cpu")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `setOutput(&_outputMu, _output, "alert"`) {
		t.Errorf("expected alert output, got:\n%s", code)
	}
	if !strings.Contains(code, "ALERT:") {
		t.Errorf("expected ALERT prefix, got:\n%s", code)
	}
}

func TestGenerateTaskStatement(t *testing.T) {
	source := `task "test" on hosts {
  let x = 1
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Task body should be inlined in AOT mode
	if !strings.Contains(code, "var x interface{} = int64(1)") {
		t.Errorf("expected task body inlined, got:\n%s", code)
	}
}

func TestGenerateListLiteral(t *testing.T) {
	source := `let x = [1, 2, 3]`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "[]interface{}") {
		t.Errorf("expected list literal, got:\n%s", code)
	}
}

func TestGenerateDictLiteral(t *testing.T) {
	source := `let x = {"key": "value"}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "map[string]interface{}") {
		t.Errorf("expected dict literal, got:\n%s", code)
	}
}

func TestGenerateUnaryExpression(t *testing.T) {
	source := `let x = !true`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "!isTruthy(true)") {
		t.Errorf("expected unary expression, got:\n%s", code)
	}
}

func TestGenerateAssignment(t *testing.T) {
	source := `let x = 1
x = 2`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "x = int64(2)") {
		t.Errorf("expected assignment, got:\n%s", code)
	}
}

func TestGenerateUserFunction(t *testing.T) {
	source := `fn greet(name) {
  print(name)
}
greet("world")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "func greet(") {
		t.Errorf("expected function definition, got:\n%s", code)
	}
	if !strings.Contains(code, `greet("world")`) {
		t.Errorf("expected function call, got:\n%s", code)
	}
}

func TestGenerateHelpersAlwaysPresent(t *testing.T) {
	source := `let x = 1`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Should always have helper functions
	if !strings.Contains(code, "func toFloat(") {
		t.Errorf("expected toFloat helper, got:\n%s", code)
	}
	if !strings.Contains(code, "func isTruthy(") {
		t.Errorf("expected isTruthy helper, got:\n%s", code)
	}
	if !strings.Contains(code, "func formatValue(") {
		t.Errorf("expected formatValue helper, got:\n%s", code)
	}
}

func TestGenerateMainFunction(t *testing.T) {
	source := `let x = 1`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "func main()") {
		t.Errorf("expected main function, got:\n%s", code)
	}
	if !strings.Contains(code, "_output") {
		t.Errorf("expected _output variable, got:\n%s", code)
	}
}

func TestGenerateSanitizedName(t *testing.T) {
	// Go reserved words should be sanitized
	source := `let type = "string"`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "_type") {
		t.Errorf("expected sanitized name '_type', got:\n%s", code)
	}
}

func TestSanitizeNameFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x", "x"},
		{"myVar", "myVar"},
		{"type", "_type"},
		{"func", "_func"},
		{"return", "_return"},
		{"normal", "normal"},
	}
	for _, tt := range tests {
		result := sanitizeName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateParallelStatement(t *testing.T) {
	source := `parallel {
  let x = 1
  let y = 2
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "sync.WaitGroup") {
		t.Errorf("expected sync.WaitGroup, got:\n%s", code)
	}
	if !strings.Contains(code, "_pWg.Add(2)") {
		t.Errorf("expected _pWg.Add(2), got:\n%s", code)
	}
	if !strings.Contains(code, "go func(") {
		t.Errorf("expected goroutine launch, got:\n%s", code)
	}
	if !strings.Contains(code, "_pWg.Wait()") {
		t.Errorf("expected _pWg.Wait(), got:\n%s", code)
	}
	if !strings.Contains(code, "\"sync\"") {
		t.Errorf("expected sync import, got:\n%s", code)
	}
}

func TestGenerateParallelEmptyBlock(t *testing.T) {
	source := `parallel {}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Empty parallel block should not emit sync import or WaitGroup
	if strings.Contains(code, "sync.WaitGroup") {
		t.Errorf("empty parallel should not use WaitGroup, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Float literal
// ---------------------------------------------------------------------------

func TestGenerateFloatLiteral(t *testing.T) {
	source := `let x = 3.14`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "float64(3.14)") {
		t.Errorf("expected 'float64(3.14)', got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Binary operators (table-driven)
// ---------------------------------------------------------------------------

func TestGenerateBinaryOperators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		contains string
	}{
		{
			name:     "subtraction",
			source:   `let x = 10 - 3`,
			contains: "int64(l) - int64(r)",
		},
		{
			name:     "multiplication",
			source:   `let x = 4 * 5`,
			contains: "int64(l) * int64(r)",
		},
		{
			name:     "division",
			source:   `let x = 10 / 2`,
			contains: "r == 0",
		},
		{
			name:     "modulo",
			source:   `let x = 10 % 3`,
			contains: "toInt",
		},
		{
			name:     "equality",
			source:   `let x = 1 == 2`,
			contains: `fmt.Sprintf("%v"`,
		},
		{
			name:     "inequality",
			source:   `let x = 1 != 2`,
			contains: `fmt.Sprintf("%v"`,
		},
		{
			name:     "less than",
			source:   `let x = 1 < 2`,
			contains: "toFloat(",
		},
		{
			name:     "greater than",
			source:   `let x = 1 > 2`,
			contains: "toFloat(",
		},
		{
			name:     "less than or equal",
			source:   `let x = 1 <= 2`,
			contains: "<=",
		},
		{
			name:     "greater than or equal",
			source:   `let x = 1 >= 2`,
			contains: ">=",
		},
		{
			name:     "logical and",
			source:   `let x = true && false`,
			contains: "!isTruthy(",
		},
		{
			name:     "logical or",
			source:   `let x = true || false`,
			contains: "isTruthy(",
		},
		{
			name:     "string concatenation via plus",
			source:   `let x = "hello" + " world"`,
			contains: "ls + formatValue(r)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateCode(tt.source, "test.ops")
			if err != nil {
				t.Fatalf("GenerateCode failed: %v", err)
			}
			if !strings.Contains(code, tt.contains) {
				t.Errorf("expected code to contain %q, got:\n%s", tt.contains, code)
			}
		})
	}
}

func TestGenerateBinaryDivisionByZero(t *testing.T) {
	// Division should generate a zero-check
	source := `let x = 10 / 0`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "r == 0") {
		t.Errorf("expected division by zero check, got:\n%s", code)
	}
	if !strings.Contains(code, "return nil") {
		t.Errorf("expected nil return on div by zero, got:\n%s", code)
	}
}

func TestGenerateBinaryModuloByZero(t *testing.T) {
	// Modulo should also generate a zero-check
	source := `let x = 10 % 0`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "r == 0") {
		t.Errorf("expected modulo by zero check, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Builtin functions: int(), float(), type()
// ---------------------------------------------------------------------------

func TestGenerateIntBuiltin(t *testing.T) {
	source := `let x = int("42")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "toInt(") {
		t.Errorf("expected toInt() call, got:\n%s", code)
	}
}

func TestGenerateFloatBuiltin(t *testing.T) {
	source := `let x = float("3.14")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "toFloat(") {
		t.Errorf("expected toFloat() call, got:\n%s", code)
	}
}

func TestGenerateTypeBuiltin(t *testing.T) {
	source := `let x = type(42)`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `fmt.Sprintf("%T"`) {
		t.Errorf("expected fmt.Sprintf(%%T) call, got:\n%s", code)
	}
}

func TestGeneratePrintNoArgs(t *testing.T) {
	source := `print()`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "fmt.Println()") {
		t.Errorf("expected fmt.Println() with no args, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// IfExpression (ternary)
// ---------------------------------------------------------------------------

func TestGenerateIfExpression(t *testing.T) {
	source := `let x = if true { 1 } else { 2 }`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "isTruthy(") {
		t.Errorf("expected isTruthy in ternary, got:\n%s", code)
	}
	if !strings.Contains(code, "int64(1)") {
		t.Errorf("expected then branch int64(1), got:\n%s", code)
	}
	if !strings.Contains(code, "int64(2)") {
		t.Errorf("expected else branch int64(2), got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// IndexExpression and MemberExpression
// ---------------------------------------------------------------------------

func TestGenerateIndexExpression(t *testing.T) {
	source := `let x = [1, 2, 3]
let y = x[0]`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "[]interface{}") {
		t.Errorf("expected slice type assertion, got:\n%s", code)
	}
}

func TestGenerateMemberExpression(t *testing.T) {
	source := `let d = {"key": "value"}
let v = d.key`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `map[string]interface{}`) {
		t.Errorf("expected map type assertion, got:\n%s", code)
	}
	if !strings.Contains(code, `"key"`) {
		t.Errorf("expected key lookup, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Unary minus
// ---------------------------------------------------------------------------

func TestGenerateUnaryMinus(t *testing.T) {
	source := `let x = -5`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "-int64(5)") {
		t.Errorf("expected unary minus expression, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// If-else-if chain
// ---------------------------------------------------------------------------

func TestGenerateIfElseIf(t *testing.T) {
	source := `if true {
  let x = 1
} else if false {
  let x = 2
} else {
  let x = 3
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "} else ") {
		t.Errorf("expected else-if chain, got:\n%s", code)
	}
	count := strings.Count(code, "if isTruthy(")
	if count < 2 {
		t.Errorf("expected at least 2 if statements, found %d, got:\n%s", count, code)
	}
}

// ---------------------------------------------------------------------------
// Bare return
// ---------------------------------------------------------------------------

func TestGenerateBareReturn(t *testing.T) {
	source := `fn noop() {
  return
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "return nil") {
		t.Errorf("expected bare return as 'return nil', got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Import statement is no-op in AOT
// ---------------------------------------------------------------------------

func TestGenerateImportIsNoop(t *testing.T) {
	source := `import "something"
let x = 1`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Import statement should not generate any code (no error, just skipped)
	if !strings.Contains(code, "x = int64(1)") {
		t.Errorf("expected let statement after import, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Empty list and dict literals
// ---------------------------------------------------------------------------

func TestGenerateEmptyListLiteral(t *testing.T) {
	source := `let x = []`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "[]interface{}{}") {
		t.Errorf("expected empty list literal, got:\n%s", code)
	}
}

func TestGenerateEmptyDictLiteral(t *testing.T) {
	source := `let x = {}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "map[string]interface{}{}") {
		t.Errorf("expected empty dict literal, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Multiple print args
// ---------------------------------------------------------------------------

func TestGeneratePrintMultipleArgs(t *testing.T) {
	source := `print("a", "b", "c")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "fmt.Println") {
		t.Errorf("expected fmt.Println, got:\n%s", code)
	}
	if !strings.Contains(code, `" "`) {
		t.Errorf("expected space separator for multiple args, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// genExpr via direct AST (unsupported type error)
// ---------------------------------------------------------------------------

func TestGenExprUnsupportedType(t *testing.T) {
	g := &CodeGenerator{}
	// Use a nil expression to trigger the default error case
	_, err := g.genExpr(nil)
	if err == nil {
		t.Fatal("expected error for unsupported expression type")
	}
	if !strings.Contains(err.Error(), "unsupported expression type") {
		t.Errorf("expected 'unsupported expression type' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// genBinary via direct AST (unsupported operator error)
// ---------------------------------------------------------------------------

func TestGenBinaryUnsupportedOperator(t *testing.T) {
	g := &CodeGenerator{}
	expr := &ast.BinaryExpression{
		Left:  &ast.IntegerLiteral{Value: 1},
		Op:    "^",
		Right: &ast.IntegerLiteral{Value: 2},
	}
	_, err := g.genBinary(expr)
	if err == nil {
		t.Fatal("expected error for unsupported binary operator")
	}
	if !strings.Contains(err.Error(), "unsupported binary operator") {
		t.Errorf("expected 'unsupported binary operator' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// genCall via direct AST (builtin arg count errors)
// ---------------------------------------------------------------------------

func TestGenCallBuiltinArgErrors(t *testing.T) {
	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "len with no args",
			call: &ast.CallExpression{
				Function: &ast.Identifier{Name: "len"},
				Args:     []ast.Expression{},
			},
		},
		{
			name: "len with two args",
			call: &ast.CallExpression{
				Function: &ast.Identifier{Name: "len"},
				Args:     []ast.Expression{&ast.StringLiteral{Value: "a"}, &ast.StringLiteral{Value: "b"}},
			},
		},
		{
			name: "str with no args",
			call: &ast.CallExpression{
				Function: &ast.Identifier{Name: "str"},
				Args:     []ast.Expression{},
			},
		},
		{
			name: "int with no args",
			call: &ast.CallExpression{
				Function: &ast.Identifier{Name: "int"},
				Args:     []ast.Expression{},
			},
		},
		{
			name: "float with no args",
			call: &ast.CallExpression{
				Function: &ast.Identifier{Name: "float"},
				Args:     []ast.Expression{},
			},
		},
		{
			name: "type with no args",
			call: &ast.CallExpression{
				Function: &ast.Identifier{Name: "type"},
				Args:     []ast.Expression{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &CodeGenerator{}
			_, err := g.genCall(tt.call)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// genIfExpr via direct AST (nil condition path)
// ---------------------------------------------------------------------------

func TestGenIfExprNilCondition(t *testing.T) {
	g := &CodeGenerator{}
	expr := &ast.IfExpression{
		Condition: nil,
		Then:      &ast.IntegerLiteral{Value: 42},
		Else:      &ast.IntegerLiteral{Value: 0},
	}
	result, err := g.genIfExpr(expr)
	if err != nil {
		t.Fatalf("genIfExpr failed: %v", err)
	}
	// With nil condition, should return just the then expression
	if !strings.Contains(result, "int64(42)") {
		t.Errorf("expected then expression for nil condition, got: %s", result)
	}
}

// ---------------------------------------------------------------------------
// resolveFuncName via direct AST
// ---------------------------------------------------------------------------

func TestResolveFuncName(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expression
		want string
	}{
		{
			name: "simple identifier",
			expr: &ast.Identifier{Name: "print"},
			want: "print",
		},
		{
			name: "member expression",
			expr: &ast.MemberExpression{
				Object: &ast.Identifier{Name: "sys"},
				Member: &ast.Identifier{Name: "cpu"},
			},
			want: "sys.cpu",
		},
		{
			name: "deep member expression",
			expr: &ast.MemberExpression{
				Object: &ast.MemberExpression{
					Object: &ast.Identifier{Name: "sys"},
					Member: &ast.Identifier{Name: "cpu"},
				},
				Member: &ast.Identifier{Name: "usage"},
			},
			want: "sys.cpu.usage",
		},
		{
			name: "unknown expression returns empty",
			expr: &ast.IntegerLiteral{Value: 42},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &CodeGenerator{}
			got := g.resolveFuncName(tt.expr)
			if got != tt.want {
				t.Errorf("resolveFuncName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// genStatementTo default error case
// ---------------------------------------------------------------------------

func TestGenStatementEnsureNowSupported(t *testing.T) {
	// Ensure used to be an unsupported statement type; it now compiles to
	// the check -> apply -> verify contract. Every real AST statement type
	// is handled, so the "unsupported" default case is no longer reachable
	// from valid programs.
	g := &CodeGenerator{}
	stmt := &ast.EnsureStatement{
		Condition: &ast.BoolLiteral{Value: true},
		Body:      &ast.BlockStatement{},
	}
	if err := g.genStatement(stmt); err != nil {
		t.Fatalf("ensure statement must generate code: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SDK import aliases (comprehensive)
// ---------------------------------------------------------------------------

func TestGenerateSDKImportAliases(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		alias   string
		sdkCall string
	}{
		{"sys uses sys alias", `let x = sys.hostname()`, "sys", "sys.Hostname()"},
		{"file uses file alias", `let x = file.read("/etc/hosts")`, "file", `file.Read(opsStr("/etc/hosts"))`},
		{"net uses opsnet alias", `let x = net.interfaces()`, "opsnet", "opsnet.Interfaces()"},
		{"process uses process alias", `let x = process.list()`, "process", "process.List()"},
		{"service uses service alias", `let x = service.status("nginx")`, "service", `service.Status(opsStr("nginx"))`},
		{"pkg uses opspkg alias", `let x = pkg.list()`, "opspkg", "opspkg.List()"},
		{"time uses opstime alias", `let x = time.now()`, "opstime", "opstime.Now()"},
		{"json uses opsjson alias", `let x = json.encode("data")`, "opsjson", `opsjson.Encode("data")`},
		{"yaml uses opsyaml alias", `let x = yaml.encode("data")`, "opsyaml", `opsyaml.Encode("data")`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateCode(tt.source, "test.ops")
			if err != nil {
				t.Fatalf("GenerateCode failed: %v", err)
			}
			if !strings.Contains(code, tt.alias) {
				t.Errorf("expected import alias %q, got:\n%s", tt.alias, code)
			}
			if !strings.Contains(code, tt.sdkCall) {
				t.Errorf("expected SDK call %q, got:\n%s", tt.sdkCall, code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// For loop structure verification
// ---------------------------------------------------------------------------

func TestGenerateForLoopStructure(t *testing.T) {
	source := `for let i = 0; i < 10; i = i + 1 {
  print(i)
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Verify all parts of the C-style for loop are present
	if !strings.Contains(code, "i := interface{}(int64(0))") {
		t.Errorf("expected init statement 'i := int64(0)', got:\n%s", code)
	}
	if !strings.Contains(code, "isTruthy(") {
		t.Errorf("expected condition check, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// User function with default params
// ---------------------------------------------------------------------------

func TestGenerateUserFunctionWithDefaultParam(t *testing.T) {
	source := `fn greet(name, greeting = "hello") {
  print(greeting)
}
greet("world")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "func greet(") {
		t.Errorf("expected function definition, got:\n%s", code)
	}
	if !strings.Contains(code, "interface{}") {
		t.Errorf("expected interface{} parameter types, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// ExpressionStatement
// ---------------------------------------------------------------------------

func TestGenerateExpressionStatement(t *testing.T) {
	source := `sys.hostname()`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// ExpressionStatement should be wrapped with _ = ...
	if !strings.Contains(code, "_ = ") {
		t.Errorf("expected '_ = ' prefix for expression statement, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// Assign to member/index target
// ---------------------------------------------------------------------------

func TestGenerateAssignToMemberTarget(t *testing.T) {
	source := `let d = {"key": "old"}
d.key = "new"`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	// Should generate an assignment to the member expression
	if !strings.Contains(code, `"new"`) {
		t.Errorf("expected assignment value, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// SanitizeName extended cases
// ---------------------------------------------------------------------------

func TestSanitizeNameExtended(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Go reserved words
		{"break", "_break"},
		{"case", "_case"},
		{"chan", "_chan"},
		{"const", "_const"},
		{"continue", "_continue"},
		{"default", "_default"},
		{"defer", "_defer"},
		{"else", "_else"},
		{"fallthrough", "_fallthrough"},
		{"for", "_for"},
		{"func", "_func"},
		{"go", "_go"},
		{"goto", "_goto"},
		{"if", "_if"},
		{"import", "_import"},
		{"interface", "_interface"},
		{"map", "_map"},
		{"package", "_package"},
		{"range", "_range"},
		{"return", "_return"},
		{"select", "_select"},
		{"struct", "_struct"},
		{"switch", "_switch"},
		{"type", "_type"},
		{"var", "_var"},

		// Special characters
		{"my-var", "my_var"},
		{"my.var", "my_var"},
		{"my var", "my_var"},

		// Leading digit (the digit is replaced, not preserved)
		{"1var", "_var"},

		// Normal names unchanged
		{"x", "x"},
		{"myVar", "myVar"},
		{"_private", "_private"},
		{"camelCase", "camelCase"},
		{"with_underscore", "with_underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// genCall fallback: unknown function (not builtin, user func, or SDK)
// ---------------------------------------------------------------------------

func TestGenCallUnknownFunctionFallback(t *testing.T) {
	g := &CodeGenerator{}
	call := &ast.CallExpression{
		Function: &ast.Identifier{Name: "unknownFunc"},
		Args: []ast.Expression{
			&ast.IntegerLiteral{Value: 42},
		},
	}
	_, err := g.genCall(call)
	if err == nil {
		t.Fatal("unknown function must fail generation (the old fallback emitted uncompilable code)")
	}
	if !strings.Contains(err.Error(), "unknown function") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// genCall: user function call
// ---------------------------------------------------------------------------

func TestGenCallUserFunction(t *testing.T) {
	g := &CodeGenerator{
		userFuncs: []userFunc{
			{
				name:   "myFunc",
				params: []ast.Parameter{{Name: &ast.Identifier{Name: "x"}}},
				body:   &ast.BlockStatement{},
			},
		},
	}
	call := &ast.CallExpression{
		Function: &ast.Identifier{Name: "myFunc"},
		Args:     []ast.Expression{&ast.StringLiteral{Value: "hello"}},
	}
	result, err := g.genCall(call)
	if err != nil {
		t.Fatalf("genCall user func failed: %v", err)
	}
	if !strings.Contains(result, `myFunc("hello")`) {
		t.Errorf("expected user function call, got: %s", result)
	}
}

// ---------------------------------------------------------------------------
// Generate with useOS flag (alert triggers os import)
// ---------------------------------------------------------------------------

func TestGenerateAlertTriggersOSImport(t *testing.T) {
	source := `alert("something went wrong")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `"os"`) {
		t.Errorf("expected os import when alert is used, got:\n%s", code)
	}
	if !strings.Contains(code, "_ = os.Stderr") {
		t.Errorf("expected os.Stderr usage marker, got:\n%s", code)
	}
}

// ---------------------------------------------------------------------------
// genExpr: error propagation from nested expressions
// ---------------------------------------------------------------------------

func TestGenExprErrorPropagation(t *testing.T) {
	g := &CodeGenerator{}
	// UnaryExpression wrapping an unsupported type (nil)
	expr := &ast.UnaryExpression{
		Op:    "-",
		Right: nil, // will trigger error in genExpr
	}
	_, err := g.genExpr(expr)
	if err == nil {
		t.Fatal("expected error from nested unsupported type")
	}
}

// ---------------------------------------------------------------------------
// genBinary: error propagation from operands
// ---------------------------------------------------------------------------

func TestGenBinaryErrorPropagation(t *testing.T) {
	g := &CodeGenerator{}
	expr := &ast.BinaryExpression{
		Left:  nil, // unsupported
		Op:    "+",
		Right: &ast.IntegerLiteral{Value: 1},
	}
	_, err := g.genBinary(expr)
	if err == nil {
		t.Fatal("expected error from unsupported left operand")
	}
}

// ---------------------------------------------------------------------------
// genIndex and genMember expressions via GenerateCode
// ---------------------------------------------------------------------------

func TestGenerateIndexExpressionViaParser(t *testing.T) {
	source := `let arr = [10, 20, 30]
let val = arr[1]`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "case []interface{}") {
		t.Errorf("expected index expression type switch, got:\n%s", code)
	}
}

func TestGenerateMemberExpressionViaParser(t *testing.T) {
	source := `let d = {"name": "ops"}
let v = d.name`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `opsToMapDeep(d)["name"]`) {
		t.Errorf("expected member expression lookup, got:\n%s", code)
	}
}

// TestGenerateUserFuncUsesSDKAndGlobals pins two related fixes: SDK calls
// inside user functions must register the package import, and user
// functions must be able to reference top-level lets (hoisted to
// package-level vars, assigned from main in source order). Before this,
// any script wrapping sys.* calls in a function failed AOT compilation
// with "undefined: sys".
func TestGenerateUserFuncUsesSDKAndGlobals(t *testing.T) {
	source := `let threshold = 80.0
fn disk_full(mount) {
	let u = sys.disk.usage(mount)
	return u.used_percent > threshold
}
if disk_full("/") {
	alert("full")
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `"github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sys"`) {
		t.Errorf("SDK sys package missing from imports; fn-body usage was not collected:\n%s", firstLines(code, 20))
	}
	if !strings.Contains(code, "threshold interface{}") {
		t.Errorf("top-level let not hoisted to package var:\n%s", code)
	}
	if !strings.Contains(code, `func disk_full(`) {
		t.Errorf("user function missing:\n%s", code)
	}
	if !strings.Contains(code, "threshold = float64(80)") {
		t.Errorf("global assignment missing from main:\n%s", firstLines(code, 5))
	}
}

func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
