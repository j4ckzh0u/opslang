# OpsLang 系统架构文档

## 1. 系统架构概览

OpsLang 是一门面向运维领域的领域特定语言（DSL），采用纯 Go 实现，支持交叉编译（`CGO_ENABLED=0`），零 cgo 依赖。系统由 CLI 工具 `opsctl`、通用执行器 `ops-runner`、独立标准库 `ops-core-sdk` 三大部分组成，提供解释执行与 AOT 编译两种运行模式。

### 核心设计原则

- **纯 Go 生态**：全部代码支持 `CGO_ENABLED=0` 交叉编译，覆盖 linux/amd64、linux/arm64 等主流架构
- **结构化返回**：标准库每个函数返回强类型结构体（带 JSON tag），禁止返回原始字符串
- **无 Shell 依赖**：所有系统操作通过 Go 标准库或直接系统调用完成，不使用 shell 解析
- **极简语法**：16 个关键字，类 Go + Python 混合风格，静态类型 + 类型推断
- **安全内置**：权限分级、审计日志、Ed25519 签名验证、资源限制、自动重试与清理

### 技术依赖

| 依赖 | 用途 |
|------|------|
| `gopsutil` | 系统信息采集（CPU、内存、磁盘、进程等） |
| `cobra` | CLI 命令行框架 |
| `pkg/sftp` | SFTP 文件传输 |
| `crypto/ed25519` | Runner 二进制签名验证 |
| `gopkg.in/yaml.v3` | YAML 配置解析 |

---

## 2. 架构图

### 2.1 总体架构

```
+---------------------------------------------------------------+
|                        opsctl (CLI)                            |
|  cmd/opsctl/                                                   |
|  main.go | run.go | build.go | exec.go | repl.go              |
+-------+----------------------------+--------------------------+
        |                            |
        v                            v
+------------------+    +-------------------------+
|   语言前端        |    |    远程执行面            |
|                  |    |                         |
| lexer/  token.go |    | exec/      执行调度      |
| parser/ parser.go|    | sshx/      SSH 客户端    |
| ast/    ast.go   |    | inventory/ 主机清单      |
|                  |    | arch/      架构检测      |
| interpreter/     |    | runner/    指令包处理    |
|   解释器          |    | security/  安全特性      |
| compiler/        |    | output/    结构化输出    |
|   编译管线        |    +-------------------------+
+------------------+
        |
        v
+-------------------------+
|   ops-core-sdk          |
|   pkg/ops-core-sdk/     |
|   sys | file | net      |
|   process | service     |
|   pkg | json | yaml     |
|   time                  |
+-------------------------+
        |
        v
+-------------------------+
|   ops-runner            |
|   cmd/ops-runner/       |
|   通用执行器（多架构）    |
+-------------------------+
```

### 2.2 项目目录结构

