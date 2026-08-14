package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var opsBin string

func TestMain(m *testing.M) {
	// 构建 ops 二进制
	tmpDir, _ := os.MkdirTemp("", "opslang-test-*")
	opsBin = filepath.Join(tmpDir, "ops")

	cmd := exec.Command("go", "build", "-o", opsBin, "../../cmd/ops")
	if err := cmd.Run(); err != nil {
		panic("构建 ops 失败: " + err.Error())
	}

	code := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

func runOps(t *testing.T, code string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test-*.ops")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(code); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cmd := exec.Command(opsBin, "run", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("运行失败: %v\n输出: %s", err, output)
	}

	return string(output)
}

func TestIntegrationBasic(t *testing.T) {
	output := runOps(t, `
print("hello world")
`)
	if strings.TrimSpace(output) != "hello world" {
		t.Errorf("期望 'hello world'，实际 %q", output)
	}
}

func TestIntegrationVariables(t *testing.T) {
	output := runOps(t, `
x = 42
name = "OpsLang"
print("{name}: {x}")
`)
	if strings.TrimSpace(output) != "OpsLang: 42" {
		t.Errorf("期望 'OpsLang: 42'，实际 %q", output)
	}
}

func TestIntegrationFunction(t *testing.T) {
	output := runOps(t, `
fn factorial(n)
    if n <= 1
        return 1
    return n * factorial(n - 1)
print(factorial(6))
`)
	if strings.TrimSpace(output) != "720" {
		t.Errorf("期望 720，实际 %q", output)
	}
}

func TestIntegrationLoop(t *testing.T) {
	output := runOps(t, `
total = 0
for i in range(1, 11)
    total = total + i
print(total)
`)
	if strings.TrimSpace(output) != "55" {
		t.Errorf("期望 55，实际 %q", output)
	}
}

func TestIntegrationArray(t *testing.T) {
	output := runOps(t, `
hosts = ["web01", "web02", "web03"]
for h in hosts
    print(h)
`)
	expected := "web01\nweb02\nweb03\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationMap(t *testing.T) {
	output := runOps(t, `
config = {"host": "localhost", "port": 8080}
print(config["host"])
print(config["port"])
`)
	expected := "localhost\n8080\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationStringMethods(t *testing.T) {
	output := runOps(t, `
s = "  Hello World  "
print(s.trim().upper())
print(s.trim().lower().len())
`)
	expected := "HELLO WORLD\n11\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationTryCatch(t *testing.T) {
	output := runOps(t, `
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

func TestIntegrationJSON(t *testing.T) {
	output := runOps(t, `
data = json.parse('{"name": "test", "value": 42}')
print(data["name"])
print(data["value"])
`)
	expected := "test\n42\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationYAML(t *testing.T) {
	output := runOps(t, `
data = yaml.parse("name: test\nvalue: 42")
print(data["name"])
print(data["value"])
`)
	expected := "test\n42\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationTripleQuote(t *testing.T) {
	output := runOps(t, `
config = """
name: test
value: 42
"""
data = yaml.parse(config.trim())
print(data["name"])
`)
	if strings.TrimSpace(output) != "test" {
		t.Errorf("期望 'test'，实际 %q", output)
	}
}

func TestIntegrationShebang(t *testing.T) {
	output := runOps(t, `#!/usr/bin/env ops run
print("shebang works")
`)
	if strings.TrimSpace(output) != "shebang works" {
		t.Errorf("期望 'shebang works'，实际 %q", output)
	}
}

func TestIntegrationEnsure(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "ensure-test-*")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	output := runOps(t, `
result = ensure.file("`+tmpFile.Name()+`", "test content")
print(result["changed"])
result2 = ensure.file("`+tmpFile.Name()+`", "test content")
print(result2["changed"])
`)
	expected := "true\nfalse\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}

	// 验证文件内容
	content, _ := os.ReadFile(tmpFile.Name())
	if string(content) != "test content" {
		t.Errorf("文件内容错误: %q", content)
	}
}

func TestIntegrationBuild(t *testing.T) {
	// 创建测试脚本
	tmpScript, _ := os.CreateTemp("", "build-test-*.ops")
	defer os.Remove(tmpScript.Name())
	tmpScript.WriteString(`print("compiled!")`)
	tmpScript.Close()

	// 编译
	tmpBin := tmpScript.Name() + ".bin"
	defer os.Remove(tmpBin)

	cmd := exec.Command(opsBin, "build", tmpScript.Name(), tmpBin)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("编译失败: %v\n%s", err, out)
	}

	// 运行编译后的二进制
	cmd = exec.Command(tmpBin)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("运行编译后二进制失败: %v", err)
	}

	if strings.TrimSpace(string(output)) != "compiled!" {
		t.Errorf("编译后输出错误: %q", output)
	}
}

func TestIntegrationClosures(t *testing.T) {
	output := runOps(t, `
fn make_adder(n)
    return fn(x) => x + n
add5 = make_adder(5)
add10 = make_adder(10)
print(add5(3))
print(add10(3))
`)
	expected := "8\n13\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationForScope(t *testing.T) {
	output := runOps(t, `
items = ["a", "b", "c"]
for x in items
    print(x)
print("done")
`)
	expected := "a\nb\nc\ndone\n"
	if output != expected {
		t.Errorf("期望 %q，实际 %q", expected, output)
	}
}

func TestIntegrationCheck(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "check-test-*.ops")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(`x = 42
print(x)`)
	tmpFile.Close()

	cmd := exec.Command(opsBin, "check", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check 失败: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "语法正确") {
		t.Errorf("check 输出异常: %q", output)
	}
}
