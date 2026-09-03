# 原子操作 SDK 模块

`pkg/ops-core-sdk` 提供可独立使用的运维原子操作。每个领域一个 Go package，返回结构化类型与显式错误，供解释器、Runner 和 AOT 生成代码复用。

## 模块范围

SDK 包覆盖系统信息、文件、网络、进程、服务、软件包及大量与 Ansible 模块对齐的运维领域。目录数量和操作列表持续增长，准确全集见 `internal/opsspec/spec.go` 和 `docs/generated/ops-index.md`。

## 契约

- 返回值可 JSON 序列化。
- 变更操作携带 changed 或等价状态。
- 参数错误包含操作名和字段上下文。
- 外部命令通过参数数组执行，避免拼接 shell 字符串。
- 控制器专用函数通过 opsspec 作用域限制执行位置。

## 文件包

`pkg/ops-core-sdk/file` 除本地文件操作外，还包含控制器侧 distribute、collect、恢复元数据、中继拓扑、HTTPS 服务、Range 客户端和 SSH 传输接线。传输测试覆盖单机协议、并发编排和万级模拟。

## 添加模块

新增 SDK 功能时同时维护实现、测试、opsspec 和三引擎接入。运行 `make docs` 生成操作索引，并以 `make docs-check` 验证提交内容。
