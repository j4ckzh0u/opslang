#!/usr/bin/env python3
"""Generate remote test report from individual scenario result files."""

import json
import os
import sys
from datetime import datetime, timezone


def load_result(path):
    """Load a single scenario result file."""
    with open(path, "r") as f:
        return json.load(f)


def parse_duration_ms(started_at, finished_at):
    """Calculate duration in milliseconds between two ISO timestamps."""
    try:
        start = datetime.fromisoformat(started_at.replace("Z", "+00:00"))
        finish = datetime.fromisoformat(finished_at.replace("Z", "+00:00"))
        return int((finish - start).total_seconds() * 1000)
    except Exception:
        return 0


def main():
    results_dir = "/Users/jackzhou/_code/AI+/OpsLang/tests/remote/results"
    report_path = "/Users/jackzhou/_code/AI+/OpsLang/tests/reports/remote_results.json"

    # Server definitions
    servers = [
        {"name": "host1", "host": "192.168.1.188", "user": "jackzhou"},
        {"name": "host2", "host": "192.188.1.193", "user": "root"},
        {"name": "host3", "host": "192.168.1.151", "user": "openclaw"},
    ]

    # Scenario ID to name mapping
    scenario_names = {
        "R01": "cpu_info",
        "R02": "memory_info",
        "R03": "disk_info",
        "R04": "hostname_load",
        "R05": "net_interfaces",
        "R06": "process_list",
        "R07": "file_ops",
        "R08": "dns_check",
        "R09": "host_info",
        "R10": "time_ops",
        "R11": "process_exec",
        "R12": "comprehensive",
    }

    # Collect all result files
    result_files = sorted(
        [f for f in os.listdir(results_dir) if f.startswith("R") and f.endswith(".json")]
    )

    # Determine server connection status from all results
    server_connections = {s["name"]: {"connected": False, "error": ""} for s in servers}

    scenarios = []
    total_successful = 0
    total_failed = 0

    for rf in result_files:
        scenario_id = rf.split("_")[0]  # e.g., "R01" from "R01_cpu_info.json"
        scenario_name = scenario_names.get(scenario_id, rf.replace(".json", ""))

        data = load_result(os.path.join(results_dir, rf))
        results = data.get("results", {})

        scenario_entry = {
            "id": scenario_id,
            "name": scenario_name,
            "overall_status": data.get("status", "unknown"),
            "hosts": {},
        }

        for server in servers:
            sname = server["name"]
            host_result = results.get(sname, {})
            status = host_result.get("status", "failed")
            error = host_result.get("error", "")
            exit_code = host_result.get("exit_code", -1)

            duration_ms = parse_duration_ms(
                host_result.get("started_at", ""),
                host_result.get("finished_at", ""),
            )

            scenario_entry["hosts"][sname] = {
                "status": status,
                "exit_code": exit_code,
                "duration_ms": duration_ms,
                "error": error,
            }

            # Track connection status
            if "failed to connect" not in error and "Exec format error" not in error:
                server_connections[sname]["connected"] = True
                if not server_connections[sname]["error"]:
                    server_connections[sname]["error"] = ""
            elif "failed to connect" in error:
                server_connections[sname]["error"] = error

            if status == "success":
                total_successful += 1
            else:
                total_failed += 1

        scenarios.append(scenario_entry)

    # Build server connection summary
    server_summary = []
    connection_success = 0
    connection_failed = 0
    for server in servers:
        sname = server["name"]
        connected = server_connections[sname]["connected"]
        error = server_connections[sname]["error"]
        if connected:
            connection_success += 1
            server_summary.append(
                {
                    "name": sname,
                    "host": server["host"],
                    "user": server["user"],
                    "connection": "success",
                    "error": "",
                }
            )
        else:
            connection_failed += 1
            server_summary.append(
                {
                    "name": sname,
                    "host": server["host"],
                    "user": server["user"],
                    "connection": "failed",
                    "error": error,
                }
            )

    # Check if runner was cached (it should be after first successful run)
    cache_dir = os.path.expanduser("~/.cache/opslang/runners")
    runner_cached = os.path.exists(os.path.join(cache_dir, "ops-runner-linux-amd64"))

    # Build report
    report = {
        "test_date": datetime.now(timezone.utc).isoformat(),
        "servers": server_summary,
        "scenarios": scenarios,
        "summary": {
            "total_scenarios": len(scenarios),
            "total_host_tests": total_successful + total_failed,
            "successful": total_successful,
            "failed": total_failed,
            "connection_success": connection_success,
            "connection_failed": connection_failed,
            "runner_cached": runner_cached,
        },
    }

    # Write report
    os.makedirs(os.path.dirname(report_path), exist_ok=True)
    with open(report_path, "w") as f:
        json.dump(report, f, indent=2)

    print(f"Report written to {report_path}")
    print(f"  Scenarios: {len(scenarios)}")
    print(f"  Successful: {total_successful}/{total_successful + total_failed}")
    print(f"  Connections: {connection_success}/3 servers reachable")
    print(f"  Runner cached: {runner_cached}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
