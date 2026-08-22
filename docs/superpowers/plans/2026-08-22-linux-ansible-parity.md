# Linux Ansible Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Make the first Linux-focused Ansible-style operations genuinely executable, idempotent, documented, and represented by honest examples.

**Architecture:** Keep `internal/opsspec` as the canonical operation table and register every operation in interpreter, Runner, and AOT codegen. SDK functions perform direct Go/system API work where possible and return structured results; Linux-only operations fail explicitly on unsupported platforms. Examples perform real checks and mutations only when prerequisites are present.

**Tech Stack:** Go 1.26, standard library, existing SDK packages, OpsLang interpreter/Runner/AOT, Markdown documentation.

---

### Task 1: Establish Linux operation contracts

**Files:** `internal/opsspec/spec.go`, `pkg/ops-core-sdk/user`, `pkg/ops-core-sdk/service`, `pkg/ops-core-sdk/pkg`

- [x] Define canonical names, argument names, mutating flags, and Linux availability for user, service, and package operations.
- [x] Add structured result types with `changed`, `present`, `running`, `enabled`, and error details.
- [x] Add unit tests for empty arguments, unsupported platform, and idempotent no-op results.

### Task 2: Wire all engines

**Files:** `internal/interpreter/sdk_bridge.go`, `internal/runner/registry.go`, `internal/compiler/codegen.go`

- [x] Register each operation in all three engines using identical argument order.
- [x] Run consistency tests and add regression coverage for the new names.

### Task 3: Replace dishonest examples

**Files:** `examples/*.ops`, `cmd/opsctl/examples_e2e_test.go`

- [x] Remove “simulation” and “伪代码” claims from runnable examples.
- [x] Add prerequisite checks for Linux, root, systemd, package manager, and network.
- [x] Make examples either execute real operations or exit with an explicit skipped/prerequisite message; never print a fake success.

### Task 4: Update documentation

**Files:** `README.md`, `docs/stdlib-reference.md`, `docs/examples.md`, `docs/getting-started.md`

- [x] Document platform, privilege, idempotency, dry-run, failure, and verification behavior for every new operation.
- [x] Add exact Linux commands and expected structured output, clearly marking values that vary by host.
- [x] Add a “real execution policy” section explaining why examples skip instead of simulating unavailable infrastructure.

### Task 5: Verify and commit

- [x] Run `gofmt`, `go test ./...`, `go vet ./...`, and Linux `CGO_ENABLED=0` builds.
- [x] Run representative examples on the current host and record skipped prerequisites honestly.
- [x] Commit implementation and documentation separately from any unrelated work.