```
opslang/
├── cmd/
│   ├── opsctl/                 # CLI 主程序
│   │   ├── main.go             # 入口，cobra 根命令
│   │   ├── run.go              # opsctl run - 解释执行脚本
│   │   ├── build.go            # opsctl build - AOT 编译
│   │   ├── exec.go             # opsctl exec - 远程执行
│   │   └── repl.go             # opsctl repl - 交互式环境
│   └── ops-runner/             # 通用 Runner
│       └── main.go             # 从 stdin 读取 JSON 指令包执行
├── internal/
│   ├── lexer/                  # 词法分析
│   │   ├── token/
│   │   │   └── token.go        # 16 个关键字 + Token 类型定义
│   │   └── lexer.go            # 词法分析器
│   ├── parser/
│   │   └── parser.go           # 递归下降语法分析器
│   ├── ast/
│   │   └── ast.go              # AST 节点定义
│   ├── interpreter/
│   │   └── interpreter.go      # AST 遍历解释器，环境作用域链
│   ├── compiler/
│   │   ├── compiler.go         # 编译器主逻辑
│   │   ├── codegen.go          # AST -> Go 源码生成
│   │   ├── cache.go            # 编译缓存（hash + arch）
│   │   └── mode_selector.go    # Runner/AOT 自动模式选择
│   ├── sshx/                   # SSH 客户端封装
│   │   ├── client.go           # 连接管理
│   │   ├── pool.go             # 连接池
│   │   ├── sftp.go             # SFTP 传输（断点续传、压缩）
│   │   ├── config.go           # 连接配置
│   │   └── errors.go           # 错误定义
│   ├── exec/
│   │   └── exec.go             # Executor 执行调度，Target 定义
│   ├── inventory/
│   │   └── inventory.go        # YAML 主机清单解析
│   ├── arch/
│   │   └── arch.go             # uname -m -> GOARCH 映射
│   ├── runner/
│   │   ├── executor.go         # Runner 指令执行引擎
│   │   ├── instruction_gen.go  # AST -> JSON 指令包生成
│   │   ├── registry.go         # 操作注册表
│   │   └── types.go            # 指令/结果类型定义
│   ├── security/
│   │   ├── privilege.go        # 权限分级（read_only / admin / root）
│   │   ├── audit.go            # 审计日志（JSON 格式）
│   │   ├── signature.go        # Ed25519 签名验证
│   │   ├── resource_limit.go   # 资源限制（systemd-run / ulimit）
│   │   ├── retry.go            # 自动重试
│   │   └── cleanup.go          # 临时目录自清理
│   └── output/
│       └── output.go           # 结构化输出处理
├── pkg/
│   └── ops-core-sdk/           # 原子操作标准库（独立 Go module）
│       ├── sys/                # 系统信息
│       ├── file/               # 文件操作
│       ├── net/                # 网络操作
│       ├── process/            # 进程管理
│       ├── service/            # systemd 服务管理
│       ├── pkg/                # 包管理（apt/yum/dnf）
│       ├── json/               # JSON 编解码
│       ├── yaml/               # YAML 编解码
│       └── time/               # 时间操作
├── tests/                      # 集成测试
├── examples/                   # 示例脚本
├── docs/                       # 文档
├── go.mod
├── Makefile
├── .gitignore
└── README.md
```

---

## 3. 模块说明

### 3.1 语言前端

#### 词法分析器（`internal/lexer`）

词法分析器将源代码转换为 Token 流。

- **`token/token.go`**：定义 16 个关键字（`let`, `fn`, `if`, `else`, `for`, `while`, `return`, `true`, `false`, `task`, `on`, `ensure`, `report`, `alert`, `import`, `nil`）以及所有 Token 类型（标识符、字面量、运算符、分隔符等）
- **`lexer.go`**：逐字符扫描源代码，支持行号/列号追踪，错误报告精确定位

#### 语法分析器（`internal/parser`）

递归下降解析器，将 Token 流转换为 AST。

- 支持变量声明、函数定义、控制流（if/else/for/while）、函数调用
- 支持 `task ... on targets` 声明
- 支持 `ensure` 声明式幂等块
- 错误恢复机制，单个语法错误不阻塞后续解析

#### AST 节点（`internal/ast`）

所有语法结构的节点定义：

```
Program          -- 程序根节点
TaskDecl         -- task 声明
LetStmt          -- let 变量声明
FnDecl           -- 函数声明
IfStmt           -- if/else 条件
ForStmt          -- for 循环
WhileStmt        -- while 循环
ReturnStmt       -- return 语句
EnsureStmt       -- ensure 幂等块
ReportStmt       -- report 输出
AlertStmt        -- alert 告警
ImportStmt       -- import 导入
AssignStmt       -- 赋值语句
ExprStmt         -- 表达式语句
BinaryExpr       -- 二元表达式
UnaryExpr        -- 一元表达式
CallExpr         -- 函数调用
IndexExpr        -- 索引访问
MemberExpr       -- 成员访问
Ident            -- 标识符
IntLit/FloatLit/StringLit/BoolLit/NilLit  -- 字面量
ListLit/DictLit  -- 列表/字典字面量
```

### 3.2 解释器（`internal/interpreter`）

基于 AST 遍历的解释执行引擎。

- **作用域链**：`Environment` 结构体持有变量映射和父环境指针，支持闭包
- **内置函数注册**：`registerDefaults()` 注册所有标准库函数映射
- **类型系统**：支持整数、浮点数、字符串、布尔、列表、字典、结构体、函数
- **错误处理**：运行时错误携带 AST 节点位置信息

