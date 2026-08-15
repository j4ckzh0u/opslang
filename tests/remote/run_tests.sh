#!/bin/bash
# Run remaining remote test scenarios with per-scenario timeout
set -e

cd /Users/jackzhou/_code/AI+/OpsLang

SCENARIOS=(
  "R03_disk_info"
  "R04_hostname_load"
  "R05_net_interfaces"
  "R06_process_list"
  "R07_file_ops"
  "R08_dns_check"
  "R09_host_info"
  "R10_time_ops"
  "R11_process_exec"
  "R12_comprehensive"
)

INSTR_DIR="tests/remote/instructions"
RESULT_DIR="tests/remote/results"

for base in "${SCENARIOS[@]}"; do
  infile="${INSTR_DIR}/${base}.json"
  outfile="${RESULT_DIR}/${base}.json"

  if [ -f "$outfile" ]; then
    echo "SKIP $base (already exists)"
    continue
  fi

  echo "=== Running $base ==="
  # Run with 5-minute timeout per scenario
  gtimeout 300 ./bin/opsctl exec \
    --inventory tests/remote/inventory.yaml \
    --instructions "$infile" \
    -o "$outfile" 2>&1 || true
  echo "=== Done $base ==="
  echo ""
done

echo "ALL SCENARIOS COMPLETE"
