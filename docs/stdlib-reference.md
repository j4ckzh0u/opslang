# OpsLang 标准库参考手册

## 目录

- [1. 概述](#1-概述)
- [2. sys 包 - 系统信息](#2-sys-包---系统信息)
- [3. file 包 - 文件操作](#3-file-包---文件操作)
- [4. net 包 - 网络操作](#4-net-包---网络操作)
- [5. process 包 - 进程管理](#5-process-包---进程管理)
- [6. service 包 - 服务管理](#6-service-包---服务管理)
- [7. pkg 包 - 包管理](#7-pkg-包---包管理)
- [8. json 包 - JSON 编解码](#8-json-包---json-编解码)
- [9. yaml 包 - YAML 编解码](#9-yaml-包---yaml-编解码)
- [10. time 包 - 时间操作](#10-time-包---时间操作)

---

## 1. 概述

OpsLang 标准库（`ops-core-sdk`）提供一组面向运维场景的原子操作函数。所有函数均为纯 Go 实现，不依赖 shell，支持 `CGO_ENABLED=0` 交叉编译。

### 核心设计原则

- **结构化返回**：每个函数返回结构体（非原始字符串），所有字段均带 JSON 标签，可直接序列化为 JSON
- **纯 Go 实现**：不依赖 shell 命令执行，信息获取直接读取 `/proc`、`/sys` 或使用 Go 标准库
- **统一错误处理**：函数签名统一为 `(T, error)` 模式，错误包含明确的错误码和消息
- **异构架构**：所有包支持 `linux/amd64` 和 `linux/arm64` 交叉编译

### 规范名称（canonical）与旧别名

每个原子操作的规范名称、参数与可用范围定义在 `internal/opsspec`（单一事实来源），三个执行引擎（解释器 bridge、runner registry、AOT codegen）由一致性测试强制对齐。本文档中的函数名均为规范名称。

Runner 指令包中仍可使用以下旧别名，查询时会透明映射到规范名称，但新生成的指令包只输出规范名称：

| 旧别名 | 规范名称 |
|---|---|
| `sys.load.avg` | `sys.load` |
| `sys.host.info` | `sys.os` |
| `net.http.get` | `net.http_get` |
| `net.http.post` | `net.http_post` |
| `net.tcp.ping` | `net.tcp_check` |
| `net.dns.resolve` | `net.dns_lookup` |
| `process.find.by_name` | `process.find_by_name` |
| `process.find.by_port` | `process.find_by_port` |
| `file.info` | `file.stat` |
| `pkg.search` | `pkg.info` |

### OpsLang 调用语法

在 OpsLang 脚本中，通过点号（`.`）表示法调用标准库函数，映射规则如下：

| Go 函数签名 | OpsLang 调用语法 |
|---|---|
| `sys.GetCPUUsage()` | `sys.cpu.usage()` |
| `file.Read("/tmp/a.txt")` | `file.read("/tmp/a.txt")` |
| `net.HTTPGet("https://...")` | `net.http_get("https://...")` |
| `process.FindByName("nginx")` | `process.find_by_name("nginx")` |

---

## 2. sys 包 - 系统信息

> Go 包路径：`pkg/ops-core-sdk/sys`
> 依赖：`gopsutil/v4`（纯 Go）
> 所有函数不依赖 shell。

### 2.1 sys.cpu.usage()

获取 CPU 使用率。以 500ms 窗口两次采样计算，反映**当前负载**而非开机以来的平均值（调用约耗时 500ms）。

**参数**：无

**返回类型**：`CPUUsage`

| 字段 | 类型 | 说明 |
|---|---|---|
| `percent` | `float64` | CPU 总使用率（百分比） |
| `user` | `float64` | 用户态使用率 |
| `system` | `float64` | 内核态使用率 |
| `idle` | `float64` | 空闲率 |

**示例**：

```ops
let cpu = sys.cpu.usage()
print("CPU 使用率: " + str(cpu.percent) + "%")
```

---

### 2.2 sys.cpu.info()

获取 CPU 详细信息。

**参数**：无

**返回类型**：`[]CPUInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `vendor_id` | `string` | 厂商标识 |
| `model_name` | `string` | 型号名称 |
| `cores` | `int32` | 核心数 |
| `cache_size` | `int32` | 缓存大小 |
| `mhz` | `float64` | 频率（MHz） |

**示例**：

```ops
let cpus = sys.cpu.info()
for let i = 0; i < len(cpus); i = i + 1 {
    print("型号: " + cpus[i].model_name + ", 核心: " + str(cpus[i].cores))
}
```

---

### 2.3 sys.cpu.count()

获取 CPU 核心数。

**参数**：无

**返回类型**：`CPUCount`

| 字段 | 类型 | 说明 |
|---|---|---|
| `logical` | `int` | 逻辑核心数 |
| `physical` | `int` | 物理核心数 |

**示例**：

```ops
let counts = sys.cpu.count()
print("逻辑核心: " + str(counts.logical) + ", 物理核心: " + str(counts.physical))
```

---

### 2.4 sys.memory.info()

获取内存使用信息。

**参数**：无

**返回类型**：`MemoryInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `total` | `uint64` | 总内存（字节） |
| `available` | `uint64` | 可用内存（字节） |
| `used` | `uint64` | 已使用内存（字节） |
| `used_percent` | `float64` | 使用率（百分比） |

**示例**：

```ops
let mem = sys.memory.info()
print("内存使用率: " + str(mem.used_percent) + "%")
```

---

### 2.5 sys.disk.usage(path)

获取指定路径的磁盘使用情况。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件系统路径 |

**返回类型**：`DiskUsage`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 查询路径 |
| `total` | `uint64` | 总空间（字节） |
| `used` | `uint64` | 已使用（字节） |
| `free` | `uint64` | 剩余空间（字节） |
| `used_percent` | `float64` | 使用率（百分比） |

**示例**：

```ops
let disk = sys.disk.usage("/")
print("根分区使用率: " + str(disk.used_percent) + "%")
```

---

### 2.6 sys.disk.partitions()

获取磁盘分区列表。

**参数**：无

**返回类型**：`[]DiskPartition`

| 字段 | 类型 | 说明 |
|---|---|---|
| `device` | `string` | 设备名 |
| `mountpoint` | `string` | 挂载点 |
| `fstype` | `string` | 文件系统类型 |
| `opts` | `string` | 挂载选项 |

**示例**：

```ops
let parts = sys.disk.partitions()
for let i = 0; i < len(parts); i = i + 1 {
    print(parts[i].device + " -> " + parts[i].mountpoint + " (" + parts[i].fstype + ")")
}
```

---

### 2.7 sys.load()

获取系统负载均值。

**参数**：无

**返回类型**：`LoadAvg`

| 字段 | 类型 | 说明 |
|---|---|---|
| `load1` | `float64` | 1 分钟负载 |
| `load5` | `float64` | 5 分钟负载 |
| `load15` | `float64` | 15 分钟负载 |

**示例**：

```ops
let load = sys.load()
print("负载: " + str(load.load1) + " / " + str(load.load5) + " / " + str(load.load15))
```

---

### 2.8 sys.hostname()

获取主机名。

**参数**：无

**返回类型**：`HostnameInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `hostname` | `string` | 短主机名 |
| `fqdn` | `string` | 完全限定域名 |

**示例**：

```ops
let host = sys.hostname()
print("主机名: " + host.hostname)
```

---

### 2.9 sys.uptime()

获取系统运行时长。

**参数**：无

**返回类型**：`UptimeInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `uptime` | `uint64` | 运行时长（秒） |
| `boot_time` | `uint64` | 启动时间（Unix 时间戳） |

**示例**：

```ops
let up = sys.uptime()
print("已运行 " + str(up.uptime) + " 秒")
```

---

### 2.10 sys.os()

获取综合主机/操作系统信息。（Runner 指令包中的旧别名 `sys.host.info` 映射到本函数。）

**参数**：无

**返回类型**：`HostInfoResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `hostname` | `string` | 主机名 |
| `uptime` | `uint64` | 运行时长（秒） |
| `boot_time` | `uint64` | 启动时间（Unix 时间戳） |
| `os` | `string` | 操作系统 |
| `platform` | `string` | 发行版名称 |
| `platform_family` | `string` | 发行版族 |
| `platform_version` | `string` | 发行版版本 |
| `kernel_version` | `string` | 内核版本 |
| `kernel_arch` | `string` | 内核架构 |

**示例**：

```ops
let info = sys.os()
print("系统: " + info.platform + " " + info.platform_version)
print("内核: " + info.kernel_version + " (" + info.kernel_arch + ")")
```

---

### 2.11 sys.users()

获取当前登录用户列表。

**参数**：无

**返回类型**：`[]UserInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `user` | `string` | 用户名 |
| `terminal` | `string` | 终端 |
| `host` | `string` | 来源主机 |
| `start_time` | `uint64` | 登录时间（Unix 时间戳） |

**示例**：

```ops
let users = sys.users()
for let i = 0; i < len(users); i = i + 1 {
    print(users[i].user + " 从 " + users[i].host + " 登录")
}
```

---

### 2.12 sys.net.interfaces()

获取网络接口列表（与 `net.interfaces()` 等价，返回相同结构）。

**参数**：无

**返回类型**：`[]NetInterface`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 接口名称 |
| `hardware_addr` | `string` | MAC 地址 |
| `mtu` | `int` | MTU 值 |
| `up` | `bool` | 是否启用 |
| `addresses` | `[]string` | IP 地址列表 |

**示例**：

```ops
let ifaces = sys.net.interfaces()
for let i = 0; i < len(ifaces); i = i + 1 {
    print(ifaces[i].name + ": " + ifaces[i].hardware_addr)
}
```

---

## 3. file 包 - 文件操作

> Go 包路径：`pkg/ops-core-sdk/file`
> 依赖：Go 标准库（纯 Go）
> 不依赖 shell。

### 3.1 file.read(path)

读取文件内容。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |

**返回类型**：`FileContent`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `content` | `string` | 文件内容 |
| `size` | `int64` | 文件大小（字节） |

**示例**：

```ops
let cfg = file.read("/etc/hostname")
print("内容: " + cfg.content)
```

---

### 3.2 file.write(path, content)

写入文件（覆盖）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |
| `content` | `string` | 是 | 写入内容 |

**返回类型**：`WriteResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `size` | `int64` | 写入大小（字节） |

**示例**：

```ops
file.write("/tmp/hello.txt", "Hello OpsLang\n")
```

---

### 3.3 file.copy(src, dst)

复制文件。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `src` | `string` | 是 | 源文件路径 |
| `dst` | `string` | 是 | 目标文件路径 |

**返回类型**：`CopyResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `src` | `string` | 源路径 |
| `dst` | `string` | 目标路径 |
| `size` | `int64` | 复制大小（字节） |

**示例**：

```ops
file.copy("/etc/nginx/nginx.conf", "/tmp/nginx.conf.bak")
```

---

### 3.4 file.move(src, dst)

移动文件（重命名）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `src` | `string` | 是 | 源文件路径 |
| `dst` | `string` | 是 | 目标文件路径 |

**返回类型**：`MoveResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `src` | `string` | 源路径 |
| `dst` | `string` | 目标路径 |

**示例**：

```ops
file.move("/tmp/new.conf", "/etc/myapp/config.conf")
```

---

### 3.5 file.delete(path)

删除文件。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |

**返回类型**：`DeleteResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `existed` | `bool` | 删除前文件是否存在 |

**示例**：

```ops
let result = file.delete("/tmp/old.log")
if result.existed {
    print("已删除旧日志")
}
```

---

### 3.6 file.exists(path)

检查文件或目录是否存在。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |

**返回类型**：`ExistsResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `exists` | `bool` | 是否存在 |
| `is_dir` | `bool` | 是否为目录 |

**示例**：

```ops
if file.exists("/etc/nginx").exists {
    print("Nginx 已安装")
}
```

---

### 3.7 file.stat(path)

获取文件详细信息。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |

**返回类型**：`FileInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `name` | `string` | 文件名 |
| `mode` | `string` | 权限模式（如 `0644`） |
| `size` | `int64` | 文件大小（字节） |
| `mod_time` | `int64` | 修改时间（Unix 时间戳） |
| `is_dir` | `bool` | 是否为目录 |

**示例**：

```ops
let info = file.stat("/var/log/syslog")
print("大小: " + str(info.size) + " 字节, 权限: " + info.mode)
```

---

### 3.8 file.chmod(path, mode)

修改文件权限。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |
| `mode` | `string` | 是 | 权限模式（八进制，如 `"0755"`） |

**返回类型**：`ChmodResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `mode` | `string` | 设置后的权限模式 |

**示例**：

```ops
file.chmod("/opt/app/run.sh", "0755")
```

---

### 3.9 file.list(dir)

列出目录内容。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `dir` | `string` | 是 | 目录路径 |

**返回类型**：`ListResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 目录路径 |
| `files` | `[]FileInfo` | 文件信息列表（每项包含 `path`, `name`, `mode`, `size`, `mod_time`, `is_dir`） |

**示例**：

```ops
let listing = file.list("/etc/nginx/conf.d")
for let i = 0; i < len(listing.files); i = i + 1 {
    let f = listing.files[i]
    print(f.name + " (" + str(f.size) + " bytes)")
}
```

---

### 3.10 file.mkdir(path)

创建目录。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目录路径 |

**返回类型**：`MkdirResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 目录路径 |
| `created` | `bool` | 是否成功创建（已存在时为 `false`） |

**示例**：

```ops
let result = file.mkdir("/opt/app/data")
if result.created {
    print("目录创建成功")
}
```

---

### 3.11 file.checksum(path, algo)

计算文件校验和。

**参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| `path` | `string` | 是 | - | 文件路径 |
| `algo` | `string` | 否 | `"sha256"` | 算法：`"md5"`, `"sha1"`, `"sha256"` |

**返回类型**：`ChecksumResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `algorithm` | `string` | 使用的算法 |
| `checksum` | `string` | 校验和值 |
| `size` | `int64` | 文件大小（字节） |

**示例**：

```ops
let hash = file.checksum("/opt/app/release.tar.gz")
print("SHA256: " + hash.checksum)

let md5 = file.checksum("/opt/app/release.tar.gz", "md5")
print("MD5: " + md5.checksum)
```

---

### 3.12 file.append(path, content)

向文件末尾追加内容；文件不存在时以 0644 创建，绝不截断已有内容。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 文件路径 |
| `content` | `string` | 是 | 追加内容 |

**返回类型**：`AppendResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 文件路径 |
| `size` | `int64` | 追加后文件总大小（字节） |

**示例**：

```ops
file.append("/var/log/app.log", "service restarted\n")
```

---

### 3.13 file.template(path, vars)

读取模板文件并渲染 `{{key}}` 占位符（用 `vars` 字典中的同名键替换；未知占位符原样保留；原文件不被修改）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 模板文件路径 |
| `vars` | `dict` | 是 | 占位符键值字典 |

**返回类型**：`TemplateResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 模板文件路径 |
| `content` | `string` | 渲染后的内容 |
| `size` | `int64` | 渲染结果长度（字节） |

**示例**：

```ops
// config.yaml.tpl 内容：listen {{port}}\nenv {{env}}
let rendered = file.template("config.yaml.tpl", {"env": "prod", "port": 8080})
file.write("/etc/myapp/config.yaml", rendered.content)
```

---

### 3.14 file.distribute(source, targets, options)

将本地文件经真实 SSH/SFTP 分发到多台远程主机。**仅控制器侧可用**（远程 runner 与 AOT 代码生成器不暴露此函数）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `source` | `string` | 是 | 本地源文件路径 |
| `targets` | `list` | 是 | 目标列表，每项为 `{host, port, user, dest}` |
| `options` | `dict` | 是 | 选项字典（见下） |

`targets[].dest` 远程目标路径**按字面使用**：结尾带 `/` 表示目录（保留源文件名），否则视为完整目标文件路径，不做目录猜测。

**options 字段**：

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `checksum` | `bool` | `false` | 传输后计算远端 SHA-256 并与本地比对（真校验） |
| `mode` | `string` | - | 传输后在远端 chmod 的八进制权限（如 `"0644"`） |
| `parallel` | `int` | `5` | 最大并发传输数 |
| `retries` | `int` | `3` | 每台主机的总尝试次数 |

**SSH 凭据**：从环境变量 `OPSLANG_SSH_PASSWORD`（密码认证）或 `OPSLANG_SSH_KEY`（私钥路径）读取。

**返回类型**：`DistributeResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `total` | `int` | 目标主机总数 |
| `succeeded` | `int` | 成功数 |
| `failed` | `int` | 失败数 |
| `skipped` | `int` | 跳过数 |
| `results` | `[]HostDistributeResult` | 每台主机的结果（`host`, `status`, `changed`, `checksum`, `size` 等） |
| `duration_ms` | `int64` | 总耗时（毫秒） |

**示例**：

```ops
let result = file.distribute(
    "/opt/release/app.tar.gz",
    [
        {"host": "web1", "port": 22, "user": "root", "dest": "/opt/app/"},
        {"host": "web2", "port": 22, "user": "root", "dest": "/opt/app/"}
    ],
    {"checksum": true, "mode": "0644", "parallel": 10, "retries": 3}
)
if result.failed > 0 {
    alert("分发有失败主机: " + str(result.failed) + " 台")
}
```

---

### 3.15 file.collect(source, targets, options)

从多台远程主机收集文件到本地目录（真实 SSH/SFTP 下载）。**仅控制器侧可用**。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `source` | `string` | 是 | 远程源文件路径（逐台收集同一路径） |
| `targets` | `list` | 是 | 目标列表，每项为 `{host, port, user}` |
| `options` | `dict` | 是 | 选项字典（见下） |

**options 字段**：

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `dest_dir` | `string` | - | 本地目标目录 |
| `parallel` | `int` | `5` | 最大并发数 |
| `retries` | `int` | `3` | 每台主机的总尝试次数 |

**SSH 凭据**：同 `file.distribute`，读取 `OPSLANG_SSH_PASSWORD` / `OPSLANG_SSH_KEY`。

**返回类型**：`CollectResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `total` | `int` | 目标主机总数 |
| `succeeded` | `int` | 成功数 |
| `failed` | `int` | 失败数 |
| `results` | `[]HostCollectResult` | 每台主机的结果 |
| `dest_dir` | `string` | 本地目标目录 |
| `duration_ms` | `int64` | 总耗时（毫秒） |

**示例**：

```ops
let result = file.collect(
    "/var/log/app/error.log",
    [
        {"host": "web1", "port": 22, "user": "root"},
        {"host": "web2", "port": 22, "user": "root"}
    ],
    {"dest_dir": "/tmp/collected", "parallel": 10}
)
report {
    succeeded: result.succeeded,
    failed: result.failed
}
```

---

## 4. net 包 - 网络操作

> Go 包路径：`pkg/ops-core-sdk/net`（Go 包名：`opsnet`）
> 依赖：Go 标准库（纯 Go）
> 不依赖 shell。

### 4.1 net.http_get(url)

发起 HTTP GET 请求。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `url` | `string` | 是 | 请求 URL |

**返回类型**：`HTTPResponse`

| 字段 | 类型 | 说明 |
|---|---|---|
| `status_code` | `int` | HTTP 状态码 |
| `status` | `string` | 状态文本（如 `"200 OK"`） |
| `body` | `string` | 响应体 |
| `headers` | `map[string]string` | 响应头 |
| `content_length` | `int64` | 内容长度 |

**示例**：

```ops
let resp = net.http_get("https://api.example.com/health")
if resp.status_code == 200 {
    print("服务正常: " + resp.body)
}
```

---

### 4.2 net.http_post(url, body)

发起 HTTP POST 请求。默认 Content-Type 为 `application/json`。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `url` | `string` | 是 | 请求 URL |
| `body` | `string` | 是 | 请求体（通常为 JSON 字符串） |

**返回类型**：`HTTPResponse`

| 字段 | 类型 | 说明 |
|---|---|---|
| `status_code` | `int` | HTTP 状态码 |
| `status` | `string` | 状态文本 |
| `body` | `string` | 响应体 |
| `headers` | `map[string]string` | 响应头 |
| `content_length` | `int64` | 内容长度 |

**示例**：

```ops
let payload = json.encode({"event": "deploy", "version": "2.1.0"})
let resp = net.http_post("https://api.example.com/webhook", payload.json)
print("结果: " + str(resp.status_code))
```

---

### 4.3 net.tcp_check(host, port)

检查 TCP 端口连通性。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `host` | `string` | 是 | 目标主机 |
| `port` | `int` | 是 | 目标端口 |

**返回类型**：`TCPResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `host` | `string` | 目标主机 |
| `port` | `int` | 目标端口 |
| `connected` | `bool` | 是否连通 |
| `latency_ms` | `float64` | 连接延迟（毫秒） |

**示例**：

```ops
let check = net.tcp_check("127.0.0.1", 3306)
if check.connected {
    print("MySQL 端口可达, 延迟: " + str(check.latency_ms) + "ms")
} else {
    alert("MySQL 端口不可达")
}
```

---

### 4.4 net.dns_lookup(host)

执行 DNS 解析。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `host` | `string` | 是 | 域名 |

**返回类型**：`DNSResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `domain` | `string` | 查询的域名 |
| `addresses` | `[]string` | 解析到的 IP 地址列表 |
| `cname` | `string` | CNAME 记录 |

**示例**：

```ops
let dns = net.dns_lookup("example.com")
for let i = 0; i < len(dns.addresses); i = i + 1 {
    print("IP: " + dns.addresses[i])
}
```

---

### 4.5 net.interfaces()

获取网络接口信息（与 `sys.net.interfaces()` 等价）。

**参数**：无

**返回类型**：`[]InterfaceInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 接口名称 |
| `hardware_addr` | `string` | MAC 地址 |
| `mtu` | `int` | MTU 值 |
| `up` | `bool` | 是否启用 |
| `addresses` | `[]string` | IP 地址列表 |

**示例**：

```ops
let ifaces = net.interfaces()
for let i = 0; i < len(ifaces); i = i + 1 {
    if ifaces[i].up {
        print(ifaces[i].name + ": " + ifaces[i].hardware_addr)
    }
}
```

---

## 5. process 包 - 进程管理

> Go 包路径：`pkg/ops-core-sdk/process`
> 依赖：`gopsutil`（纯 Go）

### 5.1 process.list()

获取所有进程列表。

**参数**：无

**返回类型**：`[]ProcessInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `pid` | `int32` | 进程 ID |
| `name` | `string` | 进程名称 |
| `exe` | `string` | 可执行文件路径 |
| `cwd` | `string` | 当前工作目录 |
| `status` | `string` | 进程状态 |
| `username` | `string` | 运行用户 |
| `cpu_percent` | `float64` | CPU 使用率 |
| `memory_percent` | `float32` | 内存使用率 |
| `create_time` | `int64` | 创建时间（Unix 时间戳） |

**示例**：

```ops
let procs = process.list()
for let i = 0; i < len(procs); i = i + 1 {
    if procs[i].cpu_percent > 50.0 {
        alert("高 CPU 进程: " + procs[i].name + " (" + str(procs[i].cpu_percent) + "%)")
    }
}
```

---

### 5.2 process.find_by_name(name)

按名称查找进程（大小写不敏感的包含匹配）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 进程名称关键字 |

**返回类型**：`[]ProcessInfo`

字段同 `process.list()`。

**示例**：

```ops
let nginx_procs = process.find_by_name("nginx")
print("找到 " + str(len(nginx_procs)) + " 个 nginx 进程")
```

---

### 5.3 process.find_by_port(port)

按 TCP 监听端口查找进程。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `port` | `int` | 是 | TCP 端口号 |

**返回类型**：`[]ProcessInfo`

字段同 `process.list()`。

**示例**：

```ops
let procs = process.find_by_port(8080)
if len(procs) > 0 {
    print("端口 8080 被 " + procs[0].name + " (PID: " + str(procs[0].pid) + ") 占用")
}
```

---

### 5.4 process.exec(command, args...)

执行外部命令（不经过 shell）。参数为可变参数，逐个传入。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `command` | `string` | 是 | 可执行文件路径或命令名 |
| `args...` | `string` | 否 | 命令参数（逐个传入，不是列表） |

**返回类型**：`ExecResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `command` | `string` | 执行的命令 |
| `args` | `[]string` | 参数列表 |
| `stdout` | `string` | 标准输出 |
| `stderr` | `string` | 标准错误 |
| `exit_code` | `int` | 退出码（0 表示成功） |
| `pid` | `int` | 进程 ID |

**示例**：

```ops
let result = process.exec("ps", "-eo", "pid,comm")
if result.exit_code == 0 {
    print(result.stdout)
}
```

---

### 5.5 process.kill(pid, signal)

向进程发送信号（不经过 shell）。

**参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| `pid` | `int` | 是 | - | 进程 ID |
| `signal` | `string` | 否 | `"TERM"` | 信号名（如 `"TERM"`, `"KILL"`, `"HUP"`） |

**返回类型**：`KillResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `pid` | `int` | 进程 ID |
| `signal` | `string` | 实际发送的信号名 |
| `sent` | `bool` | 是否发送成功 |

**示例**：

```ops
let procs = process.find_by_name("rogue-agent")
for let i = 0; i < len(procs); i = i + 1 {
    let r = process.kill(procs[i].pid, "TERM")
    if !r.sent {
        alert("kill 失败: PID " + str(procs[i].pid))
    }
}
```

---

## 6. service 包 - 服务管理

> Go 包路径：`pkg/ops-core-sdk/service`
> 依赖：`systemctl`（通过 Go 的 `os/exec` 直接调用，不经过 shell）

### 6.1 service.status(name)

获取服务状态。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 服务名称 |

**返回类型**：`ServiceStatus`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 服务名称 |
| `active_state` | `string` | 活跃状态（如 `"active"`, `"inactive"`） |
| `sub_state` | `string` | 子状态（如 `"running"`, `"dead"`） |
| `load_state` | `string` | 加载状态（如 `"loaded"`） |
| `description` | `string` | 服务描述 |
| `main_pid` | `int` | 主进程 PID |
| `enabled` | `bool` | 是否开机启动 |
| `active` | `bool` | 是否处于活跃状态 |

**示例**：

```ops
let status = service.status("nginx")
if status.active {
    print("Nginx 正在运行 (PID: " + str(status.main_pid) + ")")
} else {
    print("Nginx 未运行: " + status.active_state)
}
```

---

### 6.2 service.start(name)

启动服务。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 服务名称 |

**返回类型**：`ServiceAction`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 服务名称 |
| `action` | `string` | 执行的操作（`"start"`） |
| `message` | `string` | 操作结果消息 |
| `success` | `bool` | 是否成功 |

**示例**：

```ops
let result = service.start("nginx")
if !result.success {
    alert("Nginx 启动失败: " + result.message)
}
```

---

### 6.3 service.stop(name)

停止服务。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 服务名称 |

**返回类型**：`ServiceAction`

字段同 `service.start()`，`action` 为 `"stop"`。

**示例**：

```ops
service.stop("nginx")
```

---

### 6.4 service.restart(name)

重启服务。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 服务名称 |

**返回类型**：`ServiceAction`

字段同 `service.start()`，`action` 为 `"restart"`。

**示例**：

```ops
let result = service.restart("nginx")
if result.success {
    print("Nginx 重启成功")
}
```

---

### 6.5 service.enable(name)

设置服务开机自启。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 服务名称 |

**返回类型**：`ServiceAction`

字段同 `service.start()`，`action` 为 `"enable"`。

**示例**：

```ops
service.enable("nginx")
```

---

### 6.6 service.disable(name)

取消服务开机自启。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 服务名称 |

**返回类型**：`ServiceAction`

字段同 `service.start()`，`action` 为 `"disable"`。

**示例**：

```ops
service.disable("apache2")
```

---

## 7. pkg 包 - 包管理

> Go 包路径：`pkg/ops-core-sdk/pkg`（Go 包名：`opspkg`）
> **仅支持 Linux**：通过检测 apt-get / yum / dnf 二进制工作；未检测到受支持的包管理器时（如 macOS）调用会报错

### 7.1 pkg.install(name)

安装软件包。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 软件包名称 |

**返回类型**：`PackageAction`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 软件包名称 |
| `action` | `string` | 执行的操作（`"install"`） |
| `manager` | `string` | 使用的包管理器（`"apt"`, `"yum"`, `"dnf"`） |
| `message` | `string` | 操作结果消息 |
| `success` | `bool` | 是否成功 |

**示例**：

```ops
let result = pkg.install("htop")
if result.success {
    print("已使用 " + result.manager + " 安装 htop")
}
```

---

### 7.2 pkg.remove(name)

卸载软件包。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 软件包名称 |

**返回类型**：`PackageAction`

字段同 `pkg.install()`，`action` 为 `"remove"`。

**示例**：

```ops
pkg.remove("htop")
```

---

### 7.3 pkg.info(name)

获取软件包详细信息。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 软件包名称 |

**返回类型**：`PackageInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 软件包名称 |
| `version` | `string` | 版本号 |
| `architecture` | `string` | 架构 |
| `description` | `string` | 描述 |
| `status` | `string` | 安装状态 |
| `manager` | `string` | 包管理器 |

**示例**：

```ops
let info = pkg.info("nginx")
print("Nginx 版本: " + info.version + " (" + info.status + ")")
```

---

### 7.4 pkg.list()

获取已安装软件包列表。

**参数**：无

**返回类型**：`[]PackageInfo`

字段同 `pkg.info()`。

**示例**：

```ops
let pkgs = pkg.list()
print("已安装 " + str(len(pkgs)) + " 个软件包")
```

---

## 8. json 包 - JSON 编解码

> Go 包路径：`pkg/ops-core-sdk/json`（Go 包名：`opsjson`）

### 8.1 json.encode(value)

将数据编码为 JSON 字符串（带缩进格式化）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `value` | `interface{}` | 是 | 待编码的数据 |

**返回类型**：`EncodeResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `json` | `string` | 格式化后的 JSON 字符串（带缩进） |
| `size` | `int` | JSON 字符串长度 |

**示例**：

```ops
let data = {"name": "nginx", "port": 80}
let encoded = json.encode(data)
print(encoded.json)
```

---

### 8.2 json.decode(input)

将 JSON 字符串解码为数据。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `input` | `string` | 是 | JSON 字符串 |

**返回类型**：`DecodeResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `data` | `interface{}` | 解码后的数据 |

**示例**：

```ops
let raw = '{"host": "web01", "cpu": 45.2}'
let decoded = json.decode(raw)
print("主机: " + decoded.data.host)
```

---

## 9. yaml 包 - YAML 编解码

> Go 包路径：`pkg/ops-core-sdk/yaml`（Go 包名：`opsyaml`）

### 9.1 yaml.encode(value)

将数据编码为 YAML 字符串。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `value` | `interface{}` | 是 | 待编码的数据 |

**返回类型**：`EncodeResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `yaml` | `string` | YAML 格式字符串 |
| `size` | `int` | YAML 字符串长度 |

**示例**：

```ops
let config = {"server": {"port": 8080, "host": "0.0.0.0"}}
let output = yaml.encode(config)
print(output.yaml)
```

---

### 9.2 yaml.decode(input)

将 YAML 字符串解码为数据。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `input` | `string` | 是 | YAML 字符串 |

**返回类型**：`DecodeResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `data` | `interface{}` | 解码后的数据 |

**示例**：

```ops
let raw = "name: nginx\nport: 80"
let decoded = yaml.decode(raw)
print(decoded.data.name)
```

---

## 10. time 包 - 时间操作

> Go 包路径：`pkg/ops-core-sdk/time`（Go 包名：`opstime`）

### 10.1 time.now()

获取当前时间。

**参数**：无

**返回类型**：`TimeInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `unix` | `int64` | Unix 时间戳（秒） |
| `unix_nano` | `int64` | Unix 时间戳（纳秒） |
| `iso8601` | `string` | ISO 8601 格式时间 |
| `utc` | `string` | UTC 时间字符串 |
| `timezone` | `string` | 当前时区 |

**示例**：

```ops
let now = time.now()
print("当前时间: " + now.iso8601)
print("Unix 时间戳: " + str(now.unix))
```

---

### 10.2 time.format(ts, layout)

格式化 Unix 时间戳。

**参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| `ts` | `int64` | 是 | - | Unix 时间戳（秒） |
| `layout` | `string` | 否 | `"2006-01-02 15:04:05"` | Go 风格时间格式 |

**返回类型**：`FormatResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `formatted` | `string` | 格式化后的时间字符串 |

**示例**：

```ops
let now = time.now()
let formatted = time.format(now.unix)
print("当前: " + formatted.formatted)

let custom = time.format(now.unix, "2006/01/02")
print("日期: " + custom.formatted)
```

---

### 10.3 time.parse(layout, value)

解析时间字符串为 Unix 时间戳。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `layout` | `string` | 是 | Go 风格时间格式 |
| `value` | `string` | 是 | 时间字符串 |

**返回类型**：`ParseResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `unix` | `int64` | Unix 时间戳（秒） |
| `iso8601` | `string` | ISO 8601 格式时间 |

**示例**：

```ops
let parsed = time.parse("2006-01-02", "2026-08-15")
print("Unix: " + str(parsed.unix))
```

---

### 10.4 time.since(unix)

计算从指定时间到当前的时间差。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `unix` | `int64` | 是 | 起始 Unix 时间戳（秒） |

**返回类型**：`DurationResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `seconds` | `float64` | 总秒数 |
| `minutes` | `float64` | 总分钟数 |
| `hours` | `float64` | 总小时数 |
| `human_readable` | `string` | 人类可读格式（如 `"2h 30m 15s"`） |

**示例**：

```ops
let start = time.now()
// ... 执行某些操作 ...
let elapsed = time.since(start.unix)
print("耗时: " + elapsed.human_readable)
```

---

### 10.5 time.sleep(ms)

暂停执行指定毫秒数。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `ms` | `int` | 是 | 暂停毫秒数 |

**返回类型**：`SleepResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `milliseconds` | `int` | 实际暂停的毫秒数 |

**示例**：

```ops
print("等待 3 秒...")
time.sleep(3000)
print("完成")
```

---

### 10.6 time.diff(t1, t2)

计算两个时间戳之间的差值。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `t1` | `int64` | 是 | 起始时间（Unix 时间戳，秒） |
| `t2` | `int64` | 是 | 结束时间（Unix 时间戳，秒） |

**返回类型**：`DiffResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `seconds` | `float64` | 总秒数 |
| `minutes` | `float64` | 总分钟数 |
| `hours` | `float64` | 总小时数 |
| `human_readable` | `string` | 人类可读格式（如 `"1h 45m 30s"`） |

**示例**：

```ops
let t1 = time.parse("2006-01-02", "2026-01-01").unix
let t2 = time.parse("2006-01-02", "2026-08-15").unix
let diff = time.diff(t1, t2)
print("相差: " + diff.human_readable)
```

---

## 附录：类型速查表

### 完整类型定义汇总

```
sys.CPUUsage        → { percent, user, system, idle }
sys.CPUInfo         → { vendor_id, model_name, cores, cache_size, mhz }
sys.CPUCount        → { logical, physical }
sys.MemoryInfo      → { total, available, used, used_percent }
sys.DiskUsage       → { path, total, used, free, used_percent }
sys.DiskPartition   → { device, mountpoint, fstype, opts }
sys.LoadAvg         → { load1, load5, load15 }
sys.HostnameInfo    → { hostname, fqdn }
sys.UptimeInfo      → { uptime, boot_time }
sys.HostInfoResult  → { hostname, uptime, boot_time, os, platform, platform_family,
                        platform_version, kernel_version, kernel_arch }   // sys.os()
sys.UserInfo        → { user, terminal, host, start_time }
sys.NetInterface    → { name, hardware_addr, mtu, up, addresses }        // sys.net.interfaces()

file.FileContent    → { path, content, size }
file.WriteResult    → { path, size }
file.AppendResult   → { path, size }
file.CopyResult     → { src, dst, size }
file.MoveResult     → { src, dst }
file.DeleteResult   → { path, existed }
file.ExistsResult   → { path, exists, is_dir }
file.FileInfo       → { path, name, mode, size, mod_time, is_dir }
file.ChmodResult    → { path, mode }
file.ListResult     → { path, files }
file.MkdirResult    → { path, created }
file.ChecksumResult → { path, algorithm, checksum, size }
file.TemplateResult → { path, content, size }
file.DistributeResult → { total, succeeded, failed, skipped, results, duration_ms }
file.CollectResult    → { total, succeeded, failed, results, dest_dir, duration_ms }

net.HTTPResponse    → { status_code, status, body, headers, content_length }
net.TCPResult       → { host, port, connected, latency_ms }
net.DNSResult       → { domain, addresses, cname }
net.InterfaceInfo   → { name, hardware_addr, mtu, up, addresses }

process.ProcessInfo → { pid, name, exe, cwd, status, username,
                        cpu_percent, memory_percent, create_time }
process.ExecResult  → { command, args, stdout, stderr, exit_code, pid }
process.KillResult  → { pid, signal, sent }

service.ServiceStatus → { name, active_state, sub_state, load_state, description,
                          main_pid, enabled, active }
service.ServiceAction → { name, action, message, success }

pkg.PackageAction   → { name, action, manager, message, success }
pkg.PackageInfo     → { name, version, architecture, description, status, manager }

json.EncodeResult   → { json, size }
json.DecodeResult   → { data }

yaml.EncodeResult   → { yaml, size }
yaml.DecodeResult   → { data }

time.TimeInfo       → { unix, unix_nano, iso8601, utc, timezone }
time.FormatResult   → { formatted }
time.ParseResult    → { unix, iso8601 }
time.DurationResult → { seconds, minutes, hours, human_readable }
time.SleepResult    → { milliseconds }
time.DiffResult     → { seconds, minutes, hours, human_readable }
```

## 附录：Roadmap（未实现）

以下传输优化能力**尚未实现**，当前 `file.distribute` / `file.collect` 为直接 SSH/SFTP 传输：

- 传输压缩
- 断点续传
- 内容哈希去重分发（相同内容只传一次）
