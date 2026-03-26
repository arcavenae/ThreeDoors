#!/usr/bin/env bash
set -euo pipefail

# qa-coverage-audit.sh — QA Coverage Audit
# Runs go test -cover, compares against baseline, flags regressions >5pp.
#
# Usage:
#   ./scripts/qa-coverage-audit.sh [--baseline PATH] [--threshold 5] [--update]
#
# Designed to run weekly via system cron:
#   0 9 * * 1 cd /path/to/ThreeDoors && ./scripts/qa-coverage-audit.sh --update --report
# Or via multiclaude:
#   multiclaude work "Run QA coverage audit" --repo ThreeDoors

BASELINE_FILE="${BASELINE_FILE:-docs/quality/coverage-baseline.json}"
THRESHOLD="${THRESHOLD:-5}"
UPDATE_BASELINE="${UPDATE_BASELINE:-false}"
REPORT_TO_SUPERVISOR="${REPORT_TO_SUPERVISOR:-false}"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --baseline) BASELINE_FILE="$2"; shift 2 ;;
        --threshold) THRESHOLD="$2"; shift 2 ;;
        --update) UPDATE_BASELINE="true"; shift ;;
        --report) REPORT_TO_SUPERVISOR="true"; shift ;;
        --help|-h)
            echo "Usage: qa-coverage-audit.sh [--baseline PATH] [--threshold N] [--update] [--report]"
            echo ""
            echo "Options:"
            echo "  --baseline   Path to coverage baseline JSON (default: docs/quality/coverage-baseline.json)"
            echo "  --threshold  Coverage drop threshold in percentage points (default: 5)"
            echo "  --update     Update baseline file after reporting"
            echo "  --report     Send report to supervisor via multiclaude message send"
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Check for required tools
for cmd in go jq; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: $cmd is required but not installed." >&2
        exit 1
    fi
done

TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
TODAY="$(date -u +%Y-%m-%d)"
REGRESSIONS_FOUND=0
REPORT=""

add_to_report() {
    REPORT="${REPORT}${1}\n"
}

add_to_report "# QA Coverage Audit Report"
add_to_report "Timestamp: ${TIMESTAMP}"
add_to_report ""

# --- 1. Run go test -cover and capture output ---
add_to_report "## Running Coverage Tests"

COVERAGE_OUTPUT="$(go test -cover ./... 2>&1)" || true

# Parse coverage output into JSON
# Format: "ok  github.com/arcavenae/ThreeDoors/internal/tasks  1.234s  coverage: 78.5% of statements"
CURRENT_COVERAGE="{}"
while IFS= read -r line; do
    if echo "$line" | grep -q "coverage:.*% of statements"; then
        PKG="$(echo "$line" | awk '{print $2}')"
        PCT="$(echo "$line" | grep -oE '[0-9]+\.[0-9]+%' | tr -d '%')"
        if [[ -n "$PKG" && -n "$PCT" ]]; then
            CURRENT_COVERAGE="$(echo "$CURRENT_COVERAGE" | jq --arg pkg "$PKG" --arg pct "$PCT" --arg date "$TODAY" \
                '. + {($pkg): {"coverage": ($pct | tonumber), "updated": $date}}')"
        fi
    elif echo "$line" | grep -q "coverage: \[no statements\]"; then
        PKG="$(echo "$line" | awk '{print $2}')"
        if [[ -n "$PKG" ]]; then
            CURRENT_COVERAGE="$(echo "$CURRENT_COVERAGE" | jq --arg pkg "$PKG" --arg date "$TODAY" \
                '. + {($pkg): {"coverage": 0.0, "updated": $date, "note": "no statements"}}')"
        fi
    fi
done <<< "$COVERAGE_OUTPUT"

PKG_COUNT="$(echo "$CURRENT_COVERAGE" | jq 'length')"
add_to_report "Packages tested: ${PKG_COUNT}"
add_to_report ""

