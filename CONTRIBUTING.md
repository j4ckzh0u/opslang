# 贡献指南

感谢你考虑为 OpsLang 做出贡献！本文档将指导你如何参与项目开发。

## 行为准则

本项目遵循 [Contributor Covenant](https://www.contributor-covenant.org/zh-cn/version/2/0/code_of_conduct/) 行为准则。参与即表示你同意遵守该准则。

## 如何贡献

### 报告 Bug

1. 在 [Issues](https://github.com/opslang/opslang/issues) 中搜索是否已有相关报告
2. 如果没有，创建新的 Issue，使用 Bug 报告模板
3. 包含以下信息：
   - 操作系统和版本
   - OpsLang 版本 (`ops version`)
   - 复现步骤
   - 预期行为 vs 实际行为
   - 相关代码片段

### 提交新功能

1. 先在 Issues 中讨论你的想法
2. Fork 仓库并创建功能分支：`git checkout -b feat/my-feature`
3. 编写代码和测试
4. 确保所有测试通过：`make test`
5. 提交 PR 并描述你的更改

### 改进文档

文档改进同样重要！无论是修复拼写错误、添加示例还是重写章节，都欢迎提交 PR。

## 开发环境搭建

### 前置要求

- Go 1.21 或更高版本
- Make
- Git

### 快速开始

```bash
# 克隆仓库
git clone https://github.com/opslang/opslang.git
cd opslang

# 构建
make build

# 运行测试
make test

# 运行示例
./ops run examples/basic/hello.ops

# 启动 REPL
./ops repl
```

### 项目结构

```
opslang/
├── cmd/ops/           # CLI 入口
├── pkg/
│   ├── lexer/         # 词法分析器
│   ├── parser/        # 语法分析器
│   ├── ast/           # 抽象语法树
│   ├── vm/            # 执行引擎
│   ├── compiler/      # 编译器
│   ├── repl/          # REPL
│   ├── lsp/           # 语言服务器
│   └── packagemanager/# 包管理器
├── editors/vscode/    # VS Code 插件
├── docs/              # 文档
├── examples/          # 示例
├── scripts/           # 脚本
├── test/              # 测试
└── website/           # 官网
```

## 代码规范

### Go 代码

- 遵循标准 Go 格式化（`gofmt`）
- 所有公开函数必须有文档注释
- 新代码必须有对应的测试
- 运行 `make lint` 确保代码质量

### OpsLang 代码

- 使用 4 空格缩进
- 变量和函数使用 `snake_case`
- 模块名使用小写
- 添加注释说明复杂逻辑

## 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>: <description>

[optional body]

[optional footer]
```

类型：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更改
- `style`: 格式调整（不影响代码逻辑）
- `refactor`: 重构（非新功能、非修复）
- `test`: 测试相关
- `chore`: 构建/工具链相关

示例：
```
feat: 添加 TOML 解析支持

- 实现 toml.parse 函数
- 添加 toml.load_file 函数
- 添加 15 个测试用例

Closes #42
```

## 测试

### 运行所有测试

```bash
make test
```

### 运行特定测试

```bash
# 单元测试
go test ./pkg/lexer/ -v
go test ./pkg/vm/ -v

# 集成测试
go test ./test/integration/ -v

# LSP 测试
go test ./pkg/lsp/ -v
```

### 添加测试

- 新功能必须附带测试
- 测试覆盖率目标：80%+
- 使用表驱动测试风格
- 测试命名清晰描述场景

## 文档

### 更新文档

- 新功能需更新 `docs/language-reference.md`
- 新命令需更新 `README.md`
- 破坏性变更需更新 `CHANGELOG.md`

### 文档风格

- 使用简体中文
- 代码示例使用英文注释
- 保持术语一致（参考术语表）

## 发布流程

维护者使用以下流程发布新版本：

1. 更新 `CHANGELOG.md`
2. 更新版本号（`cmd/ops/main.go` 和 `internal/version/version.go`）
3. 运行完整测试：`make test`
4. 创建 release 分支：`git checkout -b release/v0.2.0`
5. 提交并推送到 main
6. 创建 GitHub Release 和 Tag
7. 构建并发布二进制文件

## 获取帮助

- 📖 [文档](https://github.com/opslang/opslang/tree/main/docs)
- 💬 [Discussions](https://github.com/opslang/opslang/discussions)
- 🐛 [Issues](https://github.com/opslang/opslang/issues)

## 致谢

感谢所有贡献者！你的每一份贡献都让 OpsLang 变得更好。

---

再次感谢你的贡献！🎉
