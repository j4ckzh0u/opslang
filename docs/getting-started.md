# 快速开始

本指南帮助你在 10 分钟内完成 OpsLang 的安装、编写第一个脚本、运行、编译和远程执行。

## 1. 环境要求

| 依赖 | 最低版本 | 说明 |
|------|---------|------|
| Go | 1.25.0 | 用于编译 opsctl 和 ops-runner |
| Git | 任意版本 | 用于克隆仓库 |
| SSH 客户端 | OpenSSH 7.0+ | 远程执行时需要（本地运行不需要） |

验证 Go 版本：

```bash
go version
# 输出示例：go version go1.25.0 darwin/arm64
```

如果尚未安装 Go，参考 [Go 官方安装指南](https://go.dev/doc/install)。

## 2. 安装步骤

### 从源码编译

```bash
git clone https://github.com/j4ckzh0u/opslang.git
cd opslang
go build -o opsctl ./cmd/opsctl
go build -o ops-runner ./cmd/ops-runner
```

编译完成后，当前目录会生成两个可执行文件：

- `opsctl` — CLI 主程序，用于运行、编译、远程执行脚本
- `ops-runner` — 通用 Runner，远程执行时自动上传到目标机器

### 验证安装

```bash
./opsctl version
# 输出：opsctl v0.1.0
```

### 可选：加入 PATH

```bash
sudo cp opsctl ops-runner /usr/local/bin/
opsctl version
```

## 3. 第一个脚本

创建文件 `helloworld.ops`：

```
// 第一个 OpsLang 脚本
let name = "OpsLang"
let version = "0.1.0"

print("Hello, " + name + " v" + version)

// 使用 report 输出结构化数据
report {
    lang: name,
    ver: version,
    status: "ready"
}
```

保存文件。就这么简单，不需要任何配置文件。

## 4. 运行脚本

使用 `opsctl run` 在本地解释执行脚本：

```bash
./opsctl run helloworld.ops
```

输出：

```
Hello, OpsLang v0.1.0
{
  "lang": "OpsLang",
  "ver": "0.1.0",
  "status": "ready"
}
```

### 常用标志

```bash
# 输出 JSON 格式结果
./opsctl run --json helloworld.ops

# 显示详细执行过程
./opsctl run --verbose helloworld.ops

# 干运行：ensure 的 apply 步骤只打印不执行
./opsctl run --dry-run helloworld.ops
```

### 更完整的示例

创建 `demo.ops`，展示变量、函数、条件、循环：

```
// 数据类型演示
let x = 42
let pi = 3.14
let greeting = "hello"
let flag = true
let items = [1, 2, 3]
let config = {"host": "localhost", "port": 8080}

print("x =", x)
print("type:", type(x))

// 函数定义
fn add(a, b) {
    return a + b
}
let sum = add(1, 2)
print("1 + 2 =", sum)

// 条件判断
if x > 10 {
    print("x 大于 10")
} else {
    print("x 不大于 10")
}

// 循环
for let i = 0; i < 5; i = i + 1 {
    print("i =", i)
}
```

运行：

```bash
./opsctl run demo.ops
```

输出：

```
x = 42
type: int
1 + 2 = 3
x 大于 10
i = 0
i = 1
i = 2
i = 3
i = 4
```

## 5. 编译脚本

使用 `opsctl build` 将脚本编译为独立二进制，无需目标机器安装任何运行时：

```bash
./opsctl build --source helloworld.ops --output ./helloworld --target-arch linux/amd64
```

参数说明：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--source` | 源脚本路径 | 必填 |
| `--output` | 输出二进制路径 | 必填 |
| `--target-arch` | 目标架构 | 当前架构 |

支持的目标架构：

```
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
```

交叉编译示例（在 macOS 上编译 Linux ARM64 二进制）：

```bash
./opsctl build --source demo.ops --output ./demo-arm64 --target-arch linux/arm64
```

编译出的二进制可以直接拷贝到目标机器运行，无需 Go 环境：

```bash
# 在目标 Linux 机器上
./helloworld
```

## 6. 远程执行

远程执行需要两样东西：目标主机和 JSON 指令包。

### 准备指令包

创建 `check_cpu.json`：

```json
{
  "version": "1.0",
  "task_id": "task001",
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
      "args": { "cpu": "cpu", "memory": "mem" }
    }
  ]
}
```

### 执行远程命令

```bash
./opsctl exec \
  --hosts root@192.168.1.10 \
  --instructions check_cpu.json \
  --user root \
  --key ~/.ssh/id_rsa
```

### 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `--hosts` | 目标主机，格式 `user@host` | `root@192.168.1.10` |
| `--instructions` | JSON 指令包路径 | `check_cpu.json` |
| `--user` | SSH 用户名 | `root` |
| `--key` | SSH 私钥路径 | `~/.ssh/id_rsa` |
| `--password` | SSH 密码 | - |
| `--parallel` | 并发主机数 | `10`（默认） |
| `--inventory` | 主机清单文件（YAML） | `hosts.yml` |
| `--dry-run` | 干运行模式，不实际执行变更操作 | - |
| `--runner-path` | 预构建 Runner 二进制路径（跳过自动构建） | `./ops-runner` |
| `--output` | 结果输出文件路径（默认 stdout） | `result.json` |
| `--insecure-host-key` | 跳过 SSH 主机密钥校验（仅限实验室环境） | - |

### 多主机执行

```bash
./opsctl exec \
  --hosts root@10.0.0.1,root@10.0.0.2,root@10.0.0.3 \
  --instructions check_cpu.json \
  --parallel 10
