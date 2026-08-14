# Changelog

本项目遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/) 规范。

## [Unreleased]

### 计划添加
- 模块系统 (`import` 实际工作)
- 类型注解与静态检查
- WASM 编译支持
- 更多标准库模块（数据库、HTTP 服务器）
- 包注册表 (registry)

## [0.1.0] - 2026-08-14

### 首次发布 🎉

OpsLang v0.1.0 正式发布！这是一门为运维而生的编程语言，具备完整的语言特性和丰富的运维标准库。

#### 核心语言特性
- **词法分析器** — 缩进语法、字符串插值、三引号字符串、Shebang 支持
- **语法分析器** — 递归下降，支持完整语法（变量、函数、循环、条件、闭包）
- **执行引擎** — 树遍历解释器，支持闭包、递归、异常处理
- **编译器** — 将脚本编译为 2.4MB 单二进制，支持交叉编译
- **REPL** — 交互式解释器，支持多行输入
- **LSP 服务器** — 实时诊断、智能补全、悬停信息、跳转定义

#### 标准库（10 个模块）
- `file` — 文件读写、目录操作、路径处理
- `process` — Shell 命令执行、环境变量
- `ssh` — 远程执行、SCP 传输、连通性检测
- `fleet` — 批量并行执行引擎
- `json` — JSON 解析、序列化、文件操作
- `yaml` — YAML 解析、序列化、文件操作
- `toml` — TOML 解析、文件操作
- `strings` — 字符串工具（分割、连接、替换、大小写）
- `math` — 数学函数（abs、min、max）
- `ensure` — 声明式资源管理（幂等）
- `inventory` — 主机清单加载与分组

#### 工具链
- `ops run` — 解释执行脚本
- `ops build` — 编译为单二进制
- `ops repl` — 启动交互式 REPL
- `ops check` — 语法检查
- `ops install` — 包管理器
- `ops list` / `uninstall` / `init` — 包管理命令
- `ops lsp` — 启动语言服务器

#### 开发者体验
- VS Code 插件 — 语法高亮、代码片段、LSP 集成
- 完整文档 — 语言参考、快速入门、示例代码
- 测试套件 — 26 单元测试 + 17 集成测试
- 官网落地页 — 现代深色主题设计
- GitHub Actions CI — 自动测试、lint、多平台构建

#### 统计
- Go 代码：~4500 行
- 测试用例：43 个
- 示例脚本：4 个实战 + 1 个自举演示
- 二进制大小：2.4MB (stripped, static)

[Unreleased]: https://github.com/opslang/opslang/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/opslang/opslang/releases/tag/v0.1.0
