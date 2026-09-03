# OpsLang 项目 Wiki

本 Wiki 面向使用者、集成开发者和项目贡献者，概括 OpsLang 的系统边界、公开接口、核心概念和主要代码模块。详细用户手册仍位于仓库 `docs/`，原子操作全集以 `internal/opsspec/spec.go` 及其生成索引为准。

快速链接：[系统架构](./ARCHITECTURE.md) | [接口](./INTERFACES.md) | [开发者指南](./DEVELOPER_GUIDE.md)

## 核心概念

| 概念 | 内容 |
|------|------|
| [执行引擎](./专有概念/执行引擎.md) | 解释器、Runner 和 AOT 的职责与选择 |
| [指令包](./专有概念/指令包.md) | Runner JSON 协议、状态和签名 |
| [可恢复中继传输](./专有概念/可恢复中继传输.md) | 部分文件、内容寻址、拓扑和安全边界 |

## 模块

| 模块 | 内容 |
|------|------|
| [语言前端](./模块/语言前端.md) | Lexer、Parser、AST 和 Interpreter |
| [远程执行](./模块/远程执行.md) | SSH、Runner、inventory 和多主机调度 |
| [原子操作 SDK](./模块/原子操作SDK.md) | 标准库组织、opsspec 和引擎一致性 |

## 阅读路径

1. 先读[系统架构](./ARCHITECTURE.md)，理解脚本到执行结果的数据流。
2. 使用 CLI 或 SDK 时读[接口](./INTERFACES.md)。
3. 修改代码前读[开发者指南](./DEVELOPER_GUIDE.md)。
4. 文件传输开发重点读[可恢复中继传输](./专有概念/可恢复中继传输.md)。

## 快速命令

```bash
# 本地解释执行
go run ./cmd/opsctl run examples/helloworld.ops

# 运行全量测试
make test

# 检查接口索引
make docs-check

# 交叉构建
make build-all
```
