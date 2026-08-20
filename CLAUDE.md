

牢记准则：
基础规则：
1. 不确定/查不到的直接说不知道，不许瞎编
2. 涉及数字、日期、价格、政策，主动标注信息来源是否权威
3. 用大白话说话，别整没意义的空话
4. 先给结论再讲理由，别铺垫半天没重点
5. 我需求没说清先问我，别自己瞎猜

【说话方式】直白通俗，别用空洞话术，简明扼要。
【回答规矩】不确定就说不知道，涉及敏感信息标可信度，需求不清先确认，先给结论再讲理由

代码编写准则：
# Code Quality Directives (Anti-Hallucination)

1. **No Silent Failures**: Forbid `try: ... except: pass` or `except Exception`. Must log or re-raise.

2. **Null Safety**: Explicitly handle `None`, empty lists, and out-of-bounds. Never assume valid input.

3. **No Fake APIs**: Do not invent function/class names. If unsure about a library method, ask before coding. No extra params.

4. **Tests Must Assert**: Provide `assert` statements. Include at least one edge-case test (null, overflow, empty).

5. **Design First**: For >100 lines, output pseudo-code/logic first. Wait for confirmation before full code.

6. **Honest Ignorance**: If uncertain about implementation details, add `# UNCERTAIN: verify this logic`. Do not bluff.

7. **Comments on "Why"**: Explain why, not what. Keep under 3 lines unless strictly necessary.


项目地址：https://github.com/j4ckzh0u/opslang , 开发好了,测试通过，就推送GitHub，并创建github action，进行自动化构建。
要关注一下GitHub action运行的结果，如果出现错误，请及时的修复。
一定要保证程序能在GitHub action上编译通过。

代码开发中，不使用子智能体，因为API访问的速度有限制，超过了就会出现APIError 429。
一定要记住，不使用子智能体。每次API交互前 等待 2s 
---
项目要求：
# OpsLang 项目完整开发计划

> 本计划用于指导 Claude Code 在仓库 https://github.com/j4ckzh0u/opslang/ 中开发 OpsLang 项目。  
> 请严格按照阶段顺序执行，每个阶段完成后进行验证并提交代码。

---

## 1. 项目概述

OpsLang 是一门面向运维领域的领域特定语言（DSL），旨在通过简洁的脚本完成复杂的运维操作。它彻底摆脱 Shell 字符串处理和 Python 环境依赖，提供结构化返回、声明式幂等、双执行引擎（解释执行 + 编译执行）、零依赖远程执行、异构架构支持、大规模文件分发与收集等能力。

### 核心价值
- **极简脚本**：少量代码描述完整业务意图，无需手动解析命令输出。
- **原子操作封装**：标准库提供结构化返回的系统、文件、网络、服务等操作函数。
- **双执行引擎**：通用 Runner（指令包解释执行）与 AOT 编译（静态二进制）自适应，兼顾分发效率与执行性能。
- **零依赖远程执行**：通过 SSH 下发预编译 Runner 或定制二进制，目标机无需安装 Agent 或运行时。
- **声明式幂等**：内置 `ensure` 语法，支持 dry-run 与状态收敛。
- **双向文件传输**：大规模文件分发与收集，支持分层中继、断点续传、内容哈希去重。
- **异构架构**：纯 Go 实现，交叉编译覆盖 amd64/arm64 等主流架构。
- **企业级安全**：权限分级、审批流、审计日志、签名验证、资源限制内置。
- **极简配置**：脚本即配置，自动发现主机，智能默认值，后台自动完成复杂决策。

## 2. 设计原则

1. **由内向外**：先开发原子操作 SDK，确保核心价值独立可用，再实现语言前端。
2. **纯 Go 生态**：所有依赖支持 `CGO_ENABLED=0` 交叉编译，无 cgo。
3. **最小可用闭环**：优先打通“脚本 → 编译/指令包 → SSH 下发 → 远程执行 → 结构化回传”全链路。
4. **极简与自动化**：用户只需编写描述性脚本，系统自动完成架构检测、中继选择、模式决策、重试清理等。
5. **结构化优先**：标准库函数一律返回结构体，禁止返回未经处理的原始文本。
6. **安全内置**：权限分级、审批流、审计日志、签名验证、资源限制从设计之初就纳入。

