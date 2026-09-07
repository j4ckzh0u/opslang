# Trivy 能力集成实施计划

- [x] 1. 完成 Trivy 能力范围和 OpsLang 边界设计
  - [x] 1.1 明确第一批目标为主机漏洞、文件系统依赖和 SBOM
  - [x] 1.2 明确控制端提供离线规则 bundle
  - [x] 1.3 明确容器、Kubernetes、IaC、密钥和许可证扫描为后续阶段

- [x] 2. 实现统一扫描数据模型
  - [x] 2.1 新增 `securityscan` SDK 包和 Scanner/Options/ScanResult 类型
  - [x] 2.2 新增 vulnerability finding、SBOM component、scan error 类型
  - [x] 2.3 新增规则 bundle 校验、摘要和稳定排序
  - [x] 2.4 覆盖 nil、空输入、非法 scanner、损坏 bundle 和重复组件测试

- [x] 3. 接入主机漏洞扫描
  - [x] 3.1 复用 `software.InventoryResult` 和现有版本比较逻辑
  - [x] 3.2 增加严重级别过滤、忽略规则和结果元数据
  - [x] 3.3 保持 `vulnerability.match` 兼容并新增统一扫描入口

- [x] 4. 实现文件系统清单扫描
  - [x] 4.1 实现 WalkDir 限制、忽略路径和上下文取消
  - [x] 4.2 支持 Go、Node.js、Python 清单文件
  - [x] 4.3 保留单文件解析错误并汇总组件
  - [x] 4.4 覆盖空目录、损坏文件、超大文件、深度和权限边界

- [x] 5. 实现 SBOM 输出
  - [x] 5.1 生成 OpsLang 原生对象
  - [x] 5.2 生成最小有效 CycloneDX JSON
  - [x] 5.3 生成最小有效 SPDX JSON
  - [x] 5.4 增加稳定组件 ID 和重复路径合并测试

- [x] 6. 接入解释器、Runner、AOT 和 CLI
  - [x] 6.1 注册 `security.scan`、`software.sbom`、`file.scan`
  - [x] 6.2 增加三引擎一致性和动态参数错误测试
  - [x] 6.3 新增 `opsctl scan` 本地目录和 inventory 输入
  - [x] 6.4 增加 JSON、CycloneDX、SPDX 输出和严重级别门禁

- [x] 7. 更新文档和示例
  - [x] 7.1 更新标准库参考、CLI 参考和能力边界
  - [x] 7.2 增加主机漏洞、文件系统和 SBOM 示例
  - [x] 7.3 重新生成操作索引并运行 `make docs-check`

- [x] 8. 完成质量验证和交付
  - [x] 8.1 运行格式化、全量测试、`go vet ./...`
  - [x] 8.2 运行六平台静态构建
  - [x] 8.3 提交、推送并监控 GitHub Actions

- [x] 9. 控制端远程漏洞查询
  - [x] 9.1 增加 `RemoteConfig` 和 HTTPS 规则匹配客户端
  - [x] 9.2 增加 `opsctl vulndb serve` 和 Bearer token 校验
  - [x] 9.3 增加规则版本、SHA-256、TLS 和响应体边界校验
  - [x] 9.4 增加远程查询单元测试
  - [x] 9.5 将远程配置接入 Runner 指令自动生成和下发
  - [x] 9.6 运行完整测试、构建、提交并监控 GitHub Actions
