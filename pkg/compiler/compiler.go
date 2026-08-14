// Package compiler 实现 OpsLang 编译器
// 将 OpsLang 脚本编译为单二进制文件
package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compile 将 OpsLang 脚本编译为单二进制
func Compile(sourceFile, outputFile string) error {
	// 读取源文件
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("读取源文件失败: %w", err)
	}

	// 生成临时目录
	tmpDir, err := os.MkdirTemp("", "opslang-build-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 确定输出文件名
	if outputFile == "" {
		base := filepath.Base(sourceFile)
		outputFile = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// 生成 Go 源码
	goSource := generateGoSource(string(source), outputFile)
	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(goSource), 0644); err != nil {
		return fmt.Errorf("写入 Go 源码失败: %w", err)
	}

	// 生成 go.mod
	goMod := fmt.Sprintf(`module opslang-binary

go 1.21

require github.com/opslang/opslang v0.0.0

replace github.com/opslang/opslang => %s
`, findProjectRoot())
	modFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modFile, []byte(goMod), 0644); err != nil {
		return fmt.Errorf("写入 go.mod 失败: %w", err)
	}

	// 编译
	absOutput, _ := filepath.Abs(outputFile)
	cmd := exec.Command("go", "build", "-trimpath",
		"-ldflags", "-s -w",
		"-o", absOutput, ".")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("编译失败: %w", err)
	}

	// 输出信息
	info, _ := os.Stat(absOutput)
	sizeMB := float64(info.Size()) / 1024 / 1024
	fmt.Printf("✅ 编译成功: %s (%.1f MB)\n", absOutput, sizeMB)
	return nil
}

// generateGoSource 生成嵌入 OpsLang 脚本的 Go 源码
func generateGoSource(script, name string) string {
	// 转义脚本中的反引号和特殊字符
	escaped := strings.ReplaceAll(script, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "`", "\\`")

	return fmt.Sprintf(`// 自动生成的 OpsLang 二进制
// 源脚本: %s.ops
package main

import (
	"fmt"
	"os"

	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/parser"
	"github.com/opslang/opslang/pkg/vm"
)

const script = ` + "`" + `%s` + "`" + `

func main() {
	l := lexer.New(script, "%s.ops")
	tokens := l.Tokenize()

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "语法错误: %%v\n", err)
		os.Exit(1)
	}

	machine := vm.New()
	if err := machine.Run(program); err != nil {
		fmt.Fprintf(os.Stderr, "运行错误: %%v\n", err)
		os.Exit(1)
	}
}
`, name, escaped, name)
}

// findProjectRoot 查找项目根目录（包含 go.mod 的目录）
func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}
