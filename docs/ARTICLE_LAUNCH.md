# OpsLang：让运维自动化回归本质

> 我们受够了在 Shell 脚本里解析字符串，受够了 Ansible 的 YAML 地狱，受够了在每台服务器上安装 Agent。OpsLang 想解决这个问题。

## 问题

运维自动化这件事，过去十年没有本质进步。

我们依然在用 2005 年的方式写脚本：Bash 里塞满 `awk`、`sed`、`grep`，解析命令输出靠正则，错误处理靠 `set -e` 祈祷，返回值要么是字符串要么是退出码，没有结构化数据。

然后我们发明了 Ansible、Puppet、Chef，用 YAML 描述状态，用 Jinja2 模板渲染配置。好了一点，但代价是：
- 需要学习 DSL 和模板语法
- 调试困难，报错信息晦涩
- 目标机需要安装 Agent 或 Python 运行时
- 大规模分发文件时，控制端带宽成为瓶颈

我们想要的是：**写一个脚本，在任何架构的服务器上执行，拿到结构化结果，零依赖。**

OpsLang 就是这个。

## 解决方案

OpsLang 是一门面向运维的领域特定语言（DSL）。它提供：

### 1. 结构化返回，告别字符串解析

```ops
let cpu = sys.cpu.usage()
let mem = sys.memory.info()

if cpu.percent > 90 {
    alert("CPU 过高: " + str(cpu.percent) + "%")
}

report {
    host: sys.hostname(),
    cpu: cpu,
    mem: mem
}
```

`sys.cpu.usage()` 返回的是结构体，不是字符串。你可以直接访问 `.percent`、`.user`、`.system` 字段，不需要 `awk '{print $1}'`。

### 2. 双执行引擎，自适应场景

**Runner 模式**：线性脚本编译为 JSON 指令包，通用 Runner 执行。零编译延迟，适合简单任务。

**AOT 模式**：AST 编译为 Go 源码，`go build` 生成静态二进制。支持控制流、函数、第三方库，适合复杂逻辑。

自动选择，也可以手动指定。你不需要关心底层细节。

### 3. 零依赖远程执行

```bash
opsctl deploy --hosts host1,host2,host3 script.ops
```

一条命令，脚本在远程主机执行。目标机不需要安装任何东西——OpsLang 会自动检测架构（amd64/arm64），上传对应的 Runner 或编译好的二进制，执行完回收结果。

内容寻址缓存：相同二进制只上传一次，后续缓存命中只传 ~100 bytes 校验和。

### 4. 大规模文件分发与收集

```ops
task "deploy_app" on group("env=prod") {
    file.distribute(
        source: "/data/releases/app-v2.1.0.tar.gz",
        dest: "/opt/app/releases/",
        compress: true,
        checksum: true
    )
}

task "collect_logs" on group("role=web") {
    file.collect(
        source: "/var/log/nginx/access.log",
        dest: "/data/logs/{host}/access_{date}.log",
        compress: true
    )
}
```

真实 SSH/SFTP 实现，并行传输，断点续传，校验和去重。控制端带宽占用合理，1 万主机规模可承受。

### 5. 声明式幂等

```ops
ensure service "nginx" is running {
    service.start("nginx")
    service.enable("nginx")
}
```

`ensure` 块检查状态，只在需要时执行操作。支持 `--dry-run` 预览变更，不实际执行。

## 技术细节

OpsLang 纯 Go 实现，所有依赖支持 `CGO_ENABLED=0` 交叉编译。核心组件：

- **语言前端**：手写 Lexer + Recursive Descent Parser，20 个关键字，支持闭包、默认参数、C-style for
- **解释器**：AST 遍历执行，调用 SDK，支持 dry-run
- **AOT 编译器**：AST → Go 源码 → 静态二进制，编译缓存秒级命中
- **SSH 客户端**：TOFU 主机密钥验证，连接池，SFTP 断点续传
- **安全模块**：权限分级（read_only/admin/root）、审计日志、Ed25519 签名验证、资源限制

所有标准库函数返回结构体，不返回原始字符串。60+ 原子操作覆盖系统、文件、网络、进程、服务、包管理。

## 当前状态

OpsLang 是一个**完整、可工作**的实现，不是概念验证。

- 1456 个测试通过
- 25 个包，760 个测试函数
- 23 个示例脚本
- CI 全绿，交叉编译 linux/darwin amd64/arm64 通过

没有桩代码，没有空壳。双执行引擎均可用，远程执行链路已打通。

## 适用场景

**适合**：
- 需要在异构架构（amd64/arm64）上执行相同脚本
- 需要结构化返回而非字符串解析
- 需要零依赖远程执行（目标机无 Agent）
- 需要大规模文件分发/收集
- 需要声明式幂等操作

**不适合**：
- 复杂的配置管理（用 Ansible/Terraform）
- 需要 GUI 界面（OpsLang 是 CLI）
- 需要与现有 CMDB 深度集成（当前不支持）

## 下一步

OpsLang 还在快速迭代。计划中的功能：

- 权限自动执行（当前 `CheckPrivilege()` 存在但解释器不自动调用）
- 模块导入系统（OpsLang 脚本间互相导入）
- CI 竞态检测（`-race` 在 CI 全量测试时 TSan OOM，本地已支持）
- 1 万主机真实测试（当前用 mock）

## 试用

```bash
git clone https://github.com/j4ckzh0u/opslang
cd opslang
go build -o opsctl ./cmd/opsctl

# 本地执行
./opsctl run examples/cpu_check.ops

# 远程执行
./opsctl deploy --hosts user@host1,user@host2 examples/cpu_check.ops
```

文档在 `docs/` 目录，示例在 `examples/` 目录。

---

**OpsLang 不是要取代 Ansible 或 Terraform。** 它解决的是一个更基础的问题：让运维脚本回到脚本应有的样子——简单、结构化、可预测。

如果你受够了在 Bash 里解析字符串，受够了在每台服务器上安装 Agent，试试看。

GitHub: https://github.com/j4ckzh0u/opslang
