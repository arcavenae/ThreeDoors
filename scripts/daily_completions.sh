#!/usr/bin/env bash
set -euo pipefail

# daily_completions.sh - Daily completion counts from completed.txt
# Usage: ./scripts/daily_completions.sh [path/to/completed.txt]
# Format expected: [YYYY-MM-DD HH:MM:SS] task_id | task_text

COMPLETED_FILE="${1:-${HOME}/.threedoors/completed.txt}"

# Check file exists
if [[ ! -f "$COMPLETED_FILE" ]]; then
    echo "No data found: $COMPLETED_FILE does not exist."
    exit 0
fi

# Check file not empty
if [[ ! -s "$COMPLETED_FILE" ]]; then
    echo "No data found: $COMPLETED_FILE is empty."
    exit 0
fi

echo "Daily Completion Report"
echo "======================="
printf "%-12s | %-9s | %s\n" "Date" "Completed" "Cumulative"

CUMULATIVE=0
# Extract dates from format: [YYYY-MM-DD HH:MM:SS] task_id | task_text
# Use sed for portability (macOS grep doesn't support -P)
sed -n 's/^\[\([0-9]\{4\}-[0-9]\{2\}-[0-9]\{2\}\).*/\1/p' "$COMPLETED_FILE" | sort | uniq -c | while read -r COUNT DATE; do
    CUMULATIVE=$((CUMULATIVE + COUNT))
    printf "%-12s | %-9s | %s\n" "$DATE" "$COUNT" "$CUMULATIVE"
done
