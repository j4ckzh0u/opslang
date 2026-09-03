# OpsLang

面向运维领域的领域特定语言（DSL），用简洁脚本完成复杂运维操作。控制器和目标机只需 OpsLang/ops-runner 二进制，不需要 Python 或 Shell 脚本运行时。

## 核心价值

- **结构化返回**：标准库函数全部返回带 JSON 标签的结构体，告别字符串解析
- **双执行引擎**：Runner 模式（JSON 指令包经 SSH 下发，零编译延迟，仅支持线性脚本）+ AOT 模式（按目标机架构编译静态二进制，支持包括 `ensure`/`parallel` 在内的全部语言）
- **零语言运行时远程执行**：通过 SSH 下发预编译 Runner 或 AOT 二进制，目标机无需安装 Python、Shell 或 OpsLang 运行时；具体模块仍可能要求 Linux 内核能力、系统服务或外部系统组件
- **异构架构支持**：纯 Go 实现，`CGO_ENABLED=0` 交叉编译覆盖 amd64/arm64
- **声明式幂等**：内置 `ensure` 语法（check → apply → verify → notify），支持 dry-run 与状态收敛
- **Ansible 核心模块对齐**：`pkg.ensure` / `service.ensure` / `user.ensure` / `group.ensure` / `file.ensure` 幂等收敛家族 —— 声明期望状态，重复执行零变更，`changed`/`actions` 如实报告每一次真实动作
- **可恢复文件分发/收集**：`file.distribute` / `file.collect` 经真实 SSH/SFTP 传输，支持内容哈希跳过、部分文件续传、原子替换、SHA-256 校验与并发控制
- **分层中继分发**：`file.distribute` 可按显式标签或 IP 前缀分组，通过短时令牌和 TLS 指纹固定的 HTTPS Range 服务扇出，失败目标自动回退直接 SFTP
- **SSH 安全**：主机密钥 TOFU（首次信任）校验，密钥变更即拒绝，防中间人攻击

## 快速开始

30 秒体验：`examples/check_cpu.ops`（节选）采集真实系统信息（非模拟数据）：

```ops
// check_cpu.ops - Collect real CPU usage information
// Fetch real data from the system
let hostname_info = sys.hostname()
let cpu_usage = sys.cpu.usage()
let cpu_count = sys.cpu.count()
let load_info = sys.load()

print("=== CPU Information ===")
print("Hostname: " + hostname_info.hostname)
print("Logical cores: " + str(cpu_count.logical))
print("Physical cores: " + str(cpu_count.physical))
print("CPU usage: " + str(cpu_usage.percent) + "%")
print("Load average: " + str(load_info.load1) + " / " + str(load_info.load5) + " / " + str(load_info.load15))

// Alert if CPU usage is high
if cpu_usage.percent > 80 {
    alert("CPU usage is high: " + str(cpu_usage.percent) + "%")
}

// Structured report with all collected data
report {
    host: hostname_info.hostname,
    logical_cores: cpu_count.logical,
    physical_cores: cpu_count.physical,
    cpu_percent: cpu_usage.percent,
    load_1m: load_info.load1,
    load_5m: load_info.load5,
    load_15m: load_info.load15
}
```

执行：

```bash
# 本地解释执行
opsctl run examples/check_cpu.ops

# 编译为静态二进制
opsctl build --source examples/check_cpu.ops --output check_cpu --target-arch linux/amd64

# 远程部署（自动选择 runner/aot 模式）
opsctl deploy examples/check_cpu.ops --targets user@host1,user@host2

# 直接下发 JSON 指令包
opsctl exec --hosts user@host1,user@host2 --instructions pkg.json --parallel 10
```

## 安装

### 从源码编译

```bash
git clone https://github.com/j4ckzh0u/opslang.git
cd opslang
make build
```

编译产物在 `bin/` 目录，包含 `opsctl` 和 `ops-runner`。

### 交叉编译

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/opsctl-linux-amd64 ./cmd/opsctl