## 3. 系统架构

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

### 执行模式自动选择

| 场景 | 选择模式 | 原因 |
|------|---------|------|
| 简单任务（<100行，无第三方 Go 库） | Runner 模式 | 逻辑简单，指令包分发高效 |
| 复杂业务逻辑或需要第三方 Go 库 | AOT 编译模式 | 需要 Go 生态支持 |
| 紧急故障处理 | Runner 模式 | 零编译延迟，即时下发 |
| 用户强制指定 | `--mode runner/aot` | 手动覆盖 |

## 4. 关键技术决策

- **语言语法**：类 Go + Python 混合，静态类型 + 类型推断，关键字 ≤ 15 个。
- **标准库**：独立 Go module，可单独使用；所有函数返回结构体，不返回原始字符串。
- **远程执行**：通用 Runner 首次上传缓存，后续仅发送 JSON 指令包；AOT 二进制仅在需要时编译并上传。
- **异构架构**：强制 `CGO_ENABLED=0`，提供多架构 Runner，SSH 自动检测架构。
- **大规模文件传输**：分层中继 + 压缩传输 + 断点续传 + 内容哈希去重，支持 HTTP/对象存储下载/上传通道，双向对称设计。
- **极简配置**：脚本内联目标，自动发现主机，后台自动决策，无需额外配置文件。

## 5. 开发阶段计划

### Phase 0：原子操作 SDK（ops-core-sdk）— 4 周

**目标**：开发独立可用的 Go 库，提供常用运维原子操作，返回强类型结构体，支持交叉编译。

**任务**：
1. 创建 Go module `github.com/j4ckzh0u/opslang`，初始化目录结构（见第 6 节）。
2. 定义统一错误处理规范：每个函数返回 `(T, error)`，错误包含明确错误码和消息。
3. 实现标准库包（函数列表详见附录 A）：
   - `sys`：CPU、内存、磁盘、负载、主机名、用户、进程等（基于 `gopsutil`）。
   - `file`：读写、复制、移动、删除、权限、模板渲染、校验和、原子写入。
   - `net`：HTTP GET/POST、TCP 连通性、DNS 解析、网络接口。
   - `process`：进程列表、查找、启动/停止外部命令（不使用 shell）。
   - `service`：systemd 服务管理。
   - `pkg`：包管理封装（apt/yum/dnf）。
   - `json` / `yaml`：编解码。
   - `time`：当前时间、格式化、时间差。
4. 每个函数单元测试，并在 amd64/arm64 下交叉编译验证。
5. 编写 godoc 文档。

**交付物**：
- 可独立使用的 Go 库，测试覆盖率 ≥ 80%，交叉编译通过。
- 目录结构完整，`Makefile` 可用。

**验收标准**：
- `go test ./...` 全部通过。
- 在 Linux amd64/arm64 下 `CGO_ENABLED=0 go build ./pkg/ops-core-sdk/...` 成功。
- 所有函数均不依赖 shell，信息获取直接使用 Go 库或读取 `/proc`/`/sys`。
- `README.md` 包含项目简介和标准库函数列表。

---

### Phase 1：远程执行通道（SSH 控制面 + 通用 Runner）— 4 周

**目标**：打通 SSH 远程执行链路，实现通用 Runner 与 JSON 指令协议，支持架构检测与缓存。

**任务**：
1. SSH 客户端封装（`internal/sshx`）：
   - 支持密码、密钥认证，连接超时、执行超时、重试。
   - 连接池管理，支持并发限制。
   - SFTP 上传/下载（支持断点续传、压缩传输）。
2. 架构检测：
   - SSH 会话执行 `uname -m`，映射到 GOARCH。
   - 提供 `DetectArch(conn) (string, error)`，结果缓存。
3. 通用 Runner 开发（`cmd/ops-runner`）：
   - 内嵌 `ops-core-sdk` 全部函数。
   - 从 stdin 读取 JSON 指令包，按顺序执行。
   - 输出 JSON 结果到 stdout。
   - 支持 `--dry-run` 标志。
