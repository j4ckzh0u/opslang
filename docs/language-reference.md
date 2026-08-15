# OpsLang 语言参考

## 目录

1. [概述](#1-概述)
2. [词法结构](#2-词法结构)
3. [数据类型](#3-数据类型)
4. [变量声明与赋值](#4-变量声明与赋值)
5. [运算符](#5-运算符)
6. [控制流](#6-控制流)
7. [函数定义](#7-函数定义)
8. [内置函数](#8-内置函数)
9. [结构化输出](#9-结构化输出)
10. [声明式幂等](#10-声明式幂等)
11. [任务声明](#11-任务声明)
12. [导入](#12-导入)
13. [完整语法示例](#13-完整语法示例)

---

## 1. 概述

OpsLang 是一门面向运维领域的领域特定语言（DSL），语法风格为类 Go + Python 混合。采用静态类型与类型推断，关键字数量控制在 16 个以内，力求用最少的语法表达完整的运维意图。

**关键字一览：**

| 关键字 | 用途 |
|--------|------|
| `let` | 变量声明 |
| `fn` | 函数定义 |
| `if` / `else` | 条件分支 |
| `for` | C 风格循环 |
| `while` | 条件循环 |
| `return` | 函数返回 |
| `task` / `on` | 任务声明 |
| `import` | 模块导入 |
| `true` / `false` | 布尔字面量 |
| `nil` | 空值 |
| `report` | 结构化报告输出 |
| `alert` | 告警输出 |
| `ensure` | 声明式幂等 |

---

## 2. 词法结构

### 2.1 注释

仅支持单行注释，使用 `//` 开头，到行尾结束：

```ops
// 这是一条注释
let x = 42  // 行尾注释也是合法的
```

不支持多行注释（`/* ... */`）。

### 2.2 标识符

标识符用于命名变量、函数等。规则：

- 以字母或下划线开头
- 后续字符可以是字母、数字、下划线
- 区分大小写

```ops
let hostname = "server1"     // 合法
let _count = 0               // 合法（下划线开头）
let diskUsage2 = 75.3        // 合法（包含数字）
// let 2ndDisk = 10           // 非法（数字开头）
```

### 2.3 字面量

**整数（int）：**

```ops
42
-1
0
1000000
```

**浮点数（float）：**

```ops
3.14
-0.5
0.0
1.0e3
```

**字符串（string）：** 双引号包裹

```ops
"hello"
"server-01.example.com"
""    // 空字符串
```

**布尔值（bool）：**

```ops
true
false
```

**空值（nil）：**

```ops
nil
```

**列表（list）：** 方括号包裹，元素用逗号分隔，允许混合类型

```ops
[1, 2, 3]
["a", "b", "c"]
[1, "hello", true, nil]    // 混合类型合法
[]                          // 空列表
```

**字典（dict）：** 花括号包裹，键值对用冒号分隔

```ops
{"host": "localhost", "port": 8080}
{"name": "ops", "version": 1}
{}    // 空字典
```

---

## 3. 数据类型

OpsLang 有 7 种数据类型：

| 类型 | 说明 | 示例 |
|------|------|------|
| `int` | 64 位整数（int64） | `42`, `-1` |
| `float` | 64 位浮点数（float64） | `3.14`, `-0.5` |
| `string` | 双引号字符串 | `"hello"` |
| `bool` | 布尔值 | `true`, `false` |
| `list` | 有序列表，可混合类型 | `[1, 2, 3]` |
| `dict` | 键值对集合 | `{"key": "value"}` |
| `nil` | 空值 | `nil` |

可用 `type()` 内置函数查看任意值的类型名称：

```ops
type(42)       // "int"
type(3.14)     // "float"
type("hello")  // "string"
type(true)     // "bool"
type([1, 2])   // "list"
type({"a": 1}) // "dict"
type(nil)      // "nil"
```

---

## 4. 变量声明与赋值

### 4.1 声明（let）

使用 `let` 关键字声明新变量，必须同时初始化：

```ops
let x = 42
let name = "server1"
let ratio = 0.75
let enabled = true
let items = [1, 2, 3]
let config = {"host": "localhost"}
```

`let` 在同一作用域中不可重复声明同一变量名。

### 4.2 赋值（=）

对已存在的变量重新赋值，使用 `=`。赋值操作会沿着作用域链向上查找已声明的变量并修改其值：

```ops
let x = 10    // 声明
x = 20        // 重新赋值

let count = 0
count = count + 1   // count 现在是 1
```

如果变量未在当前作用域链中声明，赋值会报错。

---

## 5. 运算符

### 5.1 算术运算符

| 运算符 | 说明 | 示例 |
|--------|------|------|
| `+` | 加法 / 字符串拼接 | `1 + 2`, `"a" + "b"` |
| `-` | 减法 | `10 - 3` |
| `*` | 乘法 | `3 * 4` |
| `/` | 除法 | `10 / 3` |
| `%` | 取模 | `10 % 3` |

**类型自动提升规则：** 如果任意操作数为 `float`，结果为 `float`：

```ops
let a = 1 + 2       // 3（int）
let b = 1.0 + 2     // 3.0（float）
let c = 10 / 3      // 3.333...（float，因为除法默认返回 float）
```

**字符串拼接：** `+` 运算符用于字符串时执行拼接。非字符串类型会自动转换为字符串再拼接：

```ops
let greeting = "Hello" + ", " + "World"   // "Hello, World"
let msg = "CPU: " + str(95.5) + "%"       // "CPU: 95.5%"
```

### 5.2 比较运算符

| 运算符 | 说明 | 示例 |
|--------|------|------|
| `==` | 等于 | `x == 10` |
| `!=` | 不等于 | `x != 0` |
| `<` | 小于 | `x < 100` |
| `>` | 大于 | `x > 0` |
| `<=` | 小于等于 | `x <= 50` |
| `>=` | 大于等于 | `x >= 1` |

比较表达式返回 `bool` 类型（`true` 或 `false`）。

### 5.3 逻辑运算符

| 运算符 | 说明 | 示例 |
|--------|------|------|
| `&&` | 逻辑与 | `a && b` |
| `\|\|` | 逻辑或 | `a \|\| b` |
| `!` | 逻辑非 | `!flag` |

逻辑运算符支持短路求值：`&&` 左侧为假时不计算右侧，`||` 左侧为真时不计算右侧。

### 5.4 赋值运算符

| 运算符 | 说明 | 示例 |
|--------|------|------|
| `=` | 赋值 | `x = 10` |

---

## 6. 控制流

### 6.1 if / else 条件分支

```ops
if x > 10 {
    print("big")
} else if x > 5 {
    print("medium")
} else {
    print("small")
}
```

`if` 条件不要求加括号。花括号 `{}` 是必需的。

**if 作为表达式：** `if/else` 可以当表达式使用，类似三元运算符：

```ops
let status = if score >= 60 { "pass" } else { "fail" }
let value = if enabled { 1 } else { 0 }
```

**Truthy 规则：** 以下值在条件判断中视为 `false`：

| 值 | 等价布尔值 |
|----|-----------|
| `nil` | false |
| `false` | false |
| `0`（int） | false |
| `0.0`（float） | false |
| `""`（空字符串） | false |
| `[]`（空列表） | false |
| `{}`（空字典） | false |

其他所有值均视为 `true`。

```ops
if "" {
    // 不会执行
}
if "hello" {
    // 会执行（非空字符串为 true）
}
if [] {
    // 不会执行
}
if [1] {
    // 会执行（非空列表为 true）
}
```

### 6.2 for 循环

C 风格的 for 循环，包含初始化、条件、更新三部分：

```ops
for let i = 0; i < 10; i = i + 1 {
    print(i)
}
```

语法结构：`for <初始化语句>; <条件表达式>; <更新语句> { <循环体> }`

```ops
// 遍历列表
let items = [10, 20, 30]
for let i = 0; i < len(items); i = i + 1 {
    print(items[i])
}

// 嵌套循环
for let i = 0; i < 3; i = i + 1 {
    for let j = 0; j < 3; j = j + 1 {
        print(i * j)
    }
}
```

### 6.3 while 循环

当条件为真时持续执行循环体：

```ops
let x = 10
while x > 0 {
    print(x)
    x = x - 1
}
```

等价于省略初始化和更新语句的 for 循环。

---

## 7. 函数定义

### 7.1 基本语法

使用 `fn` 关键字定义函数：

```ops
fn add(a, b) {
    return a + b
}

let result = add(3, 5)   // 8
```

函数参数不需要声明类型，返回类型也不需要显式标注。

### 7.2 默认参数

函数参数可以设置默认值，调用时可以省略带默认值的参数：

```ops
fn greet(name, greeting = "Hello") {
    return greeting + ", " + name
}

greet("Alice")              // "Hello, Alice"
greet("Bob", "Hi")          // "Hi, Bob"
```

### 7.3 闭包

函数是闭包，会捕获定义时的环境。内部函数可以访问外部函数的变量：

```ops
fn make_counter() {
    let count = 0
    fn increment() {
        count = count + 1
        return count
    }
    return increment
}

let counter = make_counter()
print(counter())   // 1
print(counter())   // 2
print(counter())   // 3
```

### 7.4 递归

函数可以调用自身：

```ops
fn factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}

print(factorial(5))   // 120
```

---

## 8. 内置函数

OpsLang 提供以下内置函数，无需导入即可使用：

### print

输出值到标准输出。支持多个参数：

```ops
print("Hello, World")
print("name:", "ops", "version:", 1)
print(42)
print([1, 2, 3])
```

### len

返回字符串、列表或字典的长度：

```ops
len("hello")       // 5
len([1, 2, 3])     // 3
len({"a": 1})      // 1
len("")            // 0
```

### str

将任意值转换为字符串：

```ops
str(42)        // "42"
str(3.14)      // "3.14"
str(true)      // "true"
str(nil)       // "nil"
str([1, 2])    // "[1, 2]"
```

### int

将值转换为整数（int64）：

```ops
int(3.14)      // 3（浮点截断）
int("42")      // 42（字符串解析）
int(true)      // 1
int(false)     // 0
int(nil)       // 0
```

### float

将值转换为浮点数（float64）：

```ops
float(42)      // 42.0
float("3.14")  // 3.14
float(true)    // 1.0
float(nil)     // 0.0
```

### type

返回值的类型名称字符串：

```ops
type(42)       // "int"
type(3.14)     // "float"
type("hello")  // "string"
type(true)     // "bool"
type([1, 2])   // "list"
type({"a": 1}) // "dict"
type(nil)      // "nil"
```

### log

输出日志信息。用于记录脚本运行过程中的信息：

```ops
log("Task started")
log("Processing item: " + str(i))
```

### metric

输出度量指标，支持可选的标签字典：

```ops
metric("cpu_usage", 45.2, {"host": "server1"})
metric("memory_bytes", 8192000, {"host": "server1", "env": "prod"})
metric("request_count", 1024)
```

参数说明：
- 第一个参数：指标名称（string）
- 第二个参数：指标值（数值类型）
- 第三个参数（可选）：标签字典（dict）

---

## 9. 结构化输出

OpsLang 内置结构化输出机制，用于产出可供自动化系统解析的结果。

### report

输出一组键值对作为结构化报告：

```ops
report {
    host: "server1",
    cpu: 45.2,
    mem: 8192,
    disk: "/dev/sda1"
}
```

`report` 块内的语法为 `键: 值`，多个键值对用逗号分隔。值可以是任意表达式。

### alert

输出告警信息：

```ops
if cpu_usage > 90 {
    alert("CPU 使用率过高: " + str(cpu_usage) + "%")
}

alert("磁盘空间不足")
```

### log（结构化输出模式）

除内置函数外，`log` 也属于结构化输出的一部分，用于记录过程日志：

```ops
log("开始检查系统状态")
log("检查完成，共处理 " + str(count) + " 台主机")
```

### metric（结构化输出模式）

`metric` 用于输出可被监控系统采集的指标数据：

```ops
metric("cpu_usage", 45.2, {"host": "server1"})
metric("request_latency_ms", 120, {"endpoint": "/api/health"})
```

---

## 10. 声明式幂等

`ensure` 关键字用于声明期望的系统状态。系统会自动执行"检查 → 应用 → 验证"三步流程：

```ops
ensure file.exists("/etc/myapp/config") {
    file.write("/etc/myapp/config", "default config content")
}
```

**执行逻辑：**

1. **检查（Check）：** 先执行 `ensure` 后面的条件表达式
2. **应用（Apply）：** 条件不满足时，执行 `{}` 内的操作
3. **验证（Verify）：** 操作执行后再次检查条件是否满足

```ops
// 确保服务正在运行
ensure service.status("nginx").running {
    service.start("nginx")
}

// 确保配置文件内容正确
ensure file.checksum("/etc/app.conf", "sha256") == expected_hash {
    file.write("/etc/app.conf", correct_content)
}
```

`ensure` 天然支持 `--dry-run` 模式：在 dry-run 下只输出"需要做什么"而不实际执行变更操作。

---

## 11. 任务声明

`task ... on` 语法用于声明需要在目标主机上执行的任务：

```ops
task "deploy" on targets {
    let cpu = sys.cpu.usage()
    let mem = sys.memory.info()

    report {
        host: sys.hostname(),
        cpu: cpu,
        mem: mem
    }
}
```

语法结构：`task "<任务名称>" on <目标> { <任务体> }`

**targets 的取值方式（Phase 4 完整实现）：**

```ops
// 内联主机列表
task "check" on ["host1", "host2", "host3"] {
    // ...
}

// 从变量引用
let targets = ["web1", "web2"]
task "deploy" on targets {
    // ...
}

// 分组表达式
task "collect" on group("env=prod") {
    // ...
}
```

**注意：** 当前阶段（Phase 2），任务体内的代码在本地执行。远程执行能力将在 Phase 4 中实现。

### parallel 块

`parallel` 块用于声明其中的操作可以并行执行：

```ops
task "multi_check" on targets {
    parallel {
        let cpu = sys.cpu.usage()
        let mem = sys.memory.info()
        let disk = sys.disk.usage("/")
    }

    report {
        cpu: cpu,
        mem: mem,
        disk: disk
    }
}
```

---

## 12. 导入

使用 `import` 关键字导入模块或标准库：

```ops
import sys
import file
import net
```

导入后可以调用模块中的函数：

```ops
import sys

let cpu = sys.cpu.usage()
let mem = sys.memory.info()
print("CPU: " + str(cpu.percent) + "%")
```

---

## 13. 完整语法示例

### 示例 1：系统信息采集

```ops
// 采集主机基本信息并输出报告
let hostname = sys.hostname()
let cpu = sys.cpu.usage()
let mem = sys.memory.info()
let disks = sys.disk.usage("/")

if cpu.percent > 90 {
    alert("CPU 使用率过高: " + str(cpu.percent) + "%")
}

if mem.used_percent > 85 {
    alert("内存使用率过高: " + str(mem.used_percent) + "%")
}

report {
    host: hostname,
    cpu: cpu,
    mem: mem,
    disk: disks
}
```

### 示例 2：函数与循环

```ops
// 批量检查多个端口
fn check_port(host, port) {
    let result = net.tcp_check(host, port)
    if result.connected {
        log("端口 " + str(port) + " 开放")
    } else {
        alert("端口 " + str(port) + " 不可达")
    }
    return result
}

let ports = [22, 80, 443, 8080]
let results = []

for let i = 0; i < len(ports); i = i + 1 {
    let r = check_port("localhost", ports[i])
    results = results + [r]
}

report {
    checked_ports: ports,
    results: results
}
```

### 示例 3：条件赋值与 if 表达式

```ops
// 根据环境设置不同参数
let env = "production"
let max_conn = if env == "production" { 1000 } else { 100 }
let log_level = if env == "production" { "warn" } else { "debug" }

let config = {
    "max_connections": max_conn,
    "log_level": log_level,
    "timeout": 30
}

log("当前环境: " + env)
log("最大连接数: " + str(max_conn))

report {
    config: config
}
```

### 示例 4：声明式幂等

```ops
// 确保应用配置正确
let expected_config = "worker_processes auto;\nlisten 80;\n"

ensure file.exists("/etc/nginx/conf.d/app.conf") {
    file.write("/etc/nginx/conf.d/app.conf", expected_config)
    log("已写入 nginx 配置")
}

ensure service.status("nginx").running {
    service.start("nginx")
    log("已启动 nginx 服务")
}

metric("config_managed", 1, {"file": "/etc/nginx/conf.d/app.conf"})
```

### 示例 5：任务声明

```ops
// 多主机信息采集任务
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

### 示例 6：闭包与高阶函数模式

```ops
// 使用闭包实现简单的状态跟踪
fn make_accumulator(initial) {
    let total = initial

    fn add(value) {
        total = total + value
        return total
    }

    fn get_total() {
        return total
    }

    return {"add": add, "get_total": get_total}
}

let acc = make_accumulator(0)
acc["add"](10)
acc["add"](20)
let result = acc["get_total"]()   // 30

report {
    total: result
}
```

---

## 附录：索引与成员访问

### 列表索引

使用方括号访问列表元素，索引从 0 开始：

```ops
let items = [10, 20, 30]
print(items[0])    // 10
print(items[2])    // 30
```

### 字典访问

支持两种语法访问字典的值：

```ops
let config = {"host": "localhost", "port": 8080}

// 方括号语法
print(config["host"])      // "localhost"

// 点号语法（成员访问）
print(config.host)         // "localhost"
print(config.port)         // 8080
```

两种语法等价。方括号语法支持动态键名，点号语法要求键名为合法标识符。
