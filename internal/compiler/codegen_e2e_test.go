package compiler

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests compile generated code with the real `go build` and execute
// the binaries. String-contains assertions on generated source used to pass
// while the code did not even compile (e.g. `i := int64(0)` later assigned
// an interface{}); these tests cannot lie.

// compileAndRun compiles source via the full AOT pipeline and executes the
// binary, returning its stdout and exit error.
func compileAndRun(t *testing.T, source string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	script := filepath.Join(dir, "script.ops")
	if err := os.WriteFile(script, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "script.bin")

	c, err := NewCompiler()
	if err != nil {
		t.Fatalf("NewCompiler: %v", err)
	}
	if err := c.Compile(script, "", bin); err != nil {
		return "", err
	}

	cmd := exec.Command(bin)
	out, err := cmd.Output()
	return string(out), err
}

// runAndReportJSON compiles, runs, and decodes the binary's JSON report.
func runAndReportJSON(t *testing.T, source string) map[string]interface{} {
	t.Helper()
	out, err := compileAndRun(t, source)
	if err != nil {
		t.Fatalf("compile/run failed: %v\nsource:\n%s", err, source)
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		t.Fatalf("output is not JSON: %v\noutput: %q", err, out)
	}
	return data
}

func TestAOTForLoopAndAssignment(t *testing.T) {
	data := runAndReportJSON(t, `
let total = 0
for let i = 0; i < 5; i = i + 1 {
	total = total + i
}
report { total: total }
`)
	if data["total"] != float64(10) {
		t.Errorf("total = %v, want 10", data["total"])
	}
}

func TestAOTIfElseAndUserFunction(t *testing.T) {
	data := runAndReportJSON(t, `
fn double(n) {
	return n * 2
}
let v = double(21)
if v == 42 {
	report { answer: v, correct: true }
} else {
	report { answer: v, correct: false }
}
`)
	if data["answer"] != float64(42) {
		t.Errorf("answer = %v, want 42", data["answer"])
	}
	if data["correct"] != true {
		t.Errorf("correct = %v, want true", data["correct"])
	}
}

func TestAOTWhileLoop(t *testing.T) {
	data := runAndReportJSON(t, `
let n = 1
while n < 100 {
	n = n * 2
}
report { n: n }
`)
	if data["n"] != float64(128) {
		t.Errorf("n = %v, want 128", data["n"])
	}
}

func TestAOTEnsureIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ensured-dir")

	source := `ensure file.exists("` + target + `").exists {
	file.mkdir("` + target + `")
}
report { ok: file.exists("` + target + `").exists }
`

	// First run applies, second run verifies idempotency; both must exit 0.
	for i := 0; i < 2; i++ {
		out, err := compileAndRun(t, source)
		if err != nil {
			t.Fatalf("ensure run %d failed: %v\noutput: %s", i+1, err, out)
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(out), &data); err != nil {
			t.Fatalf("run %d output not JSON: %v (%q)", i+1, err, out)
		}
		if data["ok"] != true {
			t.Errorf("run %d: ensure did not report ok=true: %v", i+1, data)
		}
	}
}

func TestAOTEnsureVerifyFailureExitsNonZero(t *testing.T) {
	// The condition can never become true: ensure must fail the binary.
	_, err := compileAndRun(t, `
ensure file.exists("/nonexistent/path/that/will/not/exist").exists {
	print("applying uselessly")
}
`)
	if err == nil {
		t.Fatal("ensure with an unsatisfiable condition must exit non-zero")
	}
}

func TestAOTParallelBlock(t *testing.T) {
	data := runAndReportJSON(t, `
let a = 0
let b = 0
parallel {
	let a = 1
	let b = 2
}
report { a: a, b: b }
`)
	if data["a"] != float64(1) || data["b"] != float64(2) {
		t.Errorf("parallel merge failed: a=%v b=%v", data["a"], data["b"])
	}
}

func TestAOTTypedSDKArgs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "typed.txt")

	data := runAndReportJSON(t, `
file.write("`+path+`", "hello")
file.chmod("`+path+`", "0600")
let info = file.stat("`+path+`")
report { size: info.size, mode: info.mode }
`)
	if data["size"] != float64(5) {
		t.Errorf("size = %v, want 5", data["size"])
	}

	// The chmod must really have applied (mode field from os.Stat).
	mode, _ := data["mode"].(string)
	if mode == "" {
		// mode may serialize as number depending on FileInfo; just make
		// sure the key exists.
		t.Errorf("stat result missing mode: %v", data)
	}

	// Verify on disk with the real filesystem.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file mode = %v, want 0600 (file.chmod did not apply)", info.Mode().Perm())
	}
}

func TestAOTTemplateRender(t *testing.T) {
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "app.conf")
	if err := os.WriteFile(tmpl, []byte("host={{host}}\nport={{port}}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	data := runAndReportJSON(t, `
let rendered = file.template("`+tmpl+`", {"host": "web-01", "port": 8080})
report { content: rendered.content }
`)
	if data["content"] != "host=web-01\nport=8080\n" {
		t.Errorf("content = %q", data["content"])
	}
}

func TestAOTSDKErrorAbortsNonZero(t *testing.T) {
	// file.read on a missing file must abort the binary. It used to become
	// the string "error: ..." and flow onward with exit code 0.
	_, err := compileAndRun(t, `
let data = file.read("/nonexistent/file/definitely/missing")
print("should never get here")
report { data: data }
`)
	if err == nil {
		t.Fatal("SDK failure must exit non-zero")
	}
}

func TestAOTUnknownFunctionFailsGeneration(t *testing.T) {
	_, err := GenerateCode(`sys.not_a_real_function()`, "test.ops")
	if err == nil {
		t.Error("unknown function must fail code generation")
	}
	if !strings.Contains(err.Error(), "unknown function") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAOTImportGoFails(t *testing.T) {
	_, err := GenerateCode(`import go "github.com/some/thirdparty"`, "test.ops")
	if err == nil {
		t.Error("third-party Go import must fail code generation")
	}
}

func TestAOTEqualitySemantics(t *testing.T) {
	data := runAndReportJSON(t, `
let a = 1
let b = 1.0
let c = "1"
report { num_eq: a == b, num_str: a == c, str_eq: "x" == "x" }
`)
	// 1 == 1.0 numerically true; 1 == "1" cross-type false; strings equal.
	if data["num_eq"] != true {
		t.Errorf("1 == 1.0 should be true, got %v", data["num_eq"])
	}
	if data["num_str"] != false {
		t.Errorf(`1 == "1" should be false, got %v`, data["num_str"])
	}
	if data["str_eq"] != true {
		t.Errorf(`"x" == "x" should be true, got %v`, data["str_eq"])
	}
}

// SDK calls return TYPED slices ([]ProcessInfo); len() and indexing used
// to silently yield 0/nil because the generated type switch only matched
// []interface{}. Found by real remote testing (top-process scenario).
func TestAOTTypedSliceLenAndIndex(t *testing.T) {
	data := runAndReportJSON(t, `
let procs = process.list()
let first = ""
let count = len(procs)
if count > 0 {
    first = procs[0].name
}
report { count: count, first: first }
`)
	if data["count"].(float64) < 10 {
		t.Errorf("count = %v, want a real process count (>10)", data["count"])
	}
	if data["first"] == "" {
		t.Error("indexing a typed SDK slice returned empty")
	}
}
