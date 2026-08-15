
对claudecode的规定：
开发新功能、测试新功能以及审查代码等等，务必派生子agent进行。保持主Agent的上下文的干净整洁

牢记准则：
基础规则：
1. 不确定/查不到的直接说不知道，不许瞎编
2. 涉及数字、日期、价格、政策，主动标注信息来源是否权威
3. 用大白话说话，别整没意义的空话
4. 先给结论再讲理由，别铺垫半天没重点
5. 我需求没说清先问我，别自己瞎猜

【说话方式】直白通俗，别用空洞话术
【回答规矩】不确定就说不知道，涉及敏感信息标可信度，需求不清先确认，先给结论再讲理由

项目地址：https://github.com/j4ckzh0u/opslang , 开发好了就推送GitHub，并创建github action，进行自动化构建。
---
项目要求：
# OpsLang 开发计划（供 Claude Code 执行）

## 1. 项目概述

OpsLang 是一门面向运维领域的领域特定语言（DSL）。它允许运维人员编写简洁的脚本，通过调用封装好的原子操作（如获取系统信息、管理文件、启停服务等）完成复杂的运维业务逻辑，彻底摆脱对 Shell 字符串解析和 Python 环境依赖。

核心特性：
- **脚本即操作**：简单易学的语法，无指针/接口/泛型，专注于描述“对哪些机器做什么”。
- **结构化返回**：所有原子操作返回强类型结构体，无需 grep/awk/sed 提取字段。
- **双执行引擎**：解释执行（Runner 模式，指令包下发）与编译执行（AOT 模式，编译为静态二进制）自适应。
- **零依赖远程执行**：目标机无需预装 Agent 或运行时，通过 SSH 下发预编译 Runner 或定制二进制即可。
- **声明式幂等**：内置 `ensure` 语法，天然支持 dry-run 与状态收敛。
- **异构架构支持**：纯 Go 实现，交叉编译覆盖 amd64/arm64 等主流服务器架构。

## 2. 设计原则

1. **由内向外**：先开发原子操作 SDK，确保核心价值独立可用，再实现语言前端。
2. **纯 Go 生态**：所有依赖必须支持 `CGO_ENABLED=0` 交叉编译，无 cgo。
3. **最小可用闭环**：优先打通“脚本 → 编译/指令包 → SSH 下发 → 远程执行 → 结构化回传”全链路。
4. **结构化优先**：标准库函数一律返回结构体，禁止返回未经处理的原始文本。
5. **安全内置**：权限分级、审计日志、签名验证、资源限制从设计之初就纳入。

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
                 ├── SSH 客户端（连接池、超时、重试）
                 ├── 架构检测（uname -m → GOARCH）
                 ├── Runner 分发（多架构预编译，缓存复用）
                 ├── 指令包生成（JSON，与架构无关）
                 └── 结果回收（stdout JSON 解析、错误聚合）