# --- 2. Compare against baseline ---
add_to_report "## Coverage Comparison"

FIRST_RUN=false
if [[ ! -f "$BASELINE_FILE" ]]; then
    FIRST_RUN=true
    add_to_report "No baseline found at ${BASELINE_FILE}. Establishing initial baseline."
    add_to_report "No regressions possible on first run."
else
    BASELINE_PACKAGES="$(jq '.packages // {}' "$BASELINE_FILE")"

    # Compare each package
    for pkg in $(echo "$CURRENT_COVERAGE" | jq -r 'keys[]'); do
        CURRENT_PCT="$(echo "$CURRENT_COVERAGE" | jq -r --arg pkg "$pkg" '.[$pkg].coverage')"
        BASELINE_PCT="$(echo "$BASELINE_PACKAGES" | jq -r --arg pkg "$pkg" '.[$pkg].coverage // empty')"

        if [[ -z "$BASELINE_PCT" ]]; then
            add_to_report "NEW: ${pkg} — ${CURRENT_PCT}% (no baseline)"
            continue
        fi

        # Calculate drop using awk for floating-point arithmetic
        DROP="$(awk "BEGIN {printf \"%.1f\", ${BASELINE_PCT} - ${CURRENT_PCT}}")"
        IS_REGRESSION="$(awk "BEGIN {print (${DROP} > ${THRESHOLD}) ? 1 : 0}")"

        if [[ "$IS_REGRESSION" -eq 1 ]]; then
            REGRESSIONS_FOUND=1
            add_to_report "REGRESSION: ${pkg} — ${BASELINE_PCT}% -> ${CURRENT_PCT}% (dropped ${DROP}pp)"
        fi
    done

    # Check for removed packages
    for pkg in $(echo "$BASELINE_PACKAGES" | jq -r 'keys[]'); do
        HAS_CURRENT="$(echo "$CURRENT_COVERAGE" | jq -r --arg pkg "$pkg" 'has($pkg)')"
        if [[ "$HAS_CURRENT" == "false" ]]; then
            add_to_report "REMOVED: ${pkg} — was in baseline but not in current run"
        fi
    done

    if [[ "$REGRESSIONS_FOUND" -eq 0 ]]; then
        add_to_report "No regressions detected (threshold: >${THRESHOLD}pp drop)."
    fi
fi

add_to_report ""

# --- 3. Per-package summary ---
add_to_report "## Current Coverage"

COVERAGE_LINES="$(echo "$CURRENT_COVERAGE" | jq -r 'to_entries | sort_by(.key) | .[] | "  \(.key): \(.value.coverage)%"')"
add_to_report "$COVERAGE_LINES"

add_to_report ""

# --- 4. Summary ---
add_to_report "## Summary"

if [[ "$REGRESSIONS_FOUND" -eq 1 ]]; then
    add_to_report "Coverage regressions detected — review items above."
elif [[ "$FIRST_RUN" == "true" ]]; then
    add_to_report "Initial baseline established. Future runs will compare against this baseline."
else
    add_to_report "All clear — coverage is within acceptable thresholds."
fi

# Output report
echo -e "$REPORT"

# --- 5. Update baseline if requested ---
if [[ "$UPDATE_BASELINE" == "true" ]]; then
    NEW_BASELINE="$(jq -n --arg date "$TODAY" --argjson pkgs "$CURRENT_COVERAGE" \
        '{"updated": $date, "packages": $pkgs}')"
    echo "$NEW_BASELINE" | jq '.' > "$BASELINE_FILE"
    echo "Baseline updated at ${BASELINE_FILE}"
fi

# Optionally send to supervisor
if [[ "$REPORT_TO_SUPERVISOR" == "true" ]] && command -v multiclaude &>/dev/null; then
    echo -e "$REPORT" | multiclaude message send --to supervisor --stdin 2>/dev/null || \
        echo "Warning: Could not send report to supervisor." >&2
fi

exit 0
