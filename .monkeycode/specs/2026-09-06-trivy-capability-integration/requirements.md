# Trivy 能力集成需求

## Introduction

OpsLang 需要吸收 Trivy 的安全扫描能力，形成面向运维主机、文件系统、软件清单、容器制品和 Kubernetes 资源的统一安全检查接口。该能力以 Trivy 官方公开能力范围为参考，采用 OpsLang 原生的结构化结果、离线执行和远程 Runner 传输模型。

第一阶段优先打通主机软件漏洞、文件系统漏洞与 SBOM 输出，复用现有 `software.inventory()`、`vulnerability.match()` 和 Runner/AOT 执行引擎。容器镜像、Kubernetes、IaC、密钥和许可证扫描作为后续可独立交付的扫描器模块。

## Glossary

- **扫描目标**：被安全扫描的主机、目录、制品、镜像或 Kubernetes 资源。
- **扫描器**：针对漏洞、SBOM、配置、密钥或许可证执行检测的模块。
- **漏洞规则源**：由调用方提供或由控制端缓存的漏洞规则数据。
- **Finding**：单条结构化安全发现结果。
- **SBOM**：软件物料清单，描述组件、版本、来源和依赖关系。
- **Trivy 兼容结果**：字段语义可映射到 Trivy JSON 结果的 OpsLang 结构化结果，不要求复制 Trivy 的内部实现。

## Requirements

### Requirement 1: 主机漏洞扫描

**User Story:** AS 运维人员，我希望扫描目标主机上的软件漏洞，以便在部署或巡检阶段识别受影响的软件包。

#### Acceptance Criteria

1. WHEN 脚本调用主机漏洞扫描操作，THE system SHALL 使用目标主机的软件清单和调用方指定的漏洞规则生成结构化 Findings。
2. WHEN 目标主机使用 apt/dpkg、yum/rpm 或 dnf 软件包管理器，THE system SHALL 使用对应的软件包版本语义判断受影响版本。
3. IF 目标主机的软件清单存在单项采集错误，THE system SHALL 保留已采集的软件项并在结果中记录采集错误。
4. IF 漏洞规则列表为空，THE system SHALL 返回初始化的空 Findings 列表。

### Requirement 2: 文件系统扫描

**User Story:** AS 安全工程师，我希望扫描指定目录中的软件包和应用依赖，以便识别未纳入系统包管理器的风险组件。

#### Acceptance Criteria

1. WHEN 用户提供一个存在的目录作为扫描目标，THE system SHALL 递归识别支持的清单文件、归档文件和二进制元数据。
2. WHEN 扫描目标包含可识别的应用依赖，THE system SHALL 输出组件名称、版本、路径、生态系统和来源信息。
3. IF 扫描路径不存在或不是目录，THE system SHALL 返回包含目标路径的明确错误。
4. IF 单个文件解析失败，THE system SHALL 保留其他文件的扫描结果并记录该文件的错误。
5. WHILE 扫描目录执行，THE system SHALL 提供可配置的最大文件大小、最大扫描深度和忽略路径规则。

### Requirement 3: SBOM 生成

**User Story:** AS 合规人员，我希望生成统一格式的 SBOM，以便进行资产盘点、漏洞关联和审计归档。

#### Acceptance Criteria

1. WHEN 用户请求生成 SBOM，THE system SHALL 输出组件名称、版本、包类型、安装路径、许可证字段和来源字段。
2. WHEN 用户指定 CycloneDX 或 SPDX 输出格式，THE system SHALL 生成对应格式的有效 JSON 文档。
3. WHEN 用户请求 OpsLang 原生格式，THE system SHALL 输出可被 `json.decode()` 读取的结构化对象。
4. IF 组件缺少许可证或来源信息，THE system SHALL 保留组件并将缺失字段表示为空值或未知状态。
5. WHEN 相同组件在多个路径出现，THE system SHALL 保留路径信息并提供稳定的组件标识。

### Requirement 4: 统一扫描接口

**User Story:** AS OpsLang 用户，我希望用统一接口选择目标和扫描器，以便在解释器、Runner、AOT 和远程部署中复用安全检查逻辑。

#### Acceptance Criteria