```

### 执行模式选择逻辑

| 场景 | 执行方式 | 说明 |
|------|---------|------|
| 简单任务（<100 行，无第三方 Go 库） | Runner 模式 | 控制端发送 JSON 指令包，目标机 Runner 解释执行 |
| 复杂业务逻辑或需要第三方 Go 库 | AOT 编译模式 | 编译为静态二进制，上传执行 |
| 用户强制指定 | `--mode runner/aot` | 手动覆盖自动选择 |
| 紧急故障处理 | Runner 模式 | 零编译延迟，即时下发 |

## 4. 技术栈与依赖

- **主语言**：Go 1.21+（利用其交叉编译、静态链接、丰富生态）
- **标准库依赖**（需纯 Go）：
  - `github.com/shirou/gopsutil/v3`（系统信息）
  - `golang.org/x/crypto/ssh`（SSH 客户端）
  - `github.com/pkg/sftp`（文件上传）
  - `gopkg.in/yaml.v3`（YAML 处理，可选）
- **构建要求**：`CGO_ENABLED=0`，目标平台 Linux amd64/arm64 优先
- **代码规范**：遵循 Go 官方风格，使用 gofmt/goimports

## 5. 开发阶段计划

### Phase 0：原子操作 SDK（ops-core-sdk） — 4 周

**目标**：开发独立可用的 Go 库，提供常用运维原子操作，返回强类型结构体，可被任何 Go 程序调用。

**任务**：
1. 创建 Go module `https://github.com/j4ckzh0u/opslang/ops-core-sdk`。
2. 定义统一返回结构体规范：每个函数返回 `(T, error)`，`T` 为可 JSON 序列化的结构体；错误返回包含明确错误码和消息。
3. 实现以下标准库包（函数列表见附录 A）：
   - `sys`：CPU、内存、磁盘、负载、主机名、用户、进程等。
   - `file`：读写、复制、移动、删除、权限、模板渲染、校验和。
   - `net`：HTTP GET/POST、TCP 连通性、DNS 解析、网络接口。
   - `process`：进程列表、按名称/端口查找、启动/停止外部命令（不使用 shell）。
   - `service`：systemd 服务管理（status/start/stop/restart/enable）。
   - `pkg`：包管理（apt/yum/dnf）安装/卸载/查询。
   - `json` / `yaml`：编解码。
   - `time`：当前时间、格式化、时间差。
4. 每个函数必须通过单元测试，并在 amd64/arm64 环境交叉编译验证。
5. 编写 godoc 文档。

**交付物**：
- `ops-core-sdk` Go module，每个子包可独立导入。
- 单元测试覆盖率 ≥ 80%。
- 示例代码展示每个函数的基本用法。

**验收标准**：
- 运行 `go test ./...` 全部通过。
- 在 Linux amd64 和 arm64 下 `CGO_ENABLED=0 go build` 成功。
- 所有函数均不调用 `/bin/sh -c`，信息获取直接使用 Go 库或读取 `/proc`/`/sys`。

---

### Phase 1：远程执行通道（SSH 控制面 + 通用 Runner） — 4 周

**目标**：打通 SSH 远程执行链路，实现通用 Runner（内嵌 SDK）与 JSON 指令协议。

**任务**：
1. **SSH 客户端封装**（包 `internal/sshx`）：
   - 支持密码、密钥认证，连接超时、执行超时、重试。
   - 连接池管理，支持并发限制。
   - SFTP 上传/下载。
2. **架构检测**：
   - 在 SSH 会话中执行 `uname -m`，映射到 GOARCH（x86_64→amd64，aarch64→arm64 等）。
   - 提供 `DetectArch(conn) (string, error)`。
3. **通用 Runner 开发**（`cmd/ops-runner`）：
   - 内嵌 `ops-core-sdk` 全部函数。
   - 从 stdin 读取 JSON 指令包（格式见附录 B），按顺序执行。
   - 输出 JSON 结果到 stdout（格式见附录 C）。
   - 支持 `--dry-run` 标志。
4. **Runner 多架构构建**：
   - 提供 Makefile 或脚本，构建 `ops-runner-linux-amd64`、`ops-runner-linux-arm64`。
   - 使用 `-ldflags "-s -w"` 减小体积，可选 zstd 压缩。
5. **opsctl exec 命令**：
   - 参数：`--hosts`、`--user`、`--key`、`--inventory`、`--instructions`（JSON 文件路径）。
   - 流程：连接 SSH → 检测架构 → 上传对应 Runner（若目标机无缓存）→ 发送指令包 → 收集输出。
   - 支持并发执行多台主机。

**交付物**：
- 可执行程序 `opsctl`，支持 `exec` 子命令。
- 通用 Runner 二进制（amd64/arm64）。
- 端到端示例：本地编写 JSON 指令，远程获取 CPU 信息并回传。

**验收标准**：
- 在远程主机（无任何预装环境）执行 `opsctl exec --instructions test.json` 成功返回结构化结果。
- 同架构第二次执行时，不重复上传 Runner（利用缓存），耗时显著降低。
- 支持并发 10 台主机执行并正确汇总结果。

