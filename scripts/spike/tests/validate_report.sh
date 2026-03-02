#!/bin/bash
# validate_report.sh - Validates spike report completeness
# Exit 0 if all sections present, non-zero otherwise

set -euo pipefail

REPORT="docs/spike-reports/2.2-apple-notes-integration.md"

if [ ! -f "$REPORT" ]; then
    echo "[FAIL] Spike report not found: $REPORT" >&2
    exit 1
fi

ERRORS=0

check_section() {
    local pattern="$1"
    local desc="$2"
    if grep -qi "$pattern" "$REPORT"; then
        echo "[OK] Found: $desc"
    else
        echo "[FAIL] Missing: $desc" >&2
        ERRORS=$((ERRORS + 1))
    fi
}

echo "Validating spike report: $REPORT"
echo "---"

check_section "executive summary" "Executive Summary"
check_section "comparison matrix" "Comparison Matrix"
check_section "recommendation" "Recommendation with Rationale"
check_section "error taxonomy" "Error Taxonomy"
check_section "ci/cd\|ci compatibility\|github actions" "CI/CD Compatibility"
check_section "performance\|benchmark\|latency" "Performance Benchmarks"
check_section "risk" "Risks and Mitigations"

# Check comparison matrix has no unfilled entries
if grep -q "| ? |" "$REPORT"; then
    echo "[FAIL] Comparison matrix has unfilled entries (? remaining)" >&2
    ERRORS=$((ERRORS + 1))
else
    echo "[OK] Comparison matrix fully filled"
fi

# Check all 3 approaches are mentioned
for approach in "AppleScript\|osascript" "SQLite" "MCP"; do
    if grep -qi "$approach" "$REPORT"; then
        echo "[OK] Approach evaluated: $approach"
    else
        echo "[FAIL] Approach missing: $approach" >&2
        ERRORS=$((ERRORS + 1))
    fi
done

echo "---"
if [ "$ERRORS" -eq 0 ]; then
    echo "[PASS] All validation checks passed"
    exit 0
else
    echo "[FAIL] $ERRORS validation check(s) failed" >&2
    exit 1
fi