# Linux arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/opsctl-linux-arm64 ./cmd/opsctl
```

## Makefile 使用说明

项目提供了完整的 Makefile，覆盖构建、测试、代码质量检查和交叉编译等常用操作。运行 `make help` 查看所有可用目标。

### 构建

```bash
make build          # 编译 opsctl 和 ops-runner 到 bin/
make build-all      # 交叉编译 linux/darwin amd64/arm64 共 8 个二进制到 dist/
make install        # 安装 opsctl 到 $GOPATH/bin
```

### 测试

```bash
make test           # 运行全部测试（含 race 检测）
make coverage       # 生成覆盖率报告（终端输出 + HTML）
make bench          # 运行基准测试
```

### 代码质量

```bash
make lint           # 运行 go vet + gofmt 检查
make fmt            # 检查格式（有未格式化文件时报错）
make fmt-fix        # 自动修复格式
make vet            # 运行 go vet
make tidy           # 整理 go.mod 依赖
make check          # 运行 lint + vet（完整检查）
```

### 开发调试

```bash
make run ARGS="run examples/helloworld.ops"   # 运行脚本
make repl                                      # 启动 REPL 交互环境
make examples                                  # 运行所有示例脚本
```

### CI 与清理

```bash
make ci             # 本地 CI：check + test + build
make clean          # 清理 bin/ dist/ coverage.out
```

## 使用示例

### 1. 变量与数据类型

```ops
// variables.ops
let count = 42
let pi = 3.14159
let name = "OpsLang"
let active = true
let items = ["alpha", "beta", "gamma"]
let config = {"host": "localhost", "port": 8080}

print("Integer: " + str(count))
print("Float: " + str(pi))
print("String: " + name)
print("Boolean: " + str(active))
print("List: " + str(items))
print("Dict host: " + config.host)
print("Dict port: " + str(config.port))

report {
    count: count,
    pi: pi,
    name: name,
    active: active,
    items: items,
    config: config
}
```

### 2. 控制流与函数

```ops
// control_flow.ops（节选）
let cpu = sys.cpu.usage()
let percent = cpu.percent

if percent >= 90 {
    print("CPU Status: CRITICAL (" + str(percent) + "%)")
} else if percent >= 70 {
    print("CPU Status: WARNING (" + str(percent) + "%)")
} else {
    print("CPU Status: NORMAL (" + str(percent) + "%)")
}

// C 风格 for 循环（唯一支持的循环形式，无 for-in 语法）
let parts = sys.disk.partitions()
for let i = 0; i < len(parts); i = i + 1 {
    let p = parts[i]
    print("  [" + str(i) + "] " + p.mountpoint + " (" + p.fstype + ")")
}

fn max_value(lst) {
    let max = lst[0]
    for let i = 1; i < len(lst); i = i + 1 {
        if lst[i] > max {
            max = lst[i]
        }
    }
    return max
}

print("Max score: " + str(max_value([85, 92, 78, 95, 88])))
```

### 3. 文件操作

```ops
// file_ops.ops
let path = "/tmp/test.txt"
let content = "Hello, OpsLang!"

// 写入文件
file.write(path, content)

// 读取文件
let result = file.read(path)
print("File content: " + result.content)

// 检查文件是否存在（返回结构体，取 .exists 字段）
if file.exists(path).exists {
    print("File exists")
}

// 修改权限（mode 为字符串）
file.chmod(path, "0644")

// 计算校验和
let checksum = file.checksum(path, "sha256")
print("SHA256: " + checksum.checksum)

// 删除文件
file.delete(path)
```

### 4. 系统信息采集

```ops
// check_memory.ops（节选）
let mem = sys.memory.info()
let total_mb = mem.total / 1024 / 1024
let usage_percent = mem.used_percent

print("Total: " + str(total_mb) + " MB")
print("Usage: " + str(usage_percent) + "%")

if usage_percent > 90 {
    alert("Memory usage critical: " + str(usage_percent) + "%")
}

report {
    total_mb: total_mb,
    usage_percent: usage_percent
}
```

### 5. 远程任务（task 声明）

`task "名字" on <目标>` 的目标路由只在 `opsctl deploy` 下生效；`opsctl run` 遇到带 `on` 的 task 会报错并提示使用 deploy。

```ops
// deploy_app.ops
task "deploy_config" on "web*" {
    let config_path = "/etc/myapp/config.yaml"
    let template_vars = {"env": "production", "port": 8080}

    let rendered = file.template("config.yaml.tpl", template_vars)
    file.write(config_path, rendered.content)

    service.restart("myapp")

    report {
        status: "deployed",
        config: config_path
    }
}
```

`on` 子句支持精确主机名 / `user@host` / glob 模式（如 `"web*"`），匹配 `opsctl deploy --targets` 提供的目标列表。

## CLI 命令

| 命令 | 说明 |
|------|------|
| `opsctl version` | 打印版本信息 |
| `opsctl run <script.ops>` | 本地解释执行脚本（`--json` / `--verbose` / `--dry-run`） |
| `opsctl repl` | 交互式 REPL 环境 |
| `opsctl build --source <script.ops> --output <binary> --target-arch <os/arch>` | AOT 编译为静态二进制 |
| `opsctl deploy <script.ops> --targets <hosts>` | 远程部署执行（`--mode auto/runner/aot`） |
| `opsctl exec --hosts <hosts> --instructions <pkg.json> --parallel <n>` | 直接下发 JSON 指令包远程执行 |

## 幂等收敛：对标 Ansible 核心模块

Ansible 的价值不在模块数量，而在**幂等收敛**：声明期望状态，重复执行安全。OpsLang 用 ensure 家族操作实现同一语义，且在三种执行引擎（本地解释、远程 Runner、AOT 编译二进制）中行为完全一致：

```ops
// examples/remote_ensure_fleet.ops —— 舰队供给剧本（真实部署节选）
privilege: admin