---

### Phase 2：语言前端与解释器 — 4 周

**目标**：定义 OpsLang 语法，实现手写词法/语法分析器和 AST 解释器，支持本地脚本执行。

**任务**：
1. **语法设计**（见附录 D）：
   - 支持 `let` 变量声明、函数定义 `fn`、`if/else`、`for` 循环、`import` 模块。
   - 数据类型：整数、浮点数、字符串、布尔、列表、字典、结构体（`{}`）。
   - 内置函数调用：如 `sys.cpu.usage()`。
2. **词法分析器**（`internal/lexer`）：
   - 识别关键字、标识符、数字、字符串、运算符、注释。
   - 错误报告包含行列号。
3. **语法分析器**（`internal/parser`）：
   - 递归下降解析，生成 AST。
   - 定义 AST 节点类型（`internal/ast`）。
4. **解释器**（`internal/interpreter`）：
   - 遍历 AST 执行，实现变量作用域、函数调用、控制流。
   - 内置函数注册机制：将 `ops-core-sdk` 函数映射为解释器可调用对象。
   - 支持 `print`、`report` 等输出语句（暂不涉及远程分发）。
5. **REPL**：
   - `opsctl repl` 进入交互式环境，支持逐行输入执行。
6. **opsctl run**：
   - `opsctl run script.ops` 本地解释执行。

**交付物**：
- 可解释执行基本脚本的 CLI。
- 至少 10 个示例脚本（获取系统信息、文件操作、进程管理等）。

**验收标准**：
- 脚本能使用变量、循环、条件判断，调用标准库函数并打印结构化结果。
- REPL 可正常交互。
- 错误提示能定位到具体行列。

---

### Phase 3：AOT 编译管线 — 4 周

**目标**：将 OpsLang 脚本编译为静态二进制，支持多架构交叉编译和缓存。

**任务**：
1. **AST → Go 源码生成器**（`internal/compiler`）：
   - 将 AST 翻译为 Go 代码，正确映射变量、函数、控制流。
   - 内置函数调用编译为对 `ops-core-sdk` 的直接调用。
   - 生成 `main()` 函数，处理输入参数（`--input` JSON 文件）、调用用户脚本逻辑、输出 JSON 结果。
2. **编译封装**：
   - 动态生成临时 Go 项目（含 go.mod），引入 `ops-core-sdk`。
   - 调用 `go build -ldflags "-s -w" -o output` 编译。
   - 支持环境变量设置 `GOOS/GOARCH` 实现交叉编译。
3. **编译缓存**：
   - 以脚本文件 hash + 标准库版本 + 目标架构为键，缓存编译产物。
   - 命中缓存直接使用。
4. **opsctl build**：
   - 参数：`--output`、`--target-arch`（默认当前架构）。
   - 输出静态二进制。

**交付物**：
- `opsctl build` 命令可将脚本编译为二进制。
- 编译缓存机制，二次编译速度显著提升。

**验收标准**：
- 编译出的二进制在无 Go 环境的目标机上可运行，输出与解释执行一致。
- 交叉编译 amd64→arm64 成功，二进制可在 ARM 服务器运行。
- 编译缓存生效（相同脚本第二次编译 < 5 秒）。

---

### Phase 4：远程编排与声明式特性 — 6 周

**目标**：实现 `task ... on targets` 远程执行、Runner 指令包生成、自动模式选择，以及 `ensure`、dry-run、结构化输出等高级特性。

**任务**：
1. **语法扩展**：
   - 支持 `task "name" on <targets>` 声明，`targets` 可从 inventory 文件或命令行传入。
   - 支持 `parallel` 块，实现多主机并行。
2. **指令包生成器**：
   - 将 AST 中远程执行部分转换为 JSON 指令序列（指令格式与 Runner 兼容）。
   - 自动识别哪些代码在本地运行（如 `report` 汇总）和远程运行（如 `sys.cpu.usage()`）。
