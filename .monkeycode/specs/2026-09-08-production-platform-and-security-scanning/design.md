# OpsLang 生产化控制面与安全扫描设计

## 分层架构

```text
opsctl -> opsd -> scheduler -> worker -> SSH -> ops-runner
                  |             |
                  |             +-> security scanner backend
                  +-> state store, policy, audit, inventory, secrets
```

`opsctl` 负责提交和查询，`opsd` 负责生命周期，scheduler 负责批量调度，runner 负责目标机执行。扫描引擎通过 provider 接口接入，控制面只依赖统一报告模型。

## 扫描后端

```go
type ScannerBackend interface {
    Name() string
    Capabilities() []Capability
    Scan(context.Context, ScanRequest) (ScanReport, error)
}
```

二进制后端负责版本校验、架构选择、临时上传、参数构造、JSON 解析和清理。服务后端负责 HTTPS、认证、数据库版本绑定和结果校验。

## 执行状态

任务状态使用事件记录：`planned`、`leased`、`started`、`applied`、`verified`、`compensated`、`failed`。状态存储必须支持幂等写入和按任务、主机、阶段查询。

## 实施顺序

1. 通用 scanner backend 和统一报告模型。
2. 控制端 scanner service 配置与 HTTPS 客户端。
3. 目标机临时二进制执行和结果转换。
4. 持久化任务状态与事件。
5. capability 权限和 Secret Provider。
6. 批量调度、限流、恢复和审计。
7. 变更事件回滚和生产策略。
