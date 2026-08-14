# OpsLang

**为运维而生的编程语言** — 类 Python 语法，Shell 般交互，单二进制部署。

```
  __  ____  ____  _____
 /  |/  \ \/ / / / ___/
/ /|/ / /\  / / __\ \
/_/ |_/ / / /_/ /___/
     Ops Automation Language
```

## 为什么选择 OpsLang？

| 痛点 | Shell | Python | **OpsLang** |
|------|-------|--------|-------------|
| 静默失败 | ✗ | ○ | ★ 默认严格 |
| 部署方式 | 复制文件 | 装运行时 | ★ **单二进制** |
| Shell 互操作 | ★ | ✗ 笨拙 | ★ **原生管道** |
| SSH 批量执行 | 手写循环 | 自己拼 | ★ **内置引擎** |
| 数据格式 | 手工解析 | 需装库 | ★ **原生 JSON/YAML/TOML** |
| 声明式管理 | ✗ | ✗ | ★ **ensure 语句** |
| 学习门槛 | 引号地狱 | 面向对象 | ★ **10 分钟上手** |

## 快速开始

### 安装

```bash
git clone https://github.com/opslang/opslang.git
cd opslang && make build
cp bin/ops /usr/local/bin/
```

### Hello World

```ops
// hello.ops
name = "运维工程师"
hosts = ["web01", "web02", "web03"]

for host in hosts
    print("检查: {host}")
```

```bash
$ ops run hello.ops
检查: web01
检查: web02
检查: web03
```

### 编译为单二进制

```bash
$ ops build hello.ops hello
✅ 编译成功: hello (2.4 MB)

$ ./hello        # 无需任何运行时依赖
检查: web01
检查: web02
检查: web03
```

### 批量运维

```ops
// batch_check.ops
hosts = ["web01", "web02", "web03"]

results = fleet.parallel(hosts, fn(h) => ssh.run(h, "uptime"))

for r in results
    print("{r.host}: ok={r.ok}")

summary = fleet.summary(results)
print("总计: {summary.total}, 成功: {summary.ok}")
```

### 声明式管理

```ops
// 确保文件、服务、包的状态
ensure.file("/etc/motd", "Welcome to {hostname}")
ensure.service("nginx", "running", true)
ensure.package("curl", "present")
```

## 功能特性

### 语言核心

- **缩进语法** — Python 风格，清晰可读
- **字符串插值** — `"Hello {name}"` 原生支持
- **三引号字符串** — `"""..."""` 嵌入 YAML/JSON/配置
- **链式方法** — `"hello".upper().trim()`
- **闭包 & Lambda** — `fn(x) => x * 2`
- **错误处理** — `try/catch`
- **REPL** — 交互式探索

### 标准库（10 模块）

| 模块 | 功能 |
|------|------|
| `file` | 读写、目录、路径操作 |
| `process` | Shell 执行、环境变量 |
| `ssh` | 远程执行、SCP 传输 |
| `fleet` | **批量并行引擎** |
| `json` | 解析、序列化、文件操作 |
| `yaml` | 解析、序列化、文件操作 |
| `toml` | 解析、文件操作 |
| `strings` | 分割、连接、查找、替换 |
| `math` | abs、min、max |
| `ensure` | **声明式资源管理** |
| `inventory` | 主机清单加载与分组 |

### 运维专用

- **SSH 远程执行** — 零配置密钥认证
- **Fleet 批量引擎** — 并发控制、结果汇总
- **Inventory 清单** — Ansible 式主机分组
- **Ensure 声明式** — 幂等资源状态保证
- **编译器** — 2.4MB 单二进制，交叉编译

## 文档

- [快速入门](docs/quickstart.md) — 5 分钟上手
- [语言参考](docs/language-reference.md) — 完整语法和 API
- [示例代码](examples/) — 实战脚本
- [设计文档](docs/design/) — 架构与决策

## 命令行

```bash
ops run <file>          # 运行脚本
ops build <file> [out]  # 编译为单二进制
ops repl                # 交互式 REPL
ops check <file>        # 语法检查
ops version             # 版本信息
```

## 实战演示

```ops
#!/usr/bin/env ops run
// deploy.ops — 完整运维工作流

// 1. 加载配置
config = yaml.load_file("app.yaml")
version = config["app"]["version"]

// 2. 加载清单
inv = inventory.load("hosts.ini")
webs = inventory.group(inv, "web_servers")

// 3. 批量部署
results = fleet.parallel(webs, fn(h) =>
    ssh.run(h, "deploy.sh " + str(version))
)

// 4. 验证
for r in results
    if r.ok
        print("✓ {r.host}")
    else
        print("✗ {r.host}")

// 5. 报告
summary = fleet.summary(results)
print("部署完成: {summary.ok}/{summary.total}")
```

## 项目结构

```
opslang/
├── cmd/ops/          # CLI 入口
├── pkg/
│   ├── lexer/        # 词法分析器
│   ├── parser/       # 语法分析器
│   ├── ast/          # 抽象语法树
│   ├── vm/           # 执行引擎
│   ├── compiler/     # 编译器
│   └── repl/         # REPL
├── stdlib/           # 标准库骨架
├── docs/             # 文档
├── examples/         # 示例
└── test/             # 测试
```

## 开发

```bash
make build        # 编译
make test         # 运行测试
make fmt          # 格式化代码
```

## 许可证

MIT License

---

<p align="center">
  <b>Shell 太弱，Python 太重，运维值得拥有一把趁手的武器。</b>
</p>