### 3.3 编译管线（`internal/compiler`）

将 OpsLang 脚本编译为静态二进制。

- **`compiler.go`**：编译流程编排 -- AST 分析 -> Go 代码生成 -> 调用 `go build` -> 输出二进制
- **`codegen.go`**：AST 到 Go 源码的翻译器
  - 变量声明 -> `var` / `:=`
  - 函数定义 -> Go `func`
  - 控制流 -> 对应 Go 语法
  - 内置函数调用 -> `ops-core-sdk` 直接调用
  - 自动生成 `main()` 函数
- **`cache.go`**：编译缓存，键为 `脚本内容 hash + 目标架构 + 标准库版本`，命中缓存直接复用
- **`mode_selector.go`**：自动模式选择启发式
  - 脚本 < 100 行且无 `import go "..."` -> Runner 模式
  - 否则 -> AOT 编译模式
  - 支持 `--mode runner|aot|auto` 覆盖

### 3.4 SSH 控制面（`internal/sshx`）

SSH 远程连接管理。

| 组件 | 职责 |
|------|------|
| `client.go` | SSH 连接建立，支持密码/密钥认证，超时控制 |
| `pool.go` | 连接池管理，复用连接，并发限制 |
| `sftp.go` | SFTP 文件传输，支持断点续传、压缩传输 |
| `config.go` | 连接配置（地址、端口、用户、认证方式） |
| `errors.go` | SSH 层错误定义 |

### 3.5 远程执行（`internal/exec` + `internal/runner`）

#### 执行调度（`internal/exec`）

- `Executor`：协调远程执行全流程
- `Target`：目标主机定义（地址、认证信息、标签）
- 并发执行多台主机，`--parallel` 参数控制并发数

#### Runner 指令处理（`internal/runner`）

- **`executor.go`**：Runner 侧指令执行引擎，从 stdin 读取 JSON 指令包，逐条执行，输出 JSON 结果
- **`instruction_gen.go`**：将 AST 中远程执行部分转换为 JSON 指令序列
- **`registry.go`**：操作注册表，映射操作名到 SDK 函数
- **`types.go`**：指令包和结果的结构体定义

#### 架构检测（`internal/arch`）

SSH 会话执行 `uname -m`，映射到 GOARCH：

```
x86_64   -> amd64
aarch64  -> arm64
armv7l   -> arm
```

结果缓存在 `~/.cache/opslang/runners/` 目录，按架构区分 Runner 二进制。

### 3.6 标准库（`pkg/ops-core-sdk`）

独立 Go module，可单独使用。每个包返回强类型结构体。

#### sys 包

| 函数 | 返回类型 | 说明 |
|------|---------|------|
| `sys.Hostname()` | `HostnameInfo` | 主机名 |
| `sys.OS()` | `OSInfo` | 操作系统信息 |
| `sys.Kernel()` | `KernelInfo` | 内核版本 |
| `sys.CPUUsage()` | `CPUUsage` | CPU 使用率 |
| `sys.CPUInfo()` | `CPUInfo` | CPU 核心数/型号 |
| `sys.MemoryInfo()` | `MemoryInfo` | 内存信息 |
| `sys.DiskUsage(path)` | `DiskUsage` | 磁盘使用率 |
| `sys.DiskPartitions()` | `[]DiskPartition` | 分区列表 |
| `sys.Load()` | `LoadAvg` | 负载均值 |
| `sys.Users()` | `[]User` | 登录用户 |
| `sys.Uptime()` | `Uptime` | 运行时长 |
| `sys.Processes()` | `[]ProcessInfo` | 进程列表 |
| `sys.NetInterfaces()` | `[]NetInterface` | 网络接口 |

#### file 包

