# opsctl CLI 命令参考

> opsctl 是 OpsLang 项目的命令行工具，负责脚本解释执行、AOT 编译、远程执行和交互式开发。

---

## 目录

1. [opsctl version](#opsctl-version)
2. [opsctl run](#opsctl-run)
3. [opsctl build](#opsctl-build)
4. [opsctl deploy](#opsctl-deploy)
5. [opsctl exec](#opsctl-exec)
6. [opsctl repl](#opsctl-repl)
7. [退出码](#退出码)
8. [环境变量](#环境变量)
9. [附录：配置文件格式](#附录配置文件格式)

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
| `--dry-run` | - | bool | `false` | 干运行模式：`ensure` 的 apply 步骤只打印将执行的动作，不实际执行 |

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
- 含 `import "go <包路径>"`（第三方 Go 库）的脚本会编译报错——该能力未实现。

---

## opsctl deploy

解析 OpsLang 脚本，按所选模式编译或生成指令包，部署到远程主机执行并汇总结果。

### 用法

```bash
opsctl deploy [flags] <script.ops>
```

必须提供 `--targets` 或 `--inventory`（至少一个）。

### 参数

| 参数 | 短选项 | 类型 | 默认值 | 说明 |
|------|--------|------|--------|------|
| `--targets` | - | string | - | 目标主机列表（逗号分隔 `user@host`） |
| `--inventory` | - | string | - | Inventory 文件路径（YAML） |
| `--parallel` | - | int | `10` | 最大并发主机数 |
| `--dry-run` | - | bool | `false` | 干运行模式，不实际执行变更操作 |
| `--mode` | - | string | `auto` | 执行模式：`auto` / `runner` / `aot` |
| `--user` | `-u` | string | `root` | 默认 SSH 用户名 |
| `--key` | `-i` | string | - | SSH 私钥文件路径 |
| `--password` | `-p` | string | - | SSH 密码 |
| `--output` | `-o` | string | stdout | 结果输出文件路径 |
| `--insecure-host-key` | - | bool | `false` | 跳过 SSH 主机密钥校验（默认启用 TOFU 校验；仅限实验室环境） |
| `--auto-approve` | - | bool | `false` | 预先批准被审批流拦截的部署（admin/root 脚本 + 生产目标）；非交互环境缺省拒绝 |

### 执行模式

| 模式 | 说明 |
|------|------|
| `runner` | 生成 JSON 指令包经 SSH 下发 ops-runner 执行。**只支持线性脚本**（调用、`let`、`report`、`alert`、`log`）；遇到 `if`/`for`/`while`/`fn`/`ensure`/`parallel` 或运行期计算表达式会**报错拒绝**（不会静默降级）。task 的 `on` 子句支持精确名 / `user@host` / glob 匹配路由主机 |
| `aot` | 编译成静态二进制，按目标机架构（`uname -m` 探测）交叉编译、真实上传并执行，失败如实报错。支持全语言（含 `ensure`/`parallel`）。**task 级 `on` 路由在 aot 模式不支持**（会报错）：自包含二进制无法知道自己落在哪台主机上，避免误路由 |
| `auto`（默认） | 先尝试 runner 指令包生成，生成失败自动转 aot |

### 输出

JSON 格式的执行汇总：

```json
{
  "task_id": "examples_check_cpu-1755330000000000000",
  "script": "check_cpu.ops",
  "started_at": "2026-08-16T10:00:00Z",
  "finished_at": "2026-08-16T10:00:12Z",
  "status": "success",
  "targets": ["host1", "host2"],
  "results": {
    "host1": {"status": "success", "exit_code": 0},
    "host2": {"status": "failed", "exit_code": 1, "error": "timeout"}
  }
}
```

### 示例

```bash
# 自动模式部署到两台主机
opsctl deploy deploy_app.ops --targets root@web1,root@web2

# 强制 runner 模式（线性脚本）
opsctl deploy collect.ops --targets web1,web2 --mode runner --key ~/.ssh/id_rsa

# 强制 AOT 模式（含 ensure/parallel 的完整语言）
opsctl deploy ensure_service.ops --inventory hosts.yaml --mode aot

# 干运行
opsctl deploy deploy_app.ops --targets web1 --dry-run
```

### 注意事项

- `status` 为 `failed` 或 `partial` 时命令返回非零退出码，且审计日志不会记录为成功。
- 脚本含 `import "go <包路径>"` 时直接报错拒绝。
- task 的 `on` 子句选不中任何 deploy 目标时报错。
- 多 task 剧本的最终 JSON 按**主机**合并所有步骤的 `data` 与 `errors`：每一步的 report 都可见，后续步骤的失败不会覆盖先前步骤的结果。

### 目标选择器与 inventory 组路由

`on` 子句支持四种选择器（glob 语法）：

| 选择器 | 匹配对象 |
|---|---|
| `"host2"` | inventory 主机名 / IP 精确匹配 |
| `"root@10.0.0.12"` | user@host 精确匹配 |
| `"web-*"` | 主机名/IP/user@host glob |
| `"root"` | **inventory 组名**（对标 Ansible `hosts:` 字段）——路由到 `group: root` 的所有主机 |

inventory 组路由示例（供给剧本只应在具备 root 能力的主机上执行变更）：

```yaml
hosts:
  - name: host1
    host: 10.0.0.11
    user: deploy
  - name: host2
    host: 10.0.0.12
    user: root
    group: root      # task "..." on "root" 选中这台
```

### Runner 二进制缓存（内容寻址）

deploy 每次会先确认远程 runner 为最新：本地 runner 缓存名由 **runner 全部源码（cmd/ops-runner、internal/runner、pkg/ops-core-sdk）的内容哈希**派生，注册表新增操作后哈希变化 → 自动重新编译上传，远程不会再出现 "unknown operation" 的陈旧 runner。远程主机侧按 SHA-256 内容寻址缓存：重复部署只做校验和比对，不重复传输 7MB 二进制。

### 审批流（生产环境保护）

**触发条件**（同时满足才拦截）：
- 脚本声明 `privilege: admin` 或 `privilege: root`；
- 目标主机来自 `--inventory` 且带生产标签（`tags: {env: prod}` 或 `env: production`）。

`--targets` 内联主机与无标签的 inventory 条目不算生产目标（生产身份只来自 inventory 元数据）；`read_only` 脚本不拦截。

**拦截后的行为**：
- 交互式终端：展示审批摘要（脚本、权限级别、变更类操作列表、生产目标数量与样例）后 `y/N` 确认，拒绝即中止且不联系任何主机，退出码非 0；
- 非交互（管道/CI）：**默认拒绝**并报错；需显式 `--auto-approve` 或环境变量 `OPSCTL_AUTO_APPROVE=1` 放行（flag 优先，显式 `--auto-approve=false` 可压掉环境变量）；
- 审批结果（批准/拒绝、来源、批准人 `$USER`、生产目标清单、变更操作）写入审计日志，与运行记录同文件可回溯。

`opsctl exec` 指令包同理：包内 `privilege` 为 `admin`/`root` 且目标为生产主机时触发审批（不带 `privilege` 字段的旧格式指令包不拦截）。

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
| `--insecure-host-key` | - | bool | `false` | 跳过 SSH 主机密钥校验（默认启用 TOFU 校验；仅限实验室环境） |
| `--output` | `-o` | string | stdout | 结果输出文件路径 |
| `--auto-approve` | - | bool | `false` | 预先批准被审批流拦截的执行（privileged 指令包 + 生产目标）；非交互环境缺省拒绝 |

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
- 部分失败（`partial`，部分主机未完成）退出码为 1；全部失败（`failed`）退出码为 2；全部成功退出码为 0。具体结果见 JSON 输出。
- `--hosts` 格式支持 `user@host` 或纯 `host`（纯 host 时使用 `--user` 指定的用户）。
- 指令包 `privilege` 为 `admin`/`root` 且 inventory 目标带生产标签时触发审批流（见 [opsctl deploy 的审批流说明](#审批流生产环境保护)）；审批被拒返回非零退出码且不联系任何主机。

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
Type OpsLang expressions. Ctrl+D to exit, Ctrl+C to cancel line.

SDK builtins loaded: sys.*, file.*, net.*, process.*, service.*, pkg.*, time.*, json.*, yaml.*
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

### opsctl 退出码

| 退出码 | 含义 |
|--------|------|
| `0` | 执行成功 |
| `1` | 部分失败：`opsctl deploy` 部分主机未完成、`opsctl exec` 部分主机失败 |
| `2` | 全部失败：`opsctl exec` 所有主机失败 |
| 其他非零 | 参数错误、脚本解析/运行错误、`opsctl deploy` 部署失败、审批被拒（未联系任何主机）等 |

### ops-runner 退出码

远程主机上的 ops-runner 进程按指令执行结果返回：

| 退出码 | 含义 |
|--------|------|
| `0` | 全部指令成功（status "ok"） |
| `1` | 部分指令失败（status "partial"） |
| `2` | 全部指令失败（status "failed"） |
| `3` | 协议/用法错误（输入损坏、不支持的版本） |

---

## 环境变量

| 变量 | 说明 |
|------|------|
| `OPSLANG_KNOWN_HOSTS` | SSH 主机密钥 TOFU 已知主机文件路径（默认 `~/.ssh/opslang_known_hosts`） |
| `OPSLANG_SSH_PASSWORD` | `file.distribute` / `file.collect` 传输使用的 SSH 密码 |
| `OPSLANG_SSH_KEY` | `file.distribute` / `file.collect` 传输使用的 SSH 私钥路径 |
| `OPSLANG_CACHE_DIR` | Runner 编译缓存目录（默认 `~/.cache/opslang/runners/`） |
| `OPSLANG_PROJECT_ROOT` | 覆盖项目根目录探测（开发调试用） |
| `OPSCTL_AUTO_APPROVE` | 设为 `1` 时放行审批流拦截的运行（CI 用）；`--auto-approve` flag 优先 |
| `OPSLANG_AUDIT_DIR` | 审计日志目录（默认 `/var/log/opsctl`，无权限时回退临时目录） |

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
