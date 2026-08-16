# OpsLang 技术内幕：双执行引擎是如何工作的

> 一个脚本，两种执行方式。Runner 模式零延迟，AOT 模式零依赖。OpsLang 如何在两者之间自动切换？

## 背景

运维脚本面临一个两难选择：

**解释执行**：启动快，无需编译，适合简单任务。但控制流受限，无法使用第三方库。

**编译执行**：性能好，支持完整语言特性，生成独立二进制。但编译有延迟，交叉编译复杂。

Ansible 选了解释（YAML → Python 执行），Terraform 选了编译（HCL → Go 二进制）。OpsLang 两个都要。

## 架构

```
                    ┌─────────────────┐
                    │   OpsLang 脚本   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Lexer → Parser │
                    │     → AST       │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  RequiresAOT()  │
                    │  自动模式选择    │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
     ┌────────▼────────┐          ┌────────▼────────┐
     │   Runner 模式    │          │    AOT 模式      │
     │  AST → JSON 指令 │          │ AST → Go → 二进制 │
     └────────┬────────┘          └────────┬────────┘
              │                             │
     ┌────────▼────────┐          ┌────────▼────────┐
     │  通用 Runner 执行 │          │  静态二进制执行   │
     └────────┬────────┘          └────────┬────────┘
              │                             │
              └──────────┬──────────────────┘
                         │
                ┌────────▼────────┐
                │  JSON 结构化结果 │
                └─────────────────┘
```

## 模式选择：RequiresAOT()

不是所有脚本都需要编译。OpsLang 的自动选择逻辑很简单：

```go
func RequiresAOT(script *ast.Program) bool {
    for _, stmt := range script.Statements {
        switch stmt.(type) {
        case *ast.IfStatement:
            return true
        case *ast.ForStatement:
            return true
        case *ast.WhileStatement:
            return true
        case *ast.FunctionDecl:
            return true
        case *ast.EnsureStatement:
            return true
        case *ast.ParallelBlock:
            return true
        }
    }
    return false
}
```

如果脚本包含 `if`、`for`、`while`、`fn`、`ensure`、`parallel`，就需要 AOT。否则用 Runner。

这意味着：
- **线性采集脚本**（获取 CPU、内存、磁盘）→ Runner 模式，秒级执行
- **带控制流的操作脚本**（如果 CPU 高就报警）→ AOT 模式，编译后执行

用户可以用 `--mode runner` 或 `--mode aot` 手动覆盖。

## Runner 模式：JSON 指令包

Runner 模式的核心思想：**脚本不是代码，是数据**。

```ops
let cpu = sys.cpu.usage()
let mem = sys.memory.info()
report { cpu: cpu, mem: mem }
```

编译为：

```json
{
  "version": "1.0",
  "task_id": "abc123",
  "dry_run": false,
  "instructions": [
    { "op": "sys.cpu.usage", "args": {}, "assign": "cpu" },
    { "op": "sys.memory.info", "args": {}, "assign": "mem" },
    { "op": "report", "args": { "cpu": "cpu", "mem": "mem" } }
  ]
}
```

通用 Runner（预编译的多架构二进制）读取 JSON，按顺序执行指令，输出 JSON 结果。

**优势**：
- 零编译延迟，指令包生成 < 10ms
- Runner 二进制内容寻址缓存，相同架构只上传一次
- 与架构无关，JSON 是通用的

**限制**：
- 只支持线性执行流（无 if/for/while）
- 不能使用用户自定义函数
- 不能导入第三方 Go 库

但对于 80% 的运维采集任务，Runner 模式够用。

## AOT 模式：AST → Go → 二进制

AOT 模式走完整编译管线。

### 1. 代码生成

AST 翻译为 Go 源码。每个 OpsLang 构造映射到对应的 Go 代码：

```ops
// OpsLang
let x = 10
if x > 5 {
    print("big")
}
```

```go
// 生成的 Go
x := 10
if x > 5 {
    fmt.Println("big")
}
```

内置函数调用映射到 `ops-core-sdk`：

```ops
// OpsLang
let cpu = sys.cpu.usage()
```

```go
// 生成的 Go
cpu, err := sys.CPUUsage()
if err != nil {
    return err
}
```

### 2. 编译

生成的 Go 代码写入临时项目，引入 `ops-core-sdk`，调用 `go build`：

```bash
go build -ldflags="-s -w" -o output main.go
```

`-ldflags="-s -w"` 去掉符号表和调试信息，减小二进制体积。`CGO_ENABLED=0` 确保纯静态链接。

### 3. 编译缓存

相同脚本不需要重复编译。缓存键：

```
SHA256(script_content) + SDK_version + target_arch
```

命中缓存直接使用，< 5 秒完成。未命中才走完整编译。

**优势**：
- 完整语言特性（控制流、函数、闭包）
- 可以使用第三方 Go 库（通过 `import go "..."`）
- 生成独立二进制，可脱离 OpsLang 运行

**限制**：
- 首次编译需要 5-30 秒（取决于脚本复杂度）
- 需要 Go 工具链（控制端）

## 远程执行链路

无论哪种模式，远程执行的流程相同：