3. **自动模式选择**：
   - 启发式规则：脚本行数 < 100 且无 `import go "..."` 时默认 Runner 模式；否则 AOT 模式。
   - 允许用户通过 `--mode` 强制指定。
4. **声明式幂等 `ensure`**：
   - 编译/解释时，将 `ensure <condition> { actions }` 转换为 `check → apply → verify` 序列。
   - `check` 阶段判断当前状态是否满足，不满足则执行 `apply`，最后 `verify` 确认。
   - 支持 `notify` 触发器（当有变更时执行）。
5. **dry-run 注入**：
   - 在解释器/编译器层面，为每个变更操作生成 dry-run 分支。
   - dry-run 模式下，只输出将要执行的操作描述，不实际执行。
6. **结构化输出**：
   - `report { key: value }` 收集结果，任务结束时统一回传。
   - `metric(name, value, labels)` 输出指标。
   - `log(msg)` 输出日志。
   - `alert(msg)` 触发告警（可接入 Webhook）。
7. **opsctl deploy**：
   - 整合 inventory 解析、远程执行、结果聚合。
   - 参数：`--targets`、`--parallel`（并发数）、`--dry-run`、`--mode`。
   - 输出 JSON 汇总结果到 stdout 和文件。

**交付物**：
- 功能完整的 `opsctl deploy` 命令。
- 示例 inventory YAML 和脚本。
- 文档说明各特性的用法。

**验收标准**：
- 一个脚本可以同时操作多台主机，支持并发，并正确聚合结果。
- `ensure` 示例：重复执行不会造成副作用（幂等）。
- `--dry-run` 模式不改变系统状态，输出操作预览。
- 结构化输出可以被其他程序解析（如导入 Elasticsearch）。

---

### Phase 5：安全与生产化 — 4 周

**目标**：加入权限分级、审计日志、签名验证、资源限制等企业级特性。

**任务**：
1. **权限分级**：
   - 脚本头部声明 `privilege: read_only | admin | root`。
   - 编译期检查：`read_only` 脚本调用变更类函数（如 `service.start`）时报错。
   - Runner 运行时二次校验。
2. **审批流**：
   - 当脚本包含 `admin`/`root` 权限且目标为主机清单中标记为生产的主机时，要求交互确认或审批 token。
3. **审计日志**：
   - 控制端记录每次任务的完整信息：用户、脚本 hash、目标主机、执行时间、结果、错误。
   - 可输出为 JSON 文件或发送到 syslog。
4. **Runner 签名验证**：
   - 控制端生成 Runner 时，使用私钥签名（Ed25519）。
   - 目标机可预置公钥，Runner 启动前验证签名（可选）。
5. **资源限制**：
   - 执行远程命令时，优先使用 `systemd-run --scope -p CPUQuota=... -p MemoryMax=...` 限制资源。
   - 无 systemd 时回退到 `ulimit` 或直接执行。
6. **临时目录自清理**：
   - Runner 和 AOT 二进制上传到 `/tmp/ops-<random>`，执行后自动删除。
   - 设置 trap 确保异常退出时也清理。

**交付物**：
- 具备完整安全特性的 opsctl。
- 审计日志示例。
- 资源限制测试报告。

**验收标准**：
- `read_only` 脚本无法执行变更操作。
- 高危操作需要审批。
- 审计日志记录完整，可回溯。
- Runner 自清理无残留。

---

### Phase 6（后续扩展，本计划暂不实施）

- 网络设备支持（SSH/NETCONF/SNMP/gNMI）。
- Docker/Kubernetes 原子操作。
- 云 API 集成（AWS/阿里云/华为云）。
- 自研 Parser 替换（tree-sitter）与 LSP。
- VS Code 插件、包管理器、Web 控制台。

## 6. 项目结构建议

