# witr 能力研究与 OpsLang 集成建议

研究对象：[`pranshuparmar/witr`](https://github.com/pranshuparmar/witr)，源码快照 `dc4fa1da82d3e266fcbd928641b4f30b3077c64f`（2026-08-08）。本文只基于上游仓库 README、Go 源码、`go.mod`、`.goreleaser.yml` 和许可证进行分析，不复制上游实现代码。

## 1. 项目定位

witr 的核心问题是回答“为什么这个进程/端口/容器/文件正在运行”，把进程祖先链、启动来源和运行上下文组合成一份因果解释。README 明确把目标归纳为：进程、端口、容器、文件最终映射到 PID，再沿 PID 祖先链解释其存在原因。[README Purpose](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/README.md#1-purpose)、[README Core Concept](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/README.md#5-core-concept)

它同时提供：

- CLI 查询：按进程名、PID、端口、文件、容器查询；目标参数可重复并混合。
- 输出模式：标准叙述、短链、树、告警、环境变量、详细信息和 JSON。
- TUI：进程、端口、容器、文件锁四个视图，自动刷新、过滤、排序和进程操作。
- 因果来源识别：systemd、launchd、FreeBSD rc.d、Supervisor/PM2 等监督器、cron、SSH、交互式 shell、容器和 init。

这些 CLI 选项和输出模式在 [`internal/app/app.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/app/app.go) 中统一编排；JSON 输出位于 [`internal/output/json.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/output/json.go)。

## 2. 能力清单

### 2.1 目标解析

`internal/target` 将用户目标解析成 PID：

| 目标 | 处理方式 | 关键实现 |
| --- | --- | --- |
| 进程名 | 模糊/精确匹配进程表 | [`target/name_*`](https://github.com/pranshuparmar/witr/tree/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/target) |
| PID | 校验正整数后直接使用 | [`target/resolve.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/target/resolve.go) |
| 端口 | 端口到 socket/PID 映射，必要时容器回退 | [`target/port_*`](https://github.com/pranshuparmar/witr/tree/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/target) |
| 文件 | 查找持有文件描述符或锁的进程 | [`target/file_*`](https://github.com/pranshuparmar/witr/tree/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/target) |
| 容器 | 遍历已注册 runtime，按名称、镜像、命令、Compose 标签匹配 | [`proc/container_runtime.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/proc/container_runtime.go) |

### 2.2 进程与因果分析

- Linux 直接读取 `/proc/<pid>/stat`、`cmdline`、`environ`、`cwd`、`cgroup`、`fd`、`/proc/net/*`，构造统一的 `model.Process`。[`proc/process_linux.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/proc/process_linux.go)
- 祖先链通过 PPID 逐级回溯，并带环检测。[`proc/ancestry.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/proc/ancestry.go)
- 来源检测采用固定优先级：容器、SSH、shell、systemd、launchd、rc.d、监督器、cron、Windows 服务、init。[`source/detect.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/source/detect.go)
- Linux systemd 信息使用 cgroup 得到 unit 名，再通过 systemd D-Bus 补充描述、unit 文件、重启次数和 timer 调度。[`source/systemd_linux.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/source/systemd_linux.go)
- 结果模型包含 PID/PPID、命令行、可执行文件、用户、启动时间、CPU/内存、工作目录、Git、socket、容器身份、环境变量、子进程、文件描述符、能力和告警。[`pkg/model`](https://github.com/pranshuparmar/witr/tree/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/pkg/model)

### 2.3 容器和 Docker 场景

witr 使用 runtime 接口注册 Docker、Podman、nerdctl/containerd、crictl/Kubernetes、Incus、LXD、经典 LXC 和 FreeBSD jail 等后端；每个后端负责列举容器、解析宿主机 PID、补充详情。[`proc/runtime_*.go`](https://github.com/pranshuparmar/witr/tree/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/proc)

Docker/Podman 类后端通过 CLI 的结构化输出获取 ID、名称、镜像、命令、状态、网络、挂载、端口和标签；单容器查询再用 inspect 补充启动时间、宿主机 PID 和健康检查。[`proc/runtime_dockerlike.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/proc/runtime_dockerlike.go)

当容器运行在独立 PID namespace（例如 Docker Desktop、WSL2 或虚拟机）而宿主机看不到 workload PID 时，witr 保留 runtime 侧元数据并渲染 fallback，而不是把“无法看到进程”误报为容器不存在。该行为在应用层容器处理逻辑和 README 的容器说明中有明确描述。[`app.go` container path](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/app/app.go)、[README Container Based Query](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/README.md#72-exit-codes)

### 2.4 诊断与交互

详细模式会读取内存映射、I/O、文件描述符、线程数、子进程、socket、环境变量和文件上下文；告警包括 root、危险 Linux capability、公共监听地址、重启过多、长期运行、高内存、删除后的可执行文件和 `LD_PRELOAD`/`DYLD_*` 注入迹象。[`pipeline/analyze.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/pipeline/analyze.go)、[`source/detect.go`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/source/detect.go)

TUI 使用 Bubble Tea/Lip Gloss，包含进程操作（kill、terminate、pause、resume、renice）。这部分适合 OpsLang 控制器交互层参考，不建议作为第一阶段远程 DSL 能力直接引入。[`internal/tui`](https://github.com/pranshuparmar/witr/tree/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/internal/tui)

## 3. 架构和依赖评估

### 3.1 上游分层

上游结构大致为：

```text
cmd/witr
  -> internal/app          CLI 参数、目标编排、退出码
  -> internal/target       name/PID/port/file/container 解析
  -> internal/pipeline     ancestry + process + source + enrichment
  -> internal/proc         OS、/proc、socket、runtime 后端
  -> internal/source       systemd/SSH/shell/container/supervisor 识别
  -> internal/output       standard/tree/short/json/env/warnings
  -> internal/tui          本地交互界面
  -> pkg/model              可序列化领域模型
```

核心分析管道和输出模型相对清晰，可以抽取为库；`cmd/witr` 和 `internal/tui` 则是产品界面，不应直接嵌入 OpsLang runner。

### 3.2 纯二进制与环境依赖

上游 GoReleaser 明确使用 `CGO_ENABLED=0`、Linux/macOS/FreeBSD/Windows 的 amd64/arm64 构建，发行物是单二进制。[`.goreleaser.yml`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/.goreleaser.yml)

但“单二进制”不等于“目标机没有系统能力依赖”：

- Linux 主流程依赖 `/proc` 和 `/sys` 可读性；systemd 增强信息依赖 systemd D-Bus。
- 容器发现依赖对应 runtime CLI（Docker、Podman、nerdctl、crictl、Incus/LXC 等）及其 socket/权限。
- macOS/FreeBSD 的部分功能调用 `ps`、`lsof`、`launchctl`、`fstat` 等系统工具。
- 受保护进程、不同用户进程、容器独立 namespace 会产生“未知/部分结果”，需要显式表达不确定性。

这与 OpsLang 的目标兼容，但 OpsLang 文档应把“无 Python/Shell 语言运行时”与“无需任何操作系统组件”区分开。

### 3.3 Go 依赖

直接依赖包括：Bubble Tea 生态（TUI）、`coreos/go-systemd` 和 `godbus/dbus`（systemd）、`golang.org/x/sys`（平台 API）、Cobra（CLI）以及少量终端输出库。[`go.mod`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/go.mod)

OpsLang 若只集成“因果分析库”，可以避免引入 Bubble Tea、Cobra 和 TUI 依赖；Linux-first 版本只保留 `/proc`、socket、cgroup、systemd D-Bus 和 Docker/Podman 适配器所需依赖。

## 4. 许可证与集成边界

witr 仓库使用 Apache License 2.0。[`LICENSE`](https://github.com/pranshuparmar/witr/blob/dc4fa1da82d3e266fcbd928641b4f30b3077c64f/LICENSE)

Apache-2.0 允许商用、修改、分发和专利授权，但分发衍生代码时必须保留许可证、版权/归属声明，并在修改文件中标明修改；如果上游包含 NOTICE，也应随分发物保留。当前仓库根目录未见单独 `NOTICE` 文件，但集成前仍应在发布检查中确认。

推荐的合规边界：

1. 不直接复制 `internal/` 源码，不把上游包改名后混入 OpsLang。
2. 先实现 OpsLang 自己的领域模型和接口，按行为重新实现所需能力。
3. 在 `THIRD_PARTY_NOTICES` 或发布文档中记录 witr 的项目名、仓库地址、Apache-2.0 许可证和集成启发来源。
4. 如果未来确实直接链接或 vendoring 上游代码，必须保留其版权和许可证文本，并对修改文件添加显著变更说明；同时固定上游 commit，避免依赖漂移。

## 5. 融入 OpsLang 的能力拆解

### 阶段 A：只读 Linux 进程因果库

新增一个独立的内部领域包（建议 `internal/causality` 或 `pkg/ops-core-sdk/process_explain`），定义与 witr 无关的 OpsLang API：

```text
process.explain(target, options) -> ProcessExplanation
```

`target` 支持 `pid`、进程名、端口、文件、容器；`options` 支持 `exact`、`verbose`、`include_env`、`include_children`、`include_sockets`。返回结构体至少包含：

- `target`、`resolved_pid`、`process`
- `ancestry[]`（PID、PPID、命令、可执行文件、用户、启动时间）
- `source`（类型、名称、描述、unit/service 信息）
- `container`（runtime、ID、名称、镜像、host PID、namespace visibility）
- `sockets[]`、`children[]`、`context`、`warnings[]`
- 每个可选字段的 `available`/`reason` 或统一 `diagnostics[]`，区分“不存在”和“权限/环境不可用”。

第一阶段直接读取 Linux `/proc`、cgroup 和 `/proc/net`，不调用 `ps`/`lsof`/shell。进程祖先链、systemd unit、SSH、shell、cron 和基础 supervisor 识别可以作为独立 detector 接口。

### 阶段 B：容器可见性和 Docker 内 Java 进程

建立 `ContainerResolver` 接口：

```text
List(ctx) -> []Container
Inspect(ctx, id) -> ContainerDetails
HostPID(ctx, id) -> (pid, visibility)
ProcessNamespace(ctx, id) -> NamespaceView
```

先实现 Docker 和 Podman：优先使用 Go Docker API/Unix socket或直接解析 `/proc` 与 cgroup；不得把用户输入拼接进 shell 命令。若 runtime 不可用，返回 `unavailable`，而不是空列表。若容器 PID namespace 不可见，结果应保留容器元数据并标记 `workload_visible=false`。

这一步直接服务当前测试需求：Java 进程查询必须同时扫描宿主机 PID 和可访问的容器 PID；库提取失败时记录权限、JDK 工具缺失或 namespace 不可见原因。

### 阶段 C：结构化运行时/依赖调查

在 `process.explain` 之上增加可选调查器：

```text
process.java_runtime(pid/container) -> JavaRuntimeReport
```

建议分层：

1. 进程识别：命令名、`/proc/<pid>/exe`、cmdline、环境变量。
2. Java 版本：优先读取 `/proc/<pid>/exe`、`java -version` 仅作为受控 fallback，不依赖 shell。
3. JVM 进程参数：解析 `jcmd`/`jps` 或 `/proc/<pid>/cmdline`；工具不可用时返回 capability 缺失。
4. Java 库：解析 JVM classpath、`-javaagent`、模块路径、JAR 文件名和 manifest `Implementation-Version`；不要把“打开的 `.jar` 文件”直接等同于实际依赖，结果需标注证据来源。
5. 容器：对 host PID 与 container PID 分别采集，合并时保留来源和 namespace。

输出应是稳定 JSON 结构，适合 DSL 的 `report`、阈值判断和批量聚合；不返回依赖 shell 文本解析的长字符串。

### 阶段 D：远程 runner/AOT 接入

- Runner 增加只读 `process.explain`/`process.java_runtime` 指令，参数为 JSON 对象。
- AOT 编译器将其生成到同一套 SDK 调用，保持解释器、runner、AOT 三种执行路径的字段和错误语义一致。
- 远程目标只上传 OpsLang runner/AOT 二进制；目标机不需要 Python、Shell 或 OpsLang 解释器。
- 远端能力探测必须返回结构化 `capabilities`，例如 `/proc`、systemd、docker socket、`jcmd`、`jps`、容器 PID 可见性。
- 只读调查默认不需要提升权限；需要读取其他用户 `/proc`、Docker socket 或 JVM attach 时，显式返回 `permission_required`，不能静默降级为“无 Java 库”。

### 阶段 E：控制器聚合和可视化

先支持 DSL 脚本批量执行、阈值判断和 JSON/CSV 汇总；不要第一阶段引入 witr 的 TUI。后续可以在 OpsLang 控制器中增加“因果链”查看器，复用结构化模型而不是嵌入 Bubble Tea。

## 6. 不建议直接搬入的部分

- witr 的 Cobra CLI、TUI、终端颜色和交互快捷键不属于 runner 领域。
- 跨平台 `ps`/`lsof`/Windows API 适配器应在 OpsLang 需要 Windows/FreeBSD 时按自身接口实现，不要为 Linux-first 版本引入全部依赖。
- 上游 runtime CLI 字符串解析（例如 Docker `--format` 文本）不应成为 OpsLang 的长期协议；应优先使用受控参数调用、Unix socket/API 或专用远端 RPC。
- witr 的“单一 primary source”展示模型适合人类叙述，但 OpsLang 报告应保留多个证据和置信度，避免 systemd、容器、SSH 等来源互相覆盖后丢失事实。
- 进程操作（kill、renice 等）属于变更能力，必须另行登记到 OpsLang 权限/审批模型，不能因为 witr TUI 支持就默认开放。

## 7. 验收标准

集成第一阶段完成时，至少应有以下测试：

1. 本机 Linux：对当前 `ops-runner`、`sshd`、一个监听端口和一个普通用户进程生成稳定 JSON。
2. 两台远端 Linux：验证 `/proc`、systemd/cgroup、端口到 PID、SSH 来源和权限不足时的结构化错误。
3. Docker：容器可见时返回 runtime、ID、名称、镜像、host PID；不可见时返回 fallback 和明确原因。
4. Java：宿主机 Java 进程和 Docker 内 Java 进程分别检测，输出 JVM 版本、classpath/JAR 名称、版本证据来源；工具缺失不能误报为空依赖。
5. 三引擎一致性：Interpreter、Runner、AOT 对同一目标返回相同字段和状态枚举。
6. 安全：所有目标值走参数化 API/安全 quoting；不接受用户输入拼接 shell；只读调查不得被标记为 mutating。
7. 回归：默认构建、`-tags opssec`、Linux amd64/arm64 交叉编译和 `go test ./...` 均通过。

## 8. 结论

witr 最值得吸收的是“进程状态 + 祖先链 + 启动来源 + 容器/服务上下文”的结构化因果模型，而不是它的 CLI/TUI 外壳。OpsLang 可以在保持纯 Go 静态二进制和远程 runner/AOT 模型的前提下，按 `causality core -> container visibility -> Java runtime/dependency evidence -> remote instruction -> controller aggregation` 的顺序分阶段实现。第一阶段应限定 Linux 只读能力，明确权限和环境能力缺失，避免再次把“单二进制”误解为“目标机完全没有系统依赖”。
