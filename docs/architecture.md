# OpsLang 系统架构文档

## 目录

1. [系统架构概览](#1-系统架构概览)
2. [架构图](#2-架构图)
3. [模块说明](#3-模块说明)
4. [执行模式](#4-执行模式)
5. [数据流](#5-数据流)
6. [扩展指南](#6-扩展指南)

---

## 1. 系统架构概览

OpsLang 是一门面向运维领域的领域特定语言（DSL），旨在通过简洁的脚本完成复杂的运维操作。系统采用纯 Go 实现，支持零依赖远程执行、异构架构、结构化返回等核心特性。

### 1.1 核心设计原则

- **纯 Go 生态**：所有依赖支持 `CGO_ENABLED=0` 交叉编译，无 cgo 依赖
- **结构化返回**：标准库函数一律返回结构体，禁止返回原始字符串
- **无 Shell 依赖**：所有操作直接使用 Go 库或读取 `/proc`/`/sys`，不依赖 shell
- **极简语法**：关键字 ≤ 15 个，类 Go + Python 混合风格
- **双执行引擎**：通用 Runner（解释执行）+ AOT 编译（静态二进制）
- **零依赖远程执行**：通过 SSH 下发预编译 Runner 或定制二进制
- **安全内置**：权限分级、审批流、审计日志、签名验证、资源限制

### 1.2 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 实现语言 | Go 1.21+ | 纯 Go，支持交叉编译 |
| CLI 框架 | cobra | 命令行参数解析 |
| SSH 客户端 | golang.org/x/crypto/ssh | SSH 连接管理 |
| SFTP | github.com/pkg/sftp | 文件传输 |
| 系统信息 | gopsutil | 跨平台系统信息采集 |
| YAML 解析 | gopkg.in/yaml.v3 | 配置文件解析 |
| 加密签名 | crypto/ed25519 | Runner 二进制签名 |

---

## 2. 架构图

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                       opsctl (CLI)                          │
│  解析用户输入，调度本地编译/解释，管理远程执行与结果汇总       │
└───────┬─────────────────────────────────────────────────────┘
        │
        ├── 语言前端（Lexer → Parser → AST）
        │        ├── 解释引擎（AST 遍历执行，用于本地调试/REPL）
        │        └── 编译管线（AST → Go 源码 → go build → 二进制）
        │
        ├── 标准库（ops-core-sdk）
        │        ├── sys / file / net / process / service / pkg ...
        │        └── 每个函数返回结构体，支持 JSON 序列化
        │
        └── 远程控制面
                 ├── SSH 客户端（连接池、超时、重试、架构检测）
                 ├── 通用 Runner（多架构预编译，缓存复用）
                 ├── 分层中继引擎（自动选择中继节点，支持大规模文件分发/收集）
                 ├── 指令包生成（JSON，与架构无关）
                 └── 结果回收（stdout JSON 解析、错误聚合、审计日志）
```

### 2.2 项目目录结构

```
opslang/
├── cmd/
│   ├── opsctl/                 # CLI 主程序
│   │   ├── main.go            # 入口，命令注册
│   │   ├── run.go             # opsctl run - 本地解释执行
│   │   ├── build.go           # opsctl build - AOT 编译
│   │   ├── exec.go            # opsctl exec - 远程执行
│   │   └── repl.go            # opsctl repl - 交互式环境
│   └── ops-runner/            # 通用 Runner 程序
│       └── main.go            # 从 stdin 读取 JSON 指令包并执行
├── internal/
│   ├── lexer/                 # 词法分析
│   │   ├── token/
│   │   │   └── token.go      # Token 定义（16 个关键字）
│   │   └── lexer.go          # 词法分析器
│   ├── parser/
│   │   └── parser.go         # 语法分析器（递归下降）
│   ├── ast/
│   │   └── ast.go            # AST 节点定义
│   ├── interpreter/
│   │   └── interpreter.go    # 解释器（AST 遍历执行）
│   ├── compiler/
│   │   ├── compiler.go       # 编译器主逻辑
│   │   ├── codegen.go        # Go 代码生成器
│   │   ├── cache.go          # 编译缓存
│   │   └── mode_selector.go  # 执行模式自动选择
│   ├── sshx/                  # SSH 客户端封装
│   │   ├── client.go         # SSH 连接管理
│   │   ├── pool.go           # 连接池
│   │   ├── sftp.go           # SFTP 文件传输
│   │   ├── config.go         # 配置管理
│   │   └── errors.go         # 错误定义
│   ├── exec/
│   │   └── exec.go           # 远程执行协调器
│   ├── inventory/
│   │   └── inventory.go      # 主机清单解析
│   ├── arch/
│   │   └── arch.go           # 架构检测（uname -m → GOARCH）
│   ├── runner/                # Runner 指令包处理
│   │   ├── executor.go       # 指令执行器
│   │   ├── instruction_gen.go # 指令包生成器
│   │   ├── registry.go       # 内置函数注册表
│   │   └── types.go          # 类型定义
│   ├── relay/                 # 分层中继调度（Phase 4）
│   ├── security/              # 安全特性（Phase 5）
│   │   ├── privilege.go      # 权限分级
│   │   ├── audit.go          # 审计日志
│   │   ├── signature.go      # 签名验证
│   │   ├── resource_limit.go # 资源限制
│   │   ├── retry.go          # 自动重试
│   │   └── cleanup.go        # 自清理
│   └── output/
│       └── output.go         # 结构化输出处理
├── pkg/
│   └── ops-core-sdk/         # 原子操作标准库（独立 Go module）
│       ├── sys/              # 系统信息
│       ├── file/             # 文件操作
│       ├── net/              # 网络操作
│       ├── process/          # 进程管理
│       ├── service/          # 服务管理
│       ├── pkg/              # 包管理
│       ├── json/             # JSON 编解码
│       ├── yaml/             # YAML 编解码
│       └── time/             # 时间操作
├── tests/                    # 集成测试
├── examples/                 # 示例脚本
├── docs/                     # 文档
├── go.mod                    # Go module 定义
├── Makefile                  # 构建脚本
└── README.md                 # 项目说明
```

---

## 3. 模块说明

### 3.1 语言前端

#### 词法分析器（internal/lexer）

**职责**：将源代码转换为 Token 流

**关键文件**：
- `token/token.go`：定义 16 个关键字和所有 Token 类型
  - 关键字：`let`, `fn`, `if`, `else`, `for`, `while`, `return`, `task`, `on`, `ensure`, `report`, `metric`, `log`, `alert`, `import`, `nil`
  - Token 类型：标识符、字面量（整数/浮点/字符串/布尔）、运算符、分隔符、注释等
- `lexer.go`：词法分析实现，支持行列号追踪，错误报告包含位置信息

**核心流程**：
```
源代码字符串 → Lexer → Token 流
```

#### 语法分析器（internal/parser）

**职责**：将 Token 流转换为抽象语法树（AST）

**实现方式**：递归下降解析

**支持的语法结构**：
- 变量声明：`let x = 10`
- 函数定义：`fn add(a, b) { return a + b }`
- 控制流：`if/else`, `for`, `while`
- 任务声明：`task "name" on targets { ... }`
- 幂等声明：`ensure condition { actions }`
- 输出语句：`report`, `metric`, `log`, `alert`
- 导入语句：`import go "package"`

#### AST 节点（internal/ast）

**职责**：定义所有语法节点的 Go 结构体

**主要节点类型**：
- `Program`：程序根节点
- `TaskDecl`：任务声明
- `FunctionDecl`：函数声明
- `LetStmt`：变量声明
- `IfStmt`：条件语句
- `ForStmt`：循环语句
- `WhileStmt`：While 循环
- `ReturnStmt`：返回语句
- `EnsureStmt`：幂等声明
- `ReportStmt`：报告语句
- `CallExpr`：函数调用
- `BinaryExpr`：二元表达式
- `UnaryExpr`：一元表达式
- `Literal`：字面量（整数/浮点/字符串/布尔）
- `Identifier`：标识符
- `ListLiteral`：列表字面量
- `MapLiteral`：字典字面量

### 3.2 执行引擎

#### 解释器（internal/interpreter）

**职责**：遍历 AST 执行脚本，用于本地调试和 REPL

**核心特性**：
- 基于 AST 节点类型的 switch-case 分发
- 支持变量作用域（词法作用域链）
- 内置函数注册机制，映射到 ops-core-sdk
- 错误处理：运行时错误包含行列号

**执行流程**：
```
AST → Interpreter → 执行结果（结构化）
```

#### 编译器（internal/compiler）

**职责**：将 AST 编译为静态二进制，用于生产部署

**核心组件**：

1. **compiler.go**：编译主流程
   - 检查编译缓存
   - 调用代码生成器
   - 调用 go build 生成二进制
   - 更新缓存

2. **codegen.go**：Go 代码生成器
   - 将 AST 翻译为 Go 源码
   - 变量声明 → Go 变量声明
   - 函数调用 → ops-core-sdk 调用
   - 控制流 → Go 控制流
   - 生成 main() 函数

3. **cache.go**：编译缓存
   - 缓存键：脚本 hash + 标准库版本 + 目标架构
   - 缓存目录：`~/.cache/opslang/compilation/`
   - 命中缓存直接返回，避免重复编译

4. **mode_selector.go**：执行模式自动选择
   - 简单任务（<100行，无第三方 Go 库）→ Runner 模式
   - 复杂任务（≥100行或有 `import go`）→ AOT 编译模式
   - 支持手动覆盖：`--mode runner|aot|auto`

**编译流程**：
```
AST → CodeGen → Go 源码 → go build → 静态二进制
```

### 3.3 标准库（pkg/ops-core-sdk）

**职责**：提供运维原子操作，返回结构化数据

**设计原则**：
- 独立 Go module，可单独使用
- 每个函数返回 `(结构体, error)`
- 结构体带 JSON tag，支持序列化
- 所有操作不依赖 shell，直接使用 Go 库或系统调用

**模块列表**：

| 模块 | 功能 | 示例函数 |
|------|------|---------|
| sys | 系统信息 | `cpu.usage()`, `memory.info()`, `disk.usage()`, `hostname()` |
| file | 文件操作 | `read()`, `write()`, `copy()`, `move()`, `delete()`, `exists()`, `checksum()` |
| net | 网络操作 | `http_get()`, `http_post()`, `tcp_check()`, `dns_lookup()`, `interfaces()` |
| process | 进程管理 | `list()`, `find_by_name()`, `find_by_port()`, `exec()`, `kill()` |
| service | 服务管理 | `status()`, `start()`, `stop()`, `restart()`, `enable()` |
| pkg | 包管理 | `install()`, `remove()`, `list()` |
| json | JSON 编解码 | `encode()`, `decode()` |
| yaml | YAML 编解码 | `encode()`, `decode()` |
| time | 时间操作 | `now()`, `format()`, `parse()`, `since()`, `sleep()`, `diff()` |

**返回结构体示例**：

```go
// sys.CPUUsage 返回
type CPUUsage struct {
    Percent float64 `json:"percent"`
    User    float64 `json:"user"`
    System  float64 `json:"system"`
    Idle    float64 `json:"idle"`
}

// file.ChecksumResult 返回
type ChecksumResult struct {
    Path   string `json:"path"`
    Algo   string `json:"algo"`
    Value  string `json:"value"`
    Size   int64  `json:"size"`
}
```

### 3.4 远程控制面

#### SSH 客户端（internal/sshx）

**职责**：管理 SSH 连接，支持认证、超时、重试、连接池

**核心功能**：
- 密码/密钥认证
- 连接超时控制
- 命令执行超时
- 自动重试（可配置次数）
- 连接池管理（复用连接）
- 并发限制

**关键文件**：
- `client.go`：SSH 连接创建和管理
- `pool.go`：连接池实现
- `sftp.go`：SFTP 文件传输（支持断点续传、压缩）
- `config.go`：连接配置
- `errors.go`：错误定义

#### 架构检测（internal/arch）

**职责**：检测远程主机架构，映射到 GOARCH

**实现方式**：
```bash
# 远程执行
uname -m
# 输出示例：x86_64, aarch64, armv7l
```

**映射规则**：
- `x86_64` → `linux/amd64`
- `aarch64` → `linux/arm64`
- `armv7l` → `linux/arm`

**缓存策略**：
- 缓存目录：`~/.cache/opslang/runners/`
- 按架构存储 Runner 二进制
- 首次上传后复用，避免重复传输

#### 通用 Runner（cmd/ops-runner）

**职责**：在远程主机上执行 JSON 指令包

**工作流程**：
1. 从 stdin 读取 JSON 指令包
2. 解析指令序列
3. 按顺序执行每条指令（调用 ops-core-sdk）
4. 收集执行结果
5. 输出 JSON 结果到 stdout

**指令包格式**：
```json
{
  "version": "1.0",
  "task_id": "abc123",
  "dry_run": false,
  "instructions": [
    {
      "op": "sys.cpu.usage",
      "args": {},
      "assign": "cpu"
    },
    {
      "op": "report",
      "args": { "cpu": "cpu" }
    }
  ]
}
```

**输出格式**：
```json
{
  "status": "ok",
  "data": {
    "cpu": { "percent": 12.5, "user": 5.0, "system": 3.0, "idle": 92.0 }
  },
  "errors": [],
  "warnings": []
}
```

**多架构构建**：
```bash
# 构建多架构 Runner
make build-runner-linux-amd64
make build-runner-linux-arm64
```

#### 执行协调器（internal/exec）

**职责**：协调远程执行全流程

**工作流程**：
1. 解析目标主机列表（`--hosts` 或 `--inventory`）
2. 加载 JSON 指令包
3. 对每台主机并发执行：
   - SSH 连接
   - 架构检测
   - 上传/复用 Runner
   - 发送指令包
   - 收集输出
4. 汇总结果，输出 JSON 摘要

**并发控制**：
- `--parallel` 参数限制并发数
- 默认并发数：10

**结果汇总格式**：
```json
{
  "task_id": "abc123",
  "script": "check_cpu.ops",
  "targets": ["host1", "host2"],
  "started_at": "2026-08-15T10:00:00Z",
  "finished_at": "2026-08-15T10:00:12Z",
  "results": {
    "host1": { "status": "success", "exit_code": 0, "data": {...} },
    "host2": { "status": "failed", "exit_code": 1, "error": "timeout" }
  },
  "audit_log": "/var/log/opsctl/abc123.json"
}
```

---

## 4. 执行模式

### 4.1 Runner 模式（解释执行）

**适用场景**：
- 简单任务（<100行，无第三方 Go 库）
- 紧急故障处理（零编译延迟）
- 本地调试和 REPL

**工作流程**：
```
脚本源码 → Lexer → Parser → AST → Interpreter → 执行结果
```

**特点**：
- 无需编译，即时执行
- 通过 SSH 下发预编译 Runner 二进制
- 发送 JSON 指令包，Runner 解释执行
- 适合快速迭代和调试

**命令示例**：
```bash
# 本地解释执行
opsctl run script.ops

# 远程解释执行
opsctl exec --hosts host1,host2 --instructions script.ops
```

### 4.2 AOT 编译模式

**适用场景**：
- 复杂业务逻辑
- 需要第三方 Go 库（`import go "..."`）
- 生产环境部署（需要高性能）

**工作流程**：
```
脚本源码 → Lexer → Parser → AST → CodeGen → Go 源码 → go build → 静态二进制
```

**特点**：
- 编译为静态二进制，零依赖
- 支持交叉编译（amd64 → arm64）
- 编译缓存加速重复构建
- 适合生产环境

**命令示例**：
```bash
# 本地编译执行
opsctl build --input script.ops --output binary
./binary

# 交叉编译
opsctl build --input script.ops --output binary --target-arch linux/arm64
```

### 4.3 模式自动选择

**选择策略**（`mode_selector.go`）：

| 条件 | 选择模式 | 原因 |
|------|---------|------|
| 脚本 < 100 行，无 `import go` | Runner | 逻辑简单，指令包分发高效 |
| 脚本 ≥ 100 行，或有 `import go` | AOT | 需要 Go 生态支持 |
| 用户强制指定 | 按用户选择 | `--mode runner|aot|auto` |

**默认行为**：`--mode auto`（自动选择）

---

## 5. 数据流

### 5.1 本地执行数据流

```
┌─────────────┐
│  脚本源码    │
│ (.ops 文件)  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Lexer     │  词法分析
│ (lexer.go)  │  源码 → Token 流
└──────┬──────┘
       │ Token 流
       ▼
┌─────────────┐
│   Parser    │  语法分析
│ (parser.go) │  Token 流 → AST
└──────┬──────┘
       │ AST
       ▼
┌─────────────┐
│ Interpreter │  解释执行
│(interpreter │  AST → 执行结果
│    .go)     │
└──────┬──────┘
       │ 结构化结果
       ▼
┌─────────────┐
│  JSON 输出  │
└─────────────┘
```

### 5.2 远程执行数据流（Runner 模式）

```
┌─────────────┐
│  脚本源码    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  语言前端    │  Lexer → Parser → AST
└──────┬──────┘
       │ AST
       ▼
┌─────────────┐
│ 指令包生成  │  AST → JSON 指令包
│(instruction │
│  _gen.go)   │
└──────┬──────┘
       │ JSON 指令包
       ▼
┌─────────────┐
│ SSH 传输    │  上传 Runner（如未缓存）
│  (sshx)     │  发送指令包到 stdin
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Runner     │  读取 stdin 指令包
│(ops-runner) │  执行 ops-core-sdk
└──────┬──────┘
       │ JSON 结果
       ▼
┌─────────────┐
│ 结果回收    │  解析 stdout JSON
│  (exec)     │  聚合多台主机结果
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  汇总输出   │  JSON 格式
└─────────────┘
```

### 5.3 远程执行数据流（AOT 编译模式）

```
┌─────────────┐
│  脚本源码    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  语言前端    │  Lexer → Parser → AST
└──────┬──────┘
       │ AST
       ▼
┌─────────────┐
│  编译器     │  AST → Go 源码 → go build
│ (compiler)  │  → 静态二进制
└──────┬──────┘
       │ 静态二进制
       ▼
┌─────────────┐
│ SSH 传输    │  上传二进制到远程主机
│  (sshx)     │  （支持多架构）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 远程执行    │  直接执行二进制
│             │  零依赖，无需 Runner
└──────┬──────┘
       │ JSON 结果
       ▼
┌─────────────┐
│ 结果回收    │  解析 stdout JSON
│  (exec)     │  聚合多台主机结果
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  汇总输出   │  JSON 格式
└─────────────┘
```

### 5.4 编译缓存数据流

```
┌─────────────┐
│  脚本源码    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  计算缓存键 │  hash(源码) + 标准库版本 + 目标架构
└──────┬──────┘
       │
       ▼
┌─────────────┐     命中      ┌─────────────┐
│  检查缓存   │ ────────────→ │  直接返回   │
└──────┬──────┘               └─────────────┘
       │ 未命中
       ▼
┌─────────────┐
│  编译生成   │  CodeGen → go build
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  写入缓存   │  存储到 ~/.cache/opslang/compilation/
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  返回二进制 │
└─────────────┘
```

---

## 6. 扩展指南

### 6.1 添加新的标准库函数

**步骤**：

1. **在对应模块中实现函数**（`pkg/ops-core-sdk/<module>/`）

```go
// pkg/ops-core-sdk/sys/swap.go
package sys

// SwapInfo 交换分区信息
type SwapInfo struct {
    Total uint64 `json:"total"`
    Used  uint64 `json:"used"`
    Free  uint64 `json:"free"`
}

// Swap 获取交换分区信息
func Swap() (*SwapInfo, error) {
    // 实现逻辑
    // 使用 gopsutil 或直接读取 /proc/meminfo
}
```

2. **编写单元测试**

```go
// pkg/ops-core-sdk/sys/swap_test.go
func TestSwap(t *testing.T) {
    info, err := Swap()
    if err != nil {
        t.Fatalf("Swap() error = %v", err)
    }
    if info.Total == 0 {
        t.Error("Swap() Total = 0, want > 0")
    }
}
```

3. **在 Runner 中注册**（`internal/runner/registry.go`）

```go
func init() {
    Registry["sys.swap"] = func(args map[string]interface{}) (interface{}, error) {
        return sys.Swap()
    }
}
```

4. **在解释器中注册**（`internal/interpreter/interpreter.go`）

```go
func (i *Interpreter) registerDefaults() {
    i.registerBuiltin("sys.swap", func(args ...Object) (Object, error) {
        result, err := sys.Swap()
        if err != nil {
            return nil, err
        }
        return structToObject(result), nil
    })
}
```

### 6.2 添加新的 AST 节点类型

**步骤**：

1. **定义节点结构**（`internal/ast/ast.go`）

```go
// CustomStmt 自定义语句
type CustomStmt struct {
    Token token.Token
    Name  string
    Body  []Statement
}
```

2. **在 Parser 中实现解析逻辑**（`internal/parser/parser.go`）

```go
func (p *Parser) parseCustomStatement() *ast.CustomStmt {
    stmt := &ast.CustomStmt{Token: p.curToken}
    // 解析逻辑
    return stmt
}
```

3. **在 Interpreter 中实现执行逻辑**（`internal/interpreter/interpreter.go`）

```go
func (i *Interpreter) evalCustomStmt(stmt *ast.CustomStmt) (Object, error) {
    // 执行逻辑
    return nil, nil
}
```

4. **在 CodeGen 中实现代码生成**（`internal/compiler/codegen.go`）

```go
func (g *Generator) genCustomStmt(stmt *ast.CustomStmt) {
    // 生成 Go 代码
}
```

### 6.3 添加新的 CLI 命令

**步骤**：

1. **创建命令文件**（`cmd/opsctl/<command>.go`）

```go
package main

import (
    "github.com/spf13/cobra"
)

var customCmd = &cobra.Command{
    Use:   "custom [args]",
    Short: "自定义命令描述",
    Long:  `自定义命令详细描述`,
    RunE: func(cmd *cobra.Command, args []string) error {
        // 命令实现
        return nil
    },
}

func init() {
    rootCmd.AddCommand(customCmd)
    // 添加命令参数
    customCmd.Flags().StringP("flag", "f", "default", "flag 描述")
}
```

2. **实现命令逻辑**

```go
func runCustom(args []string, flag string) error {
    // 实现逻辑
    return nil
}
```

### 6.4 添加新的安全特性

**权限分级扩展**（`internal/security/privilege.go`）：

```go
// 添加新的权限级别
const (
    PrivilegeReadOnly  PrivilegeLevel = "read_only"
    PrivilegeAdmin     PrivilegeLevel = "admin"
    PrivilegeRoot      PrivilegeLevel = "root"
    PrivilegeCustom    PrivilegeLevel = "custom" // 新增
)

// 添加权限检查函数
func CheckCustomPrivilege(script *ast.Program) error {
    // 检查逻辑
    return nil
}
```

**审计日志扩展**（`internal/security/audit.go`）：

```go
// 添加新的审计事件类型
type AuditEventType string

const (
    EventTaskExec    AuditEventType = "task_exec"
    EventFileAccess  AuditEventType = "file_access"
    EventCustomEvent AuditEventType = "custom_event" // 新增
)

// 记录自定义事件
func LogCustomEvent(event *AuditEvent) error {
    // 记录逻辑
    return nil
}
```

### 6.5 测试指南

**单元测试**：
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./pkg/ops-core-sdk/sys/...

# 带覆盖率
go test -cover ./...
```

**集成测试**：
```bash
# 运行集成测试
go test ./tests/...

# 端到端测试
make e2e-test
```

**交叉编译测试**：
```bash
# 测试 amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./...

# 测试 arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./...
```

### 6.6 贡献流程

1. Fork 项目仓库
2. 创建功能分支：`git checkout -b feature/new-feature`
3. 编写代码和测试
4. 确保所有测试通过：`go test ./...`
5. 确保代码格式化：`go fmt ./...`
6. 提交更改：`git commit -m 'feat: add new feature'`
7. 推送分支：`git push origin feature/new-feature`
8. 创建 Pull Request

**提交信息规范**：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

---

## 附录

### A. 关键字列表

| 关键字 | 用途 | 示例 |
|--------|------|------|
| `let` | 变量声明 | `let x = 10` |
| `fn` | 函数定义 | `fn add(a, b) { return a + b }` |
| `if` | 条件语句 | `if x > 0 { ... }` |
| `else` | 条件分支 | `else { ... }` |
| `for` | 循环语句 | `for i in list { ... }` |
| `while` | 循环语句 | `while x > 0 { ... }` |
| `return` | 返回语句 | `return result` |
| `task` | 任务声明 | `task "name" on targets { ... }` |
| `on` | 目标声明 | `on ["host1", "host2"]` |
| `ensure` | 幂等声明 | `ensure condition { actions }` |
| `report` | 报告输出 | `report { key: value }` |
| `metric` | 指标输出 | `metric(name, value, labels)` |
| `log` | 日志输出 | `log("message")` |
| `alert` | 告警输出 | `alert("warning")` |
| `import` | 导入语句 | `import go "package"` |
| `nil` | 空值 | `let x = nil` |

### B. 标准库函数完整列表

详见 `CLAUDE.md` 附录 A。

### C. 数据类型

**基本类型**：
- 整数：`42`
- 浮点数：`3.14`
- 字符串：`"hello"`
- 布尔：`true`, `false`
- 空值：`nil`

**复合类型**：
- 列表：`[1, 2, 3]`
- 字典：`{"key": "value"}`
- 结构体：由标准库函数返回

### D. 运算符

**算术运算符**：`+`, `-`, `*`, `/`, `%`

**比较运算符**：`==`, `!=`, `<`, `>`, `<=`, `>=`

**逻辑运算符**：`&&`, `||`, `!`

**赋值运算符**：`=`

---

## 参考资源

- 项目仓库：https://github.com/j4ckzh0u/opslang
- Go 文档：https://pkg.go.dev/github.com/j4ckzh0u/opslang
- 示例脚本：`examples/` 目录
- 开发计划：`CLAUDE.md`
