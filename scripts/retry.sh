#!/usr/bin/env bash
# retry.sh — run a command up to RETRY_MAX times; succeed if ANY attempt succeeds.
#
# Why this exists: the `internal/security` test binary (and occasionally other
# Go test binaries) intermittently dies with:
#   fatal error: runtime: cannot allocate memory
# at process startup/heap-growth, inside the Go 1.26 runtime's own mmap
# (mheap.allocSpan -> allocMSpanLocked -> fixalloc.alloc -> persistentalloc).
#
# This is NOT an OpsLang bug and NOT memory starvation:
#   - the host has 25+ GB free RAM and RSS peak is ~4.5 MB
#   - vm.max_map_count and overcommit are far from limits
#   - a hand-written C mmap probe (16KB x200k and 64MB x5000) succeeds 100%
#   - only Go's runtime mmap fails, ~45% of the time, reproducibly on both
#     the bare-metal 32GB host (Go 1.26.0) and CI (Go 1.26.7)
#
# The crash is a process-level abort, not a test-assertion failure, so a fresh
# `go test` process has an independent chance of passing. Retrying keeps CI
# green without masking real test failures (a genuine logic failure fails every
# attempt). Set RETRY_MAX to tune reliability vs. wall-clock cost.
set -o pipefail

n="${RETRY_MAX:-4}"
for i in $(seq 1 "$n"); do
  echo "=== attempt $i/$n ==="
  "$@"
  rc=$?
  if [ "$rc" -eq 0 ]; then
    echo "=== succeeded on attempt $i ==="
    exit 0
  fi
  echo "attempt $i failed (exit $rc), retrying"
done
echo "=== all $n attempts failed ==="
exit 1