```

### 使用主机清单

创建 `hosts.yml`：

```yaml
hosts:
  - host: 10.0.0.1
    user: root
  - host: 10.0.0.2
    user: deploy
```

```bash
./opsctl exec \
  --inventory hosts.yml \
  --instructions check_cpu.json
```

### 工作原理

1. SSH 连接目标主机
2. 自动检测目标架构（`uname -m`）
3. 上传对应架构的 `ops-runner`（首次上传，后续缓存复用）
4. 发送 JSON 指令包到 `ops-runner` 的 stdin
5. 收集 stdout 的 JSON 结果并汇总

### 部署脚本（opsctl deploy）

更常用的远程执行方式是直接部署 `.ops` 脚本（无需手写指令包）：

```bash
./opsctl deploy examples/check_cpu.ops --targets root@192.168.1.10
```

`--mode` 选择执行模式（默认 `auto`）：

| 模式 | 说明 |
|------|------|
| `runner` | 生成 JSON 指令包下发 ops-runner；**只支持线性脚本**（调用、`let`、`report`、`alert`、`log`），控制流等会报错拒绝 |
| `aot` | 按目标机架构交叉编译静态二进制后上传执行，支持全语言（含 `ensure`/`parallel`） |
| `auto` | 先试 runner 生成，失败自动转 aot |

task 语句的 `on` 子句在 deploy 下生效（精确名 / glob 匹配目标）；`opsctl run` 遇到带 `on` 的 task 会报错提示改用 deploy。

## 7. REPL 交互环境

启动交互式环境：

```bash
./opsctl repl
```

进入 REPL 后看到提示符 `ops>`：

```
OpsLang REPL v0.1.0
Type OpsLang expressions. Ctrl+D to exit, Ctrl+C to cancel line.

SDK builtins loaded: sys.*, file.*, net.*, process.*, service.*, pkg.*, time.*, json.*, yaml.*

ops> let x = 42
ops> print(x)
42
ops> fn double(n) { return n * 2 }
ops> print(double(21))
42
```

### REPL 支持的功能

- 单行表达式求值
- 多行代码块（行尾为 `{` 时自动续行，空行触发执行）
- 变量定义和函数定义
- `help` — 查看帮助
- `exit` 或 `quit` — 退出
- `Ctrl+D` — 退出

多行示例：

```
ops> fn max(a, b) {
...     if a > b {
...         return a
...     } else {
...         return b
...     }
... }
ops> print(max(10, 20))
20
```

## 8. 常见问题

### Q: 编译报错 `go: command not found`

确认已安装 Go 1.25.0 或以上版本，并加入 PATH：

```bash
export PATH=$PATH:/usr/local/go/bin
go version
```

### Q: `opsctl run` 报语法错误

检查脚本文件是否存在，路径是否正确。OpsLang 错误提示会包含行号和列号：

```
错误: helloworld.ops:3:15: 未预期的 token
```

根据提示定位到对应行修复即可。

### Q: 远程执行时 SSH 连接失败

排查清单：

1. 确认目标主机 SSH 端口可达：`ssh root@10.0.0.1`
2. 确认密钥权限正确：`chmod 600 ~/.ssh/id_rsa`
3. 如果目标主机使用非默认端口，检查 SSH 配置
4. 若报主机密钥不匹配：OpsLang 默认做 TOFU 校验（记录在 `~/.ssh/opslang_known_hosts`），目标机重装后密钥变更会被拒绝；确认安全后删除该文件中对应条目，或（仅实验室）加 `--insecure-host-key`

### Q: 交叉编译后的二进制无法运行

确认 `--target-arch` 与目标机器架构一致：

```bash
# 在目标机器上查看架构
ssh root@target-host "uname -m"
# x86_64 → linux/amd64
# aarch64 → linux/arm64
```

### Q: ops-runner 上传失败

首次执行远程命令时会上传 `ops-runner` 到目标机器。如果上传失败：

1. 确认目标机器磁盘空间充足
2. 确认用户对目标目录有写权限（默认上传到 `/tmp/ops-<random>/`）
3. 手动上传测试：`scp ops-runner root@target:/tmp/`

### Q: 脚本中文注释乱码

OpsLang 源文件使用 UTF-8 编码。确保编辑器保存时使用 UTF-8：

```bash
file helloworld.ops
# 应显示：UTF-8 Unicode text
```

## 下一步

- 查看 `examples/` 目录了解更多脚本示例
- 阅读 `docs/` 目录了解语言语法详细规范
- 了解标准库函数：`sys`、`file`、`net`、`process`、`service` 等