4. Runner 多架构构建：
   - 构建 `ops-runner-linux-amd64`、`ops-runner-linux-arm64`。
   - 使用 `-ldflags "-s -w"` 减小体积，可选 zstd 压缩。
5. `opsctl exec` 命令：
   - 参数：`--hosts`、`--user`、`--key`、`--inventory`、`--instructions`。
   - 流程：SSH 连接 → 架构检测 → 上传/复用 Runner → 发送指令包 → 收集输出。
   - 并发执行多台主机，自动限流。

**交付物**：
- 可执行 `opsctl`，通用 Runner 二进制，端到端示例。
- `opsctl exec` 可用。

**验收标准**：
- 远程主机无预装环境，执行 `opsctl exec --instructions test.json` 成功返回结构化结果。
- 同架构第二次执行不重复上传 Runner（缓存生效）。
- 支持并发 10 台主机执行并正确汇总。

---

### Phase 2：语言前端与解释器 — 4 周

**目标**：定义 OpsLang 语法，实现手写词法/语法分析器和 AST 解释器，支持本地脚本执行。

**任务**：
1. 语法设计（关键字 ≤ 15）：
   - 变量声明 `let`，函数 `fn`，控制流 `if/else`、`for`、`while`。
   - 数据类型：整数、浮点数、字符串、布尔、列表、字典、结构体。
   - 内置函数调用：如 `sys.cpu.usage()`。
   - 支持 `task ... on targets` 声明（解析后暂不远程执行）。
   - 支持函数默认参数。
2. 词法分析器（`internal/lexer`）：
   - 识别关键字、标识符、数字、字符串、运算符、注释。
   - 错误报告包含行列号。
3. 语法分析器（`internal/parser`）：
   - 递归下降解析，生成 AST。
4. 解释器（`internal/interpreter`）：
   - 遍历 AST 执行，实现变量作用域、函数调用、控制流。
   - 内置函数注册机制，映射到 `ops-core-sdk`。
   - 支持 `print`、`report` 等输出。
5. REPL：
   - `opsctl repl` 交互式环境。
6. `opsctl run`：
   - `opsctl run script.ops` 本地解释执行。

**交付物**：
- 可解释执行基本脚本的 CLI。
- 至少 10 个示例脚本（获取系统信息、文件操作、进程管理等）。

**验收标准**：
- 脚本能使用变量、循环、条件判断，调用标准库函数并打印结构化结果。
- REPL 正常交互。
- 错误提示能定位到具体行列。

---

### Phase 3：AOT 编译管线 — 4 周

**目标**：将 OpsLang 脚本编译为静态二进制，支持多架构交叉编译和缓存。

**任务**：
1. AST → Go 源码生成器（`internal/compiler`）：
   - 将 AST 翻译为 Go 代码，正确映射变量、函数、控制流。
   - 内置函数调用编译为对 `ops-core-sdk` 的直接调用。
   - 生成 `main()` 函数，处理输入参数、调用用户逻辑、输出 JSON 结果。
2. 编译封装：
   - 动态生成临时 Go 项目，引入 `ops-core-sdk`。
   - 调用 `go build -ldflags "-s -w" -o output`。
   - 支持 `GOOS/GOARCH` 设置实现交叉编译。
3. 编译缓存：
   - 以脚本 hash + 标准库版本 + 目标架构为键缓存。
   - 命中缓存直接使用。
4. `opsctl build`：
   - 参数：`--output`、`--target-arch`（默认当前架构）。
   - 输出静态二进制。

**交付物**：
- `opsctl build` 命令，编译缓存机制。

**验收标准**：
- 编译出的二进制在无 Go 环境目标机运行，输出与解释执行一致。
- 交叉编译 amd64→arm64 成功，二进制可在 ARM 服务器运行。
- 相同脚本第二次编译 < 5 秒（缓存生效）。

---

### Phase 4：远程编排与声明式特性 — 6 周

**目标**：实现 `task ... on targets` 远程执行、Runner 指令包生成、自动模式选择、`ensure` 幂等、dry-run、结构化输出、大规模文件分发与收集。

