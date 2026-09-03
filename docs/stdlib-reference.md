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
- [18. user 包 - 用户管理](#18-user-包---用户管理)
- [19. group 包 - 组管理](#19-group-包---组管理)
- [20. 幂等收敛（ensure 家族）——对标 Ansible 核心模块](#20-幂等收敛ensure-家族对标-ansible-核心模块)

---

## 1. 概述

OpsLang 标准库（`ops-core-sdk`）提供一组面向运维场景的原子操作函数。核心信息与文件操作均由纯 Go 实现，不需要 Python 或 Shell 脚本运行时；包管理、服务管理和 `command` 等系统集成模块会按平台直接调用目标机已有组件。所有模块支持 `CGO_ENABLED=0` 交叉编译。

### 核心设计原则

- **结构化返回**：每个函数返回结构体（非原始字符串），所有字段均带 JSON 标签，可直接序列化为 JSON
- **优先纯 Go**：信息获取直接读取 `/proc`、`/sys` 或使用 Go 标准库；需要系统集成时通过参数化进程调用，不拼接不可信 shell 字符串
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

获取**承载数据的真实挂载点**列表，自动屏蔽机器差异。

**过滤规则**（黑名单优先）：

1. 排除内核伪文件系统：`proc`、`sysfs`、`devtmpfs`、`tmpfs`、`overlay`、`squashfs`（snap 循环挂载）、`efivarfs`、`cgroup*`、`autofs` 等
2. 保留三类真实数据挂载：本地块设备（`/dev/sd*`、`/dev/nvme*`、`/dev/vd*`、LVM、mdraid 等）、网络存储（nfs/cifs/ceph/glusterfs）、ZFS 数据集

每台服务器的挂载环境差异很大（容器 overlay、snap、/boot/efi……），本函数只返回运维需要关注的数据挂载；需要完整原始挂载表时使用 `sys.list_mounts()`。

**参数**：无

**返回类型**：`[]DiskPartition`

| 字段 | 类型 | 说明 |
|---|---|---|
| `device` | `string` | 设备名 |
| `mountpoint` | `string` | 挂载点 |
| `fstype` | `string` | 文件系统类型 |
| `opts` | `string` | 挂载选项 |
| `total_bytes` | `int` | 总容量（字节）；单个挂载点探测失败时为 0，不影响整体调用 |
| `used_bytes` | `int` | 已用容量 |
| `free_bytes` | `int` | 可用容量 |
| `used_percent` | `float` | 使用率百分比 |

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

### 2.10a sys.virt()

识别当前执行环境类型：容器、虚拟机还是物理机。安装 agent、选择备份策略、资源评估等决策的第一问。

**参数**：无

**返回类型**：`VirtInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `system` | `string` | 虚拟化层标识（`docker`/`kvm`/`vmware`/`xen`…）；空字符串表示未探测到（物理机） |
| `role` | `string` | `guest`（被虚拟化）/ `host`；探测不到时为空 |
| `is_container` | `bool` | 是否容器运行时（docker/podman/lxc/systemd-nspawn 等的归一判断，脚本直接分支用） |

探测在部分平台上不支持（如 macOS），此时**显式报错**而不是猜测默认值。

**示例**：

```ops
let v = sys.virt()
if v.is_container {
    log("跳过：容器环境不安装监控 agent")
} else {
    print("裸机/VM，system = " + str(v.system))
}
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

获取**业务网卡**列表（与 `net.interfaces()` 等价，返回相同结构）。

**过滤规则**：只返回满足全部条件的接口——状态为 up、至少有一个地址、且不是虚拟设备。自动排除的虚拟设备包括：

- 回环（`lo`/`lo0`）
- 容器/虚拟化：`docker0`、`veth*`、Docker 网桥 `br-<十六进制id>`、`cni*`、`cali*`、`flannel*`、`virbr*`、`ovs*`
- 隧道：`tun*`、`utun*`、`tap*`、`wg*`、`awdl*`、`llw*`
- 其他内核虚拟设备：`vxlan*`、`bridge*`、`ifb*`

注意区分：路由器的真实 LAN 网桥（`br0`、`br-lan`）**不会**被过滤。需要完整原始接口表时使用 `sys.net.all_interfaces()`。

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

### 2.13 sys.net.all_interfaces()

返回**全量**网络接口（含回环与虚拟设备），是 `sys.net.interfaces()` 语义化过滤的逃生门，字段结构完全相同。

**参数**：无　　**返回类型**：`[]NetInterface`

---

### 2.14 sys.net.primary_ip()

返回本机对外宣告的 IP：第一个持有**全局可路由 IPv4** 的业务网卡。这是"应用该绑定哪个地址""防火墙该放行哪个 IP"的标准答案。

选择规则：

1. 遍历业务网卡（同 `sys.net.interfaces()` 过滤规则）
2. 取第一个非回环、非链路本地（169.254.\*）的 IPv4
3. 找不到全局地址时回退到第一个非回环 IPv4（容器内返回内网地址而不是报错）
4. 完全没有可用 IPv4 时报错

**参数**：无

**返回类型**：`PrimaryIPResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `interface` | `string` | 网卡名 |
| `address` | `string` | 选中的 IPv4 地址 |

**示例**：

```ops
let ip = sys.net.primary_ip()
print("对外IP: " + ip.address + " (网卡 " + ip.interface + ")")
```

---

### 2.15 sys.net.rate(seconds)

在指定窗口内采集发送/接收字节数并计算平均比特率。函数会进行两次内核计数器采样，返回总量及逐接口明细；计数器重置的接口会被忽略，避免产生负速率。

**参数**：`seconds`（`int`，1–3600）

**返回类型**：`NetRate`，包含 `bytes_sent`、`bytes_recv`、`bits_per_second`、`window_seconds` 和 `interfaces`。

这是窗口平均值，不是持久化的滚动指标。例如 `sys.net.rate(3)` 表示 3 秒实测平均，五分钟测量应传入 `300`。

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
| `targets` | `list` | 是 | 目标列表，每项为 `{host, port, user, dest, relay_group?, tags?}` |
| `options` | `dict` | 是 | 选项字典（见下） |

`targets[].dest` 远程目标路径**按字面使用**：结尾带 `/` 表示目录（保留源文件名），否则视为完整目标文件路径，不做目录猜测。

**options 字段**：

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `checksum` | `bool` | `false` | 传输后计算远端 SHA-256 并与本地比对（真校验） |
| `mode` | `string` | - | 传输后在远端 chmod 的八进制权限（如 `"0644"`） |
| `parallel` | `int` | `5` | 最大并发传输数 |
| `retries` | `int` | `3` | 每台主机的总尝试次数 |
| `resume` | `bool` | `false` | 启用内容哈希跳过、部分文件续传与原子替换 |
| `part_retention` | `int` | `86400000` | 部分文件元数据保留时间，单位毫秒 |
| `relay` | `bool` | `false` | 启用确定性拓扑分组和 HTTPS 中继扇出 |
| `relay_group` | `string` | - | 所有目标的中继组后备值 |
| `relay_threshold` | `int` | `20` | 组内启用中继的最少目标数 |
| `relay_max_targets` | `int` | `100` | 每个中继最多服务的下游目标数 |

目标中继组按 `targets[].relay_group`、`targets[].tags.relay_group`、全局 `relay_group`、IPv4 `/24` 或 IPv6 `/64` 的优先级确定。无法解析为 IP 且没有显式组的目标直接使用 SFTP。`targets[].tags.relay = "true"` 可提高该目标的候选优先级。

恢复文件使用最终路径旁的 `.opslang.part` 和 `.opslang.part.json`。只有源大小、SHA-256、确认偏移和确认块一致时才从 Range 偏移继续；完整校验后原子替换最终文件。

**SSH 凭据**：从环境变量 `OPSLANG_SSH_PASSWORD`（密码认证）或 `OPSLANG_SSH_KEY`（私钥路径）读取。

**返回类型**：`DistributeResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `total` | `int` | 目标主机总数 |
| `succeeded` | `int` | 成功数 |
| `failed` | `int` | 失败数 |
| `skipped` | `int` | 跳过数 |
| `results` | `[]HostDistributeResult` | 每台主机的结果，包含 `transfer_source`、`resumed_bytes`、`transferred_bytes` 和 `warnings` |
| `duration_ms` | `int64` | 总耗时（毫秒） |

**示例**：

```ops
let result = file.distribute(
    "/opt/release/app.tar.gz",
    [
        {"host": "web1", "port": 22, "user": "root", "dest": "/opt/app/"},
        {"host": "web2", "port": 22, "user": "root", "dest": "/opt/app/"}
    ],
    {
        "checksum": true,
        "mode": "0644",
        "parallel": 10,
        "retries": 3,
        "resume": true,
        "relay": true,
        "relay_group": "prod-web"
    }
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
| `resume` | `bool` | `false` | 启用内容哈希跳过、部分文件续传与原子替换 |
| `part_retention` | `int` | `86400000` | 部分文件元数据保留时间，单位毫秒 |

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
    {"dest_dir": "/tmp/collected", "parallel": 10, "resume": true}
)
report {
    succeeded: result.succeeded,
    failed: result.failed
}
```

---

### 3.16 file.ensure(path, state, mode)

幂等收敛文件/目录状态，对标 Ansible `file` 模块。这是"声明期望状态"的写法：
无论当前是什么状态，执行后一定收敛到期望状态；已满足期望时**零动作**（`changed=false`，`actions` 为空）。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `path` | `string` | 是 | 目标路径 |
| `state` | `string` | 是 | 期望状态：`"directory"` \| `"file"` \| `"touch"` \| `"absent"` |
| `mode` | `string` | 否 | 期望权限位，八进制字符串如 `"0755"`；空串表示不管理权限 |

**状态语义**：

| state | 行为 |
|---|---|
| `directory` | 不存在则递归创建（mode 缺省 0755）；存在但类型是文件则**报错**（绝不静默覆盖）；权限漂移则 chmod |
| `file` | 必须已存在（**不会创建**，与 Ansible 一致，需要创建请用 `touch`）；权限漂移则 chmod |
| `touch` | 不存在则创建空文件；已存在则**保持不动**（与 Ansible 不同：Ansible 每次刷新 mtime，我们选择让收敛运行严格幂等） |
| `absent` | 存在则删除（目录递归删除）；不存在则零动作 |

**返回类型**：`EnsureResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `path` | `string` | 目标路径 |
| `state` | `string` | 期望状态 |
| `type` | `string` | 磁盘上的实际类型：`"directory"` / `"file"` / `"other"` / `""`(不存在) |
| `mode` | `string` | 收敛后的权限位（未管理权限时为空） |
| `changed` | `bool` | 本次是否产生了真实变更 |
| `actions` | `[]string` | 实际执行的动作列表：`mkdir` / `create` / `chmod` / `remove`；幂等运行为空列表 |
| `message` | `string` | 人类可读结果 |
| `error` | `string` | 错误详情（仅失败时） |

**平台与权限**：任何平台可运行；变更系统路径需要相应写权限。`Mutating`（需要 `privilege: admin`）。

**示例**（摘自 `examples/ensure_idempotency_proof.ops`，真实运行输出）：

```ops
let r1 = file.ensure("/srv/app/conf.d", "directory", "0750")
// r1 = { "path": "/srv/app/conf.d", "state": "directory", "type": "directory",
//        "mode": "0750", "changed": true, "actions": ["mkdir"], ... }

let r2 = file.ensure("/srv/app/conf.d", "directory", "0750")
// r2 = { ..., "changed": false, "actions": [], "message": "directory already up to date" }
```

---

## 4. net 包 - 网络操作

> Go 包路径：`pkg/ops-core-sdk/net`（Go 包名：`opsnet`）
> 依赖：Go 标准库（纯 Go）
> 不依赖 shell。

### 4.0 net.connections(kind)

枚举内核 socket 表并归属到进程 —— 纯 Go 实现（gopsutil 直读 `/proc/net/tcp(6)` + `/proc/<pid>/fd`），等价于 `ss -tuanp` / `netstat -tuanp` 但**不调用它们**。端口监听↔进程、TCP 连接↔进程两类巡检都基于此操作。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `kind` | `string` | 否 | 过滤器：`"inet"`（默认，IPv4+IPv6 的 TCP+UDP）、`"tcp"`、`"tcp4"`、`"tcp6"`、`"udp"`、`"inet6"` 等 |

**返回类型**：`[]ConnectionInfo`

| 字段 | 类型 | 说明 |
|---|---|---|
| `fd` | `int` | 套接字描述符 |
| `protocol` | `string` | `"tcp"` 或 `"udp"` |
| `local_addr` | `string` | 本端 `ip:port`，IPv6 加方括号（如 `[::1]:22`） |
| `remote_addr` | `string` | 对端地址（监听套接字为空） |
| `status` | `string` | `LISTEN` / `ESTABLISHED` / `TIME_WAIT` ... |
| `pid` | `int` | 归属进程 PID |
| `process_name` | `string` | 归属进程名（归属失败为空） |
| `uid` | `int` | 套接字属主 UID |

**权限语义（重要）**：socket 表人人可读，但归属到进程需要 root —— 非 root 时其他用户的套接字 `pid=0`、`process_name=""`，**如实返回而不是隐藏**，调用方可统计 unattributed 数量。

**示例**：

```ops
let conns = net.connections("inet")
let listeners = []
for c in conns {
    if c.status == "LISTEN" && c.pid > 0 {
        listeners = listeners + { "listen": c.local_addr, "pid": c.pid, "process": c.process_name }
    }
}
```

---

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

获取所有用户态进程列表。自动过滤内核线程（Linux 上的 `[kthreadd]`、`[kworker/*]` 等方括号命名进程）——它们每台机器数量不同且不承载运维信号。

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

### 5.1.1 process.java_apps()

扫描 Linux `/proc` 中正在运行的 Java 进程，解析命令行中的 `.jar`/classpath 条目并提取库名称、版本和路径。进程所属 cgroup 若包含 Docker、containerd 或 CRI-O ID，则同时返回 `container_runtime` 和 `container_id`。

**参数**：无　　**返回类型**：`[]JavaApp`

`JavaApp` 字段包括 `pid`、`user`、`command`、`executable`、`libraries`；`libraries` 每项包含 `name`、`version`、`path`。宿主机只有在与容器共享 PID namespace 时才能看到容器内 Java 进程；否则应在容器内执行脚本。

---

### 5.1.2 causal.trace_pid(pid)

追踪指定 Linux PID 的父进程链，返回从目标进程到根进程的节点和 `parent` 边。节点包含 PID、PPID、名称、可执行文件、命令行、UID，以及可识别时的 Docker/containerd/CRI-O cgroup 信息。单个进程因权限或退出而不可读时，已采集链会保留在结果中，并在 `warnings` 中说明。

**参数**：`pid`（`int`，正整数）　　**返回类型**：`CausalTrace`

---

### 5.1.3 causal.find(name)

按大小写不敏感的名称包含匹配查找进程，并为每个匹配项返回一条 `CausalTrace`。结果按 PID 排序；Linux 之外返回空列表，便于跨平台脚本保留同一数据契约。

**参数**：`name`（`string`）　　**返回类型**：`[]CausalTrace`

---

### 5.1.4 causal.trace_port(port)

查找占用指定 TCP/UDP 端口的进程，并为每个 socket 返回进程 ancestry。结果包含本地/远端地址、协议、PID 和嵌套的 `CausalTrace`；无权限归属的内核 socket 会被跳过，不会伪造进程信息。

**参数**：`port`（`int`，1–65535）　　**返回类型**：`[]PortConnection`

---

### 5.1.5 causal.trace_file(path)

扫描 `/proc/<pid>/fd`，查找解析到指定路径的打开文件描述符，并返回持有者进程的 ancestry。单个进程的 fd 目录不可读时按 best-effort 处理。

**参数**：`path`（`string`）　　**返回类型**：`[]FileTrace`

---

### 5.1.6 causal.trace_container(id)

按 Docker/containerd/CRI-O cgroup ID 查找进程并返回各自 ancestry。此 API 只依赖内核 cgroup 信息，不要求目标机安装 Docker CLI 或访问容器运行时 socket。

**参数**：`id`（`string`，6–64 位十六进制）　　**返回类型**：`[]CausalTrace`

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

### 6.7 service.ensure(name, state)

幂等收敛服务运行状态，对标 Ansible `service`/`systemd` 模块的 `state=` 参数。先读取真实状态再决定动作：已满足期望则**零动作**。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | systemd 单元名 |
| `state` | `string` | 是 | 期望状态：`"started"` \| `"stopped"` \| `"restarted"` \| `"reloaded"` |

**状态语义**：

| state | 当前 active | 动作 | changed |
|---|---|---|---|
| `started` | true | 无 | false |
| `started` | false | `systemctl start` | true |
| `stopped` | false | 无 | false |
| `stopped` | true | `systemctl stop` | true |
| `restarted` | 任意 | `systemctl restart` | 恒为 true（重启永不幂等） |
| `reloaded` | 任意 | `systemctl reload`；单元不支持 reload 时**自动回退 restart**（与 Ansible 行为一致） | 恒为 true |

**返回类型**：`EnsureResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 单元名 |
| `state` | `string` | 期望状态 |
| `active` | `bool` | 收敛后的运行状态 |
| `enabled` | `bool` | 当前开机自启状态（ensure 不改动它） |
| `changed` | `bool` | 是否产生真实变更 |
| `actions` | `[]string` | 实际执行的 systemctl 动词（如 `["start"]`、reload 回退时 `["reload","restart"]`）；幂等运行为空 |
| `message` | `string` | 结果说明 |
| `error` | `string` | 错误详情（仅失败时） |

**平台与权限**：Linux + systemd + root。`Mutating`。

**示例**（真机 Ubuntu 22.04 上的真实输出，先手工 `systemctl stop cron`）：

```ops
let run = service.ensure("cron", "started")
```

```json
{ "name": "cron", "state": "started", "active": true, "enabled": true,
  "changed": true, "actions": ["start"],
  "message": "service \"cron\" converged to started" }
```

再次执行（cron 已 active）：

```json
{ "name": "cron", "state": "started", "active": true, "enabled": true,
  "changed": false, "actions": [],
  "message": "service \"cron\" already active" }
```

---

### 6.8 service.ensure_enabled(name, enabled)

幂等收敛开机自启，对标 Ansible 的 `enabled=yes/no`。已满足期望则零动作。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | systemd 单元名 |
| `enabled` | `bool` | 是 | `true`=开机自启，`false`=禁用自启 |

**返回类型**：同 `service.ensure`（`state` 字段报告 `"enabled"`/`"disabled"`）。

**语义**：

| enabled | 当前 is-enabled | 动作 | changed |
|---|---|---|---|
| true | enabled | 无 | false |
| true | 其他 | `systemctl enable` | true |
| false | disabled | 无 | false |
| false | 其他 | `systemctl disable` | true |

**平台与权限**：Linux + systemd + root。`Mutating`。

**示例**：

```ops
// 典型组合：确保 nginx 运行且开机自启（重复执行零变更）
let run  = service.ensure("nginx", "started")
let boot = service.ensure_enabled("nginx", true)
if run.changed || boot.changed {
    log("nginx 本轮发生变更: start=" + str(run.actions) + " enable=" + str(boot.actions))
}
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

### 7.5 pkg.ensure(name)

幂等安装软件包，对标 Ansible `package` 模块（`state=present`）。已安装则直接返回 `changed=false`，**不再触碰包管理器**；未安装则自动探测系统包管理器（apt → dnf → yum）执行安装。

**参数**：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `name` | `string` | 是 | 包名 |

**返回类型**：`PackageAction`

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 包名 |
| `action` | `string` | `"ensure"` |
| `manager` | `string` | 实际使用的包管理器（apt/dnf/yum） |
| `success` | `bool` | 操作是否成功 |
| `changed` | `bool` | `false` = 已安装（零动作）；`true` = 本次执行了安装 |
| `message` | `string` | 结果说明 |
| `error` | `string` | 错误详情（仅失败时） |

**平台与权限**：Linux（macOS/无包管理器环境显式报错）。安装需要 root。`Mutating`。

**示例**：

```ops
let r = pkg.ensure("htop")
if r.changed {
    log("htop 新安装完成")
} else {
    log("htop 已存在，零动作")   // 第二次运行必然走这里
}
```

---

### 7.6 pkg.owner(path)

反查"这个文件由哪个软件包安装"（`dpkg -S` / `rpm -qf` 的结构化版本）。典型场景：进程巡检发现一个吃 CPU 的进程，`pkg.owner(process.exe)` 直接回答"这个二进制是谁装的"。

**参数**：`path`（string，必填）—— 文件绝对路径。

**返回类型**：`OwnerResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `file` | `string` | 查询路径（原样回显） |
| `package` | `string` | 归属包名；无归属时为空 |
| `found` | `bool` | `true`=归属于某包；`false`=不属于任何包 |
| `manager` | `string` | 使用的包管理器（apt/yum/dnf） |
| `output` | `string` | 原始查询输出（命中变体路径可见） |

**语义要点**：

- **无归属不是错误**：自编译二进制、`/usr/local/bin` 下手工放置的工具、容器内解包的文件 —— `found=false` 且不返回 error。这恰是安全审计的关键信号（不在包数据库管控内的可执行文件）
- **usrmerge 兼容**：dpkg 数据库按登记路径精确匹配，Ubuntu 22.04 等 usrmerge 系统上 `/usr/bin/ls` 登记为 `/bin/ls` —— 查询 miss 时自动回退 `/usr/bin→/bin`、`/usr/sbin→/sbin`、`/usr/lib→/lib` 变体重查
- **diversion 处理**：`dpkg -S` 的 diversion 提示行会被解析器跳过

**平台**：Linux（apt/yum/dnf）。macOS 无包管理器时报错。只读操作。

**真实输出**（Ubuntu 22.04 实测）：

```json
{ "file": "/usr/bin/dockerd", "package": "docker-ce", "found": true, "manager": "apt" }
{ "file": "/opt/kube/bin/kube-apiserver", "package": "", "found": false, "manager": "apt" }
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

## 11. archive 包 - 归档操作

### 11.1 archive.create(dest, sources)

创建归档文件。格式由目标路径扩展名决定：`.tar`、`.tar.gz`/`.tgz`、`.zip`。

**参数**：
| 名称 | 类型 | 说明 |
|------|------|------|
| dest | string | 目标归档路径 |
| sources | list[string] | 源文件/目录列表 |

**返回**：`archive.CreateResult`

```
let result = archive.create("/tmp/backup.tar.gz", ["/etc/nginx", "/var/log/app.log"])
print("归档大小: " + str(result.size) + " 字节")
```

### 11.2 archive.extract(src, dest)

解压归档文件到目标目录。支持 tar、tar.gz、zip。包含路径遍历防护。

**参数**：
| 名称 | 类型 | 说明 |
|------|------|------|
| src | string | 归档文件路径 |
| dest | string | 解压目标目录 |

**返回**：`archive.ExtractResult`

```
let result = archive.extract("/tmp/backup.tar.gz", "/tmp/restored")
print("解压文件数: " + str(result.count))
```

---

## 12. ssh 包 - SSH 密钥管理

### 12.1 ssh.authorized_key_add(user, key, exclusive)

为用户添加 SSH 公钥到 authorized_keys。exclusive=true 时清除其他密钥。

**参数**：
| 名称 | 类型 | 说明 |
|------|------|------|
| user | string | 用户名 |
| key | string | SSH 公钥内容 |
| exclusive | bool | 是否独占（清除其他密钥） |

**返回**：`ssh.AuthorizedKeyResult`

```
ssh.authorized_key_add("deploy", "ssh-rsa AAAA...", false)
```

### 12.2 ssh.authorized_key_remove(user, key)

从用户 authorized_keys 中移除指定公钥。

**返回**：`ssh.AuthorizedKeyResult`

### 12.3 ssh.authorized_key_list(user)

列出用户的所有 SSH 公钥。

**返回**：`ssh.AuthorizedKeyListResult` — 包含 `keys`（列表）和 `count`。

---

## 13. kernel 包 - 内核模块管理

### 13.1 kernel.module_list()

列出当前已加载的内核模块（读取 /proc/modules）。

**返回**：`kernel.ModuleListResult` — 包含 `modules`（模块名列表）和 `count`。

```
let mods = kernel.module_list()
print("已加载模块: " + str(mods.count))
```

### 13.2 kernel.module_load(name)

加载指定内核模块（使用 modprobe）。

**返回**：`kernel.ModuleLoadResult`

```
kernel.module_load("br_netfilter")
```

### 13.3 kernel.module_unload(name)

卸载指定内核模块（使用 modprobe -r）。

**返回**：`kernel.ModuleLoadResult`

---

## 14. disk 包 - 磁盘管理

### 14.1 disk.filesystem(device, fs_type)

在设备上创建文件系统。支持 ext2/ext3/ext4/xfs/btrfs/vfat/swap。

**参数**：
| 名称 | 类型 | 说明 |
|------|------|------|
| device | string | 设备路径（如 /dev/sdb1） |
| fs_type | string | 文件系统类型 |

**返回**：`disk.FilesystemResult`

```
disk.filesystem("/dev/sdb1", "ext4")
```

### 14.2 disk.part_list(device)

列出设备的分区信息（使用 lsblk）。

**返回**：`disk.PartListResult` — 包含 `partitions` 列表，每项含 name、size、type、fstype、mountpoint。

```
let parts = disk.part_list("/dev/sda")
for p in parts.partitions {
    print(p.name + " " + p.size + " " + p.fstype)
}
```

---

## 15. file 扩展操作

### 15.1 file.find(paths, patterns, regex, file_type, max_depth, age, size)

在指定路径中查找匹配的文件/目录。

**参数**：
| 名称 | 类型 | 说明 |
|------|------|------|
| paths | list[string] | 搜索根目录列表 |
| patterns | list[string] | glob 模式列表（如 `["*.log"]`） |
| regex | string | 正则表达式过滤（可选，传 `""` 跳过） |
| file_type | string | `"file"` / `"directory"` / `"any"` |
| max_depth | int | 最大递归深度（0=无限） |
| age | int64 | 仅匹配 N 秒前修改的（0=不限） |
| size | int64 | 仅匹配大于 N 字节的（0=不限） |

**返回**：`file.FindResult` — 包含 `matched`（匹配路径列表）和 `count`。

```
let logs = file.find(["/var/log"], ["*.log"], "", "file", 3, 0, 0)
print("找到 " + str(logs.count) + " 个日志文件")
```

### 15.2 file.replace(path, pattern, replacement, after, before)

正则替换文件内容。可选 `after`/`before` 限定替换范围。

**返回**：`file.ReplaceResult` — 包含 `replacements`（替换次数）和 `changed`。

```
let r = file.replace("/etc/app.conf", "old_value", "new_value", "", "")
```

### 15.3 file.blockinfile(path, marker, content, present, insert_after, insert_before)

在文件中插入/更新一个由标记包裹的文本块。标记支持 `{mark}` 占位符（Ansible 风格）。

**返回**：`file.BlockInFileResult`

```
file.blockinfile(
    "/etc/ssh/sshd_config",
    "# {mark} MANAGED BY OPSLANG",
    "PermitRootLogin no\nPasswordAuthentication no",
    true, "", ""
)
```

### 15.4 file.ini_get(path, section, key)

读取 INI 文件中指定 section/key 的值。

**返回**：`file.IniGetResult` — 包含 `value` 和 `found`。

### 15.5 file.ini_set(path, section, key, value)

设置 INI 文件中指定 section/key 的值。自动创建缺失的 section 和 key。

**返回**：`file.IniSetResult`

```
file.ini_set("/etc/my.cnf", "mysqld", "max_connections", "500")
```

---

## 16. net 扩展操作

### 16.1 net.download(url, dest, checksum_algo, checksum_expected)

下载文件到本地。支持 checksum 校验（md5/sha1/sha256）。

**参数**：
| 名称 | 类型 | 说明 |
|------|------|------|
| url | string | 下载 URL |
| dest | string | 本地目标路径 |
| checksum_algo | string | 校验算法（`""` 跳过校验） |
| checksum_expected | string | 期望的校验值 |

**返回**：`net.DownloadResult` — 包含 `dest`、`size`、`status_code`、`checksum`。

```
let dl = net.download("https://example.com/app.tar.gz", "/tmp/app.tar.gz", "sha256", "abc123...")
```

### 16.2 net.wait_for_connection(host, port, timeout)

等待 TCP 端口可达。轮询间隔 1 秒，超时后报错。

**返回**：`net.WaitForConnectionResult` — 包含 `connected`（bool）和 `elapsed_ms`。

```
net.wait_for_connection("db.internal", 5432, 60)
```

---

## 17. sys 扩展操作

### 17.1 sys.timezone_get()

获取当前系统时区。

**返回**：`sys.TimezoneResult` — 包含 `timezone` 字段。

### 17.2 sys.timezone_set(timezone)

设置系统时区（需要 root）。写入 /etc/timezone 并更新 /etc/localtime 软链接。

```
sys.timezone_set("Asia/Shanghai")
```

### 17.3 sys.reboot()

重启系统。使用 `shutdown -r now`，回退到 `reboot`。

**返回**：`sys.RebootResult`

---

## 18. user 包 - 用户管理

直接调用 `useradd`/`usermod`/`userdel`（无 shell 包装）。读取 `/etc/passwd`，任何平台可查询，变更需要 Linux + root。除 `18.1`-`18.5` 的底层操作外，推荐日常使用 **18.6 `user.ensure` / 18.7 `user.absent`** 的幂等收敛形式。

### 18.1 user.list()

返回 `/etc/passwd` 全量用户。返回 `ListResult { users: []UserInfo{username, uid, gid, comment, home, shell} }`。

### 18.2 user.exists(username)

返回 `ExistsResult { exists bool }`。

### 18.3 user.info(username)

返回 `InfoResult { user UserInfo }`；用户不存在时报错。

### 18.4 user.add(username, opts)

创建用户（已存在则 `changed=false`）。`opts` 支持键：`shell`、`home`、`uid`、`groups`、`create_home`（`"true"`/`"1"`）。

**同名组处理**：当 `/etc/group` 已存在同名组时（常见于先 `group.ensure` 再建服务账号的剧本），自动使用 `useradd -g <gid>` 绑定既有组，而不是让 useradd 因"组已存在"而失败。

返回 `AddResult { changed, username, uid, error }`。

### 18.5 user.remove(username, remove_home)

删除用户（不存在则 `changed=false`）。`remove_home=true` 时连带删除家目录（`userdel -r`）。返回 `RemoveResult { changed, username, error }`。

### 18.6 user.ensure(username, opts)

幂等收敛用户状态（present + 属性），对标 Ansible `user` 模块 `state=present`：

| 当前状态 | 行为 | changed |
|---|---|---|
| 用户不存在 | `useradd`（带 opts） | true |
| 存在，shell/home 与期望一致 | **零动作** | false |
| 存在，shell 或 home 漂移 | `usermod -s` / `usermod -d` 收敛 | true |

**参数**：`username`（string，必填）、`opts`（dict，可选；支持 `shell`/`home`/`uid`/`groups`/`create_home`，其中 `shell`/`home` 参与漂移收敛）。

**返回类型**：`EnsureResult`

| 字段 | 类型 | 说明 |
|---|---|---|
| `username` | `string` | 用户名 |
| `present` | `bool` | 收敛后的存在状态（ensure 后恒 true） |
| `changed` | `bool` | 是否产生真实变更 |
| `action` | `string` | `"ensure"` |
| `shell` / `home` | `string` | 收敛后的属性 |
| `uid` | `int` | 用户 UID |
| `message` | `string` | 结果说明（`"user created"` / `"user already up to date"` / `"user attributes converged"`） |
| `error` | `string` | 错误详情（仅失败时） |

**平台与权限**：Linux + root。`Mutating`。

**示例**（真机 Ubuntu 22.04 真实输出，AOT 模式）：

```ops
let g = group.ensure("opslang-demo", {})
let u = user.ensure("opslang-demo", { "shell": "/usr/sbin/nologin", "create_home": "false" })
print("changed=" + str(u.changed) + " shell=" + u.shell)
```

```text
changed=true shell=/usr/sbin/nologin        # 第一次：创建
changed=false shell=/usr/sbin/nologin       # 第二次：零动作
changed=true shell=/usr/sbin/nologin        # 手工 usermod 改成 /bin/bash 后：漂移被收敛回 nologin
```

### 18.7 user.absent(username, remove_home)

幂等删除用户（对标 `state=absent`）：不存在则零动作；**拒绝删除 root**（远程主机上几乎不可能是本意且不可恢复）。`remove_home=true` 连带删除家目录。

返回 `EnsureResult`（`present=false`，`message` 为 `"user removed"` / `"user already absent"`）。

> 真实行为说明：`userdel` 在删除用户时会顺带删除其同名主组（无其他成员时），因此随后执行 `group.absent` 常见 `"group already absent"` —— 这不是 bug，是 shadow-utils 的真实语义。

---

## 19. group 包 - 组管理

直接调用 `groupadd`/`groupdel`，读取 `/etc/group`。

### 19.1 group.list()

返回 `[]GroupInfo{gid, name, members}`。

### 19.2 group.exists(name)

返回 `ExistsResult { exists bool }`。

### 19.3 group.info(name)

返回 `GroupInfo`；组不存在时报错。

### 19.4 group.add(name, opts)

创建组。**已幂等**（对标 Ansible）：组已存在时 `changed=false` 且不执行 groupadd。`opts` 支持 `gid`、`system`（`"true"` 时 `groupadd -r`）。返回 `AddResult { changed, gid, error }`。

### 19.5 group.remove(name)

删除组（不存在则 `changed=false`）。返回 `RemoveResult { changed, error }`。

### 19.6 group.ensure(name, opts)

幂等收敛组存在性（对标 `state=present`）：不存在则按 opts 创建（`changed=true`），存在则零动作并回报现有 `gid`。不支持对既有组重编 GID（`groupmod -g` 几乎从不是运维想要的收敛行为）。

返回 `EnsureResult { name, present, changed, action, gid, message, error }`。

**平台与权限**：Linux + root。`Mutating`。

### 19.7 group.absent(name)

幂等删除组（对标 `state=absent`）：不存在则零动作。返回同上（`present=false`）。

---

## 20. 幂等收敛（ensure 家族）——对标 Ansible 核心模块

OpsLang 的 ensure 家族是**收敛式运维**的核心：你声明期望状态，操作先读真实状态、再决定是否动作。重复执行同一脚本是安全的——第二次及以后全部 `changed=false`、`actions` 为空。这与 `install/start/add` 这类"动作式" API 的本质区别在于：动作式 API 无法安全重复执行。

**与 Ansible 模块的对应关系**：

| OpsLang | Ansible 模块 | Ansible 参数 | 语义差异 |
|---|---|---|---|
| `pkg.ensure(name)` | `package` | `state=present` | 无 |
| `service.ensure(name, state)` | `service` / `systemd` | `state=started/stopped/restarted/reloaded` | reload 失败回退 restart，与 Ansible 一致 |
| `service.ensure_enabled(name, enabled)` | `service` / `systemd` | `enabled=yes/no` | 拆成独立操作，组合使用 |
| `user.ensure(name, opts)` | `user` | `state=present` + `shell`/`home` | 单次调用只收敛 shell/home 漂移 |
| `user.absent(name, remove_home)` | `user` | `state=absent` | 额外拒绝删除 root |
| `group.ensure(name, opts)` | `group` | `state=present` | 不支持 GID 重编 |
| `group.absent(name)` | `group` | `state=absent` | 无 |
| `file.ensure(path, state, mode)` | `file` | `state/path/mode` | `touch` 不刷新 mtime（保持严格幂等）；`state=file` 不创建（一致） |
| `file.lineinfile(path, line, present, rx)` | `lineinfile` | `line/regexp/state` | 已有 |

**通用返回约定**：每个 ensure 操作都返回 `changed`（本次是否真实变更）与 `actions`（实际执行的动作列表）。**审计这两 个字段**是判断"这次部署到底改了什么"的唯一可信来源。

**三引擎一致性**：ensure 家族在解释器（`opsctl run`）、远程 Runner（deploy runner 模式）与 AOT 编译二进制（deploy aot 模式）中语义完全一致，由 `internal/opsspec` 规范表与跨引擎一致性测试保证。

**真机验证记录**（3 台 Ubuntu 22.04，runner 与 AOT 双模式）：

1. 首次部署：`group_changed=true, user_changed=true`（真实创建）
2. 立即重复部署：**全部 `changed=false`**（幂等实证）
3. `systemctl stop cron` 后部署：`start_changed=true, actions=["start"]`，终态 `active=true`（真实收敛）
4. `user.absent`/`group.absent` 清理，重复执行零动作

完整可复现示例见 `examples/remote_ensure_fleet.ops`（舰队供给）与 `examples/ensure_idempotency_proof.ops`（幂等性自证，带断言）。

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

## 附录：文件传输 Roadmap

当前剩余的传输优化能力：

- 传输压缩
- `file.collect` 的分层中继汇聚
