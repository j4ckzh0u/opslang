module github.com/opslang/opslang

go 1.21

// 核心依赖（MVP 阶段保持最小化）
require (
	// YAML 支持
	gopkg.in/yaml.v3 v3.0.1

	// SSH 客户端
	golang.org/x/crypto v0.21.0

	// TOML 支持
	github.com/BurntSushi/toml v1.3.2

	// 彩色终端输出
	github.com/fatih/color v1.16.0

	// 命令行解析
	github.com/spf13/cobra v1.8.0

	// 进度条
	github.com/schollz/progressbar/v3 v3.14.2
)