| 函数 | 返回类型 | 说明 |
|------|---------|------|
| `file.Read(path)` | `FileContent` | 读取文件 |
| `file.Write(path, content, mode)` | `FileInfo` | 写入文件 |
| `file.Copy(src, dst)` | `FileInfo` | 复制文件 |
| `file.Move(src, dst)` | `FileInfo` | 移动文件 |
| `file.Delete(path)` | `DeleteResult` | 删除文件 |
| `file.Exists(path)` | `bool` | 是否存在 |
| `file.Stat(path)` | `FileInfo` | 文件信息 |
| `file.Chmod(path, mode)` | `FileInfo` | 修改权限 |
| `file.List(dir)` | `[]FileInfo` | 目录列表 |
| `file.Mkdir(path)` | `FileInfo` | 创建目录 |
| `file.Checksum(path, algo)` | `ChecksumResult` | 校验和 |

#### net 包

| 函数 | 返回类型 | 说明 |
|------|---------|------|
| `net.HTTPGet(url, headers)` | `HTTPResponse` | HTTP GET |
| `net.HTTPPost(url, body, headers)` | `HTTPResponse` | HTTP POST |
| `net.TCPConnect(host, port)` | `TCPResult` | TCP 连通检测 |
| `net.DNSLookup(domain)` | `DNSResult` | DNS 解析 |
| `net.Interfaces()` | `[]NetInterface` | 网络接口 |

#### process 包

| 函数 | 返回类型 | 说明 |
|------|---------|------|
| `process.List()` | `[]ProcessInfo` | 进程列表 |
| `process.FindByName(name)` | `[]ProcessInfo` | 按名称查找 |
| `process.FindByPort(port)` | `ProcessInfo` | 按端口查找 |
| `process.Exec(cmd, args...)` | `ExecResult` | 执行命令（不经 shell） |

#### service 包

| 函数 | 返回类型 | 说明 |
|------|---------|------|
| `service.Status(name)` | `ServiceStatus` | 服务状态 |
| `service.Start(name)` | `ServiceResult` | 启动 |
| `service.Stop(name)` | `ServiceResult` | 停止 |
| `service.Restart(name)` | `ServiceResult` | 重启 |
| `service.Enable(name)` | `ServiceResult` | 开机启动 |
| `service.Disable(name)` | `ServiceResult` | 禁用启动 |

#### pkg 包

| 函数 | 返回类型 | 说明 |
|------|---------|------|
| `pkg.Install(name, version)` | `PkgResult` | 安装 |
| `pkg.Remove(name)` | `PkgResult` | 卸载 |
| `pkg.Info(name)` | `PkgInfo` | 包信息 |
| `pkg.List()` | `[]PkgInfo` | 已安装包列表 |

#### json / yaml / time 包

- `json.Encode(v)` / `json.Decode(data, &v)` -- JSON 编解码
- `yaml.Encode(v)` / `yaml.Decode(data, &v)` -- YAML 编解码
- `time.Now()` / `time.Format(t, layout)` / `time.Parse(s, layout)` / `time.Since(t)` / `time.Sleep(d)` / `time.Diff(a, b)` -- 时间操作

### 3.7 安全模块（`internal/security`）

| 组件 | 职责 |
|------|------|
| `privilege.go` | 权限分级：`read_only` / `admin` / `root`，编译期 + 运行时双重校验 |
| `audit.go` | 审计日志：每次任务完整记录，JSON 输出，默认存储 `/var/log/opsctl/` |
| `signature.go` | Ed25519 签名验证：Runner 二进制签名，目标机可选校验 |
| `resource_limit.go` | 资源限制：优先 `systemd-run --scope`，回退 `ulimit` |
| `retry.go` | 自动重试：可配置次数和间隔 |
| `cleanup.go` | 自清理：临时目录 `/tmp/ops-<random>`，trap 保证异常退出也清理 |

---

## 4. 执行模式

### 4.1 Runner 模式（解释执行）

```
+------------------+
| source.ops       |
+--------+---------+
         |
         v
+------------------+
| Lexer            |   token/token.go
| 词法分析          |
+--------+---------+
         | Token 流
         v
+------------------+
| Parser           |   parser.go
| 语法分析          |
+--------+---------+
         | AST
         v
+------------------+
| Interpreter      |   interpreter.go
| 解释执行          |   环境作用域链
+--------+---------+
         | 结构化结果
         v
+------------------+
| JSON 输出         |
+------------------+
```