**任务**：
1. 语法扩展：
   - 支持 `task "name" on <targets>`，`targets` 可为内联主机列表、环境变量、`group()` 表达式。
   - 支持 `parallel` 块。
2. 指令包生成器：
   - 将 AST 中远程执行部分转换为 JSON 指令序列。
   - 自动识别本地/远程代码边界（标准库函数带 `scope` 元数据）。
3. 自动模式选择：
   - 启发式规则：脚本行数 <100 且无 `import go "..."` 默认 Runner 模式，否则 AOT。
   - 支持 `--mode auto` 默认。
4. 声明式幂等 `ensure`：
   - 将 `ensure <condition> { actions }` 转换为 `check → apply → verify`。
   - 支持 `notify` 触发器。
5. dry-run 注入：
   - 为每个变更操作生成 dry-run 分支，只输出操作描述。
6. 结构化输出：
   - `report { key: value }`、`metric(name, value, labels)`、`log(msg)`、`alert(msg)`。
7. `opsctl deploy`：
   - 整合 inventory 解析、远程执行、结果聚合。
   - 参数：`--targets`、`--parallel`、`--dry-run`、`--mode`。
   - 输出 JSON 汇总结果。
8. **大规模文件分发与收集专项**（对称设计）：
   - 实现 `file.distribute(source, dest, ...)` 原子操作，自动压缩、去重、断点续传、原子替换。
   - 实现 `file.collect(source, dest, ...)` 原子操作，自动打包、压缩、断点续传、按主机归档。
   - 分层中继选择算法：根据 IP 前缀/网络拓扑自动选择中继节点，支持双向传输。
   - 支持 HTTP/对象存储下载/上传通道（可选）。
   - 自动并发调度与限流，失败重试。
   - 1 万级模拟测试（分发与收集场景）。

**交付物**：
- 功能完整的 `opsctl deploy`。
- 大规模文件分发与收集能力。
- 示例与文档。

**验收标准**：
- 脚本可操作多台主机，支持并发，正确聚合结果。
- `ensure` 重复执行无副作用。
- `--dry-run` 不改变系统状态，输出预览。
- `file.distribute` 在 1 万主机模拟中，控制端带宽占用合理，成功率 >99.9%。
- `file.collect` 在 1 万主机模拟中，高效收集文件并自动归档，成功率 >99.9%。
- 结构化输出可被解析（如导入 Elasticsearch）。

---

### Phase 5：安全与生产化 — 4 周

**目标**：加入权限分级、审批流、审计日志、签名验证、资源限制、自动清理。

**任务**：
1. 权限分级：
   - 脚本头部 `privilege: read_only | admin | root`。
   - 编译期检查：`read_only` 脚本调用变更类函数报错。
   - Runner 运行时二次校验。
2. 审批流：
   - 当脚本权限为 `admin`/`root` 且目标主机标签含生产环境时，自动要求审批。
3. 审计日志：
   - 记录每次任务完整信息，输出 JSON 文件或 syslog。
   - 默认存储到本地固定目录。
4. Runner 签名验证：
   - 控制端生成 Runner 时使用 Ed25519 签名。
   - 目标机可预置公钥，Runner 启动前验证（可选）。
5. 资源限制：
   - 优先使用 `systemd-run --scope` 限制 CPU/内存。
   - 无 systemd 时回退到 `ulimit`。
6. 临时目录自清理：
   - 上传文件到 `/tmp/ops-<random>`，执行后删除，设置 trap 确保异常退出也清理。
7. 自动重试与回滚：
   - 任务失败自动重试，变更操作失败尝试回滚。

**交付物**：
- 具备完整安全特性的 `opsctl`。
- 审计日志示例。
- 资源限制测试报告。

**验收标准**：
- `read_only` 脚本无法执行变更操作。
- 高危操作需审批。
- 审计日志完整可回溯。
- 自清理无残留，资源限制生效。

---

## 6. 项目结构

