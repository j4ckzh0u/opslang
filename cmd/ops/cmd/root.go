package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/opslang/opslang/internal/version"
	"github.com/opslang/opslang/pkg/lexer"
	"github.com/opslang/opslang/pkg/parser"
	"github.com/opslang/opslang/pkg/vm"
)

var rootCmd = &cobra.Command{
	Use:   "ops",
	Short: "OpsLang - 为运维而生的编程语言",
	Long: `OpsLang 是一门专为运维场景设计的编程语言。

特性:
  • 类 Python 语法，Shell 般的系统交互
  • 解释执行 + 编译为单二进制
  • 内置 SSH、YAML、JSON、批量执行
  • 声明式资源管理 (ensure)
  • 大规模并行运维引擎 (fleet)`,
}

// runCmd 运行 OpsLang 脚本
var runCmd = &cobra.Command{
	Use:   "run [file]",
	Short: "运行 OpsLang 脚本",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		// 读取源文件
		source, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}

		// 词法分析
		l := lexer.New(string(source), filename)
		tokens := l.Tokenize()

		// 语法分析
		p := parser.New(tokens)
		program, err := p.Parse()
		if err != nil {
			return fmt.Errorf("语法错误: %w", err)
		}

		// 执行
		machine := vm.New()
		return machine.Run(program)
	},
}

// replCmd 启动交互式 REPL
var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "启动交互式 REPL",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Full())
		fmt.Println("输入 :help 查看帮助，:quit 退出")
		fmt.Println()
		// TODO: 实现 REPL 循环
		fmt.Println("REPL 即将上线...")
	},
}

// versionCmd 显示版本信息
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Full())
	},
}

// checkCmd 语法检查（不执行）
var checkCmd = &cobra.Command{
	Use:   "check [file]",
	Short: "语法检查（不执行）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		source, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}

		l := lexer.New(string(source), filename)
		tokens := l.Tokenize()

		p := parser.New(tokens)
		_, err = p.Parse()
		if err != nil {
			return fmt.Errorf("语法错误: %w", err)
		}

		fmt.Printf("✅ %s 语法正确\n", filename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(replCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(checkCmd)
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}
