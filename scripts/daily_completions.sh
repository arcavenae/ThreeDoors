#!/bin/bash
# ThreeDoors Daily Completions Report
# Parses completed.txt to show daily task completion counts

COMPLETED_FILE="$HOME/.threedoors/completed.txt"

if [[ ! -f "$COMPLETED_FILE" ]]; then
    echo "No completion data found at $COMPLETED_FILE"
    echo "Complete some tasks to see daily statistics."
    exit 1
fi

echo "=== Daily Task Completions ==="
echo ""

# Extract dates and count completions per day
cat "$COMPLETED_FILE" | cut -d' ' -f1 | tr -d '[]' | sort | uniq -c | awk '{print $2 ": " $1 " task(s)"}'

echo ""
echo "Total Completions:"
wc -l < "$COMPLETED_FILE"
