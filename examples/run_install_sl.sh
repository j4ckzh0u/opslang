#!/bin/bash
# run_install_sl.sh - 执行 install_sl.ops 并生成汇总报告
# 用法: ./run_install_sl.sh host1 host2 host3 ...

set -euo pipefail

if [ $# -eq 0 ]; then
    echo "用法: $0 host1 host2 host3 ..."
    echo "示例: $0 user@192.168.1.10 user@192.168.1.11"
    exit 1
fi

HOSTS=$(IFS=,; echo "$*")
SCRIPT="examples/install_sl.ops"
REPORT_DIR="reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${REPORT_DIR}/install_sl_${TIMESTAMP}.json"
SUMMARY_FILE="${REPORT_DIR}/install_sl_${TIMESTAMP}_summary.md"

# 创建报告目录
mkdir -p "$REPORT_DIR"

echo "========================================"
echo "批量安装 sl"
echo "========================================"
echo "目标主机: $HOSTS"
echo "开始时间: $(date -Iseconds)"
echo ""

# 执行 OpsLang 脚本
echo "执行脚本: $SCRIPT"
./opsctl deploy --hosts "$HOSTS" "$SCRIPT" > "$REPORT_FILE" 2>&1

EXIT_CODE=$?

echo ""
echo "执行完成，退出码: $EXIT_CODE"
echo "详细报告: $REPORT_FILE"
echo ""

# 生成 Markdown 汇总报告
cat > "$SUMMARY_FILE" << EOF
# sl 安装报告

**执行时间**: $(date -Iseconds)
**目标主机**: $HOSTS
**退出码**: $EXIT_CODE

## 执行结果

\`\`\`json
$(cat "$REPORT_FILE" | python3 -m json.tool 2>/dev/null || cat "$REPORT_FILE")
\`\`\`

## 统计

EOF

# 解析 JSON 并生成统计
if command -v python3 >/dev/null 2>&1; then
    python3 << PYEOF >> "$SUMMARY_FILE"
import json
import sys

try:
    with open('$REPORT_FILE') as f:
        data = json.load(f)

    results = data.get('results', {})
    total = len(results)
    success = sum(1 for r in results.values() if r.get('status') == 'success')
    skipped = sum(1 for r in results.values() if r.get('status') == 'skipped')
    failed = sum(1 for r in results.values() if r.get('status') == 'failed')

    print(f"- **总主机数**: {total}")
    print(f"- **成功**: {success}")
    print(f"- **跳过（已安装）**: {skipped}")
    print(f"- **失败**: {failed}")
    print()

    if failed > 0:
        print("## 失败详情")
        print()
        for host, result in results.items():
            if result.get('status') == 'failed':
                print(f"### {host}")
                print(f"- 错误: {result.get('error', 'unknown')}")
                print(f"- 包管理器: {result.get('package_manager', 'unknown')}")
                print()

    # 按操作系统统计
    os_counts = {}
    for result in results.values():
        os_name = result.get('os', 'unknown')
        os_counts[os_name] = os_counts.get(os_name, 0) + 1

    if os_counts:
        print("## 操作系统分布")
        print()
        for os_name, count in sorted(os_counts.items()):
            print(f"- {os_name}: {count}")
        print()

    # 平均安装耗时
    durations = [r.get('duration_ms', 0) for r in results.values() if r.get('status') == 'success']
    if durations:
        avg_duration = sum(durations) / len(durations)
        print(f"## 性能")
        print()
        print(f"- **平均安装耗时**: {avg_duration:.0f}ms")
        print(f"- **最快**: {min(durations)}ms")
        print(f"- **最慢**: {max(durations)}ms")
        print()

except Exception as e:
    print(f"解析报告失败: {e}", file=sys.stderr)
    sys.exit(1)
PYEOF
else
    echo "_（需要 python3 生成统计）_" >> "$SUMMARY_FILE"
fi

echo "汇总报告: $SUMMARY_FILE"
echo ""
echo "========================================"
echo "完成"
echo "========================================"

exit $EXIT_CODE