**使用场景**：
- `opsctl run script.ops` -- 本地解释执行
- `opsctl repl` -- 交互式环境
- 简单脚本（< 100 行，无第三方 Go 库依赖）
- 紧急故障处理（零编译延迟）

**特点**：
- 启动快，无需编译
- 直接遍历 AST 执行
- 适合调试和交互

### 4.2 AOT 编译模式

```
+------------------+
| source.ops       |
+--------+---------+
         |
         v
+------------------+
| Lexer + Parser   |
| 词法 + 语法分析   |
+--------+---------+
         | AST
         v
+------------------+
| Compiler         |   compiler.go
| 编译编排          |
+--------+---------+
         |
         v
+------------------+
| CodeGen          |   codegen.go
| AST -> Go 源码    |   映射到 ops-core-sdk
+--------+---------+
         | Go 源码
         v
+------------------+
| 编译缓存检查      |   cache.go
| key: hash+arch   |   命中则直接返回
+--------+---------+
         | 未命中
         v
+------------------+
| go build         |   CGO_ENABLED=0
| -ldflags "-s -w" |   GOOS/GOARCH 交叉编译
+--------+---------+
         |
         v
+------------------+
| 静态二进制        |   零依赖，可分发
+------------------+
```

**使用场景**：
- `opsctl build --output bin --target-arch linux/arm64`
- 复杂业务逻辑
- 需要第三方 Go 库（`import go "..."`）
- 生产环境部署

**特点**：
- 生成静态二进制，目标机零依赖
- 支持交叉编译（amd64 -> arm64）
- 编译缓存加速重复构建

### 4.3 模式自动选择

```
                    +-------------------+
                    | 脚本输入           |
                    +--------+----------+
                             |
                             v
                    +-------------------+
                    | mode_selector.go  |
                    | 启发式分析         |
                    +--------+----------+
                             |
              +--------------+--------------+
              |                             |
              v                             v
    +------------------+          +------------------+
    | 脚本 < 100 行     |          | 脚本 >= 100 行   |
    | 无 import go     |          | 或有 import go   |
    | -> Runner 模式    |          | -> AOT 编译模式  |
    +------------------+          +------------------+
```

用户可通过 `--mode runner|aot|auto` 强制指定。

### 4.4 远程执行流程

```
opsctl exec --hosts h1,h2,h3 --instructions task.json
    |
    v
+---------------------------+
| 解析目标主机               |
| --hosts 或 --inventory    |
| (inventory.go YAML 解析)  |
+------------+--------------+
             |
             v
+---------------------------+
| 加载 JSON 指令包          |
+------------+--------------+
             |
             v
+---------------------------+
| 并发执行（--parallel 限流）|
+------------+--------------+
             |
     +-------+-------+-------+
     |               |       |
     v               v       v
  [host1]         [host2] [host3]
     |               |       |
     v               v       v
  1. SSH 连接    (sshx 连接池)
  2. 架构检测    (arch.go: uname -m -> GOARCH)
  3. 上传 Runner (缓存: ~/.cache/opslang/runners/)
  4. stdin 发送 JSON 指令包
  5. stdout 收集 JSON 结果
     |               |       |
     v               v       v
+---------------------------+
| 结果聚合 -> JSON 汇总     |
+---------------------------+
```

Runner 二进制按架构缓存，同架构第二次执行不重复上传。

---

## 5. 数据流

### 5.1 本地执行数据流

```
source.ops (源代码)
    |
    | Lexer: 逐字符扫描
    v
[]token.Token (Token 流)
    |
    | Parser: 递归下降
    v
*ast.Program (AST)
    |
    | Interpreter: 深度优先遍历
    v
*Object (运行时值)
    |
    | output.go: 序列化
    v
JSON 输出 (stdout)
```

### 5.2 远程执行数据流

```
[控制端]                              [目标端]

task.json                             ops-runner (缓存)
    |                                      |
    |  SSH/SFTP 上传 Runner（如未缓存）      |
    |------------------------------------->|
    |                                      |
    |  stdin: JSON 指令包                   |
    |------------------------------------->|
    |                                      |
    |                            +---------+---------+
    |                            | Runner Executor   |
    |                            | 逐条执行指令       |
    |                            | 调用 ops-core-sdk |
    |                            +---------+---------|
    |                                      |
    |  stdout: JSON 结果                    |
    |<-------------------------------------|
    |                                      |
    v
结果聚合器
JSON 汇总输出
```

