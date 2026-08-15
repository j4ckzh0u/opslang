package compiler

import (
	"strings"
	"testing"
)

func TestGenerateLetStatement(t *testing.T) {
	source := `let x = 42`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "x := int64(42)") {
		t.Errorf("expected 'x := int64(42)', got:\n%s", code)
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
	if !strings.Contains(code, "flag := true") {
		t.Errorf("expected 'flag := true', got:\n%s", code)
	}
}

func TestGenerateNilLiteral(t *testing.T) {
	source := `let x = nil`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "x := nil") {
		t.Errorf("expected 'x := nil', got:\n%s", code)
	}
}

func TestGenerateBinaryExpression(t *testing.T) {
	source := `let x = 1 + 2`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, "x :=") {
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
	if !strings.Contains(code, `sys.GetDiskUsage("/")`) {
		t.Errorf("expected sys.GetDiskUsage(\"/\"), got:\n%s", code)
	}
}

func TestGenerateNetSDKImport(t *testing.T) {
	source := `let r = net.http.get("http://example.com")`
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
	if !strings.Contains(code, `_output["host"]`) {
		t.Errorf("expected report output for host, got:\n%s", code)
	}
	if !strings.Contains(code, `_output["cpu"]`) {
		t.Errorf("expected report output for cpu, got:\n%s", code)
	}
}

func TestGenerateAlertStatement(t *testing.T) {
	source := `alert("high cpu")`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if !strings.Contains(code, `_output["alert"]`) {
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
	if !strings.Contains(code, "x := int64(1)") {
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
