# OpsLang 语言规范（草案）

## 1. 词法结构

### 1.1 注释

```ops
// 单行注释

/*
  多行注释
*/
```

### 1.2 标识符

标识符以字母或下划线开头，后跟字母、数字或下划线：

```ops
name
host_name
_web01
```

### 1.3 字面量

```ops
// 整数
42
0xFF        // 十六进制
0b1010      // 二进制

// 浮点
3.14
1.5e10

// 字符串
"hello"
"Hello {name}"   // 插值字符串

// 布尔
true
false

// nil
nil
```

---

## 2. 变量与类型

### 2.1 变量声明

```ops
// 类型推断
name = "web01"
count = 42
pi = 3.14
active = true

// 数组
hosts = ["web01", "web02", "web03"]

// 字典
config = {
    host: "localhost",
    port: 8080,
}
```

### 2.2 类型系统

渐进类型：脚本模式动态，生产模式可声明类型。

```ops
// 动态类型（默认）
x = 42
x = "hello"   // 允许

// 静态类型（显式声明）
let name: string = "web01"
let count: int = 42
let hosts: [string] = ["web01", "web02"]
```

---

## 3. 控制流

### 3.1 条件语句

```ops
if disk_usage > 80 {
    alert("磁盘空间不足")
} else if disk_usage > 60 {
    warn("磁盘空间警告")
} else {
    print("磁盘正常")
}
```

### 3.2 循环

```ops
// for-in 循环
for host in hosts {
    print(host)
}

// while 循环
while retry_count < 3 {
    result = ssh.run(host, "uptime")
    if result.exitCode == 0 {
        break
    }
    retry_count += 1
}
```

---

## 4. 函数

### 4.1 函数定义

```ops
fn check_disk(host, threshold = 80) {
    usage = ssh.run(host, "df -h /")
    if usage.to_int() > threshold {
        alert("磁盘使用率过高: {host}")
    }
}
```

### 4.2 Lambda

```ops
// Lambda 表达式
double = fn(x) => x * 2

// 在集合操作中使用
hosts.filter(fn(h) => h.starts_with("web"))
    .map(fn(h) => h.to_upper())
```

---

## 5. 错误处理

### 5.1 try/catch

```ops
try {
    result = ssh.run("unknown-host", "uptime")
} catch SSHError as e {
    print("SSH 连接失败: {e.message}")
} catch TimeoutError as e {
    print("超时: {e.message}")
}
```

---

## 6. 运维原生特性

### 6.1 Shell 互操作

```ops
// 执行 Shell 命令
output = sh("ls -la /tmp")

// 管道
output = sh("cat /var/log/syslog | grep error | wc -l")

// 重定向
sh("echo 'hello' > /tmp/test.txt")
```

### 6.2 SSH 批量执行

```ops
// 批量执行
results = fleet.parallel(hosts, fn(host) {
    return ssh.run(host, "uptime")
})

// 滚动执行
fleet.rolling_update(hosts, batch_size=2, fn(host) {
    ssh.run(host, "yum update -y")
    ssh.run(host, "systemctl restart app")
})
```

### 6.3 声明式资源管理

```ops
ensure.file("/etc/motd", content="Welcome!")
ensure.service("nginx", state="running", enabled=true)
ensure.package("curl", state="present", version="7.68")
```

---

## 7. 模块系统

```ops
// 导入标准库
import net.ssh
import data.yaml

// 导入本地模块
import "./lib/utils"

// 别名导入
import cloud.aws as aws
```

---

## 8. 运算符

| 运算符 | 说明 | 示例 |
|--------|------|------|
| `=` | 赋值 | `x = 1` |
| `==` | 等于 | `x == 1` |
| `!=` | 不等于 | `x != 1` |
| `<` `<=` `>` `>=` | 比较 | `x > 1` |
| `&&` `||` | 逻辑 | `a && b` |
| `+` `-` `*` `/` `%` | 算术 | `1 + 2` |
| `\|>` | 管道 | `list \|> filter(fn)` |
| `=>` | Lambda | `fn(x) => x + 1` |
