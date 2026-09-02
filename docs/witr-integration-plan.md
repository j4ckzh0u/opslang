# WITR 能力拆解与 OpsLang 集成计划

本文基于 WITR 公共仓库当前 README/`llms.txt` 的能力描述整理。WITR 的核心问题是回答“why is this running?”：把进程、端口、容器或文件追溯到启动它的因果链，并提供人类可读树、JSON 和交互式 TUI 输出（[项目 README](https://github.com/pranshuparmar/witr)，[llms.txt](https://raw.githubusercontent.com/pranshuparmar/witr/main/llms.txt)）。

## 能力拆解

1. **目标解析层**：按 PID、进程名、端口、容器、文件句柄解析查询目标；支持模糊/精确匹配及多目标组合。
2. **事实采集层**：读取 `/proc`、进程父子关系、启动时间、工作目录、环境变量、打开文件和网络套接字；结合 systemd、Docker/Podman/containerd、Kubernetes、LXC 等运行时元数据。
3. **因果图层**：将目标统一映射为 PID，沿 PPID、service/supervisor、容器 init 等边界构建有向链，记录缺失权限和不确定来源。
4. **呈现层**：文本摘要、树形 ancestry、JSON 机器接口、warnings/verbose 模式和 Bubble Tea TUI。
5. **动作层**：TUI 中对进程执行 signal、pause/resume、renice 等操作；这部分必须与 OpsLang 的权限、审批和审计模型绑定。
6. **平台适配层**：WITR 发布单静态二进制并覆盖 Linux、macOS、FreeBSD、Windows；OpsLang 当前优先 Linux，Windows 适配放在同一接口之后。

## 融入 OpsLang 的边界

不要把 WITR CLI 作为远程 shell 命令调用。应复用其能力模型，在 `pkg/ops-core-sdk/causal` 中实现纯 Go、只读的事实和图构建 API，再由三条现有执行路径注册：

- `internal/opsspec`：增加 `causal.find`, `causal.trace_pid`, `causal.trace_port`, `causal.trace_container`, `causal.trace_file` 以及 `causal.snapshot` 规范；声明每个调用的只读/变更属性。
- `internal/runner/registry.go` 与 `internal/interpreter/sdk_bridge.go`：注册同名 SDK 函数，统一返回可 JSON 序列化的 `CausalTrace`，避免解释器和远程 runner 产生不同语义。
- `internal/compiler/codegen.go`：加入 AOT 映射，确保脚本编译为目标机静态二进制后仍只依赖 Linux 内核接口，不依赖 Python、shell 或目标机上的 WITR 安装。
- `cmd/opsctl`：增加 `opsctl trace` 便捷命令（底层仍调用 SDK），并支持 `--json`, `--tree`, `--short`, `--warnings`；脚本中的 `report` 继续作为机器接口。

## 分阶段实施

### Phase 1：Linux 只读核心

- 先实现 `/proc` 进程快照、PPID 链、systemd unit 推断、监听端口到 PID 的映射。
- 定义稳定模型：`Target`, `ProcessNode`, `Edge`, `Evidence`, `Warning`, `CausalTrace`，每条证据带来源和采集时间。
- 加入 fixture 驱动测试（伪造 proc 树、权限拒绝、孤儿进程、PID 重用），并验证 interpreter/runner/AOT 三路径结果一致。

### Phase 2：容器与文件关联

- 解析 Docker/containerd/CRI-O cgroup ID，并通过运行时 socket（可选、超时、只读）补充容器名、镜像和 compose/k8s 标签。
- 解析 `/proc/<pid>/fd`、`/proc/net/*` 与 Unix socket，支持文件和端口反查；运行时不可用时返回 warning 而不是失败。

### Phase 3：查询与输出体验

- 提供 `causal.find` 的多目标批量接口和分页/节点上限，防止海量主机上结果失控。
- 在 `opsctl` 中实现文本树和 JSON；TUI 作为独立前端包，不能进入无终端的远程 runner。

### Phase 4：安全动作与跨平台

- 将 signal/renice 等动作接入 `ensure`、审批、签名和审计；默认关闭，显式 capability 才能执行。
- 抽象 `ProcessProvider`/`SocketProvider`，Linux 使用 `/proc`，后续 Windows 使用 ETW/WMI，保持脚本 API 不变。

## 依赖与许可决策

WITR 是 Apache-2.0 项目；引入代码前应保留版权和 NOTICE，并优先采用“接口/数据模型借鉴 + 独立实现”。WITR 当前 Go 模块还包含 Bubble Tea、go-systemd、DBus 等依赖（[go.mod](https://raw.githubusercontent.com/pranshuparmar/witr/main/go.mod)）；这些 UI/运行时依赖不应无条件进入 OpsLang 核心二进制。可选的 TUI 模块和 Linux 运行时探测模块应通过构建标签或独立命令隔离。

## 验收标准

- 同一目标在 interpreter、runner、AOT 三模式下产生等价 JSON。
- 无权限读取单个 `/proc` 文件时，结果保留已知链并附 `warnings`，不导致整批主机失败。
- 对宿主机可见的 Docker PID namespace 能显示容器 ID；隔离 PID namespace 明确报告“不可见”，并提示在容器内执行。
- 远程执行不调用 `witr`、Python 或 shell；仅上传 OpsLang 静态 runner/AOT 二进制。