1. WHEN 用户调用安全扫描接口，THE system SHALL 支持目标类型、扫描器列表、规则源、忽略规则和输出格式参数。
2. WHEN 用户选择 `vuln`、`sbom`、`misconfig`、`secret` 或 `license` 扫描器，THE system SHALL 返回对应类型的结构化结果或明确的能力错误。
3. WHEN 同一脚本在解释器、Runner 和 AOT 中执行，THE system SHALL 使用一致的操作名称、参数名称和结果字段。
4. IF 扫描器或目标类型尚未实现，THE system SHALL 在执行前返回可定位到名称的错误。
5. WHEN 用户选择 JSON 输出，THE system SHALL 返回稳定的 JSON 字段结构并包含扫描目标、扫描时间、扫描器和结果状态。

### Requirement 5: 离线与远程执行

**User Story:** AS 无 Python 和 Shell 环境的运维人员，我希望在目标主机直接执行安全扫描，以便满足受限环境的运行要求。

#### Acceptance Criteria

1. THE system SHALL provide scanner execution in a statically linked Go binary with `CGO_ENABLED=0`.
2. WHEN Runner 在远程目标执行主机扫描时，THE system SHALL 通过结构化指令传递扫描参数并通过 JSON 返回结果。
3. WHEN 漏洞规则源未在目标主机缓存时，THE system SHALL allow the controller to provide a signed or content-addressed rule bundle.
4. IF 规则源校验失败、规则包损坏或规则版本不兼容，THE system SHALL stop the affected scan and return a security error.
5. THE system SHALL keep vulnerability matching and SBOM generation local to supplied data unless the user explicitly enables a rule update operation.

### Requirement 6: Trivy 兼容能力边界

**User Story:** AS 熟悉 Trivy 的用户，我希望 OpsLang 的扫描结果具有可迁移性，以便接入现有安全流水线。

#### Acceptance Criteria

1. THE system SHALL document the mapping between OpsLang scan operations and Trivy target/scanner concepts.
2. THE system SHALL document supported targets and scanners separately from planned targets and scanners.
3. WHEN exporting vulnerability or SBOM results, THE system SHALL provide field mappings for Trivy-compatible JSON consumers.
4. THE system SHALL preserve OpsLang-specific host, task, audit and execution metadata alongside compatible security fields.
5. IF a Trivy feature requires a runtime dependency or behavior outside OpsLang's security model, THE system SHALL expose an explicit compatibility limitation.

### Requirement 7: 安全、性能和资源保护

**User Story:** AS 平台管理员，我希望大规模扫描具备边界和审计能力，以便控制资源消耗并追踪安全结果来源。

#### Acceptance Criteria

1. WHEN 扫描超过配置的文件数、文件大小、目录深度或执行时间限制，THE system SHALL stop the affected target and return a limit result。
2. WHEN 多台主机并行扫描，THE system SHALL obey the existing deployment concurrency limit and aggregate results by target.
3. WHEN a scan completes, THE system SHALL include rule source identity, rule version, scanner list, target identity and execution timestamps in the audit record.
4. IF a scan returns findings at or above the configured severity threshold, THE system SHALL expose a non-success status suitable for CI or deployment gating.
5. THE system SHALL provide tests for empty targets, malformed rules, oversized files, permission errors, duplicate components, cancellation and partial results.

## Initial Scope Proposal

第一批实现建议包含：

1. `vulnerability.scan(inventory, rules, options)`，扩展现有本地漏洞匹配能力。
2. `software.sbom(inventory, format)`，生成 OpsLang 原生 JSON，并提供 CycloneDX/SPDX 字段映射。
3. `file.scan(path, scanners, options)` 的目录清单识别基础层，先支持常见 Linux/Go/Node/Python 清单文件。
4. 统一 `ScanResult`、`VulnerabilityFinding`、`SBOMComponent`、`ScanError` 数据结构。
5. 解释器、Runner、AOT、CLI `opsctl scan` 的一致性接入。

容器镜像、Kubernetes、IaC、密钥和许可证扫描先完成接口和能力边界设计，再按独立任务实现。

## References

1. Trivy GitHub repository: https://github.com/aquasecurity/trivy
2. Trivy documentation: https://trivy.dev/docs/latest/
3. OpsLang existing vulnerability SDK: `pkg/ops-core-sdk/vulnerability`
4. OpsLang operation source of truth: `internal/opsspec/spec.go`
