package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/opslang/opslang/internal/parser"
)

// Compiler encapsulates the AOT compilation pipeline.
type Compiler struct {
	cache *Cache
}

// NewCompiler creates a new Compiler with caching enabled.
func NewCompiler() (*Compiler, error) {
	cache, err := NewCache("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}
	return &Compiler{cache: cache}, nil
}

// Compile takes a source file path and target architecture, and produces a static binary.
// targetArch format: "os/arch" (e.g., "linux/amd64", "linux/arm64").
// If targetArch is empty, uses the current GOOS/GOARCH.
// outputPath is where the binary will be written.
func (c *Compiler) Compile(sourcePath string, targetArch string, outputPath string) error {
	// Read source file
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	sourceStr := string(source)

	// Parse target architecture
	goos, goarch, err := parseTargetArch(targetArch)
	if err != nil {
		return err
	}

	// Check cache
	cacheKey := c.cache.Key(sourceStr, goos+"/"+goarch)
	if cached := c.cache.Get(cacheKey); cached != "" {
		if err := copyFile(cached, outputPath); err != nil {
			return fmt.Errorf("failed to copy cached binary: %w", err)
		}
		return nil
	}

	// Parse source
	p := parser.New(sourceStr, sourcePath)
	prog, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Generate Go code
	gen := &CodeGenerator{}
	goCode, err := gen.Generate(prog)
	if err != nil {
		return fmt.Errorf("code generation error: %w", err)
	}

	// Create build directory inside the project
	// This allows the generated code to use the project's go.mod for SDK imports
	projectRoot, err := findProjectRoot(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}

	buildDir := filepath.Join(projectRoot, ".opslang-build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	defer os.RemoveAll(buildDir)

	// Write generated Go source
	mainFile := filepath.Join(buildDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(goCode), 0644); err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	// Build output path
	buildOutput := filepath.Join(buildDir, "output")

	// Run go build using the project's go.mod
	cmd := exec.Command("go", "build",
		"-ldflags", "-s -w",
		"-o", buildOutput,
		"./.opslang-build/",
	)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// Verify binary was produced
	if _, err := os.Stat(buildOutput); err != nil {
		return fmt.Errorf("compiled binary not found at %s", buildOutput)
	}

	// Cache the result
	if err := c.cache.Put(cacheKey, buildOutput); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to cache binary: %v\n", err)
	}

	// Copy to final output path
	if err := copyFile(buildOutput, outputPath); err != nil {
		return fmt.Errorf("failed to copy binary to output: %w", err)
	}

	return nil
}

// GenerateCode parses the source and returns generated Go code without compiling.
// Useful for testing.
func GenerateCode(source string, filename string) (string, error) {
	p := parser.New(source, filename)
	prog, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	gen := &CodeGenerator{}
	return gen.Generate(prog)
}

// parseTargetArch parses "os/arch" format. If empty, returns current runtime values.
func parseTargetArch(targetArch string) (string, string, error) {
	if targetArch == "" {
		return runtime.GOOS, runtime.GOARCH, nil
	}

	parts := strings.SplitN(targetArch, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid target architecture %q: expected format os/arch (e.g., linux/amd64)", targetArch)
	}

	return parts[0], parts[1], nil
}

// findProjectRoot walks up from the source file to find the go.mod.
func findProjectRoot(sourcePath string) (string, error) {
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(absPath)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fall back to current working directory
	return os.Getwd()
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if dir := filepath.Dir(dst); dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(dst, data, 0755)
}
