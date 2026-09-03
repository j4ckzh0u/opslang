# OpsLang 开发者指南

## 环境要求

- Go 1.26，与 `go.mod` 和 GitHub Actions 保持一致。
- Git。
- 涉及真实远程执行时，需要可访问的 SSH 测试目标。
- 核心构建强制 `CGO_ENABLED=0`，确保静态交叉编译。

仓库没有数据库或常驻服务依赖。Go module 下载由标准 Go 工具链处理。

## 常用命令

```bash
# 运行全部测试和 race detector
make test

# 运行静态检查
make vet

# 检查生成文档是否最新
make docs-check

# 构建当前平台二进制
make build

# 构建 Linux 和 Darwin 的 amd64 与 arm64 二进制
make build-all

# 运行 10000 主机文件传输模拟
make scale-test
```

`make fmt` 会检查整个仓库的 Go 格式。修改代码时应对本次涉及的 Go 文件运行 `gofmt -w`。

## 本地运行

```bash
# 解释执行示例
go run ./cmd/opsctl run examples/helloworld.ops

# 启动 REPL
go run ./cmd/opsctl repl

# 查看命令帮助
go run ./cmd/opsctl --help
```

## 开发工作流

1. 从 `internal/opsspec/spec.go` 确认操作名称、参数和执行范围。
2. 找到对应 SDK 包及解释器、Runner、AOT 的接入位置。
3. 先写能暴露行为差异的测试，包含空输入、越界或取消等边界。
4. 实现最小改动，并对本次文件执行 `gofmt`。
5. 运行目标包测试，再运行全量测试、vet、文档检查和交叉构建。
6. 推送后跟踪 GitHub Actions 的测试分片和四平台构建。

## 添加原子操作

涉及的契约层通常包括：

| 位置 | 作用 |
|------|------|
| `pkg/ops-core-sdk/<name>/` | 强类型业务实现和单元测试 |
| `internal/opsspec/spec.go` | 名称、参数、变更属性和作用域 |
| `internal/interpreter/sdk_bridge.go` 或模块注册 | 解释器调用 |
| `internal/runner/registry.go` | JSON 指令执行 |
| `internal/compiler/codegen.go` | AOT 调用生成 |
| `docs/generated/ops-index.md` | 由 docgen 自动生成的索引 |

完成后运行 `make docs` 更新索引，再运行 `make docs-check` 检查结果。

## 修改语言语法

语法改动从 lexer token 和 AST 类型开始，再更新 parser、interpreter 与 compiler。错误必须保留行列位置。控制流和表达式语义需要同时覆盖解释执行和 AOT 测试。

## 修改远程执行

SSH 代码位于 `internal/sshx`，多主机调度位于 `internal/exec`。所有网络操作需要明确超时、上下文取消和资源关闭。主机密钥策略与认证错误必须保留上下文，避免将密码、私钥或短时令牌写入日志。

文件传输改动还要验证：

- 最终文件原子性和 SHA-256 完整性。
- 部分元数据损坏、源变化、偏移超界和空文件。
- 中继令牌、TLS 指纹、路径、方法、过期和并发限制。
- 候选切换、直接回退、每目标唯一结果和流量上界。
- 1000 主机 CI 层与 10000 主机完整规模层。

## 测试组织

测试与源码同包放置，主要使用表驱动子测试。真实 SSH/SFTP 测试通过仓库内受控服务器验证协议行为；规模测试只替换传输边界，保留真实编排、并发、重试和结果聚合代码。

测试必须断言结果。涉及列表和输入解析时至少覆盖空值、错误类型或越界。使用包级可替换函数的测试不能并行运行，并需通过 `t.Cleanup` 恢复原值。

## CI

`.github/workflows/ci.yml` 在 push 和 pull request 到 `main` 时运行：

- Core 测试分片与生成文档检查。
- SDK 测试分片。
- Linux/Darwin、amd64/arm64 的 `opsctl` 与 `ops-runner` 静态构建。
- 最终 gate 汇总测试与构建结果。

CI 对已知 Go 1.26 runtime mmap 波动使用有界重试脚本。业务断言失败会在所有尝试中持续失败并阻止构建 gate。

## 关键文档

- `README.md`：项目入口和能力状态。
- `docs/architecture.md`：完整架构说明。
- `docs/cli-reference.md`：命令参数。
- `docs/stdlib-reference.md`：标准库调用说明。
- `docs/generated/ops-index.md`：opsspec 生成索引。
- `.monkeycode/specs/`：功能需求、设计和任务清单。
