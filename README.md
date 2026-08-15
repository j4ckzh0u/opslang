# OpsLang

面向运维领域的领域特定语言（DSL），用简洁脚本完成复杂运维操作，彻底摆脱 Shell 字符串处理和 Python 环境依赖。

## 核心价值

- **结构化返回**：标准库函数全部返回强类型结构体，支持 JSON 序列化，告别字符串解析
- **双执行引擎**：Runner 模式（JSON 指令包解释执行，零编译延迟）+ AOT 编译模式（静态二进制，支持第三方 Go 库）
- **零依赖远程执行**：通过 SSH 下发预编译 Runner，目标机无需安装任何运行时
- **异构架构支持**：纯 Go 实现，`CGO_ENABLED=0` 交叉编译覆盖 amd64/arm64
- **声明式幂等**：内置 `ensure` 语法，支持 dry-run 与状态收敛
- **大规模文件传输**：分层中继 + 压缩传输 + 断点续传 + 内容哈希去重
- **企业级安全**：权限分级、审批流、审计日志、Ed25519 签名验证、资源限制

## 快速开始

30 秒体验：编写一个采集系统信息的脚本。

```ops
// check_cpu.ops
let cpu = {"percent": 0, "cores": 4}
let hostname = "localhost"

print("=== CPU Information ===")
print("Hostname: " + hostname)
print("CPU cores: " + str(cpu.cores))
print("CPU usage: " + str(cpu.percent) + "%")

if cpu.percent > 80 {
    alert("CPU usage is high: " + str(cpu.percent) + "%")
}

report {
    host: hostname,
    cpu: cpu
}
```

执行：

```bash
# 本地解释执行
opsctl run examples/check_cpu.ops

# 编译为静态二进制
opsctl build --source examples/check_cpu.ops --output check_cpu --target-arch linux/amd64

# 远程执行（SSH 下发 Runner）
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
// control_flow.ops
let scores = [85, 92, 78, 95, 88]
let total = 0
let count = 0

for score in scores {
    total = total + score
    count = count + 1
}

let avg = total / count
print("Average score: " + str(avg))

if avg >= 90 {
    print("Excellent")
} else if avg >= 80 {
    print("Good")
} else {
    print("Need improvement")
}

fn max_value(lst) {
    let max = lst[0]
    for v in lst {
        if v > max {
            max = v
        }
    }
    return max
}

print("Max score: " + str(max_value(scores)))
```

### 3. 文件操作

```ops
// file_ops.ops
let path = "/tmp/test.txt"
let content = "Hello, OpsLang!"

// 写入文件
file.write(path, content, "0644")

// 读取文件
let result = file.read(path)
print("File content: " + result.content)

// 检查文件是否存在
if file.exists(path) {
    print("File exists")
}

// 计算校验和
let checksum = file.checksum(path, "sha256")
print("SHA256: " + checksum.checksum)

// 删除文件
file.delete(path)
```

### 4. 系统信息采集

```ops
// check_memory.ops
let mem = sys.memory.info()

print("=== Memory Information ===")
print("Total: " + str(mem.total) + " bytes")
print("Available: " + str(mem.available) + " bytes")
print("Used: " + str(mem.used) + " bytes")
print("Used Percent: " + str(mem.used_percent) + "%")

if mem.used_percent > 90 {
    alert("Memory usage is critical: " + str(mem.used_percent) + "%")
}

report {
    memory: mem
}
```

### 5. 远程任务（task 声明）

```ops
// deploy_app.ops
task "deploy_config" on targets {
    let config_path = "/etc/myapp/config.yaml"
    let template_vars = {"env": "production", "port": 8080}

    let rendered = file.template("config.yaml.tpl", template_vars)
    file.write(config_path, rendered.content, "0644")

    service.restart("myapp")

    report {
        status: "deployed",
        config: config_path
    }
}
```

## CLI 命令

| 命令 | 说明 |
|------|------|
| `opsctl version` | 打印版本信息 |
| `opsctl run <script.ops>` | 本地解释执行脚本 |
| `opsctl repl` | 交互式 REPL 环境 |
| `opsctl build --source <script.ops> --output <binary> --target-arch <os/arch>` | AOT 编译为静态二进制 |
| `opsctl exec --hosts <hosts> --instructions <pkg.json> --parallel <n>` | 远程执行指令包 |

## 语言特性

### 关键字（16 个）

`let`, `fn`, `if`, `else`, `for`, `while`, `return`, `task`, `on`, `import`, `true`, `false`, `nil`, `report`, `alert`, `ensure`

### 数据类型

- `int` - 整数
- `float` - 浮点数
- `string` - 字符串
- `bool` - 布尔值
- `list` - 列表
- `dict` - 字典
- `nil` - 空值

### 运算符

- 算术：`+`, `-`, `*`, `/`, `%`
- 比较：`==`, `!=`, `<`, `>`, `<=`, `>=`
- 逻辑：`&&`, `||`, `!`
- 赋值：`=`

### 内置函数

- `print(value)` - 打印输出
- `len(value)` - 获取长度
- `str(value)` - 转为字符串
- `int(value)` - 转为整数
- `float(value)` - 转为浮点数
- `type(value)` - 获取类型
- `log(msg)` - 日志输出
- `metric(name, value, labels)` - 指标上报

## 标准库

### ops-core-sdk

位于 `pkg/ops-core-sdk/`，提供以下包：

| 包 | 说明 |
|----|------|
| `sys` | CPU、内存、磁盘、负载、主机名、用户、进程等系统信息 |
| `file` | 文件读写、复制、移动、删除、权限、模板渲染、校验和 |
| `net` | HTTP GET/POST、TCP 连通性、DNS 解析、网络接口 |
| `process` | 进程列表、查找、启动/停止外部命令 |
| `service` | systemd 服务管理 |
| `pkg` | 包管理封装（apt/yum/dnf） |
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
| Phase 0 | 原子操作 SDK（ops-core-sdk） | ✅ 已完成 |
| Phase 1 | 远程执行通道（SSH + Runner） | ✅ 已完成 |
| Phase 2 | 语言前端与解释器（Lexer/Parser/Interpreter） | ✅ 已完成 |
| Phase 3 | AOT 编译管线 | ✅ 已完成 |
| Phase 4 | 远程编排与声明式特性 | 🚧 进行中（deploy/task/ensure 已完成，分层中继待实现） |
| Phase 5 | 安全与生产化 | 🚧 进行中（权限/审计/签名/资源限制已完成） |

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
│   ├── interpreter/            # 解释器
│   ├── compiler/               # Go 代码生成器
│   ├── sshx/                   # SSH 客户端封装
│   ├── inventory/              # 主机清单
│   ├── runner/                 # Runner 指令包处理
│   └── security/               # 安全特性
├── pkg/
│   └── ops-core-sdk/           # 原子操作标准库
├── examples/                   # 示例脚本
├── go.mod
├── Makefile
└── README.md
```

## 文档

详细文档请参考 `docs/` 目录（待补充）：

- 语言语法参考
- 标准库 API 文档
- Runner 指令包协议
- 远程执行架构

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

Apache License 2.0
