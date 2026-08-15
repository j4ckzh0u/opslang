#!/bin/bash
# OpsLang Integration Test Runner
# Executes all scenarios in tests/scenarios/*.ops and produces JSON reports

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OPSCTL="$PROJECT_ROOT/bin/opsctl"
SCENARIOS_DIR="$SCRIPT_DIR/scenarios"
REPORTS_DIR="$SCRIPT_DIR/reports"
TIMEOUT_SEC=60

mkdir -p "$REPORTS_DIR"

# Clean any previous leftover temp files
rm -f /tmp/opslang-test-* 2>/dev/null || true
rm -rf /tmp/opslang-test-B /tmp/opslang-test-C /tmp/opslang-test-F /tmp/opslang-test-H /tmp/opslang-test-I 2>/dev/null || true

TEST_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [ ! -x "$OPSCTL" ]; then
    echo "ERROR: opsctl binary not found at $OPSCTL"
    exit 1
fi

# Collect scenario files sorted into array
SCENARIO_FILES=()
while IFS= read -r f; do
    SCENARIO_FILES+=("$f")
done < <(ls "$SCENARIOS_DIR"/*.ops 2>/dev/null | sort)

TOTAL=${#SCENARIO_FILES[@]}
echo "Found $TOTAL scenario files"
echo "Test date: $TEST_DATE"
echo "---"

# Arrays to hold results
declare -a RESULTS_JSON=()
PASSED=0
FAILED=0
ERRORED=0
SKIPPED=0
TOTAL_DURATION=0
MIN_DURATION=999999
MAX_DURATION=0
SLOWEST_NAME=""
SLOWEST_MS=0
FASTEST_NAME=""
FASTEST_MS=999999

# Per-batch accumulators
declare -A BATCH_TOTAL
declare -A BATCH_PASS
declare -A BATCH_FAIL
declare -A BATCH_ERR
declare -A BATCH_DUR

for f in "${SCENARIO_FILES[@]}"; do
    basename=$(basename "$f")
    scenario="${basename%%_*}"         # e.g. A1 from A1_cpu_info.ops
    name_part="${basename%.ops}"       # e.g. A1_cpu_info
    name_part="${name_part#*_}"        # e.g. cpu_info

    batch_letter="${scenario:0:1}"

    # Init batch counters
    BATCH_TOTAL[$batch_letter]=$(( ${BATCH_TOTAL[$batch_letter]:-0} + 1 ))
    BATCH_PASS[$batch_letter]=${BATCH_PASS[$batch_letter]:-0}
    BATCH_FAIL[$batch_letter]=${BATCH_FAIL[$batch_letter]:-0}
    BATCH_ERR[$batch_letter]=${BATCH_ERR[$batch_letter]:-0}
    BATCH_DUR[$batch_letter]=${BATCH_DUR[$batch_letter]:-0}

    output_file="$REPORTS_DIR/${scenario}.txt"
    printf "[%s] %-40s " "$scenario" "$name_part"

    start_ts=$(python3 -c 'import time; print(int(time.time()*1000))')

    # Run with timeout, capture stdout+stderr
    set +e
    timeout_output=$(timeout "$TIMEOUT_SEC" "$OPSCTL" run "$f" 2>&1)
    exit_code=$?
    set -e

    end_ts=$(python3 -c 'import time; print(int(time.time()*1000))')

    duration_ms=$(( end_ts - start_ts ))

    # Save output
    echo "$timeout_output" > "$output_file"

    # Determine status
    status="fail"
    error_msg=""
    notes=""

    # timeout exit code is 124
    if [ $exit_code -eq 124 ]; then
        status="error"
        error_msg="timeout after ${TIMEOUT_SEC}s"
        notes="TIMEOUT"
        ERRORED=$((ERRORED + 1))
        BATCH_ERR[$batch_letter]=$(( ${BATCH_ERR[$batch_letter]} + 1 ))
    elif [ $exit_code -gt 128 ]; then
        status="error"
        error_msg="signal $((exit_code - 128))"
        notes="CRASH"
        ERRORED=$((ERRORED + 1))
        BATCH_ERR[$batch_letter]=$(( ${BATCH_ERR[$batch_letter]} + 1 ))
    elif [ $exit_code -eq 0 ]; then
        # Check output contains scenario name (e.g. "A1" in A1's output)
        if grep -q "$scenario" <<< "$timeout_output" 2>/dev/null; then
            status="pass"
            PASSED=$((PASSED + 1))
            BATCH_PASS[$batch_letter]=$(( ${BATCH_PASS[$batch_letter]} + 1 ))
        else
            status="fail"
            error_msg="exit 0 but scenario name '$scenario' not in output"
            notes="NO_MATCH"
            FAILED=$((FAILED + 1))
            BATCH_FAIL[$batch_letter]=$(( ${BATCH_FAIL[$batch_letter]} + 1 ))
        fi
    else
        status="fail"
        # Take first line of output as error
        error_msg=$(head -n 1 <<< "$timeout_output" | tr -d '\n' | cut -c 1-120)
        notes="exit=$exit_code"
        FAILED=$((FAILED + 1))
        BATCH_FAIL[$batch_letter]=$(( ${BATCH_FAIL[$batch_letter]} + 1 ))
    fi

    # Escape error_msg for JSON
    error_msg_escaped=$(printf '%s' "$error_msg" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read())[1:-1])')

    # Stats
    TOTAL_DURATION=$((TOTAL_DURATION + duration_ms))
    BATCH_DUR[$batch_letter]=$(( ${BATCH_DUR[$batch_letter]} + duration_ms ))
    if [ $duration_ms -lt $MIN_DURATION ]; then
        MIN_DURATION=$duration_ms
        FASTEST_NAME=$scenario
    fi
    if [ $duration_ms -gt $MAX_DURATION ]; then
        MAX_DURATION=$duration_ms
        SLOWEST_NAME=$scenario
        SLOWEST_MS=$duration_ms
    fi
    if [ $duration_ms -gt $SLOWEST_MS ]; then
        SLOWEST_MS=$duration_ms
    fi
    if [ $duration_ms -lt $FASTEST_MS ]; then
        FASTEST_MS=$duration_ms
    fi

    # Print progress
    case "$status" in
        pass) printf "PASS  %6dms\n" "$duration_ms" ;;
        fail) printf "FAIL  %6dms  %s\n" "$duration_ms" "$error_msg" ;;
        error) printf "ERROR %6dms  %s\n" "$duration_ms" "$error_msg" ;;
    esac

    # Build JSON entry
    RESULTS_JSON+=("{\"scenario\":\"$scenario\",\"name\":\"$name_part\",\"status\":\"$status\",\"exit_code\":$exit_code,\"duration_ms\":$duration_ms,\"output_file\":\"${scenario}.txt\",\"error\":\"$error_msg_escaped\"}")
done

echo "---"
echo "Total: $TOTAL | Passed: $PASSED | Failed: $FAILED | Errored: $ERRORED | Skipped: $SKIPPED"
echo "Avg duration: $(( TOTAL_DURATION / TOTAL ))ms | Slowest: $SLOWEST_NAME (${SLOWEST_MS}ms)"

# Build results.json
{
    echo "{"
    echo "  \"test_date\": \"$TEST_DATE\","
    echo "  \"total\": $TOTAL,"
    echo "  \"passed\": $PASSED,"
    echo "  \"failed\": $FAILED,"
    echo "  \"errored\": $ERRORED,"
    echo "  \"skipped\": $SKIPPED,"
    echo "  \"results\": ["
    for i in "${!RESULTS_JSON[@]}"; do
        if [ $i -gt 0 ]; then echo "    ,"; fi
        echo "    ${RESULTS_JSON[$i]}"
    done
    echo "  ]"
    echo "}"
} > "$REPORTS_DIR/results.json"

# Build summary.json with batch stats
{
    echo "{"
    echo "  \"test_date\": \"$TEST_DATE\","
    echo "  \"total\": $TOTAL,"
    echo "  \"passed\": $PASSED,"
    echo "  \"failed\": $FAILED,"
    echo "  \"errored\": $ERRORED,"
    echo "  \"skipped\": $SKIPPED,"
    echo "  \"average_duration_ms\": $(( TOTAL_DURATION / TOTAL )),"
    echo "  \"min_duration_ms\": $MIN_DURATION,"
    echo "  \"max_duration_ms\": $MAX_DURATION,"
    echo "  \"slowest_scenario\": \"$SLOWEST_NAME\","
    echo "  \"fastest_scenario\": \"$FASTEST_NAME\","
    echo "  \"batch_stats\": {"
    batch_keys=($(echo "${!BATCH_TOTAL[@]}" | tr ' ' '\n' | sort))
    for i in "${!batch_keys[@]}"; do
        k=${batch_keys[$i]}
        if [ $i -gt 0 ]; then echo "    ,"; fi
        avg=0
        if [ ${BATCH_TOTAL[$k]} -gt 0 ]; then
            avg=$(( ${BATCH_DUR[$k]} / ${BATCH_TOTAL[$k]} ))
        fi
        echo "    \"$k\": {\"total\": ${BATCH_TOTAL[$k]}, \"passed\": ${BATCH_PASS[$k]}, \"failed\": ${BATCH_FAIL[$k]}, \"errored\": ${BATCH_ERR[$k]}, \"avg_duration_ms\": $avg}"
    done
    echo "  }"
    echo "}"
} > "$REPORTS_DIR/summary.json"

echo ""
echo "Results written to:"
echo "  $REPORTS_DIR/results.json"
echo "  $REPORTS_DIR/summary.json"
echo "  $REPORTS_DIR/<scenario>.txt (per-scenario output)"
