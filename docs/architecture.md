# OpsLang 系统架构文档

## 目录

1. [系统架构概览](#1-系统架构概览)
2. [架构图](#2-架构图)
3. [模块说明](#3-模块说明)
4. [执行模式](#4-执行模式)
5. [数据流](#5-数据流)
6. [SSH 安全](#6-ssh-安全)
7. [扩展指南](#7-扩展指南)
8. [Roadmap（未实现）](#8-roadmap未实现)

---

## 1. 系统架构概览

OpsLang 是一门面向运维领域的领域特定语言（DSL），旨在通过简洁的脚本完成复杂的运维操作。系统采用纯 Go 实现，支持零依赖远程执行、异构架构、结构化返回等核心特性。

### 1.1 核心设计原则

- **纯 Go 生态**：所有依赖支持 `CGO_ENABLED=0` 交叉编译，无 cgo 依赖
- **结构化返回**：标准库函数一律返回结构体，禁止返回原始字符串
- **无 Shell 依赖**：所有操作直接使用 Go 库或读取 `/proc`/`/sys`，不依赖 shell
- **极简语法**：关键字 20 个，类 Go + Python 混合风格，动态类型
- **双执行引擎**：Runner（JSON 指令包，线性脚本）+ AOT 编译（静态二进制，全语言）
- **单一事实来源**：`internal/opsspec` 定义全部原子操作的名称、参数与可用范围，三个引擎由一致性测试强制对齐
- **零依赖远程执行**：通过 SSH 下发预编译 Runner 或 AOT 二进制
- **安全内置**：权限分级、审计日志、SSH 主机密钥 TOFU 校验

### 1.2 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 实现语言 | Go 1.25+ | 纯 Go，支持交叉编译 |
| CLI 框架 | cobra | 命令行参数解析 |
| SSH 客户端 | golang.org/x/crypto/ssh | SSH 连接管理（含 TOFU 主机密钥校验） |
| SFTP | github.com/pkg/sftp | 文件传输 |
| 系统信息 | gopsutil/v4 | 跨平台系统信息采集 |
| YAML 解析 | gopkg.in/yaml.v3 | 配置文件解析 |
| 加密签名 | crypto/ed25519 | 审计/签名工具（internal/security） |

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
                 ├── opsspec（原子操作单一事实来源，三引擎一致性测试）
                 ├── SSH 客户端（连接池、超时、重试、TOFU 主机密钥校验、架构检测）
                 ├── 通用 Runner（多架构预编译，缓存复用）
                 ├── 指令包生成（JSON，与架构无关；只支持线性脚本）
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
│   │   ├── deploy.go          # opsctl deploy - 远程部署（runner/aot/auto）
│   │   ├── exec.go            # opsctl exec - 远程执行指令包
│   │   └── repl.go            # opsctl repl - 交互式环境
│   └── ops-runner/            # 通用 Runner 程序
│       └── main.go            # 从 stdin 读取 JSON 指令包并执行
├── internal/
│   ├── lexer/                 # 词法分析
│   │   ├── token/
│   │   │   └── token.go      # Token 定义（20 个关键字）
│   │   └── lexer.go          # 词法分析器
│   ├── parser/
│   │   └── parser.go         # 语法分析器（递归下降）
│   ├── ast/
│   │   └── ast.go            # AST 节点定义
│   ├── interpreter/
│   │   ├── interpreter.go    # 解释器（AST 遍历执行）
│   │   └── sdk_bridge.go     # SDK 内置函数注册（解释器侧）
│   ├── compiler/
│   │   ├── compiler.go       # 编译器主逻辑
│   │   ├── codegen.go        # Go 代码生成器
│   │   ├── cache.go          # 编译缓存
│   │   └── mode_selector.go  # 执行模式选择（RequiresAOT）
│   ├── opsspec/               # 原子操作单一事实来源
│   │   ├── spec.go           # canonical 名称 + 参数 + 可用范围 + 旧别名
│   │   └── consistency_test.go # 三引擎一致性测试
│   ├── sshx/                  # SSH 客户端封装
│   │   ├── client.go         # SSH 连接管理（SFTP 上传/下载）
│   │   ├── pool.go           # 连接池
│   │   ├── hostkey.go        # TOFU 主机密钥校验
│   │   ├── config.go         # 配置管理
│   │   └── errors.go         # 错误定义
│   ├── exec/
│   │   └── exec.go           # 远程执行协调器（并行编排、架构探测、
│   │                          #   runner/AOT 二进制上传、buildOnce 去重）
│   ├── inventory/
│   │   └── inventory.go      # 主机清单解析
│   ├── arch/
│   │   └── arch.go           # 架构检测（uname -m → GOARCH）
│   ├── runner/                # Runner 指令包处理
│   │   ├── executor.go       # 指令执行器
│   │   ├── instruction_gen.go # 指令包生成器（线性脚本，越界即报错）
│   │   ├── registry.go       # 内置函数注册表（runner 侧）
│   │   └── types.go          # 类型定义
│   ├── security/              # 安全特性
│   │   ├── privilege.go      # 权限分级
│   │   ├── audit.go          # 审计日志
│   │   ├── signature.go      # Ed25519 签名工具
│   │   ├── resource_limit.go # 资源限制
│   │   ├── resources.go      # 资源抽象
│   │   ├── retry.go          # 自动重试
│   │   └── cleanup.go        # 自清理
│   └── output/
│       └── output.go         # 结构化输出处理
├── pkg/
│   └── ops-core-sdk/         # 原子操作标准库
│       ├── sys/              # 系统信息
│       ├── file/             # 文件操作（含 distribute/collect SSH 传输）
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
- `token/token.go`：定义 20 个关键字和所有 Token 类型
  - 关键字：`let`, `fn`, `if`, `else`, `for`, `while`, `return`, `task`, `on`, `import`, `privilege`, `true`, `false`, `nil`, `report`, `alert`, `ensure`, `metric`, `log`, `parallel`
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
- 控制流：`if/else`，C 风格 `for`，`while`（无 for-in 遍历语法）
- 任务声明：`task "name" on <目标> { ... }`
- 幂等声明：`ensure condition { actions } notify <表达式>`
- 并行块：`parallel { ... }`
- 输出语句：`report`, `metric`, `log`, `alert`
- 导入语句：`import`（声明性；`import "go ..."` 第三方库在执行入口报错拒绝）

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
   - 缓存键：codegen 版本 + 脚本源码 hash + 目标架构
   - 缓存目录：`~/.opsctl/cache/`
   - 命中缓存直接返回，避免重复编译

4. **mode_selector.go**：执行模式选择
   - `RequiresAOT()` 检测线性 runner 指令 VM 无法精确表达的语句（`if`/`for`/`while`/`fn`/`ensure`/`parallel`/`return`），命中即必须 AOT
   - `opsctl deploy --mode auto`（默认）：先做一次真实指令包生成试验，生成器拒绝的脚本自动转 AOT，保证模式判定与生成器永不漂移
   - 支持手动覆盖：`--mode runner|aot|auto`

**编译流程**：
```
AST → CodeGen → Go 源码 → go build → 静态二进制
```

codegen 生成的代码由 `codegen_e2e_test.go` 经真实 `go build` 编译并验证。

### 3.3 标准库（pkg/ops-core-sdk）

**职责**：提供运维原子操作，返回结构化数据

**设计原则**：
- 每个函数返回 `(结构体, error)`
- 结构体带 JSON tag，支持序列化
- 所有操作不依赖 shell，直接使用 Go 库或系统调用

**模块列表**：

| 模块 | 功能 | 示例函数 |
|------|------|---------|
| sys | 系统信息 | `cpu.usage()`, `cpu.count()`, `memory.info()`, `disk.usage()`, `disk.partitions()`, `hostname()`, `load()`, `os()`, `uptime()`, `users()`, `net.interfaces()` |
| file | 文件操作 | `read()`, `write()`, `append()`, `copy()`, `move()`, `delete()`, `exists()`, `stat()`, `list()`, `mkdir()`, `chmod()`, `checksum()`, `template()` |
| file（控制器专用） | SSH 分发/收集 | `distribute()`, `collect()`（真实 SFTP 传输，传输后 SHA-256 校验） |
| net | 网络操作 | `http_get()`, `http_post()`, `tcp_check()`, `dns_lookup()`, `interfaces()` |
| process | 进程管理 | `list()`, `find_by_name()`, `find_by_port()`, `kill()`, `exec()` |
| service | 服务管理 | `status()`, `start()`, `stop()`, `restart()`, `enable()`, `disable()` |
| pkg | 包管理（仅 Linux） | `install()`, `remove()`, `info()`, `list()` |
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
    Path      string `json:"path"`
    Algorithm string `json:"algorithm"`
    Checksum  string `json:"checksum"`
    Size      int64  `json:"size"`
}
```

### 3.4 原子操作单一事实来源（internal/opsspec）

`opsspec.Funcs` 表定义全部原子操作的 canonical 名称、位置参数名和可用范围（`All` / `ControllerOnly`），是三个执行引擎共同的依据：

- **解释器**（`internal/interpreter/sdk_bridge.go`）— 本地执行时通过 bridge 调用 SDK
- **Runner 注册表**（`internal/runner/registry.go`）— 远程 runner 按 canonical 名（或旧别名，查询时透明映射）分发
- **AOT 代码生成器**（`internal/compiler/codegen.go`）— 只生成 canonical 名调用

`internal/opsspec/consistency_test.go` 强制三个引擎与 spec 表完全一致；新增函数必须三处同步注册，否则测试失败。`ControllerOnly` 的函数（如 `file.distribute`/`file.collect`）只在控制器侧暴露，runner 与 codegen 拒绝。

### 3.5 远程控制面

#### SSH 客户端（internal/sshx）

**职责**：管理 SSH 连接，支持认证、超时、重试、连接池

**核心功能**：
- 密码/密钥认证
- 连接超时控制
- 命令执行超时
- 自动重试（可配置次数）
- 连接池管理（复用连接）
- 并发限制
- TOFU 主机密钥校验（见第 6 节）

**关键文件**：
- `client.go`：SSH 连接创建和管理（含 SFTP 上传/下载）
- `pool.go`：连接池实现
- `hostkey.go`：TOFU 主机密钥校验
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
1. 解析目标主机列表（`--hosts`/`--targets` 或 `--inventory`）
2. 加载 JSON 指令包（exec）或按 deploy 步骤生成指令包
3. 对每台主机并发执行（goroutine + 信号量限流）：
   - SSH 连接（TOFU 主机密钥校验）
   - 架构检测（`uname -m` → GOARCH）
   - 按架构构建并上传 Runner / AOT 应用二进制（首次上传后本地缓存复用）
   - 发送指令包
   - 收集输出与退出码
4. 汇总结果，输出 JSON 摘要

**并发与构建去重**：
- `--parallel` 参数限制并发主机数（默认 10），实现为信号量
- `buildOnce`（sync.Once 按构建键去重，singleflight 模式）保证同一架构的二进制在并发主机间只编译一次
- AOT 模式下 `binary.exec` 指令的失败会如实上报；deploy 汇总为 `partial`/`failed` 时返回非零退出码，审计日志不记成功

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

### 4.1 本地解释执行（opsctl run）

**适用场景**：本地调试、REPL、快速验证

**工作流程**：
```
脚本源码 → Lexer → Parser → AST → Interpreter → 执行结果
```

**能力边界**：全语言支持（`if`/`for`/`while`/`fn`/`ensure`/`parallel`）。但带 `on` 子句的 `task` 会报错，提示改用 `opsctl deploy`（本地执行无法路由远程主机）。

```bash
opsctl run script.ops [--dry-run]
```

`--dry-run` 下 `ensure` 的 apply 步骤只打印不执行。

### 4.2 Runner 模式（opsctl deploy --mode runner）

**适用场景**：线性脚本的零编译远程执行

**工作流程**：
```
脚本源码 → Lexer → Parser → AST → 指令包生成（JSON）→ SSH 下发 → ops-runner 执行
```

**能力边界（重要）**：
- 只支持**线性脚本**：调用、`let`、`report`、`alert`、`log`
- `if`/`for`/`while`/`fn`/`ensure`/`parallel` 和运行期计算表达式会**报错拒绝**——不会静默降级或误翻译
- task 的 `on` 子句在此模式下生效：支持精确名 / `user@host` / glob（`path.Match`）匹配 `--targets` 目标；变量与动态选择器报错
- 旧指令包中的历史别名（`sys.load.avg`、`net.http.get` 等）在 runner 侧透明解析为 canonical 名

**退出码**（ops-runner）：0=全部成功，1=部分失败，2=全部失败，3=协议错误。

### 4.3 AOT 编译模式（opsctl deploy --mode aot / opsctl build）

**适用场景**：完整语言（含 `ensure`/`parallel`）的远程执行、生产部署

**工作流程**：
```
脚本源码 → Lexer → Parser → AST → CodeGen → Go 源码 → go build（按目标架构交叉编译）
→ SSH 上传 → 远程执行 → 结果回收（binary.exec 失败如实上报）
```

**特点**：
- 按目标机架构（`uname -m` 探测）交叉编译，真实上传执行
- 支持全语言（`if`/`for`/`while`/`fn`/`ensure`/`parallel`）
- 编译缓存加速重复构建
- **task 级 `on` 路由不支持**（会报错）：自包含二进制无法知道自己落在哪台主机，为避免误路由到全部目标而拒绝
- 不支持第三方 Go 库（`import "go ..."` 直接报错）

```bash
# 手动编译
opsctl build --source script.ops --output binary --target-arch linux/arm64
```

### 4.4 模式自动选择（--mode auto，默认）

| 条件 | 选择模式 | 原因 |
|------|---------|------|
| 脚本含控制流/函数/ensure/parallel（`RequiresAOT` 命中） | AOT | 线性指令 VM 无法精确表达 |
| 指令包生成试验成功 | Runner | 线性脚本，指令包分发最高效 |
| 指令包生成试验失败（生成器拒绝） | AOT | 与生成器保持一致，永不漂移 |
| 用户强制指定 | 按用户选择 | `--mode runner|aot|auto` |

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
│  写入缓存   │  存储到 ~/.opsctl/cache/（编译缓存）
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  返回二进制 │
└─────────────┘
```

> Runner 二进制缓存目录为 `~/.cache/opslang/runners/`（可用 `OPSLANG_CACHE_DIR` 覆盖）。

---

## 6. SSH 安全

### 6.1 主机密钥 TOFU 校验

默认对每条 SSH 连接执行 **TOFU（Trust On First Use，首次信任）** 主机密钥校验（`internal/sshx/hostkey.go`）：

- 已知主机 + 密钥匹配 → 通过
- 已知主机 + **密钥不一致 → 拒绝连接**（可能存在中间人攻击）
- 未知主机 → 记录密钥后放行（首次信任）

已知主机文件独立维护，不污染用户 OpenSSH 状态：

- 默认：`~/.ssh/opslang_known_hosts`
- 可用环境变量 `OPSLANG_KNOWN_HOSTS` 覆盖

历史版本使用 `InsecureIgnoreHostKey` 完全跳过校验，现已移除；仅当显式传入 `--insecure-host-key` 逃生开关（仅限实验室环境）时才跳过校验。

### 6.2 分发/收集凭据

`file.distribute` / `file.collect` 的 SSH 凭据从环境变量读取：

- `OPSLANG_SSH_PASSWORD` — 密码认证
- `OPSLANG_SSH_KEY` — 私钥路径

---

## 7. 扩展指南

### 7.1 添加新的标准库函数

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

3. **在 opsspec 表中登记**（`internal/opsspec/spec.go`）——canonical 名称、参数名、可用范围，这是一致性测试的依据

```go
// internal/opsspec/spec.go
{Name: "sys.swap"},
```

4. **在 Runner 中注册**（`internal/runner/registry.go`）

```go
r.Register("sys.swap", func(args map[string]interface{}) (interface{}, error) {
    return sys.Swap()
})
```

5. **在解释器 bridge 中注册**（`internal/interpreter/sdk_bridge.go`）

```go
interp.builtins["sys.swap"] = func(args ...interface{}) (interface{}, error) {
    r, err := sys.Swap()
    if err != nil {
        return nil, err
    }
    return structToMap(r)
}
```

6. **在 AOT codegen 中支持**（`internal/compiler/codegen.go`）——三引擎一致性测试会强制上述注册与 opsspec 表对齐，缺一处即测试失败

### 7.2 添加新的 AST 节点类型

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

### 7.3 添加新的 CLI 命令

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

### 7.4 添加新的安全特性

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

### 7.5 测试指南

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

### 7.6 贡献流程

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

## 8. Roadmap（未实现）

以下能力**尚未实现**，架构文档不以现有功能描述它们：

- **`for ... in ...` 遍历循环语法** — 只支持 C 风格 for 循环
- **`import "go <包路径>"` 第三方 Go 库** — 所有引擎报错拒绝
- **分层中继（relay）架构** — 原 `internal/relay` 已删除
- **文件传输压缩、断点续传、内容哈希去重分发** — `file.distribute`/`file.collect` 为直接 SFTP 传输

## 附录

### A. 关键字列表

| 关键字 | 用途 | 示例 |
|--------|------|------|
| `let` | 变量声明 | `let x = 10` |
| `fn` | 函数定义 | `fn add(a, b) { return a + b }` |
| `if` | 条件语句 | `if x > 0 { ... }` |
| `else` | 条件分支 | `else { ... }` |
| `for` | C 风格循环 | `for let i = 0; i < 10; i = i + 1 { ... }` |
| `while` | 循环语句 | `while x > 0 { ... }` |
| `return` | 返回语句 | `return result` |
| `task` | 任务声明 | `task "name" on "web*" { ... }` |
| `on` | 目标声明 | `on "host1"` |
| `ensure` | 幂等声明 | `ensure cond { actions } notify expr` |
| `report` | 报告输出 | `report { key: value }` |
| `metric` | 指标输出 | `metric(name, value, labels)` |
| `log` | 日志输出 | `log("message")` |
| `alert` | 告警输出 | `alert("warning")` |
| `parallel` | 并行块 | `parallel { ... }` |
| `privilege` | 权限声明 | `privilege: admin` |
| `import` | 导入（声明性） | `import sys` |
| `nil` | 空值 | `let x = nil` |
| `true` / `false` | 布尔字面量 | `let ok = true` |

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