### 5.3 Runner 指令包 JSON 格式

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

### 5.4 Runner 输出 JSON 格式

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

### 5.5 deploy 结果汇总格式

```json
{
  "task_id": "abc123",
  "script": "check_cpu.ops",
  "targets": ["host1", "host2"],
  "started_at": "2026-08-15T10:00:00Z",
  "finished_at": "2026-08-15T10:00:12Z",
  "results": {
    "host1": { "status": "success", "exit_code": 0, "data": {} },
    "host2": { "status": "failed", "exit_code": 1, "error": "timeout" }
  },
  "audit_log": "/var/log/opsctl/abc123.json"
}
```

---

## 6. 扩展指南

### 6.1 添加新的标准库函数

1. 在 `pkg/ops-core-sdk/<包名>/` 下创建函数，返回带 JSON tag 的结构体：

```go
// pkg/ops-core-sdk/sys/swap.go

package sys

// SwapInfo 交换分区信息
type SwapInfo struct {
    Total uint64  `json:"total"`
    Used  uint64  `json:"used"`
    Free  uint64  `json:"free"`
}

// Swap 返回交换分区信息
func Swap() (*SwapInfo, error) {
    // 实现...
}
```

2. 编写单元测试，确保覆盖率 >= 80%
3. 验证交叉编译：`CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./pkg/ops-core-sdk/...`

### 6.2 添加新的内置函数

在 `internal/interpreter/interpreter.go` 的 `registerDefaults()` 中注册：

```go
func (i *Interpreter) registerDefaults() {
    // ... 现有注册 ...

    i.registerBuiltin("sys.swap", func(args ...Object) (Object, error) {
        info, err := sys.Swap()
        if err != nil {
            return nil, err
        }
        return structToMap(info), nil
    })
}
```

### 6.3 添加新的语句类型

按以下顺序完成：

1. **定义 AST 节点**（`internal/ast/ast.go`）：

```go
type MyNewStmt struct {
    Token    token.Token
    Name     string
    Body     *BlockStmt
}
```

2. **添加解析规则**（`internal/parser/parser.go`）：

```go
func (p *Parser) parseMyNewStmt() Statement {
    // 解析逻辑...
}
```

在 `parseStatement()` 中增加分支。

3. **实现解释执行**（`internal/interpreter/interpreter.go`）：

```go
func (i *Interpreter) evalMyNewStmt(stmt *ast.MyNewStmt) (Object, error) {
    // 执行逻辑...
}
```

在 `eval()` 的 switch 中增加分支。

4. **实现代码生成**（`internal/compiler/codegen.go`）：

```go
func (g *Generator) genMyNewStmt(stmt *ast.MyNewStmt) {
    // Go 代码生成...
}
```

### 6.4 添加新的 CLI 命令

在 `cmd/opsctl/` 下新增文件，使用 cobra：

```go
// cmd/opsctl/mycommand.go

package cmd

import "github.com/spf13/cobra"

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "命令描述",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 实现...
        return nil
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

### 6.5 模块依赖关系

```
opsctl (cmd/opsctl)
├── internal/lexer
├── internal/parser
│   ├── internal/lexer
│   └── internal/ast
├── internal/interpreter
│   └── internal/ast
├── internal/compiler
│   ├── internal/ast
│   └── pkg/ops-core-sdk
├── internal/exec
│   ├── internal/sshx
│   ├── internal/inventory
│   └── internal/runner
├── internal/sshx
├── internal/runner
│   └── pkg/ops-core-sdk
└── internal/security

ops-runner (cmd/ops-runner)
├── internal/runner
│   └── pkg/ops-core-sdk
└── pkg/ops-core-sdk (独立 module，可单独使用)
```

扩展时注意依赖方向：`internal/*` 依赖 `pkg/ops-core-sdk`，但 `pkg/ops-core-sdk` 不依赖任何 `internal/*` 包。标准库保持独立可用。