```
opslang/
├── cmd/
│   ├── opsctl/                 # CLI 主程序
│   │   └── main.go
│   └── ops-runner/             # 通用 Runner 程序
│       └── main.go
├── internal/
│   ├── lexer/                  # 词法分析
│   ├── parser/                 # 语法分析
│   ├── ast/                    # AST 节点定义
│   ├── interpreter/            # 解释器
│   ├── compiler/               # Go 代码生成器
│   ├── sshx/                   # SSH 客户端封装
│   ├── inventory/              # 主机清单与自动发现
│   ├── runner/                 # Runner 指令包处理
│   ├── relay/                  # 分层中继调度
│   ├── security/               # 签名、权限
│   └── output/                 # 结构化输出处理
├── pkg/
│   └── ops-core-sdk/           # 原子操作标准库
│       ├── sys/
│       ├── file/
│       ├── net/
│       ├── process/
│       ├── service/
│       ├── pkg/
│       ├── json/
│       ├── yaml/
│       └── time/
├── tests/                      # 集成测试
├── examples/                   # 示例脚本
├── docs/                       # 文档
├── go.mod                      # module github.com/j4ckzh0u/opslang
├── Makefile
├── .gitignore
└── README.md
```

## 7. 关键数据格式定义

### 7.1 标准库函数返回结构体示例

```go
// sys.CPUUsage 返回
type CPUUsage struct {
    Percent float64 `json:"percent"`
    User    float64 `json:"user"`
    System  float64 `json:"system"`
    Idle    float64 `json:"idle"`
}

// sys.MemoryInfo 返回
type MemoryInfo struct {
    Total       uint64  `json:"total"`
    Available   uint64  `json:"available"`
    Used        uint64  `json:"used"`
    UsedPercent float64 `json:"used_percent"`
}

// file.DistributeResult 返回
type DistributeResult struct {
    Host        string `json:"host"`
    Status      string `json:"status"` // success/skipped/failed
    Changed     bool   `json:"changed"`
    Checksum    string `json:"checksum"`
    DurationMs  int64  `json:"duration_ms"`
    Error       string `json:"error,omitempty"`
}

// file.CollectResult 返回
type CollectResult struct {
    Host        string `json:"host"`
    Status      string `json:"status"` // success/skipped/failed
    Source      string `json:"source"`
    Dest        string `json:"dest"`
    Checksum    string `json:"checksum"`
    Size        int64  `json:"size"`
    DurationMs  int64  `json:"duration_ms"`
    Error       string `json:"error,omitempty"`
}
```

### 7.2 Runner 指令包 JSON 格式

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

### 7.3 Runner 输出 JSON 格式

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

### 7.4 opsctl deploy 结果汇总格式

```json
{
  "task_id": "abc123",
  "script": "check_cpu.ops",
  "targets": ["host1", "host2"],
  "started_at": "2026-08-15T10:00:00Z",
  "finished_at": "2026-08-15T10:00:12Z",
  "results": {
    "host1": { "status": "success", "exit_code": 0, "data": { ... } },
    "host2": { "status": "failed", "exit_code": 1, "error": "timeout" }
  },
  "audit_log": "/var/log/opsctl/abc123.json"
}
```

## 8. 极简配置与自动化要求（贯穿各阶段）

- **默认值丰富**：标准库函数支持默认参数，调用时省略大部分参数可正常工作。
- **自动发现主机**：支持从 SSH config、云 API、Kubernetes、CMDB 自动获取主机列表，避免手动 inventory。
- **自动架构检测**：SSH 连接后自动执行 `uname -m`，缓存结果，无需用户指定。
- **自动模式选择**：根据脚本特征自动选择 Runner 或 AOT，用户可强制覆盖。
- **自动分层分发/收集**：大规模文件传输时自动选择中继节点，无需人工规划。
- **自动重试与清理**：失败自动重试，临时文件自动清理，无需配置。
- **审计自动生成**：每次任务自动生成审计日志，无需额外开关。

## 9. 整体验收标准

