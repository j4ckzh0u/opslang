// Package repl 实现 OpsLang 的交互式解释器
package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/parser"
	"github.com/opslang/opslang/pkg/vm"
)

// Start 启动 REPL
func Start() {
	fmt.Println("OpsLang 0.1.0 — 交互式解释器")
	fmt.Println("输入 :help 查看帮助，:quit 退出")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	machine := vm.New()

	for {
		// 读取一行
		fmt.Print("ops>>> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		// 处理特殊命令（在任何解析之前）
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if handleCommand(trimmed) {
			continue
		}

		// 收集多行输入（如果检测到块结构）
		input := collectInput(scanner, line)

		// 执行
		executeREPL(input, machine)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "REPL 错误: %v\n", err)
	}
	fmt.Println()
}

// collectInput 收集多行输入
func collectInput(scanner *bufio.Scanner, firstLine string) string {
	lines := []string{firstLine}
	indentLevel := 0

	// 检查第一行是否开启块
	if opensBlock(firstLine) {
		indentLevel++
	}

	for indentLevel > 0 {
		fmt.Print("...    ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		lines = append(lines, line)

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// 计算缩进变化
		indent := countIndent(line)
		if indent > 0 && opensBlock(trimmed) {
			indentLevel++
		}
		// 当缩进回到 0 或遇到非缩进行，块结束
		if indent == 0 && trimmed != "" {
			indentLevel--
		}
	}

	return strings.Join(lines, "\n")
}

// opensBlock 判断一行是否开启代码块
func opensBlock(line string) bool {
	trimmed := strings.TrimSpace(line)

	// 以 { 结尾 → 块开始
	if strings.HasSuffix(trimmed, "{") {
		return true
	}

	// 关键字开头且没有 { → 隐式块（缩进风格）
	blockKeywords := []string{"for ", "while ", "if ", "else", "fn ", "try", "catch "}
	for _, kw := range blockKeywords {
		if strings.HasPrefix(trimmed, kw) {
			return true
		}
	}

	return false
}

// countIndent 计算行首缩进空格数
func countIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

// handleCommand 处理特殊命令，返回是否已处理
func handleCommand(line string) bool {
	switch line {
	case ":quit", ":q", ":exit":
		fmt.Println("再见！")
		os.Exit(0)
		return true

	case ":help", ":h":
		printHelp()
		return true

	case ":clear":
		fmt.Print("\033[H\033[2J")
		return true
	}

	return false
}

// executeREPL 执行 REPL 输入
func executeREPL(input string, machine *vm.VM) {
	// 词法分析
	l := lexer.New(input, "<repl>")
	tokens := l.Tokenize()

	// 语法分析
	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		fmt.Printf("语法错误: %v\n", err)
		return
	}

	// 执行
	if err := machine.Run(program); err != nil {
		fmt.Printf("运行错误: %v\n", err)
		return
	}
}

// printHelp 打印帮助
func printHelp() {
	fmt.Print(`
OpsLang REPL 帮助:
  :help, :h      显示帮助
  :quit, :q      退出
  :clear         清屏

示例:
  ops>>> name = "OpsLang"
  ops>>> print("Hello, {name}!")
  Hello, OpsLang!

  ops>>> hosts = ["web01", "web02"]
  ops>>> for h in hosts
  ...        print(h)
  web01
  web02

  ops>>> fn factorial(n)
  ...        if n <= 1
  ...            return 1
  ...        return n * factorial(n - 1)
  ops>>> factorial(5)
  120
`)
}
