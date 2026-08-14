# OpsLang 语言参考

## 1. 基础语法

### 1.1 变量

```ops
// 类型推断
name = "OpsLang"
count = 42
pi = 3.14
active = true
nothing = nil

// 数组
hosts = ["web01", "web02", "web03"]

// 字典
config = {"host": "localhost", "port": 8080}
```

### 1.2 字符串

```ops
// 双引号：支持插值
greeting = "Hello {name}"

// 单引号：不做插值
raw = 'Hello {name}'  // 字面量

// 三引号：多行原始字符串
config = """
server:
  host: localhost
  port: 8080
"""

// 转义字符（仅双引号）
"hello\nworld"    // 换行
"tab\there"       // 制表符
"quote\"here"     // 引号
```

### 1.3 运算符

```ops
// 算术
1 + 2     // 3
10 - 3    // 7
3 * 4     // 12
15 / 3    // 5
10 % 3    // 1

// 比较
a == b    // 等于
a != b    // 不等于
a < b     // 小于
a <= b    // 小于等于
a > b     // 大于
a >= b    // 大于等于

// 逻辑
a && b    // 与
a || b    // 或
!a        // 非

// 管道
value |> fn    // 将 value 作为 fn 的第一个参数
```

---

## 2. 控制流

### 2.1 条件

```ops
if x > 10
    print("big")
else if x > 5
    print("medium")
else
    print("small")
```

### 2.2 循环

```ops
// for 循环
for host in hosts
    print(host)

// while 循环
i = 0
while i < 10
    print(i)
    i = i + 1

// break / continue
for i in range(100)
    if i == 5
        break
    if i % 2 == 0
        continue
    print(i)
```

### 2.3 函数

```ops
// 基本函数
fn greet(name)
    print("Hello, {name}")

// 默认参数
fn connect(host, port)
    print("Connecting to {host}:{port}")

// 返回值
fn add(a, b)
    return a + b

// Lambda
double = fn(x) => x * 2
list.map(fn(x) => x * 2)

// 闭包
fn make_counter()
    n = 0
    return fn() => n + 1
```

### 2.4 错误处理

```ops
try
    result = risky_operation()
catch e
    print("Error: {e}")

// try/catch 可以嵌套
try
    try
        inner_risky()
    catch e
        print("Inner error: {e}")
catch e
    print("Outer error: {e}")
```

---

## 3. 标准库

### 3.1 file — 文件操作

```ops
content = file.read("/etc/hosts")
file.write("/tmp/test.txt", "hello")
exists = file.exists("/tmp/test.txt")
file.mkdir("/tmp/newdir")
files = file.list("/tmp")
base = file.basename("/tmp/test.txt")   // "test.txt"
dir = file.dirname("/tmp/test.txt")     // "/tmp"
```

### 3.2 process — 进程与环境

```ops
result = process.shell("ls -la")        // 执行 shell 命令
print(result.stdout)                    // 输出
print(result.exitCode)                  // 退出码
print(result.ok)                        // 是否成功

result = process.run("git", "status")   // 直接执行命令
cwd = process.cwd()                     // 当前目录
host = process.hostname()               // 主机名
path = process.env("PATH")              // 环境变量
```

### 3.3 ssh — 远程执行

```ops
result = ssh.run("web01", "uptime")           // SSH 远程执行
result = ssh.run("web01", "ls", "admin")      // 指定用户

ssh.copy("/local/file", "web01", "/remote/")  // SCP 传输
alive = ssh.ping("web01")                      // 连通性检测
```

### 3.4 fleet — 批量执行引擎

```ops
// 串行执行
results = fleet.serial(hosts, fn(h) => ssh.run(h, "uptime"))

// 并行执行（可指定并发数）
results = fleet.parallel(hosts, fn(h) => ssh.run(h, "uptime"), 10)

// 批量 SSH 命令
results = fleet.exec(hosts, "uptime", "root")

// 结果汇总
summary = fleet.summary(results)
print("total: {summary.total}, ok: {summary.ok}, fail: {summary.fail}")
```

