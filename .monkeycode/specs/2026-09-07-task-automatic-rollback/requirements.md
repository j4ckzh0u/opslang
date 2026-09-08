# Requirements Document

## Introduction

本功能为 OpsLang 远程任务提供显式补偿语义。脚本作者在任务中声明失败处理动作，部署控制器根据远程主机的结构化结果自动触发补偿，并将主操作与补偿结果一并写入部署输出和审计日志。

## Glossary

- **主操作**：`task` 主体中按声明顺序执行的操作。
- **补偿块**：任务主体失败后执行的 `rescue` 块。
- **收尾块**：主操作和补偿处理结束后执行的 `always` 块。
- **失败主机**：主操作返回失败状态的目标主机。
- **补偿结果**：补偿块在单台目标主机上的状态、错误、时间和结构化数据。
- **部署控制器**：`opsctl deploy` 中负责目标选择、调度和结果聚合的组件。

## Requirements

### Requirement 1: Explicit Compensation Declaration

**User Story:** AS an operations engineer, I want to declare compensation actions next to a task, so that recovery intent remains reviewable with the change.

#### Acceptance Criteria

1. WHEN a script contains `task "name" on targets { ... } rescue { ... }`, the parser SHALL attach the compensation block to the task AST node.
2. WHEN a task contains an `always` clause, the parser SHALL attach the cleanup block to the task AST node.
3. IF `rescue` or `always` appears without a preceding task body, the parser SHALL return an error containing the source line and column.
4. WHEN the AST is rendered for diagnostics, the task representation SHALL indicate the presence of `rescue` and `always` clauses.

### Requirement 2: Host-Scoped Compensation

**User Story:** AS an operations engineer, I want failed task executions to trigger declared compensation, so that each affected host can return to a known state.

#### Acceptance Criteria

1. WHEN a task main package fails on a target host, the deployment controller SHALL execute the task compensation package on that failed host.
2. WHEN a task main package succeeds on a target host, the deployment controller SHALL preserve the successful main result.
3. WHEN a compensation package finishes, the deployment controller SHALL retain both the main result and compensation result for the target host.
4. IF compensation execution fails, the deployment controller SHALL classify the target result as `rollback_failed` and preserve both error chains.
5. WHEN any target host triggers compensation, the deployment controller SHALL stop before starting the next task.

### Requirement 3: Cleanup Execution

**User Story:** AS an operations engineer, I want cleanup actions to run after task handling, so that temporary resources receive deterministic treatment.

#### Acceptance Criteria

1. WHEN task handling reaches a terminal main or compensation state, the deployment controller SHALL execute the declared `always` package on the applicable host.
2. IF cleanup execution fails, the deployment controller SHALL preserve the preceding main and compensation results and record the cleanup failure separately.
3. WHEN no `always` clause exists, the deployment controller SHALL complete task handling after main or compensation execution.

### Requirement 4: Execution Engine Consistency

**User Story:** AS a script author, I want compensation behavior to remain consistent across supported execution modes, so that mode selection preserves task intent.

#### Acceptance Criteria

1. WHEN Runner mode receives a task with compensation, the instruction generator SHALL produce separately identifiable main, compensation, and cleanup packages.
2. WHEN AOT mode compiles a task with compensation, the generated binary SHALL preserve main, compensation, cleanup, and error propagation semantics.
3. WHEN automatic mode selects an engine, the mode selector SHALL choose an engine that can execute the declared task semantics.
4. IF a forced execution mode cannot represent the task semantics, `opsctl deploy` SHALL return a source-oriented validation error before contacting target hosts.

### Requirement 5: Privilege, Approval, and Signing

**User Story:** AS a security administrator, I want recovery actions governed by the same controls as primary changes, so that compensation remains auditable and authorized.

#### Acceptance Criteria

1. WHEN privilege validation evaluates a task, the validator SHALL inspect main, compensation, and cleanup blocks.
2. WHEN approval analysis identifies mutating compensation or cleanup operations for production targets, the approval summary SHALL include those operations.
3. WHEN package signing is enabled, the controller SHALL sign every main, compensation, and cleanup package after final configuration injection.
4. WHEN the Runner verifies a compensation or cleanup package, the Runner SHALL apply the same signature and privilege rules used for a main package.

### Requirement 6: Dry Run and Structured Output

**User Story:** AS an operations engineer, I want rollback plans visible before deployment and complete results after deployment, so that I can assess and audit recovery behavior.

#### Acceptance Criteria

1. WHILE dry-run mode is active, the deployment controller SHALL return previews for main, compensation, and cleanup packages without applying mutating operations.
2. WHEN a deployment result is serialized, each task and host SHALL expose main, compensation, and cleanup results as distinct fields.
3. WHEN audit logging records a deployment, the audit entry SHALL include the rollback trigger, execution order, package identifiers, and all terminal statuses.
4. IF context cancellation prevents compensation or cleanup execution, the deployment controller SHALL record a cancellation result with the preceding failure context.

### Requirement 7: Retry Boundary

**User Story:** AS an operations engineer, I want retries limited to safe transport preparation, so that mutating operations do not repeat unexpectedly.

#### Acceptance Criteria

1. WHEN connection, architecture detection, upload, or command setup encounters a retryable transport failure, the executor SHALL apply the existing bounded transport retry policy.
2. WHEN a main operation starts and returns a failure, the deployment controller SHALL proceed to compensation without replaying the main operation package.
3. WHEN compensation starts and returns a failure, the deployment controller SHALL record `rollback_failed` without replaying the compensation package.
