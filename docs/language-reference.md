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
14. [Roadmap（未实现特性）](#14-roadmap未实现特性)

---

## 1. 概述

OpsLang 是一门面向运维领域的领域特定语言（DSL），语法风格为类 Go + Python 混合。OpsLang 是**动态类型**语言（变量无类型标注，运行期决定类型），关键字共 20 个，力求用最少的语法表达完整的运维意图。

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
| `import` | 模块导入（声明性） |
| `privilege` | 脚本权限级别声明（`privilege: read_only \| admin \| root`，须置于脚本顶部；未声明默认 `read_only`。变更类函数（file.write、process.exec、service.* 等，清单见 opsspec）需要至少 `admin`，违反时解释器运行时报错、AOT 编译期报错、Runner 执行前二次校验拒绝） |
| `true` / `false` | 布尔字面量 |
| `nil` | 空值 |
| `report` | 结构化报告输出 |
| `alert` | 告警输出 |
| `ensure` | 声明式幂等 |
| `metric` | 指标输出 |
| `log` | 日志输出 |
| `parallel` | 并行块 |

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

**字符串（string）：** 双引号包裹（支持 `\n`、`\t`、`\"`、`\\` 转义）

```ops
"hello"
"server-01.example.com"
""              // 空字符串
"line1\nline2"  // 换行
```

反引号包裹的**原始字符串**不做任何转义，适合书写正则、Windows 路径与多行文本：

```ops
`C:\Users\admin`      // 就是字面的 C:\Users\admin，反斜杠不转义
`d{4}-d{2}-d{2}`     // 正则模式常用
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

`let` 在同一作用域中不可重复声明同一变量名（重复声明报错；重新赋值用 `=`）。

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

**类型提升与整除规则：** 两个 `int` 相除为**整除**；任意操作数为 `float` 时结果为 `float`：

```ops
let a = 1 + 2       // 3（int）
let b = 1.0 + 2     // 3.0（float）
let c = 10 / 3      // 3（int 整除，舍去小数）
let d = 10 / 3.0    // 3.333...（float）
```

除零和模零均为运行时错误。

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

**相等比较是严格的：** 类型不同的值不相等（`1 != "1"`）；数值在 int/float 之间可以跨类型按数值比较（`1 == 1.0` 为 `true`）。

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

C 风格的 for 循环是唯一支持的 for 形式，包含初始化、条件、更新三部分：

```ops
for let i = 0; i < 10; i = i + 1 {
    print(i)
}
```

语法结构：`for <初始化语句>; <条件表达式>; <更新语句> { <循环体> }`

```ops
// 遍历列表（按下标）
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

### 6.3 for-in 遍历循环

遍历 list、dict（迭代键）或 string（迭代字符）：

```ops
let services = ["nginx", "redis", "postgres"]
for svc in services {
    print("checking " + svc)
}

let config = {"host": "localhost", "port": 8080}
for key in config {
    print(key + " = " + str(config[key]))
}
```



### 6.4 while 循环

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

### 6.5 block / rescue / always 错误处理

对标 Ansible 的 `block/rescue/always`。`block` 内任意语句出错时：跳到 `rescue`（可读取 `_error` 字符串获取真实错误），`always` 无论成败都执行：

```ops
block {
    let conf = file.read("/etc/myapp/app.conf")
    print("读取成功: " + conf.content)
}
rescue {
    print("读取失败: " + _error)   // _error 是真实错误信息，不是占位符
}
always {
    print("清理临时资源")
}
```

`rescue`/`always` 可省略；块可嵌套。这是编写"探测-降级"逻辑的标准写法（示例库大量使用：前提不满足时进入 rescue 打印真实原因并显式跳过）。

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

将值转换为整数（int64）。字符串采用**严格解析**，含非数字字符即报错：

```ops
int(3.14)      // 3（浮点截断）
int("42")      // 42（字符串严格解析）
int(true)      // 1
int(false)     // 0
int(nil)       // 0
int("42abc")   // 运行时错误（严格解析，不截断前缀）
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

### 数据操作内置函数

以下 12 个纯函数用于字符串、列表、字典的日常变换。它们是**语言级函数**而非原子操作：解释器与 AOT 编译产物行为完全一致；runner 模式部署会明确拒绝（线性指令 VM 无表达式求值器），需要它们的脚本请用 `--mode aot`。

#### split / join

```ops
split("10.0.0.1", ".")     // ["10", "0", "0", "1"]
join(["a", "b"], "-")      // "a-b"（元素自动 str 化）
```

#### replace / upper / lower / trim

```ops
replace("/etc/nginx/nginx.conf", "/", "_")  // "_etc_nginx_nginx.conf"
upper("web-01")            // "WEB-01"
lower("WEB-01")            // "web-01"
trim("  prod \n")          // "prod"
```

#### contains / index_of

`contains` 与 `index_of` 对字符串和列表多态；`contains` 还支持字典键检查。

```ops
contains("nginx.conf", "conf")   // true（子串）
contains([1, 2, 3], 2)           // true（深相等逐项比较）
contains({"env": "prod"}, "env") // true（键存在）
index_of("hello world", "world") // 6；找不到返回 -1
index_of(["a", "b"], "b")        // 1
```

#### sort / reverse

两者都返回**新列表**，绝不修改入参。`sort` 要求全部数值或全部字符串；混合类型是显式错误，不会给出静默顺序。

```ops
sort([3, 1, 2])              // [1, 2, 3]（原列表不变）
sort(["nginx", "apache"])    // ["apache", "nginx"]
reverse([1, 2, 3])           // [3, 2, 1]
sort([1, "a"])               // 运行时错误：混合列表不可排序
```

#### keys / values

字典无序，两者都按**排序后的键**输出且相互对应，解释器与 AOT 结果一致。

```ops
let d = {"z": 26, "a": 1}
keys(d)     // ["a", "z"]
values(d)   // [1, 26]
```

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

`ensure` 关键字用于声明期望的系统状态。完整语法：

```
ensure <条件表达式> {
    <应用动作>
} notify <表达式>
```

`notify` 子句可选。系统按"检查 → 应用 → 验证 → 通知"四步执行：

1. **检查（Check）：** 先执行 `ensure` 后面的条件表达式
2. **应用（Apply）：** 条件不满足时，执行 `{}` 内的操作
3. **验证（Verify）：** 操作执行后再次检查条件；若仍不满足，**脚本报错退出**
4. **通知（Notify）：** `notify` 表达式只在本次**实际发生了变更**（执行过 apply）之后才求值；条件本来就满足时不触发

```ops
// 确保目录存在，创建成功后输出通知
ensure file.exists("/etc/myapp").exists {
    file.mkdir("/etc/myapp")
} notify log("created /etc/myapp")

// 确保配置文件存在
ensure file.exists("/etc/app.conf").exists {
    file.write("/etc/app.conf", "default")
} notify alert("app.conf was missing and has been recreated")
```

**dry-run 支持：** `opsctl run --dry-run` 下，`ensure` 的 apply 步骤只打印将要执行的动作而不实际执行，check/verify 照常评估。

```ops
// 判断条件可以是任意表达式，例如校验和比较
ensure file.checksum("/etc/app.conf", "sha256").checksum == expected_hash {
    file.write("/etc/app.conf", correct_content)
} notify log("app.conf rewritten")
```

---

## 11. 任务声明

`task ... on` 语法用于声明需要在目标主机上执行的任务：

```ops
task "deploy" on "web1" {
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

**执行方式（重要）：**

- `opsctl run` 本地执行时，带 `on` 子句的 task 会**报错**并提示改用 `opsctl deploy`（本地执行无法路由到远程主机）；不带 `on` 的 task 体在本地执行。
- `opsctl deploy` 下目标路由生效：task 体只发送到 `on` 子句选中的目标主机；task 之外的顶层语句发送到全部目标。

**on 子句的选择器（deploy 模式）：**

```ops
// 精确主机名 / user@host，或 glob 模式
task "check" on "web1" { }
task "check" on "deploy@web1.example.com" { }
task "check" on "web*" { }
```

选择器按精确名、主机地址、`user@host` 三种形式与 deploy 目标列表匹配，支持 `path.Match` 语法的 glob。以下写法在 deploy 时会**报错拒绝**：

```ops
let targets = ["web1", "web2"]
task "deploy" on targets { }        // 错误：变量选择器无法在部署期解析

task "collect" on group("env=prod") { }  // 错误：动态选择器不支持
```

若 `on` 子句选不中任何 deploy 目标，deploy 同样报错。

### parallel 块

`parallel` 块内的语句并发执行（每个语句一个 goroutine）：

```ops
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
```

**语义：**

- 块内各语句并发执行，互不共享中间写入
- 块内的 `let` 声明按**源码顺序**确定性合并回外层作用域，因此执行结束后外层可以访问这些变量
- 任一语句出错则整个 parallel 块报错

### parallel for 扇出循环

对列表的每个元素**并发**执行同一组语句——批量检查、批量探测的标准写法：

```ops
let servers = ["web-01", "web-02", "db-01"]
parallel for s in servers {
    print("checking " + s + " ...")
}
```

**语义：**

- 每个元素一个 goroutine，循环变量在迭代内可见
- 迭代环境相互隔离：`let` 与标识符赋值按**源码顺序**确定性合并（最后一次迭代的值胜出），与顺序执行的 for-in 结果一致
- 循环变量在语句结束后不泄漏到外层
- 列表为空时整体跳过；非列表是显式错误
- **禁止在 body 内对循环外声明的字典/列表做索引赋值**（如 `results[k] = v`）——那会跨 goroutine 写共享对象，属于数据竞争。解释器与 AOT 都会明确报错拒绝

runner 模式部署时 `parallel` 块与 `parallel for` 都会明确拒绝（线性指令 VM 无并发能力）；需要并发的脚本请用 `--mode auto` 或 `--mode aot`。

---

## 12. 导入

`import` 是**声明性**的：标准库内置函数全局注册，无需导入即可直接调用（`import sys` 之类的写法只是显式声明意图，不产生副作用）：

```ops
import sys

let cpu = sys.cpu.usage()   // 不写 import sys 也可以直接调用
print("CPU: " + str(cpu.percent) + "%")
```

**第三方 Go 库导入未实现：** `import "go <包路径>"` 在 `opsctl run` 与 `opsctl deploy/build` 中都会报错拒绝（见 Roadmap）。

### 用户模块：import "./path/to/file.ops"

以 `.ops` 结尾的导入路径是**文件模块**——把可复用的 `fn` 和 `let` 声明抽到独立文件，供多个脚本引用：

```ops
// lib/checks.ops —— 可复用的检查函数库
let warn_threshold = 80.0

fn disk_full(mount) {
    let u = sys.disk.usage(mount)
    return u.used_percent > warn_threshold
}
```

```ops
// main.ops —— 入口脚本
import "./lib/checks.ops"

if disk_full("/") {
    alert("根分区超过 " + str(warn_threshold) + "%")
}
```

**模块规则（刻意严格，保证团队库行为可预测）：**

- 模块顶层只允许 `fn`、`let` 和进一步的 `import`；`task`/`ensure` 等属于入口脚本，出现在模块里会报错并指出文件与行号
- 导入的声明在 import 语句的位置拼接入主程序；相对路径相对于**当前文件**解析
- 跨文件重名是错误，不会静默覆盖
- 循环导入报错并列出完整导入链；同一模块被多次导入只合并一次
- 解释器、runner 指令生成与 AOT 编译消费同一个链接后的程序，三引擎结果一致

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

ensure file.exists("/etc/nginx/conf.d/app.conf").exists {
    file.write("/etc/nginx/conf.d/app.conf", expected_config)
    log("已写入 nginx 配置")
}

ensure service.status("nginx").active {
    service.start("nginx")
    log("已启动 nginx 服务")
}

metric("config_managed", 1, {"file": "/etc/nginx/conf.d/app.conf"})
```

### 示例 5：任务声明

```ops
// 多主机信息采集任务（通过 opsctl deploy --targets web1,web2 执行）
task "collect_info" on "web*" {
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

## 14. Roadmap（未实现特性）

以下语法特性**尚未实现**，本文档不以现有功能描述它们：

- **`import "go <包路径>"` 第三方 Go 库**：所有执行引擎均报错拒绝
- **task `on` 变量 / 动态选择器**：deploy 只支持字面量主机选择器（精确名 / `user@host` / glob / inventory 组名）

> `for ... in ...` 遍历循环与 `block/rescue/always` 错误处理**已实现**（见 6.3 节与 `examples/block_rescue_example.ops`）。

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

**关键字成员名**：点号之后允许出现语言关键字（关键字在此位置没有语句含义）。标准库依赖这一规则提供了与关键字同名的操作，例如 `file.ensure(...)`、`user.absent(...)`、`group.ensure(...)`。
