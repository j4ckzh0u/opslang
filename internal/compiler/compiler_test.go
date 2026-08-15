package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseTargetArchEmpty(t *testing.T) {
	goos, goarch, err := parseTargetArch("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goos != runtime.GOOS {
		t.Errorf("expected GOOS=%s, got %s", runtime.GOOS, goos)
	}
	if goarch != runtime.GOARCH {
		t.Errorf("expected GOARCH=%s, got %s", runtime.GOARCH, goarch)
	}
}

func TestParseTargetArchValid(t *testing.T) {
	goos, goarch, err := parseTargetArch("linux/amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goos != "linux" {
		t.Errorf("expected linux, got %s", goos)
	}
	if goarch != "amd64" {
		t.Errorf("expected amd64, got %s", goarch)
	}
}

func TestParseTargetArchInvalid(t *testing.T) {
	_, _, err := parseTargetArch("invalid")
	if err == nil {
		t.Error("expected error for invalid arch format")
	}
}

func TestGenerateCodeSimple(t *testing.T) {
	source := `let x = 42
print(x)`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty generated code")
	}
}

func TestGenerateCodeWithSDK(t *testing.T) {
	source := `let cpu = sys.cpu.usage()
report {
  cpu: cpu
}`
	code, err := GenerateCode(source, "test.ops")
	if err != nil {
		t.Fatalf("GenerateCode failed: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty generated code")
	}
}

func TestGenerateCodeParseError(t *testing.T) {
	source := `let = invalid syntax`
	_, err := GenerateCode(source, "test.ops")
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestCompileEndToEnd(t *testing.T) {
	// Skip if go compiler is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go compiler not available")
	}

	// Find project root by walking up from this test file
	_, testFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(testFile)))

	// Verify we found the right project root (should contain go.mod)
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err != nil {
		t.Skipf("could not find project root from %s", testFile)
	}

	// Create test source file inside the project tree
	exampleDir := filepath.Join(projectRoot, "examples")
	os.MkdirAll(exampleDir, 0755)
	testSource := filepath.Join(exampleDir, "_compiler_test.ops")
	source := `let x = 42
print(x)`
	if err := os.WriteFile(testSource, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}
	defer os.Remove(testSource)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-output")

	c, err := NewCompiler()
	if err != nil {
		t.Fatalf("NewCompiler failed: %v", err)
	}

	err = c.Compile(testSource, "", outputPath)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Verify binary exists
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output binary not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output binary is empty")
	}

	// Run the binary and check output
	cmd := exec.Command(outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run compiled binary: %v\noutput: %s", err, output)
	}
	if len(output) == 0 {
		t.Error("compiled binary produced no output")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "src")
	dst := filepath.Join(tmpDir, "dst")

	content := []byte("test content")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}
	if string(data) != "test content" {
		t.Errorf("expected 'test content', got %q", string(data))
	}
}
