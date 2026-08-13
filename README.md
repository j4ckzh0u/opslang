# OpsLang

**为运维而生的编程语言** — 类 Python 语法，Shell 般交互，单二进制部署。

[![Go Report Card](https://goreportcard.com/badge/github.com/opslang/opslang)](https://goreportcard.com/report/github.com/opslang/opslang)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## 为什么需要 OpsLang？

| 痛点 | Shell | Python | **OpsLang** |
|------|-------|--------|-------------|
| 静默失败 | ✗ | ○ | ★ 默认严格 |
| 类型系统 | ✗ 全字符串 | ○ 动态 | ★ 渐进类型 |
| 部署方式 | 复制文件 | 装运行时 | ★ **单二进制** |
| Shell 互操作 | ★ | ✗ 笨拙 | ★ **原生管道** |
| SSH 批量执行 | 手写循环 | 自己拼 | ★ **内置引擎** |
| 数据格式 | 手工解析 | 需装库 | ★ **原生 JSON/YAML** |
| 声明式管理 | ✗ | ✗ | ★ **ensure 语句** |

---

## 快速开始

### 安装

```bash
# 从源码编译
git clone https://github.com/opslang/opslang.git
cd opslang
make build

# 复制到 PATH
cp bin/ops /usr/local/bin/
```

### 第一个脚本

```ops
// hello.ops - 你好，运维
hosts = ["web01", "web02", "web03"]

for host in hosts {
    result = ssh.run(host, "uptime")
    print("{host}: {result.stdout}")
}
```

```bash
$ ops run hello.ops
web01: 10:30:01 up 30 days, 2:15, 1 user, load average: 0.01, 0.05, 0.01
web02: 10:30:01 up 15 days, 5:30, 0 users, load average: 0.12, 0.08, 0.03
web03: 10:30:02 up 45 days, 0:45, 2 users, load average: 0.50, 0.35, 0.20
```

### 批量检查与修复

```ops
// nginx_check.ops - 批量检查 Nginx 配置合规性

nginx_compliance = {
    worker_processes: "auto",
    keepalive_timeout: 65,
    server_tokens: "off",
}

report = fleet.check_and_fix(
    hosts = inventory.group("web_servers"),
    parallel = 20,
    target = "nginx",
    compliance = nginx_compliance,
)

report.summary()
report.notify(channels=["wechat:ops-group"])
```

### 声明式资源管理

```ops
// ensure_server.ops - 确保服务器状态

ensure.file("/etc/motd", content="Welcome to {hostname}")
ensure.service("nginx", state="running", enabled=true)
ensure.package("curl", state="present")
ensure.user("deploy", shell="/bin/bash", groups=["sudo", "docker"])
```

---

## 语言特性

### 语法设计

- **Python 风格**：缩进分块、简洁可读
- **Shell 互操作**：原生管道、重定向、后台执行
- **渐进类型**：脚本模式动态类型，生产模式静态类型
- **字符串插值**：`"Hello {name}"` 原生支持
- **模式匹配**：强大的模式匹配（规划中）

### 运维原生

- **SSH 内置**：无需额外库，直接远程执行
- **数据格式**：JSON、YAML、TOML、INI 原生支持
- **批量引擎**：`fleet` 模块并行执行、滚动更新
- **声明式**：`ensure` 语句管理资源状态
- **幂等执行**：重复运行不产生副作用

### 部署友好

- **单二进制**：编译后一个文件，零依赖
- **跨平台**：macOS/Linux/Windows，交叉编译
- **体积小**：目标 < 20MB
- **启动快**：< 5ms 解释执行，0ms 编译执行

---

## 项目结构

```
opslang/
├── cmd/                 # 命令行入口
│   └── ops/             # ops 主命令
├── pkg/                 # 核心库（可被外部引用）
│   ├── lexer/           # 词法分析器
│   ├── parser/          # 语法分析器
│   ├── ast/             # 抽象语法树
│   ├── compiler/        # 编译器
│   ├── vm/              # 字节码虚拟机
│   └── builtins/        # 内置函数
├── stdlib/              # 标准库
│   ├── os/              # 操作系统
│   ├── net/             # 网络 (SSH/HTTP)
│   ├── storage/         # 存储
│   ├── db/              # 数据库
│   ├── middleware/       # 中间件
│   ├── container/       # 容器 (Docker/K8s)
│   ├── fleet/           # 批量执行引擎
│   ├── ensure/          # 声明式资源管理
│   └── data/            # 数据格式
├── internal/            # 内部工具
├── docs/                # 文档
├── examples/            # 示例代码
└── test/                # 测试
```

---

## 开发指南

### 环境要求

- Go 1.21+
- Make

### 构建

```bash
make build        # 编译
make test         # 测试
make lint         # 代码检查
make fmt          # 格式化
```

### REPL 模式

```bash
$ ops repl
OpsLang 0.1.0
>>> print("Hello, Ops!")
Hello, Ops!
>>> 1 + 2 * 3
7
>>> hosts = ["web01", "web02"]
>>> for h in hosts { print(h) }
web01
web02
```

---

## 路线图

### Phase 1: MVP (3-6 个月)

- [x] 项目结构
- [ ] 词法分析器
- [ ] 语法分析器
- [ ] 字节码 VM
- [ ] 基础语法（变量/函数/循环/条件）
- [ ] 标准库 MVP（process/fs/ssh/json）
- [ ] `ops run` 解释执行
- [ ] 基础 REPL

### Phase 2: 编译器 (3-6 个月)

- [ ] OpsLang → Go 转译器
- [ ] `ops build` 编译到单二进制
- [ ] 跨平台编译
- [ ] 标准库完善（yaml/toml/http）

### Phase 3: 运维特性 (6-12 个月)

- [ ] `fleet` 批量并行执行引擎
- [ ] `ensure` 声明式资源管理
- [ ] Inventory 管理系统
- [ ] 模板引擎
- [ ] Dry-run 模式
- [ ] 审计日志

### Phase 4: 生态 (12-18 个月)

- [ ] 包管理器 `ops install`
- [ ] LSP 支持
- [ ] IDE 插件（VS Code）
- [ ] 完整文档和教程
- [ ] 社区建设

---

## 设计哲学

1. **简单胜于强大**：运维人员需要的是简单，不是图灵完备
2. **显式胜于隐式**：错误不应该被静默吞掉
3. **内置胜于第三方**：SSH/YAML/JSON 应该是语言的一部分
4. **单二进制胜于运行时**：复制一个文件就能跑
5. **渐进式学习**：10 分钟上手，不需要知道 class/decorator/generator

---

## 贡献

欢迎贡献！请先阅读 [贡献指南](CONTRIBUTING.md)。

## 许可证

MIT License

---

<p align="center">
  <strong>Shell 太弱，Python 太重，运维值得拥有一把趁手的武器。</strong>
</p>