1. **端到端闭环**：用户编写一个 OpsLang 脚本，执行 `opsctl deploy`，即可在多种架构的远程主机上完成操作并返回结构化结果。
2. **零配置演示**：仅用一个脚本文件和一条命令，完成向 100 台以上异构主机分发文件并触发后续操作，全程无需 YAML/JSON 配置。
3. **文件收集**：从 100 台以上主机收集指定文件，自动归档到控制端，全程无需额外配置。
4. **安全合规**：权限分级、审批、审计、签名验证均正常工作。
5. **性能达标**：1 万主机文件分发/收集模拟中，控制端带宽占用合理，成功率 >99.9%，耗时在可接受范围。
6. **代码质量**：所有包测试覆盖率 ≥ 80%，`go vet` 无警告，交叉编译测试通过。

## 10. 里程碑与时间估算

| 阶段 | 内容 | 预计时间 | 关键交付物 |
|------|------|---------|-----------|
| Phase 0 | 原子操作 SDK | 4 周 | ops-core-sdk |
| Phase 1 | 远程执行通道 | 4 周 | opsctl exec + Runner |
| Phase 2 | 语言前端与解释器 | 4 周 | opsctl run + REPL |
| Phase 3 | AOT 编译管线 | 4 周 | opsctl build |
| Phase 4 | 远程编排与声明式 | 6 周 | opsctl deploy + 文件分发/收集 |
| Phase 5 | 安全与生产化 | 4 周 | 企业级特性 |
| **合计** | | **26 周** | 可产品化 MVP |


# 对标产品
最重要：OpsLang对标产品是ansible，功能一定要覆盖且增强于ansbile。OpsLang是可以在没有python、shell的环境中正常运行。而且速度很快，支持海量服务器并发操作。

---

## 附录 A：标准库函数列表

> **权威来源**：`internal/opsspec/spec.go` 是全部原子操作的单一事实来源（名称、参数、可用范围），三套执行引擎（解释器、runner registry、AOT codegen）由一致性测试强制对齐。下表与之一致；发现不一致以 opsspec 为准。

### sys 包
| 函数 | 返回结构体 | 说明 |
|------|-----------|------|
| `sys.hostname()` | `HostnameInfo` | 主机名 |
| `sys.os()` | `HostInfoResult` | 操作系统/平台/内核信息（含内核版本字段） |
| `sys.cpu.usage()` | `CPUUsage` | CPU 使用率（500ms 两次采样，反映当前值） |
| `sys.cpu.count()` | `CPUCount` | 逻辑/物理核心数 |
| `sys.cpu.info()` | `[]CPUInfo` | CPU 型号列表 |
| `sys.memory.info()` | `MemoryInfo` | 内存信息 |
| `sys.disk.usage(path)` | `DiskUsage` | 磁盘使用率 |
| `sys.disk.partitions()` | `[]DiskPartition` | 磁盘分区列表 |
| `sys.load()` | `LoadAvg` | 负载均值 |
| `sys.users()` | `[]UserInfo` | 当前登录用户 |
| `sys.uptime()` | `UptimeInfo` | 运行时长 |
| `sys.net.interfaces()` | `[]NetInterface` | 网络接口 |

### file 包
| 函数 | 返回 | 说明 |
|------|------|------|
| `file.read(path)` | `FileContent` | 读取文件内容 |
| `file.write(path, content)` | `WriteResult` | 写入文件（0644） |
| `file.append(path, content)` | `AppendResult` | 追加内容 |
| `file.copy(src, dst)` | `CopyResult` | 复制文件 |
| `file.move(src, dst)` | `MoveResult` | 移动文件 |
| `file.delete(path)` | `DeleteResult` | 删除文件 |
| `file.exists(path)` | `ExistsResult` | 文件是否存在（.exists 字段） |
| `file.stat(path)` | `FileInfo` | 文件元信息 |
| `file.list(dir)` | `ListResult` | 列目录 |
| `file.mkdir(path)` | `MkdirResult` | 创建目录（幂等） |
| `file.chmod(path, mode)` | `ChmodResult` | 改权限（mode 为八进制字符串） |
| `file.checksum(path, algo)` | `ChecksumResult` | 计算校验和（md5/sha1/sha256） |
| `file.template(path, vars)` | `TemplateResult` | 渲染 {{key}} 占位符，不修改源文件 |
| `file.distribute(source, targets, options)` | `DistributeResult` | 多主机分发（仅控制器侧，真实 SSH/SFTP） |
| `file.collect(source, targets, options)` | `CollectResult` | 多主机收集（仅控制器侧） |

