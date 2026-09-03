# 批量安装 sl 命令

这个示例演示如何使用 OpsLang 批量安装 `sl`（Steam Locomotive）命令，并生成详细的执行报告。

`sl` 是一个有趣的命令行工具，当你误输入 `ls` 时会显示一个蒸汽火车动画。虽然不实用，但作为演示批量安装和报告生成的例子很合适。

## 文件说明

- `install_sl.ops` - OpsLang 脚本，执行安装逻辑
- `run_install_sl.sh` - Shell 包装脚本，执行 opsctl 并生成汇总报告

## 使用方法

### 方法 1：直接使用 opsctl

```bash
# 编译 opsctl（如果还没编译）
go build -o opsctl ./cmd/opsctl

# 在单台主机执行
./opsctl deploy examples/install_sl.ops --targets user@192.168.1.10

# 在多台主机执行
./opsctl deploy examples/install_sl.ops --targets user@host1,user@host2,user@host3

# 从 inventory 文件读取主机列表
./opsctl deploy --inventory hosts.txt examples/install_sl.ops
```

### 方法 2：使用包装脚本（推荐）

```bash
# 自动执行并生成 Markdown 报告
./examples/run_install_sl.sh user@host1 user@host2 user@host3

# 报告会生成在 reports/ 目录：
# - install_sl_20260817_123456.json      (原始 JSON 结果)
# - install_sl_20260817_123456_summary.md (Markdown 汇总报告)
```

## 脚本功能

### install_sl.ops

- **检查已安装**：如果 sl 已存在，跳过安装（幂等）
- **执行安装**：使用 `pkg.install("sl")`，自动检测 apt/yum/dnf
- **验证安装**：检查 sl 路径和版本
- **记录详细数据**：
  - 主机名、操作系统、内核版本
  - 安装状态（success/skipped/failed）
  - 安装耗时（毫秒）
  - 包管理器类型
  - 磁盘和内存使用率
  - 时间戳

### run_install_sl.sh

- 执行 opsctl deploy
- 保存原始 JSON 结果
- 生成 Markdown 汇总报告，包含：
  - 执行统计（成功/跳过/失败数量）
  - 失败详情
  - 操作系统分布
  - 性能统计（平均/最快/最慢安装耗时）

## 输出示例

### JSON 报告

```json
{
  "task_id": "abc123",
  "results": {
    "host1": {
      "host": "web-01",
      "os": "ubuntu",
      "arch": "x86_64",
      "kernel": "5.15.0",
      "action": "install_sl",
      "status": "success",
      "changed": true,
      "path": "/usr/games/sl",
      "version": "sl version 5.02",
      "duration_ms": 2340,
      "package_manager": "apt",
      "disk_usage_percent": 45.2,
      "memory_usage_percent": 68.3,
      "timestamp": "2026-08-17T12:34:56+08:00"
    }
  }
}
```

### Markdown 汇总报告

```markdown
# sl 安装报告

**执行时间**: 2026-08-17T12:34:56+08:00
**目标主机**: host1,host2,host3
**退出码**: 0

## 统计

- **总主机数**: 3
- **成功**: 2
- **跳过（已安装）**: 1
- **失败**: 0

## 操作系统分布

- ubuntu: 2
- centos: 1

## 性能

- **平均安装耗时**: 2340ms
- **最快**: 1800ms
- **最慢**: 2880ms
```

## 注意事项

1. **权限要求**：安装软件包需要 root 或 sudo 权限
2. **包管理器**：支持 apt（Debian/Ubuntu）、yum/dnf（RHEL/CentOS/Fedora）
3. **网络要求**：目标主机需要能访问包管理器的软件源
4. **幂等性**：已安装 sl 的主机会跳过，不会重复安装

## 自定义

如果需要安装其他软件，修改 `install_sl.ops`：

```ops
// 把 "sl" 改成你想要的软件包名
let result = pkg.install("nginx")

// 修改验证命令
let verify = process.exec("which", "nginx")
```

## 故障排查

### 安装失败

检查报告中的 `error` 字段，常见原因：
- 权限不足（需要用 root 或 sudo）
- 网络不通（无法访问软件源）
- 包名不存在（不同发行版包名可能不同）

### 报告生成失败

`run_install_sl.sh` 需要 python3 来解析 JSON 和生成统计。如果没有 python3，只会生成原始 JSON 报告。

## 下一步

这个示例展示了：
- 批量软件安装
- 幂等操作（检查已安装）
- 详细数据记录
- 结构化报告生成
- 多主机并行执行

可以基于这个模式创建其他批量操作脚本：
- 批量安装监控 Agent
- 批量配置 NTP
- 批量更新系统包
- 批量部署应用
