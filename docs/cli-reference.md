# opsctl CLI 命令参考

> opsctl 是 OpsLang 项目的命令行工具，负责脚本解释执行、AOT 编译、远程执行和交互式开发。

---

## 目录

1. [opsctl version](#opsctl-version)
2. [opsctl run](#opsctl-run)
3. [opsctl build](#opsctl-build)
4. [opsctl exec](#opsctl-exec)
5. [opsctl repl](#opsctl-repl)
6. [退出码](#退出码)
7. [环境变量](#环境变量)
8. [附录：配置文件格式](#附录配置文件格式)

---

## opsctl version

打印当前 opsctl 版本号。

### 用法

```bash
opsctl version
```

### 输出

```
opsctl v0.1.0
```

### 示例

```bash
$ opsctl version
opsctl v0.1.0
```

---

## opsctl run

本地解释执行 OpsLang 脚本。执行流程：读取文件 → 词法分析 → 语法分析 → 解释执行 → 输出结果。

### 用法

```bash
opsctl run [flags] <script.ops>
```

必须提供一个参数：脚本文件路径。

### 参数

| 参数 | 短选项 | 类型 | 默认值 | 说明 |
|------|--------|------|--------|------|
| `--json` | - | bool | `false` | 以 JSON 格式输出结果 |
| `-v` | `--verbose` | bool | `false` | 打印执行细节（如解析的语句数量） |

### 输出模式

**文本模式（默认）**

| 语句类型 | 输出目标 | 格式 |
|----------|----------|------|
| `print` / `log` | stdout | 直接输出文本 |
| `report` | stdout | JSON 格式化输出 |
| `alert` | stderr | `ALERT: <消息>` |
| `metric` | stdout | `METRIC: {"name": ..., "value": ...}` |

**JSON 模式（`--json`）**

输出单个 JSON 对象：

```json
{
  "output": [...],
  "return": <value>
}
```

### 示例

**示例 1：执行简单脚本**

```bash
$ cat check_system.ops
let cpu = sys.cpu.usage()
print("CPU 使用率: " + str(cpu.percent) + "%")

$ opsctl run check_system.ops
CPU 使用率: 12.5%
```

**示例 2：使用 verbose 模式**

```bash
$ opsctl run -v check_system.ops
Parsed 2 statements from file
CPU 使用率: 12.5%
```

**示例 3：JSON 输出**

```bash
$ opsctl run --json report.ops
{"output":[{"type":"report","data":{"host":"web1","cpu":{"percent":12.5}}}],"return":null}
```

### 注意事项

- 脚本语法错误会输出行号和列号，方便定位问题。
- `--json` 模式下，`alert` 仍然写入 stderr，不混入 JSON 输出。
- 解释器适用于本地调试和快速验证，生产环境建议使用 `opsctl build` 编译后执行。

---

## opsctl build

将 OpsLang 脚本通过 AOT 编译为静态二进制文件。编译结果缓存（基于源文件 hash + 目标架构），相同脚本二次编译直接使用缓存。

### 用法

```bash
opsctl build [flags]
```

### 参数

| 参数 | 短选项 | 类型 | 默认值 | 说明 |
|------|--------|------|--------|------|
| `--source` | `-s` | string | **必填** | OpsLang 源文件路径 |
| `--output` | `-o` | string | `./output` | 输出二进制文件路径 |
| `--target-arch` | - | string | 当前平台 | 目标架构，格式 `os/arch`，如 `linux/amd64`、`linux/arm64` |

### 输出

```
Compiling source -> output
Target: linux/amd64
Build successful!
```

### 示例

**示例 1：编译为当前架构**

```bash
$ opsctl build -s check_system.ops -o ./bin/check_system
Compiling check_system.ops -> ./bin/check_system
Target: linux/amd64
Build successful!
```

**示例 2：交叉编译到 ARM64**

```bash
$ opsctl build -s deploy.ops --target-arch linux/arm64 -o ./bin/deploy-arm64
Compiling deploy.ops -> ./bin/deploy-arm64
Target: linux/arm64
Build successful!
```

**示例 3：使用默认输出路径**

```bash
$ opsctl build -s my_script.ops
Compiling my_script.ops -> ./output
Target: linux/amd64
Build successful!
```

### 注意事项

- 编译需要本机安装 Go 工具链。
- 交叉编译使用 `CGO_ENABLED=0`，生成的二进制为纯静态链接，可直接在目标机器运行。
- 相同脚本 + 相同目标架构，第二次编译 < 5 秒（缓存命中）。
- 编译产物大小已通过 `-ldflags "-s -w"` 优化。

---

## opsctl exec

通过 SSH 在远程主机上执行 JSON 指令包。流程：SSH 连接 → 架构检测 → 上传/复用 Runner → 发送指令包 → 收集结果。

### 用法

```bash
opsctl exec [flags]
```

必须同时满足：
- 提供 `--instructions`
- 提供 `--hosts` 或 `--inventory`（至少一个）

### 参数

| 参数 | 短选项 | 类型 | 默认值 | 说明 |
|------|--------|------|--------|------|
| `--hosts` | - | []string | - | 目标主机列表，格式 `user@host` 或 `host` |
| `--user` | `-u` | string | `root` | 默认 SSH 用户名 |
| `--key` | `-i` | string | - | SSH 私钥文件路径 |
| `--password` | `-p` | string | - | SSH 密码 |
| `--inventory` | - | string | - | Inventory 文件路径（YAML 格式） |
| `--instructions` | - | string | **必填** | JSON 指令包文件路径 |
| `--parallel` | - | int | `10` | 最大并发主机数 |
| `--dry-run` | - | bool | `false` | 干运行模式，不实际执行变更操作 |
| `--runner-path` | - | string | - | 预构建 Runner 二进制路径（跳过自动构建） |
| `--output` | `-o` | string | stdout | 结果输出文件路径 |

### 输出

JSON 格式的执行汇总：

```json
{
  "task_id": "abc123",
  "targets": ["host1", "host2"],
  "results": {
    "host1": {"status": "success", "exit_code": 0, "data": {}},
    "host2": {"status": "failed", "exit_code": 1, "error": "timeout"}
  }
}
```

### 示例

**示例 1：指定主机列表执行**

```bash
$ opsctl exec --hosts root@192.168.1.10,root@192.168.1.11 \
    --instructions tasks.json --key ~/.ssh/id_rsa
```

**示例 2：使用 inventory 文件**

```bash
$ opsctl exec --inventory hosts.yaml --instructions tasks.json --parallel 20
```

**示例 3：干运行模式**

```bash
$ opsctl exec --hosts root@web1 --instructions deploy.json --dry-run
```

**示例 4：结果输出到文件**

```bash
$ opsctl exec --hosts root@web1 --instructions tasks.json -o result.json
```

### 注意事项

- Runner 二进制在同一架构下只上传一次，后续复用缓存。
- 按 Ctrl+C 或发送 SIGTERM 信号可优雅中断，已启动的远程任务会尝试清理。
- 任意一台主机执行失败，整体退出码为 1。
- `--hosts` 格式支持 `user@host` 或纯 `host`（纯 host 时使用 `--user` 指定的用户）。

---

## opsctl repl

启动交互式 REPL（Read-Eval-Print Loop）环境，适合调试和快速实验。

### 用法

```bash
opsctl repl
```

无参数。

### 启动界面

```
OpsLang REPL v0.1.0
Type 'help' for help, 'exit' or 'quit' to leave.
ops>
```

### 交互特性

| 特性 | 说明 |
|------|------|
| 提示符 | `ops> ` |
| 多行输入 | 行尾为 `{` 时自动续行，空行触发执行 |
| Ctrl+C | 取消当前输入行 |
| Ctrl+D | 退出 REPL |
| `exit` / `quit` | 退出 REPL |
| `help` | 显示帮助信息 |

### 输出格式

| 语句类型 | 显示方式 |
|----------|----------|
| `print` / `log` | ` <文本>` |
| `report` | ` <JSON>` |
| `metric` | ` METRIC: {...}` |
| `alert` | ` ALERT: <消息>`（stderr） |
| 表达式返回值 | ` => <value>` |

### 示例

**示例 1：基本表达式求值**

```
ops> let x = 42
ops> x
 => 42
```

**示例 2：调用标准库函数**

```
ops> sys.hostname()
 => {"hostname": "web1", "fqdn": "web1.example.com"}
```

**示例 3：多行输入**

```
ops> if true {
...   print("hello")
... }
 hello
```

### 注意事项

- REPL 使用解释执行模式，适合调试，不适合性能敏感场景。
- 每次输入完整语句或表达式后即时执行并返回结果。

---

## 退出码

| 退出码 | 含义 |
|--------|------|
| `0` | 执行成功 |
| `1` | 执行失败（脚本错误、远程主机部分失败等） |

`opsctl exec` 只要有一台主机执行失败，整体返回 1。

---

## 环境变量

| 变量 | 说明 |
|------|------|
| `OPS_SSH_KEY` | SSH 私钥路径（`--key` 未指定时使用） |
| `OPS_SSH_USER` | 默认 SSH 用户（覆盖 `--user` 默认值 `root`） |
| `OPS_RUNNER_PATH` | 预构建 Runner 路径（`--runner-path` 未指定时使用） |
| `OPS_CACHE_DIR` | 编译缓存目录（默认 `$HOME/.opsctl/cache`） |

---

## 附录：配置文件格式

### Inventory 文件（YAML）

```yaml
hosts:
  - name: web1
    host: 192.168.1.10
    user: root
    port: 22
  - name: web2
    host: 192.168.1.11
    user: deploy
    port: 22
  - name: db1
    host: 192.168.1.20
    user: root
    port: 22
```

字段说明：

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | 否 | 主机别名，用于结果展示 |
| `host` | 是 | 主机 IP 或域名 |
| `user` | 否 | SSH 用户，缺省使用 `--user` 或 `root` |
| `port` | 否 | SSH 端口，默认 22 |

### JSON 指令包

```json
{
  "version": "1.0",
  "task_id": "abc123",
  "dry_run": false,
  "instructions": [
    {
      "op": "sys.cpu.usage",
      "args": {},
      "assign": "cpu"
    },
    {
      "op": "sys.memory.info",
      "args": {},
      "assign": "mem"
    },
    {
      "op": "report",
      "args": {
        "cpu": "cpu",
        "memory": "mem"
      }
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `version` | string | 协议版本，当前为 `"1.0"` |
| `task_id` | string | 任务唯一标识 |
| `dry_run` | bool | 是否以干运行模式执行 |
| `instructions` | array | 指令序列，按顺序执行 |
| `instructions[].op` | string | 操作名称，对应 `ops-core-sdk` 函数 |
| `instructions[].args` | object | 操作参数 |
| `instructions[].assign` | string | 可选，将结果赋值给变量名 |