```
opslang/
├── cmd/
│   ├── opsctl/                 # CLI 主程序
│   └── ops-runner/             # 通用 Runner 程序
├── internal/
│   ├── lexer/                  # 词法分析
│   ├── parser/                 # 语法分析
│   ├── ast/                    # AST 节点定义
│   ├── interpreter/            # 解释器
│   ├── compiler/               # Go 代码生成器
│   ├── sshx/                   # SSH 客户端封装
│   ├── inventory/              # 主机清单解析
│   ├── runner/                 # Runner 指令包处理
│   ├── security/               # 签名、权限
│   └── output/                 # 结构化输出处理
├── pkg/
│   ├── ops-core-sdk/           # 原子操作标准库
│   │   ├── sys/
│   │   ├── file/
│   │   ├── net/
│   │   ├── process/
│   │   ├── service/
│   │   ├── pkg/
│   │   ├── json/
│   │   ├── yaml/
│   │   └── time/
│   └── ...
├── tests/                      # 集成测试
├── examples/                   # 示例脚本
├── docs/                       # 文档
├── go.mod
└── Makefile
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

// sys.DiskUsage 返回
type DiskUsage struct {
    Path        string  `json:"path"`
    Total       uint64  `json:"total"`
    Used        uint64  `json:"used"`
    Free        uint64  `json:"free"`
    UsedPercent float64 `json:"used_percent"`
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
      "op": "sys.memory.info",
      "args": {},
      "assign": "mem"
    },
    {
      "op": "report",
      "args": {
        "cpu": "cpu",
        "mem": "mem"
      }
    }
  ]
}
```

### 7.3 Runner 输出 JSON 格式

```json
{
  "status": "ok",
  "data": {
    "cpu": { "percent": 12.5, "user": 5.0, "system": 3.0, "idle": 92.0 },
    "mem": { "total": 16777216, "available": 8388608, "used": 8388608, "used_percent": 50.0 }
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
  }
}
```

## 8. 开发协作约定

- **模块化开发**：每个 Phase 内按包拆分，每个 PR 聚焦一个功能点。
- **测试驱动**：每个新增函数/模块必须包含单元测试；集成测试覆盖端到端链路。
- **代码风格**：遵循 Go 官方指南，使用 `gofmt`、`go vet`。
- **错误处理**：标准库函数返回明确错误，不 panic；错误信息包含上下文。
- **日志**：CLI 使用 `logrus` 或标准库 `log`，区分 debug/info/warn/error 级别。
- **版本管理**：标准库遵循语义化版本；Runner 与指令包协议包含版本号。

## 9. 风险与应对

| 风险 | 应对措施 |
|------|---------|
| 标准库覆盖不足 | 提供 `process.exec()` 逃生舱，允许直接调用外部命令（但默认禁用） |
| 异构架构兼容问题 | 强制纯 Go，CI 交叉编译测试；标准库函数避免架构特定假设 |
| Runner 分发体积大 | 首次上传后缓存，后续仅发送指令包；二进制压缩 |
| 用户学习成本 | 提供大量示例和文档；语法保持简洁；REPL 辅助调试 |
| 安全漏洞 | 权限分级、审批、审计、签名验证、资源限制，定期安全审计 |
| 编译时间过长 | 缓存机制；Runner 模式避免不必要的编译 |

## 10. 里程碑与时间估算

| 阶段 | 内容 | 预计时间 | 关键交付物 |
|------|------|---------|-----------|
| Phase 0 | 原子操作 SDK | 4 周 | ops-core-sdk |
| Phase 1 | 远程执行通道 | 4 周 | opsctl exec + Runner |
| Phase 2 | 语言前端与解释器 | 4 周 | opsctl run + REPL |
| Phase 3 | AOT 编译管线 | 4 周 | opsctl build |
| Phase 4 | 远程编排与声明式 | 6 周 | opsctl deploy |
| Phase 5 | 安全与生产化 | 4 周 | 企业级特性 |
| **合计** | | **26 周** | 可产品化 MVP |

以上计划为 Claude Code 提供了清晰的执行路径。每个 Phase 均可独立验证，逐步构建完整的 OpsLang 系统。
