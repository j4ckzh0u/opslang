// OpsLang - 为运维而生的编程语言
package main

import (
	"fmt"
	"os"

	"github.com/opslang/opslang/pkg/compiler"
	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/packagemanager"
	"github.com/opslang/opslang/pkg/parser"
	"github.com/opslang/opslang/pkg/repl"
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
	case "build":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: ops build <file.ops> [output]")
			os.Exit(1)
		}
		output := ""
		if len(os.Args) >= 4 {
			output = os.Args[3]
		}
		buildFile(os.Args[2], output)
	case "repl":
		repl.Start()
	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: ops check <file.ops>")
			os.Exit(1)
		}
		checkFile(os.Args[2])
	case "version":
		fmt.Printf("OpsLang %s\n", version)
	case "install":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: ops install <user/repo> 或 ops install ./local-path")
			os.Exit(1)
		}
		installPackage(os.Args[2])
	case "list":
		listPackages()
	case "uninstall":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "用法: ops uninstall <package>")
			os.Exit(1)
		}
		uninstallPackage(os.Args[2])
	case "init":
		name := "my-package"
		if len(os.Args) >= 3 {
			name = os.Args[2]
		}
		initPackage(name)
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
  ops run <file>          运行 OpsLang 脚本
  ops build <file> [out]  编译为单二进制
  ops repl                启动交互式 REPL
  ops check <file>        语法检查（不执行）

包管理:
  ops install <user/repo> 从 GitHub 安装包
  ops install <./path>    从本地路径安装
  ops list                列出已安装的包
  ops uninstall <name>    卸载包
  ops init [name]         初始化新包

其他:
  ops version             显示版本信息
  ops help                显示帮助`)
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

func buildFile(filename, output string) {
	if err := compiler.Compile(filename, output); err != nil {
		fmt.Fprintf(os.Stderr, "编译失败: %v\n", err)
		os.Exit(1)
	}
}

func installPackage(name string) {
	mgr, err := packagemanager.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化包管理器失败: %v\n", err)
		os.Exit(1)
	}

	// 判断是本地路径还是远程包
	if name[0] == '.' || name[0] == '/' {
		if err := mgr.InstallLocal(name); err != nil {
			fmt.Fprintf(os.Stderr, "安装失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := mgr.Install(name); err != nil {
			fmt.Fprintf(os.Stderr, "安装失败: %v\n", err)
			os.Exit(1)
		}
	}
}

func listPackages() {
	mgr, err := packagemanager.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化包管理器失败: %v\n", err)
		os.Exit(1)
	}

	packages, err := mgr.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "列出包失败: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("没有已安装的包")
		return
	}

	fmt.Println("已安装的包:")
	for _, pkg := range packages {
		fmt.Printf("  %s v%s — %s\n", pkg.Name, pkg.Version, pkg.Description)
	}
}

func uninstallPackage(name string) {
	mgr, err := packagemanager.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化包管理器失败: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.Uninstall(name); err != nil {
		fmt.Fprintf(os.Stderr, "卸载失败: %v\n", err)
		os.Exit(1)
	}
}

func initPackage(name string) {
	mgr, err := packagemanager.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化包管理器失败: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.InitPackage(name, "0.1.0", "A OpsLang package"); err != nil {
		fmt.Fprintf(os.Stderr, "初始化包失败: %v\n", err)
		os.Exit(1)
	}
}