### 3.5 json — JSON 处理

```ops
data = json.parse('{"name": "test"}')
dumped = json.dump(data)
pretty = json.prettify(data)

// 文件操作
data = json.load_file("config.json")
json.save_file("output.json", data)
```

### 3.6 yaml — YAML 处理

```ops
data = yaml.parse("name: test\nversion: 1")
dumped = yaml.dump(data)

// 文件操作
data = yaml.load_file("config.yaml")
yaml.save_file("output.yaml", data)
```

### 3.7 toml — TOML 处理

```ops
data = toml.parse('[server]\nhost = "localhost"')
data = toml.load_file("config.toml")
```

### 3.8 strings — 字符串工具

```ops
parts = strings.split("a,b,c", ",")
joined = strings.join(["a", "b", "c"], ", ")
has = strings.contains("hello world", "world")
replaced = strings.replace("hello", "hello", "hi")
trimmed = strings.trim("  hello  ")
upper = strings.upper("hello")
lower = strings.lower("HELLO")
```

### 3.9 ensure — 声明式资源管理

```ops
// 文件
result = ensure.file("/etc/motd", "Welcome!")
// result.changed: 是否做了更改
// result.ok: 是否成功

// 目录
ensure.dir("/opt/app")

// 文件行
ensure.line("/etc/hosts", "10.0.0.1 web01")

// 服务
ensure.service("nginx", "running", true)

// 包
ensure.package("curl", "present")

// 用户
ensure.user("deploy", "/bin/bash", "sudo,docker")
```

### 3.10 inventory — 主机清单

```ops
// 从 INI 文件加载
inv = inventory.load("/etc/ops/hosts.ini")

// 从数组创建
inv = inventory.from_list(["web01", "web02", "db01"])

// 获取分组
webs = inventory.group(inv, "web_servers")
all = inventory.all(inv)
```

---

## 4. 链式方法

### 4.1 字符串方法

```ops
"hello".upper()                    // "HELLO"
"HELLO".lower()                    // "hello"
"  hello  ".trim()                 // "hello"
"a,b,c".split(",")                 // ["a", "b", "c"]
"hello world".contains("world")    // true
"hello".replace("hello", "hi")     // "hi"
"hello".starts_with("he")          // true
"hello".ends_with("lo")            // true
"hello".len()                      // 5
"42".to_int()                      // 42

// 链式调用
"  HELLO  ".trim().lower().len()   // 5
```

### 4.2 数组方法

```ops
[1,2,3].len()                      // 3
[1,2,3].append(4)                  // [1,2,3,4]
["a","b"].join(",")                // "a,b"
[1,2,3].reverse()                  // [3,2,1]
```

---

## 5. 内置函数

```ops
print("hello")             // 输出
len("hello")               // 5
len([1,2,3])              // 3
type(42)                   // "int"
str(42)                    // "42"
int("42")                  // 42
range(5)                   // [0,1,2,3,4]
range(1, 5)               // [1,2,3,4]
range(0, 10, 2)           // [0,2,4,6,8]
append([1,2], 3)          // [1,2,3]
map([1,2,3], fn(x) => x*2) // [2,4,6]
filter([1,2,3,4], fn(x) => x > 2) // [3,4]
```

---

## 6. 命令行

```bash
ops run script.ops          # 运行脚本
ops build script.ops        # 编译为单二进制
ops build script.ops out    # 指定输出文件名
ops repl                    # 交互式 REPL
ops check script.ops        # 语法检查
ops version                 # 版本信息
```

---

## 7. Shebang

```ops
#!/usr/bin/env ops run
// 第一行 shebang 会被自动跳过
print("可以直接执行!")
```

```bash
chmod +x script.ops
./script.ops
```
