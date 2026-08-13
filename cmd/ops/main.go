// OpsLang - 为运维而生的编程语言
package main

import (
	"fmt"
	"os"

	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/parser"
	"github.com/opslang/opslang/pkg/vm"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: ops run <file.ops>")
			os.Exit(1)
		}
		runFile(os.Args[2])
	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: ops check <file.ops>")
			os.Exit(1)
		}
		checkFile(os.Args[2])
	case "version":
		fmt.Printf("OpsLang %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`OpsLang - 为运维而生的编程语言

用法:
  ops run <file>     运行 OpsLang 脚本
  ops check <file>   语法检查（不执行）
  ops version        显示版本信息
  ops help           显示帮助`)
}

func runFile(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		os.Exit(1)
	}

	// 词法分析
	l := lexer.New(string(source), filename)
	tokens := l.Tokenize()

	// 语法分析
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "语法错误: %v\n", err)
		os.Exit(1)
	}

	// 执行
	machine := vm.New()
	if err := machine.Run(program); err != nil {
		fmt.Fprintf(os.Stderr, "运行错误: %v\n", err)
		os.Exit(1)
	}
}

func checkFile(filename string) {
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(source), filename)
	tokens := l.Tokenize()

	p := parser.New(tokens)
	_, err = p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "语法错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s 语法正确\n", filename)
}
