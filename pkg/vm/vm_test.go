package vm

import (
	"strings"
	"testing"

	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/parser"
)

// helper: 运行 OpsLang 代码并捕获输出
func runCode(t *testing.T, code string) string {
	t.Helper()

	l := lexer.New(code, "test")
	tokens := l.Tokenize()

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("语法错误: %v", err)
	}

	vm := New()
	vm.output = &strings.Builder{}
	if err := vm.Run(program); err != nil {
		t.Fatalf("运行错误: %v", err)
	}

	return vm.output.String()
}

func TestVariableAssignment(t *testing.T) {
	output := runCode(t, `
x = 42
print(x)
`)
	if strings.TrimSpace(output) != "42" {
		t.Errorf("期望 42，实际 %q", output)
	}
}

func TestStringInterpolation(t *testing.T) {
	output := runCode(t, `
name = "World"
print("Hello {name}")
`)
	if strings.TrimSpace(output) != "Hello World" {
		t.Errorf("期望 Hello World，实际 %q", output)
	}
}

func TestArithmetic(t *testing.T) {
	output := runCode(t, `
print(2 + 3)
print(10 - 4)
print(3 * 7)
print(15 / 3)
print(10 % 3)
`)
	expected := "5\n6\n21\n5\n1\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIfElse(t *testing.T) {
	output := runCode(t, `
x = 10
if x > 5
    print("big")
else
    print("small")
`)
	if strings.TrimSpace(output) != "big" {
		t.Errorf("期望 big，实际 %q", output)
	}
}

func TestForLoop(t *testing.T) {
	output := runCode(t, `
items = [1, 2, 3]
for x in items
    print(x)
`)
	expected := "1\n2\n3\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestWhileLoop(t *testing.T) {
	output := runCode(t, `
i = 0
while i < 3
    print(i)
    i = i + 1
`)
	expected := "0\n1\n2\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestFunction(t *testing.T) {
	output := runCode(t, `
fn add(a, b)
    return a + b
print(add(3, 4))
`)
	if strings.TrimSpace(output) != "7" {
		t.Errorf("期望 7，实际 %q", output)
	}
}

func TestRecursion(t *testing.T) {
	output := runCode(t, `
fn factorial(n)
    if n <= 1
        return 1
    return n * factorial(n - 1)
print(factorial(5))
`)
	if strings.TrimSpace(output) != "120" {
		t.Errorf("期望 120，实际 %q", output)
	}
}

func TestArrayOperations(t *testing.T) {
	output := runCode(t, `
arr = [1, 2, 3]
print(len(arr))
print(arr[0])
print(arr[2])
`)
	expected := "3\n1\n3\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestMapOperations(t *testing.T) {
	output := runCode(t, `
m = {"a": 1, "b": 2}
print(m["a"])
print(m["b"])
`)
	expected := "1\n2\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestStringMethods(t *testing.T) {
	output := runCode(t, `
s = "  Hello  "
print(s.trim().upper())
`)
	if strings.TrimSpace(output) != "HELLO" {
		t.Errorf("期望 HELLO，实际 %q", output)
	}
}

func TestTryCatch(t *testing.T) {
	output := runCode(t, `
try
    x = 10 / 0
catch e
    print("caught")
print("done")
`)
	expected := "caught\ndone\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestBreakContinue(t *testing.T) {
	output := runCode(t, `
for i in range(10)
    if i == 3
        break
    print(i)
`)
	expected := "0\n1\n2\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestClosures(t *testing.T) {
	output := runCode(t, `
fn make_adder(n)
    return fn(x) => x + n
add5 = make_adder(5)
print(add5(10))
`)
	if strings.TrimSpace(output) != "15" {
		t.Errorf("期望 15，实际 %q", output)
	}
}

func TestLambda(t *testing.T) {
	output := runCode(t, `
double = fn(x) => x * 2
print(double(7))
`)
	if strings.TrimSpace(output) != "14" {
		t.Errorf("期望 14，实际 %q", output)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	output := runCode(t, `
print(len("hello"))
print(type(42))
print(str(3.14))
print(int("42"))
`)
	expected := "5\nint\n3.14\n42\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestTripleQuoteString(t *testing.T) {
	output := runCode(t, `
s = """hello world"""
print(s)
`)
	if strings.TrimSpace(output) != "hello world" {
		t.Errorf("期望 hello world，实际 %q", output)
	}
}

func TestJSONModule(t *testing.T) {
	output := runCode(t, `
data = json.parse('{"name": "test"}')
print(data["name"])
`)
	if strings.TrimSpace(output) != "test" {
		t.Errorf("期望 test，实际 %q", output)
	}
}

func TestForLoopVariableScope(t *testing.T) {
	output := runCode(t, `
items = ["a", "b", "c"]
for x in items
    print(x)
// x 不应泄漏到外层
print("done")
`)
	expected := "a\nb\nc\ndone\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}