```
1. SSH 连接目标主机
2. 检测架构：uname -m → GOARCH
3. 准备二进制：Runner 或 AOT 编译产物
4. 上传二进制（内容寻址缓存）
5. 执行，传入脚本或指令包
6. 回收 JSON 结果
7. 清理临时文件
```

### 内容寻址缓存

传统方式：每次执行都上传二进制，浪费带宽。

OpsLang 方式：

```go
func (e *Executor) ensureRemoteBinary(conn *sshx.Connection, localPath string) (string, error) {
    // 1. 计算本地 SHA256
    localHash, err := fileSHA256(localPath)
    if err != nil { return "", err }

    remoteName := "/tmp/ops-runner-" + localHash

    // 2. 探测远程是否已存在（~100 bytes）
    out, err := conn.Run("sha256sum " + remoteName + " 2>/dev/null")
    if err == nil && strings.Contains(out, localHash) {
        return remoteName, nil  // 缓存命中
    }

    // 3. 上传
    if err := conn.Upload(localPath, remoteName); err != nil {
        return "", err
    }

    // 4. 校验
    remoteHash, _ := conn.Run("sha256sum " + remoteName)
    if !strings.Contains(remoteHash, localHash) {
        return "", fmt.Errorf("checksum mismatch after upload")
    }

    return remoteName, nil
}
```

**效果**：
- 首次执行：上传完整二进制（~10MB）
- 后续执行：只传 ~100 bytes 校验和
- 二进制更新：hash 变化，自动重新上传

## 一致性保证

两套引擎必须行为一致。OpsLang 通过一致性测试强制对齐：

```go
func TestRunnerAOTConsistency(t *testing.T) {
    scripts := []string{
        "examples/cpu_check.ops",
        "examples/disk_report.ops",
        // ...
    }

    for _, script := range scripts {
        // Runner 模式执行
        runnerResult := runWithRunner(script)

        // AOT 模式执行
        aotResult := runWithAOT(script)

        // 结果必须一致
        if runnerResult != aotResult {
            t.Errorf("inconsistent results for %s", script)
        }
    }
}
```

SDK 操作注册也强制对齐——解释器和 Runner registry 必须包含相同的操作集。

## 安全考虑

### 权限分级

```ops
// 脚本头部声明权限
privilege "read_only"
```

操作分类：
- `read`：只读（sys.cpu.usage, file.read）
- `write`：写文件（file.write, file.delete）
- `exec`：执行命令（process.exec）
- `admin`：系统管理（service.restart, pkg.install）
- `system`：root 操作（sys.users, process.kill）

`read_only` 脚本调用 `write` 以上操作，编译期报错。

### 资源限制

```go
limits := &security.ResourceLimits{
    MemoryMB: 1024,
    CPUPercent: 100,
}
security.ApplyResourceLimits(limits)
```

通过 `setrlimit(2)` 限制内存使用。CPU 配额当前未强制执行（ulimit 无法设置 CPU quota，需要 systemd-run --scope，但那是外部编排器的职责）。

### 审计日志

每次执行生成 JSON 审计日志：

```json
{
  "task_id": "abc123",
  "script": "cpu_check.ops",
  "privilege": "read_only",
  "targets": ["host1", "host2"],
  "user": "ops",
  "mode": "aot",
  "started_at": "2026-08-16T10:00:00Z",
  "finished_at": "2026-08-16T10:00:05Z",
  "status": "success"
}
```

## 性能数据

本地测试（macOS M2, 16GB）：

| 操作 | 耗时 |
|------|------|
| Lexer + Parser | < 5ms |
| Runner 模式（线性脚本） | < 10ms |
| AOT 首次编译 | 5-30s |
| AOT 缓存命中 | < 5s |
| 远程执行（缓存命中） | SSH 延迟 + 执行时间 |
| 远程执行（首次上传） | + 1-3s（10MB 二进制） |

远程文件分发（100 主机，1GB 文件）：

| 模式 | 控制端带宽 | 耗时 |
|------|-----------|------|
| 直接分发 | 100GB | ~10min |
| 压缩传输 | ~60GB | ~6min |
| 校验和去重 | ~30GB（假设 50% 重复） | ~3min |

## 已知限制

诚实地说：

1. **Runner 模式不支持控制流**：if/for/while 必须用 AOT
2. **CPU 配额未强制执行**：只有内存限制生效
3. **权限未自动执行**：`CheckPrivilege()` 存在但解释器不自动调用
4. **CI 未跑 `-race`**：TSan 在全量测试时 OOM，本地已支持
5. **无模块系统**：OpsLang 脚本间无法互相导入

这些是有意识的简化，不是遗漏。代码中有 `ponytail:` 注释标明升级路径。

## 总结

OpsLang 的双执行引擎不是过度设计，而是务实选择：

- **Runner 模式**覆盖 80% 的简单采集任务，零延迟
- **AOT 模式**覆盖 20% 的复杂逻辑，零依赖
- **自动选择**让用户不需要关心底层细节
- **一致性测试**保证两种模式行为相同

这是 Ponytail 方法论的体现：停在第一个够用的台阶上。

---

代码：https://github.com/j4ckzh0u/opslang  
文档：`docs/` 目录  
示例：`examples/` 目录
