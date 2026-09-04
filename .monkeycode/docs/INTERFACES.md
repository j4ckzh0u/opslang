# OpsLang 接口

## CLI

| 命令 | 作用 |
|------|------|
| `opsctl run <script.ops>` | 本地解释执行脚本 |
| `opsctl repl` | 启动交互式解释环境 |
| `opsctl build <script.ops>` | AOT 编译为静态二进制 |
| `opsctl exec` | 向远端 Runner 发送 JSON 指令包 |
| `opsctl deploy <script.ops>` | 选择目标、执行 task 并聚合结果 |
| `opsctl version` | 输出 CLI 版本 |
| `opsctl keygen` | 生成指令包签名密钥，由安全构建标签接入 |

完整参数以 `opsctl <command> --help` 和 `docs/cli-reference.md` 为准。

## Runner 协议

Runner 从标准输入读取协议版本 `1.0` 的 JSON：

```json
{
  "version": "1.0",
  "task_id": "example-task",
  "dry_run": false,
  "instructions": [
    {
      "op": "sys.hostname",
      "args": {},
      "assign": "host"
    }
  ]
}
```

输出包含 `status`、结构化 `data`、`errors` 和 `warnings`。进程退出码为：`0` 全部成功、`1` 部分失败、`2` 全部失败、`3` 协议或命令用法错误。

使用 `--pubkey <path>` 时，Runner 在执行前验证 Ed25519 签名。公钥文件支持原始 32 字节或十六进制编码。

## Runner 中继子命令

```text
ops-runner relay serve --file <path> --listen <address> --advertise-host <host> --ttl <duration> --max-concurrent <count> [--detach]
ops-runner relay fetch --url <url> --token <token> --fingerprint <sha256> --sha256 <sha256> --size <bytes> --dest <path> [--wire-sha256 <sha256> --wire-size <bytes> --decompress]
```

`serve` 暴露固定 `/file` 路径，支持 `GET`、`HEAD` 和 Range 请求。`fetch` 固定 TLS 叶证书 SHA-256 指纹，拒绝重定向，并在大小和内容哈希验证后原子提交。

## file.distribute

OpsLang 调用形式：

```go
let result = file.distribute(
    "/data/release.bin",
    [
        {
            "host": "10.0.1.10",
            "user": "ops",
            "port": 22,
            "dest": "/opt/app/release.bin",
            "relay_group": "zone-a",
            "tags": {"relay": "true"}
        }
    ],
    {
        "checksum": true,
        "resume": true,
        "compress": true,
        "relay": true,
        "relay_threshold": 20,
        "relay_max_targets": 100,
        "part_retention": 86400000
    }
)
```

选项：

| 字段 | 类型 | 默认行为 |
|------|------|----------|
| `checksum` | bool | 控制是否执行传输后 SHA-256 校验 |
| `mode` | string | 非空时应用八进制权限字符串 |
| `parallel` | integer | 非正数时使用 5 |
| `retries` | integer | 非正数时总尝试次数 3 |
| `resume` | bool | 默认关闭；启用内容寻址和部分文件恢复 |
| `compress` | bool | 默认关闭；使用 gzip 传输并在目标端校验后解压 |
| `relay` | bool | 默认关闭；启用分层中继计划 |
| `relay_group` | string | 全局中继组后备值 |
| `relay_threshold` | integer | 默认 20 |
| `relay_max_targets` | integer | 默认 100 个中继下游目标 |
| `part_retention` | 毫秒数 | 默认 24 小时；负值返回参数错误 |

每主机结果包含 `host`、`status`、`changed`、`checksum`、`size`、`transfer_source`、`resumed_bytes`、`transferred_bytes`、`warnings`、`duration_ms` 和可选 `error`。

中继组优先级依次为目标 `relay_group`、目标 `tags.relay_group`、全局 `relay_group`、IPv4 `/24` 或 IPv6 `/64`。无法解析为 IP 且没有显式组的目标使用直接 SFTP。

## file.collect

```go
let result = file.collect(
    "/var/log/app.log",
    [{"host": "10.0.1.10", "user": "ops", "port": 22}],
    {
        "dest_dir": "/data/collected",
        "parallel": 10,
         "resume": true,
         "compress": true,
         "relay": true,
         "relay_threshold": 20,
         "relay_max_targets": 100,
         "part_retention": 86400000
    }
)
```

收集文件按目标主机归档到 `dest_dir/<host>/<basename>`。`resume` 使用本地部分文件和元数据继续 SFTP 下载；`compress` 使用 gzip 传输、解压到临时文件并在原子提交前计算 SHA-256。`relay` 启用方案 2 时，每个源主机提供短时 HTTPS 文件服务，中继节点通过 `relay fetch` 拉取后由控制端从中继下载。最终结果包含大小、SHA-256、恢复字节和实际传输字节；中继候选失败后会切换候选，候选全部失败后回退控制端直连。

## SSH 配置

控制器侧文件传输读取项目定义的 SSH 配置和凭据入口。常用环境变量包括：

| 变量 | 用途 |
|------|------|
| `OPSLANG_SSH_PASSWORD` | 文件分发与收集的 SSH 密码 |
| `OPSLANG_SSH_KEY` | SSH 私钥路径 |
| `OPSLANG_REMOTE_RUNNER` | 远端 `ops-runner` 可执行文件路径 |

凭据只用于控制端建立目标连接。中继会话参数包含短时令牌、证书指纹、文件哈希和大小。

## Go SDK

公开原子操作位于 `pkg/ops-core-sdk/<module>`。函数遵循强类型结构化返回和显式 `error`。控制器专用操作由 `internal/opsspec` 标记作用域，Runner 与 AOT 对不支持的作用域返回错误。

操作全集由 `docs/generated/ops-index.md` 生成。新增操作时需要同步 SDK 实现、opsspec、解释器桥接、Runner registry 和 AOT codegen，并通过一致性测试。