task "fleet_facts" on "*" {
    let os = sys.os()
    report { platform: os.platform, platform_version: os.platform_version }
}

task "provision_service_account" on "root" {     // on "root" 路由到 inventory 中 group=root 的主机
    let g = group.ensure("opslang-fleet", {})
    let u = user.ensure("opslang-fleet", { "shell": "/usr/sbin/nologin", "create_home": "false" })
    report { group_changed: g.changed, user_changed: u.changed, user_shell: u.shell }
}

task "converge_cron" on "root" {
    let run = service.ensure("cron", "started")
    let boot = service.ensure_enabled("cron", true)
    report { start_changed: run.changed, start_actions: run.actions, active: run.active }
}
```

部署与幂等性实证（3 台真实 Ubuntu 22.04 主机）：

```bash
opsctl deploy examples/remote_ensure_fleet.ops --inventory hosts.yaml --parallel 3 --auto-approve
# 第一次：group_changed=true, user_changed=true（真实创建）
# 第二次：全部 changed=false（幂等 —— 这正是 Ansible playbook 的核心承诺）
```

| OpsLang | Ansible 模块 | 语义 |
|---|---|---|
| `pkg.ensure(name)` | `package` (`state=present`) | 已安装零动作 |
| `service.ensure(name, state)` | `service`/`systemd` (`state=`) | started/stopped 幂等；restarted 恒变更；reload 失败回退 restart |
| `service.ensure_enabled(name, enabled)` | `service` (`enabled=`) | 自启收敛 |
| `user.ensure(name, opts)` / `user.absent(...)` | `user` | 创建/漂移收敛（shell、home）/删除（拒绝删 root） |
| `group.ensure(name, opts)` / `group.absent(name)` | `group` | 存在性收敛 |
| `file.ensure(path, state, mode)` | `file` | directory/file/touch/absent + 权限位收敛 |

所有 ensure 操作返回 `changed` 与 `actions`（实际执行的动作列表）——审计这两个字段是判断"这次部署到底改了什么"的唯一可信来源。完整文档见 `docs/stdlib-reference.md` 第 20 节。

舰队级主机巡检（CPU/内存/磁盘、软件包名+版本、用户态进程 Top、端口监听↔进程、TCP 连接↔进程、进程二进制路径↔软件包归属）见 `examples/remote_fleet_audit.ops` —— 六个维度全部纯 Go 采集（`net.connections` 直读内核 socket 表并归属 pid，等价 `ss -tlnp` 但不调用它），已在 3 台真实 Ubuntu 主机验证。

### 真实执行政策（Real Execution Policy）

本项目对示例与文档执行一条铁律：**要么真实执行，要么显式说明为何未执行，绝不伪造成功**。

- 所有示例中的数据来自真实系统调用；不存在"模拟数据/伪代码演示成功"的路径
- 变更类示例内置真实前提探测（平台、root、systemd 单元、包管理器）：条件不满足时打印 `SKIP: <真实原因>` 并在 report 中如实标注 `skipped: true`
- 破坏性/侵入性操作（改主机名、改防火墙、装软件包）在示例中不自动执行，但明确列出真实用法与前提，不假装已执行
- 每个示例都被 `cmd/opsctl/examples_e2e_test.go` 通过真实解释器跑一遍，坏示例会直接挂 CI

## 语言特性

OpsLang 是动态类型语言。关键字共 20 个：

`let`, `fn`, `if`, `else`, `for`, `while`, `return`, `task`, `on`, `import`, `privilege`, `true`, `false`, `nil`, `report`, `alert`, `ensure`, `metric`, `log`, `parallel`

### 数据类型

- `int` - 整数
- `float` - 浮点数
- `string` - 字符串
- `bool` - 布尔值
- `list` - 列表
- `dict` - 字典
- `nil` - 空值

### 运算符

- 算术：`+`, `-`, `*`, `/`, `%`（两个 int 相除为整除，如 `10 / 3 == 3`）
- 比较：`==`, `!=`, `<`, `>`, `<=`, `>=`（严格比较：`1 != "1"`；数值跨 int/float 可比：`1 == 1.0`）
- 逻辑：`&&`, `||`, `!`
- 赋值：`=`（`let` 同一作用域不可重复声明，重新赋值用 `=`）

### 内置函数

- `print(value)` - 打印输出
- `len(value)` - 获取长度
- `str(value)` - 转为字符串
- `int(value)` - 转为整数（严格解析，`int("42abc")` 报错）
- `float(value)` - 转为浮点数
- `type(value)` - 获取类型
- `log(msg)` - 日志输出
- `metric(name, value, labels)` - 指标上报

## 标准库

### ops-core-sdk

位于 `pkg/ops-core-sdk/`，提供以下包：

| 包 | 说明 |
|----|------|
| `sys` | CPU、内存、磁盘、负载、主机名、用户、网络接口等系统信息 |
| `file` | 文件读写、复制、移动、删除、权限、模板渲染、校验和、SSH 分发/收集 |
| `net` | HTTP GET/POST、TCP 连通性、DNS 解析、网络接口 |
| `process` | 进程列表、查找、kill、执行外部命令 |
| `service` | systemd 服务管理 |
| `pkg` | 包管理封装（仅 Linux，apt/yum/dnf） |
| `json` | JSON 编解码 |
| `yaml` | YAML 编解码 |
| `time` | 时间操作、格式化、时间差 |

所有函数返回强类型结构体，支持 JSON 序列化。示例：

```go
type CPUUsage struct {
    Percent float64 `json:"percent"`
    User    float64 `json:"user"`
    System  float64 `json:"system"`
    Idle    float64 `json:"idle"`
}

