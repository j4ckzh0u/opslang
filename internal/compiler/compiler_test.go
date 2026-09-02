package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

func TestCompileConcurrentUsesIsolatedBuildDirectories(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go compiler not available")
	}
	_, testFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(testFile)))
	scriptPath := filepath.Join(projectRoot, "examples", "_compiler_concurrent_test.ops")
	if err := os.WriteFile(scriptPath, []byte("report { value: 42 }\n"), 0644); err != nil {
		t.Fatalf("write concurrent source: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(scriptPath) })

	c, err := NewCompiler()
	if err != nil {
		t.Fatalf("NewCompiler: %v", err)
	}
	tmp := t.TempDir()
	outputs := []string{filepath.Join(tmp, "one"), filepath.Join(tmp, "two")}
	errs := make(chan error, len(outputs))
	var wg sync.WaitGroup
	for _, output := range outputs {
		wg.Add(1)
		go func(output string) {
			defer wg.Done()
			errs <- c.Compile(scriptPath, "", output)
		}(output)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent compile failed: %v", err)
		}
	}
	for _, output := range outputs {
		if info, err := os.Stat(output); err != nil || info.Size() == 0 {
			t.Fatalf("compiled output %s invalid: info=%v err=%v", output, info, err)
		}
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

// ---------------------------------------------------------------------------
// findProjectRoot
// ---------------------------------------------------------------------------

func TestFindProjectRoot(t *testing.T) {
	// Use this test file itself to find the project root
	_, testFile, _, _ := runtime.Caller(0)
	root, err := findProjectRoot(testFile)
	if err != nil {
		t.Fatalf("findProjectRoot failed: %v", err)
	}
	// The project root should contain go.mod
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Errorf("expected go.mod in project root %s", root)
	}
}

func TestFindProjectRootFromFileInSubdir(t *testing.T) {
	// Create a temp directory structure with go.mod
	tmpDir := t.TempDir()
	goMod := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create nested directory
	subDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create a file in the nested directory
	testFile := filepath.Join(subDir, "test.ops")
	if err := os.WriteFile(testFile, []byte("let x = 1"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	root, err := findProjectRoot(testFile)
	if err != nil {
		t.Fatalf("findProjectRoot failed: %v", err)
	}
	if root != tmpDir {
		t.Errorf("expected root %s, got %s", tmpDir, root)
	}
}

func TestFindProjectRootFallbackToCwd(t *testing.T) {
	// Create a temp file outside any go.mod hierarchy
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "isolated.ops")
	if err := os.WriteFile(testFile, []byte("let x = 1"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	root, err := findProjectRoot(testFile)
	if err != nil {
		t.Fatalf("findProjectRoot failed: %v", err)
	}
	// Should fall back to cwd (which contains go.mod for this project)
	if root == "" {
		t.Error("expected non-empty fallback root")
	}
}

// ---------------------------------------------------------------------------
// Compile error paths
// ---------------------------------------------------------------------------

func TestCompileMissingSourceFile(t *testing.T) {
	c, err := NewCompiler()
	if err != nil {
		t.Fatalf("NewCompiler failed: %v", err)
	}
	err = c.Compile("/nonexistent/path/test.ops", "", "/tmp/output")
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
	if !strings.Contains(err.Error(), "failed to read source file") {
		t.Errorf("expected 'failed to read source file' error, got: %v", err)
	}
}

func TestCompileInvalidTargetArch(t *testing.T) {
	// Create a temporary source file
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "test.ops")
	if err := os.WriteFile(src, []byte("let x = 1"), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	c, err := NewCompiler()
	if err != nil {
		t.Fatalf("NewCompiler failed: %v", err)
	}
	err = c.Compile(src, "invalid-arch", filepath.Join(tmpDir, "output"))
	if err == nil {
		t.Fatal("expected error for invalid target arch")
	}
	if !strings.Contains(err.Error(), "invalid target architecture") {
		t.Errorf("expected 'invalid target architecture' error, got: %v", err)
	}
}

func TestCompileParseError(t *testing.T) {
	// Create a source file with syntax error
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "bad.ops")
	if err := os.WriteFile(src, []byte("let = invalid"), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	c, err := NewCompiler()
	if err != nil {
		t.Fatalf("NewCompiler failed: %v", err)
	}
	err = c.Compile(src, "", filepath.Join(tmpDir, "output"))
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("expected 'parse error', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// copyFile edge cases
// ---------------------------------------------------------------------------

func TestCopyFileCreatesDestDir(t *testing.T) {
	tmpDir := t.TempDir()

	src := filepath.Join(tmpDir, "src")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	// Destination in a non-existent subdirectory
	dst := filepath.Join(tmpDir, "a", "b", "dst")
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("expected 'content', got %q", string(data))
	}
}

func TestCopyFileMissingSource(t *testing.T) {
	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "dst")
	err := copyFile("/nonexistent/file", dst)
	if err == nil {
		t.Fatal("expected error for missing source file")
	}
}

// ---------------------------------------------------------------------------
// Compile cache-hit path
// ---------------------------------------------------------------------------

func TestCompileCacheHit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	source := `let x = 42`
	src := filepath.Join(tmpDir, "test.ops")
	if err := os.WriteFile(src, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	// Create a cache and pre-populate it
	cacheDir := filepath.Join(tmpDir, "cache")
	cache, err := NewCache(cacheDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Compute the cache key that Compile will use
	goos, goarch, _ := parseTargetArch("")
	key := cache.Key(source, goos+"/"+goarch)

	// Create a fake cached binary
	fakeBinary := filepath.Join(tmpDir, "fake-binary")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho cached"), 0755); err != nil {
		t.Fatalf("failed to write fake binary: %v", err)
	}
	if err := cache.Put(key, fakeBinary); err != nil {
		t.Fatalf("cache.Put failed: %v", err)
	}

	// Create compiler with the pre-populated cache
	c := &Compiler{cache: cache}

	outputPath := filepath.Join(tmpDir, "output")
	err = c.Compile(src, "", outputPath)
	if err != nil {
		t.Fatalf("Compile with cache hit failed: %v", err)
	}

	// Verify the output was copied from cache
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if string(data) != "#!/bin/sh\necho cached" {
		t.Errorf("expected cached content, got: %q", string(data))
	}
}
