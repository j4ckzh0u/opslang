# OpsLang 示例教程

本目录包含 OpsLang 的实用示例脚本。每个示例都包含完整代码、运行命令和预期输出。

## 运行方式

所有示例通过 `opsctl run` 命令执行：

```bash
opsctl run examples/示例文件.ops
```

## 示例列表

| 示例 | 说明 | 关键语法 |
|------|------|---------|
| [系统信息采集](#1-系统信息采集) | 采集 CPU/内存信息并输出结构化报告 | `report`, `alert`, `log`, `if` |
| [文件批量操作](#2-文件批量操作) | 批量处理文件列表 | `for`, `list`, `len()` |
| [服务健康检查](#3-服务健康检查) | 检查多个服务端口状态 | `fn`, `list` of `dict`, `return` |
| [应用部署流程](#4-应用部署流程) | 多步骤应用部署 | `dict`, 字符串拼接, `report` |
| [监控告警脚本](#5-监控告警脚本) | 指标采集与多级告警 | `metric()`, 嵌套 `if/else` |
| [ensure 幂等操作](#6-ensure-幂等操作) | 声明式幂等配置管理 | `ensure` |
| [函数与递归](#7-函数与递归) | 函数定义、默认参数、递归 | `fn`, 默认参数, 递归 |
| [远程任务声明](#8-远程任务声明) | task/on 远程任务语法 | `task`, `on` |

---

## 1. 系统信息采集

采集系统基本信息，超过阈值自动告警，输出结构化报告。

**文件**: `examples/collect_system_info.ops`

```ops
// 采集系统基本信息并输出结构化报告
let hostname = "server-01"

// 模拟 CPU/内存数据（实际环境调用 sys.cpu.usage() 等）
let cpu_percent = 45.2
let mem_total = 16384
let mem_used = 8192
let mem_percent = 50.0

// 条件告警
if cpu_percent > 80 {
    alert("CPU 使用率过高: " + str(cpu_percent) + "%")
}

if mem_percent > 90 {
    alert("内存使用率过高: " + str(mem_percent) + "%")
}

// 结构化报告
report {
    host: hostname,
    cpu: cpu_percent,
    mem_total: mem_total,
    mem_used: mem_used,
    mem_percent: mem_percent,
    status: "ok"
}

log("系统信息采集完成")
```

**运行命令**:

```bash
opsctl run examples/collect_system_info.ops
```

**预期输出**:

```
[REPORT] {
  "host": "server-01",
  "cpu": 45.2,
  "mem_total": 16384,
  "mem_used": 8192,
  "mem_percent": 50.0,
  "status": "ok"
}
[LOG] 系统信息采集完成
```

> 说明：CPU 45.2% 未超过 80% 阈值，内存 50% 未超过 90% 阈值，所以不会触发 alert。

---

## 2. 文件批量操作

遍历文件列表，逐个处理。

**文件**: `examples/batch_file_ops.ops`

```ops
// 批量文件操作示例
let files = ["/tmp/a.txt", "/tmp/b.txt", "/tmp/c.txt"]
let count = len(files)

print("准备处理 " + str(count) + " 个文件")

for let i = 0; i < count; i = i + 1 {
    let path = files[i]
    print("处理文件: " + path)
    // 实际环境: file.write(path, "content")
    // 实际环境: let checksum = file.checksum(path, "sha256")
}

print("所有文件处理完成")

report {
    total: count,
    operation: "batch_process",
    status: "completed"
}
```

**运行命令**:

```bash
opsctl run examples/batch_file_ops.ops
```

**预期输出**:

```
准备处理 3 个文件
处理文件: /tmp/a.txt
处理文件: /tmp/b.txt
处理文件: /tmp/c.txt
所有文件处理完成
[REPORT] {
  "total": 3,
  "operation": "batch_process",
  "status": "completed"
}
```

---

## 3. 服务健康检查

定义检查函数，遍历服务列表，输出汇总报告。

**文件**: `examples/service_health_check.ops`

```ops
// 服务健康检查脚本
fn check_service(name, port) {
    // 模拟检查（实际环境用 net.TCPConnect）
    let result = {"name": name, "port": port, "status": "running"}
    return result
}

let services = [
    {"name": "nginx", "port": 80},
    {"name": "mysql", "port": 3306},
    {"name": "redis", "port": 6379}
]

let total = len(services)
let healthy = 0

for let i = 0; i < total; i = i + 1 {
    let svc = services[i]
    let result = check_service(svc["name"], svc["port"])
    print("检查 " + svc["name"] + ":" + str(svc["port"]) + " -> " + result["status"])
    healthy = healthy + 1
}

report {
    total_services: total,
    healthy: healthy,
    unhealthy: total - healthy,
    timestamp: "2026-08-15T10:00:00Z"
}

if healthy < total {
    alert("有服务不健康！")
}
```

**运行命令**:

```bash
opsctl run examples/service_health_check.ops
```

**预期输出**:

```
检查 nginx:80 -> running
检查 mysql:3306 -> running
检查 redis:6379 -> running
[REPORT] {
  "total_services": 3,
  "healthy": 3,
  "unhealthy": 0,
  "timestamp": "2026-08-15T10:00:00Z"
}
```

> 说明：3 个服务全部健康，`healthy == total`，不触发 alert。

---

## 4. 应用部署流程

模拟多步骤应用部署，包含配置管理和结构化输出。

**文件**: `examples/deploy_app.ops`

```ops
// 应用部署脚本
let app_name = "myapp"
let app_version = "2.1.0"
let deploy_dir = "/opt/app"
let config = {
    "worker": 4,
    "timeout": 30,
    "log_level": "info"
}

print("开始部署 " + app_name + " v" + app_version)

// 步骤 1: 检查环境
print("[1/4] 检查部署目录...")
// 实际环境: file.exists(deploy_dir)

// 步骤 2: 创建备份
print("[2/4] 创建备份...")
let backup_name = app_name + "_backup_" + app_version
print("  备份名: " + backup_name)

// 步骤 3: 部署新版本
print("[3/4] 部署新版本...")
print("  配置: workers=" + str(config["worker"]) + " timeout=" + str(config["timeout"]))

// 步骤 4: 验证
print("[4/4] 验证部署...")

report {
    app: app_name,
    version: app_version,
    deploy_dir: deploy_dir,
    config: config,
    status: "deployed",
    steps_completed: 4
}

log("部署完成: " + app_name + " v" + app_version)
```

**运行命令**:

```bash
opsctl run examples/deploy_app.ops
```

**预期输出**:

```
开始部署 myapp v2.1.0
[1/4] 检查部署目录...
[2/4] 创建备份...
  备份名: myapp_backup_2.1.0
[3/4] 部署新版本...
  配置: workers=4 timeout=30
[4/4] 验证部署...
[REPORT] {
  "app": "myapp",
  "version": "2.1.0",
  "deploy_dir": "/opt/app",
  "config": {
    "worker": 4,
    "timeout": 30,
    "log_level": "info"
  },
  "status": "deployed",
  "steps_completed": 4
}
[LOG] 部署完成: myapp v2.1.0
```

---

## 5. 监控告警脚本

采集多项指标，输出 metric 数据点，按阈值生成多级告警。

**文件**: `examples/monitoring_alert.ops`

```ops
// 监控告警脚本
fn get_metric(name) {
    // 模拟指标获取（实际环境调用 sys.* 函数）
    let metrics = {
        "cpu": 75.5,
        "memory": 82.3,
        "disk": 45.0,
        "load": 2.5
    }
    return metrics[name]
}

// 阈值配置
let thresholds = {
    "cpu_warn": 70,
    "cpu_crit": 90,
    "mem_warn": 80,
    "mem_crit": 95,
    "disk_warn": 80
}

// 采集指标
let cpu = get_metric("cpu")
let mem = get_metric("memory")
let disk = get_metric("disk")

// 输出指标
metric("cpu_usage", cpu, {"host": "server-01"})
metric("memory_usage", mem, {"host": "server-01"})
metric("disk_usage", disk, {"host": "server-01"})

// 检查告警
if cpu > thresholds["cpu_crit"] {
    alert("[CRITICAL] CPU 使用率: " + str(cpu) + "%")
} else if cpu > thresholds["cpu_warn"] {
    alert("[WARNING] CPU 使用率: " + str(cpu) + "%")
}

if mem > thresholds["mem_crit"] {
    alert("[CRITICAL] 内存使用率: " + str(mem) + "%")
} else if mem > thresholds["mem_warn"] {
    alert("[WARNING] 内存使用率: " + str(mem) + "%")
}

if disk > thresholds["disk_warn"] {
    alert("[WARNING] 磁盘使用率: " + str(disk) + "%")
}

report {
    cpu: cpu,
    memory: mem,
    disk: disk,
    alerts_generated: true
}
```

**运行命令**:

```bash
opsctl run examples/monitoring_alert.ops
```

**预期输出**:

```
[METRIC] cpu_usage = 75.5 {"host": "server-01"}
[METRIC] memory_usage = 82.3 {"host": "server-01"}
[METRIC] disk_usage = 45.0 {"host": "server-01"}
[ALERT] [WARNING] CPU 使用率: 75.5%
[ALERT] [WARNING] 内存使用率: 82.3%
[REPORT] {
  "cpu": 75.5,
  "memory": 82.3,
  "disk": 45.0,
  "alerts_generated": true
}
```

> 说明：
> - CPU 75.5% > 70 (warn) 但 < 90 (crit)，触发 WARNING
> - 内存 82.3% > 80 (warn) 但 < 95 (crit)，触发 WARNING
> - 磁盘 45.0% < 80 (warn)，不触发告警

---

## 6. ensure 幂等操作

声明式幂等：条件为真时跳过后执行动作体，条件为假时执行动作恢复期望状态。

**文件**: `examples/ensure_example.ops`

```ops
// 声明式幂等示例
// ensure 语法: ensure <条件> { <动作> }
// 语义: 检查条件 -> 如果为假则执行动作 -> 再次验证条件

let config_path = "/etc/myapp/config.yaml"
let expected_content = "workers: 4\ntimeout: 30"

// 确保配置文件存在且内容正确
// ensure file.exists(config_path) {
//     file.write(config_path, expected_content)
// }

// 简化演示
let file_exists = true

ensure file_exists {
    print("创建配置文件: " + config_path)
}

print("配置文件已就绪")

// ensure 的 check -> apply -> verify 语义:
// 1. CHECK: 评估条件 (file_exists == true)
// 2. APPLY: 条件为假时执行动作体
// 3. VERIFY: 重新评估条件，若仍为假则报错
```

**运行命令**:

```bash
opsctl run examples/ensure_example.ops
```

**预期输出**:

```
配置文件已就绪
```

> 说明：`file_exists` 为 `true`，条件成立，`ensure` 块内的 `print` 不会执行。
>
> 如果把 `let file_exists = true` 改为 `let file_exists = false`，输出会变为：
> ```
> 创建配置文件: /etc/myapp/config.yaml
> 配置文件已就绪
> ```

---

## 7. 函数与递归

演示基本函数、默认参数、递归和列表操作。

**文件**: `examples/functions_closures.ops`

```ops
// 函数特性演示

// 基本函数
fn add(a, b) {
    return a + b
}

// 默认参数
fn greet(name, prefix = "Hello") {
    return prefix + ", " + name + "!"
}

// 递归函数
fn factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

// 使用
print(add(3, 4))
print(greet("OpsLang"))
print(greet("World", "Hi"))
print(factorial(5))

// 列表操作
fn sum_list(items) {
    let total = 0
    for let i = 0; i < len(items); i = i + 1 {
        total = total + items[i]
    }
    return total
}

let numbers = [1, 2, 3, 4, 5]
print("Sum: " + str(sum_list(numbers)))
```

**运行命令**:

```bash
opsctl run examples/functions_closures.ops
```

**预期输出**:

```
7
Hello, OpsLang!
Hi, World!
120
Sum: 15
```

---

## 8. 远程任务声明

`task ... on targets` 语法声明远程执行任务。当前版本 task 体内代码在本地执行，Phase 4 将实现真正的远程执行。

**文件**: `examples/remote_task.ops`

```ops
// 远程任务声明示例
// task 语法定义要在哪些目标上执行什么操作
// 当前版本: task 体内代码在本地执行
// Phase 4 将实现真正的远程执行

task "collect_info" on targets {
    let hostname = "remote-server"
    let cpu = 25.0
    let mem = 4096

    report {
        host: hostname,
        cpu: cpu,
        memory: mem
    }
}

task "deploy_config" on targets {
    let config = "/etc/app/config.yaml"
    print("部署配置到: " + config)

    report {
        target: config,
        status: "deployed"
    }
}
```

**运行命令**:

```bash
opsctl run examples/remote_task.ops
```

**预期输出**:

```
[REPORT] {
  "host": "remote-server",
  "cpu": 25.0,
  "memory": 4096
}
部署配置到: /etc/app/config.yaml
[REPORT] {
  "target": "/etc/app/config.yaml",
  "status": "deployed"
}
```

---

## 快速参考

### 数据类型

| 类型 | 示例 |
|------|------|
| int | `42`, `-1`, `0` |
| float | `3.14`, `-0.5`, `1.0` |
| string | `"hello"`, `"world"` |
| bool | `true`, `false` |
| list | `[1, 2, 3]`, `["a", "b"]` |
| dict | `{"key": "value"}`, `{"a": 1, "b": 2}` |
| nil | `nil` |

### 关键字

```
let  fn  if  else  for  while  return
task  on  import  true  false  nil
report  alert  ensure
```

### 内置函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `print(x)` | 打印到控制台 | `print("hello")` |
| `len(x)` | 获取长度（list/dict/string） | `len([1, 2, 3])` -> `3` |
| `str(x)` | 转换为字符串 | `str(42)` -> `"42"` |
| `int(x)` | 转换为整数 | `int("42")` -> `42` |
| `float(x)` | 转换为浮点数 | `float("3.14")` -> `3.14` |
| `type(x)` | 获取类型名 | `type(42)` -> `"int"` |
| `log(msg)` | 输出日志 | `log("操作完成")` |
| `metric(name, value, labels)` | 输出指标数据点 | `metric("cpu", 75.5, {"host": "s1"})` |
| `report { ... }` | 输出结构化报告 | `report { status: "ok" }` |
| `alert(msg)` | 输出告警 | `alert("异常!")` |

### 运算符

| 类别 | 运算符 |
|------|--------|
| 算术 | `+` `-` `*` `/` `%` |
| 比较 | `==` `!=` `<` `>` `<=` `>=` |
| 逻辑 | `&&` `\|\|` `!` |
| 赋值 | `=` |

### 结构化输出

```ops
// 报告 - 输出 JSON 结构化数据
report {
    host: "server-01",
    cpu: 45.2,
    status: "ok"
}

// 告警 - 高亮显示
alert("CPU 使用率过高!")

// 日志 - 带时间戳
log("操作完成")

// 指标 - 可被监控系统采集
metric("cpu_usage", 45.2, {"host": "server-01"})
```