type MemoryInfo struct {
    Total       uint64  `json:"total"`
    Available   uint64  `json:"available"`
    Used        uint64  `json:"used"`
    UsedPercent float64 `json:"used_percent"`
}
```

## 开发阶段

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase 0 | 原子操作 SDK（ops-core-sdk） | 已完成 |
| Phase 1 | 远程执行通道（SSH + Runner） | 已完成 |
| Phase 2 | 语言前端与解释器（Lexer/Parser/Interpreter） | 已完成 |
| Phase 3 | AOT 编译管线 | 已完成 |
| Phase 4 | 远程编排与声明式特性（deploy/task/ensure/parallel） | 部分完成：基础编排、断点续传和分层中继分发可用，传输压缩待实现 |
| Phase 5 | 安全与生产化（权限分级、审计、签名、资源限制） | 部分完成：核心安全链路可用，资源限制回退与自动回滚接入待实现 |

## Roadmap

以下能力**尚未实现**，文档中不再作为现有功能描述：

- `import "go <包路径>"` 引用第三方 Go 库（当前会报错拒绝）
- 文件传输压缩
- 分层中继文件收集；当前中继扇出用于 `file.distribute`，`file.collect` 使用可恢复 SFTP
- SSH 连接在多次 deploy 之间的跨进程复用（单次部署内并发正确；架构检测结果已通过 `~/.opsctl/arch-cache.json` 跨部署缓存）
- 无 `systemd-run` 目标机上的资源限制回退，以及部署失败后的自动回滚接入

> 注：`for ... in ...` 遍历循环与 `block/rescue` 错误处理**已实现**（见 docs/language-reference.md 第 6.3 节）。

## 项目结构

```
opslang/
├── cmd/
│   ├── opsctl/                 # CLI 主程序
│   └── ops-runner/             # 通用 Runner
├── internal/
│   ├── lexer/                  # 词法分析
│   ├── parser/                 # 语法分析
│   ├── ast/                    # AST 节点定义
│   ├── interpreter/            # 解释器（含 SDK bridge）
│   ├── compiler/               # AOT 编译器（Go 代码生成）
│   ├── opsspec/                # 原子操作单一事实来源（名称+参数+可用范围）
│   ├── exec/                   # 并行 SSH 编排（架构探测、二进制上传）
│   ├── sshx/                   # SSH 客户端封装（TOFU 主机密钥校验）
│   ├── inventory/              # 主机清单
│   ├── arch/                   # 远程架构检测
│   ├── runner/                 # Runner 指令包处理与注册表
│   ├── output/                 # 结构化输出处理
│   └── security/               # 安全特性
├── pkg/
│   └── ops-core-sdk/           # 原子操作标准库
├── examples/                   # 示例脚本
├── go.mod
├── Makefile
└── README.md
```

## 文档

详细文档请参考 `docs/` 目录：

- `docs/getting-started.md` - 快速开始
- `docs/language-reference.md` - 语言语法参考
- `docs/stdlib-reference.md` - 标准库 API 文档
- `docs/cli-reference.md` - CLI 命令参考（含 Runner 指令包协议）
- `docs/architecture.md` - 系统架构
- `docs/examples.md` - 示例教程

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

Apache License 2.0