### net 包
| 函数 | 返回 | 说明 |
|------|------|------|
| `net.http_get(url)` | `HTTPResponse` | HTTP GET |
| `net.http_post(url, body)` | `HTTPResponse` | HTTP POST（JSON body） |
| `net.tcp_check(host, port)` | `TCPResult` | TCP 连通性 |
| `net.dns_lookup(domain)` | `DNSResult` | DNS 解析 |
| `net.interfaces()` | `[]NetInterface` | 网络接口列表 |

### process 包
| 函数 | 返回 | 说明 |
|------|------|------|
| `process.list()` | `[]ProcessInfo` | 进程列表 |
| `process.find_by_name(name)` | `[]ProcessInfo` | 按名称查找 |
| `process.find_by_port(port)` | `[]ProcessInfo` | 按端口查找 |
| `process.exec(command, args...)` | `ExecResult` | 执行外部命令（不经 shell） |
| `process.kill(pid, signal)` | `KillResult` | 发送信号（默认 TERM；TERM/KILL/HUP/INT/USR1/USR2） |

### service 包
| 函数 | 返回 | 说明 |
|------|------|------|
| `service.status(name)` | `ServiceStatus` | 服务状态 |
| `service.start(name)` | `ServiceResult` | 启动服务 |
| `service.stop(name)` | `ServiceResult` | 停止服务 |
| `service.restart(name)` | `ServiceResult` | 重启服务 |
| `service.enable(name)` | `ServiceResult` | 设置开机启动 |
| `service.disable(name)` | `ServiceResult` | 取消开机启动 |

### pkg 包
| 函数 | 返回 | 说明 |
|------|------|------|
| `pkg.install(name)` | `PkgResult` | 安装软件包（apt/yum/dnf，仅 Linux） |
| `pkg.remove(name)` | `PkgResult` | 卸载软件包 |
| `pkg.info(name)` | `PkgInfo` | 查询软件包 |
| `pkg.list()` | `[]PkgInfo` | 已安装包列表 |

### json / yaml / time 包
| 函数 | 返回 | 说明 |
|------|------|------|
| `json.encode(value)` / `json.decode(input)` | `EncodeResult` / 任意值 | JSON 编解码 |
| `yaml.encode(value)` / `yaml.decode(input)` | `EncodeResult` / 任意值 | YAML 编解码 |
| `time.now()` | `TimeInfo` | 当前时间 |
| `time.format(ts, layout)` | `FormatResult` | 格式化时间戳 |
| `time.parse(layout, value)` | `TimeInfo` | 解析时间字符串 |
| `time.diff(t1, t2)` | `DiffResult` | 时间差 |
| `time.since(ts)` | `DurationInfo` | 距今时长 |
| `time.sleep(ms)` | `SleepResult` | 睡眠 |

## 附录 B：OpsLang 语法示例

```go
// 简单采集脚本
task "collect_info" on targets {
    let cpu = sys.cpu.usage()
    let mem = sys.memory.info()
    let disks = sys.disk.usage("/")

    if cpu.percent > 90 {
        alert("CPU 使用率过高: " + str(cpu.percent) + "%")
    }

    report {
        host: sys.hostname(),
        cpu: cpu,
        mem: mem,
        disk: disks
    }
}
```

```go
// 文件分发脚本（大规模）
task "deploy_app" on group("env=prod") {
    file.distribute(
        source: "/data/releases/app-v2.1.0.tar.gz",
        dest: "/opt/app/releases/",
        mode: "0644",
        owner: "appuser",
        on_changed: {
            sys.service.restart("myapp")
        }
    )
}
```

```go
// 文件收集脚本
task "collect_logs" on group("role=web") {
    file.collect(
        source: "/var/log/nginx/access.log",
        dest: "/data/logs/{host}/access_{date}.log",
        compress: true,
        if_changed: true
    )
}
```

---

