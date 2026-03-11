#!/usr/bin/env bash
set -euo pipefail

# test-ci-metrics.sh — Tests for ci-metrics.sh
# Run from the repository root: ./scripts/test-ci-metrics.sh
#
# Tests script argument handling, output format, and edge cases.
# Live API tests require gh auth; structural tests work offline.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$SCRIPT_DIR/ci-metrics.sh"
PASS=0
FAIL=0

pass() {
    PASS=$((PASS + 1))
    echo "  PASS: $1"
}

fail() {
    FAIL=$((FAIL + 1))
    echo "  FAIL: $1" >&2
}

echo "=== CI Metrics Script Tests ==="

# Test 1: Help flag exits 0
echo "Test: --help exits successfully"
if "$SCRIPT" --help >/dev/null 2>&1; then
    pass "--help exits 0"
else
    fail "--help should exit 0"
fi

# Test 2: Help output contains usage info
echo "Test: --help shows usage information"
HELP_OUTPUT="$("$SCRIPT" --help 2>&1)"
if echo "$HELP_OUTPUT" | grep -q "Usage"; then
    pass "--help shows usage"
else
    fail "--help should show usage info"
fi

# Test 3: Invalid option fails
echo "Test: unknown option fails"
if "$SCRIPT" --bogus 2>/dev/null; then
    fail "unknown option should fail"
else
    pass "unknown option exits non-zero"
fi

# Test 4: Invalid --days value fails
echo "Test: --days 0 fails"
if "$SCRIPT" --days 0 2>/dev/null; then
    fail "--days 0 should fail"
else
    pass "--days 0 exits non-zero"
fi

# Test 5: Invalid --days non-numeric fails
echo "Test: --days abc fails"
if "$SCRIPT" --days abc 2>/dev/null; then
    fail "--days abc should fail"
else
    pass "--days abc exits non-zero"
fi

# --- Live API tests (require gh auth) ---
echo ""
echo "=== Live API Tests ==="

# Check gh auth
if ! gh auth status &>/dev/null; then
    echo "  SKIP: gh not authenticated — skipping live API tests"
    echo ""
    echo "=== Results ==="
    echo "Passed: $PASS"
    echo "Failed: $FAIL"
    echo "Total:  $((PASS + FAIL))"
    if [[ "$FAIL" -gt 0 ]]; then exit 1; fi
    echo "All offline tests passed (live tests skipped)."
    exit 0
fi

# Test 6: Human-readable output contains all required sections
echo "Test: human-readable output has required sections"
OUTPUT="$("$SCRIPT" --days 7 --repo arcaven/ThreeDoors 2>&1)" || {
    fail "script execution failed"
    echo "$OUTPUT" >&2
}
SECTIONS_OK=true
for section in "Total CI runs" "Merged PRs" "CI runs per merged PR" "Skipped" "Push-to-main" "Baseline Comparison" "ADR-0030"; do
    if ! echo "$OUTPUT" | grep -q "$section"; then
        fail "missing section: $section"
        SECTIONS_OK=false
    fi
done
if [[ "$SECTIONS_OK" == "true" ]]; then
    pass "all required sections present"
fi

# Test 7: Human-readable output includes baseline numbers
echo "Test: output includes baseline comparison"
if echo "$OUTPUT" | grep -q "5-10 runs/PR"; then
    pass "baseline comparison present"
else
    fail "should include baseline comparison (5-10 runs/PR)"
fi

# Test 8: JSON output is valid JSON
echo "Test: --json produces valid JSON"
JSON_OUT="$("$SCRIPT" --days 7 --json --repo arcaven/ThreeDoors 2>&1)" || {
    fail "script --json execution failed"
    echo "$JSON_OUT" >&2
}
if echo "$JSON_OUT" | jq empty 2>/dev/null; then
    pass "--json output is valid JSON"
else
    fail "--json output should be valid JSON"
fi

# Test 9: JSON output contains all required fields
echo "Test: JSON output has required fields"
FIELDS_OK=true
for field in "total_ci_runs" "merged_prs" "runs_per_merged_pr" "skipped_runs" "main_push_failures" "baseline_runs_per_pr" "improvement_pct" "adr_0030_triggered"; do
    if ! echo "$JSON_OUT" | jq -e "has(\"$field\")" >/dev/null 2>&1; then
        fail "missing JSON field: $field"
        FIELDS_OK=false
    fi
done
if [[ "$FIELDS_OK" == "true" ]]; then
    pass "all required JSON fields present"
fi

# Test 10: JSON fields have correct types
echo "Test: JSON field types are correct"
TYPES_OK=true
# total_ci_runs should be a number
if ! echo "$JSON_OUT" | jq -e '.total_ci_runs | type == "number"' >/dev/null 2>&1; then
    fail "total_ci_runs should be a number"
    TYPES_OK=false
fi
# merged_prs should be a number
if ! echo "$JSON_OUT" | jq -e '.merged_prs | type == "number"' >/dev/null 2>&1; then
    fail "merged_prs should be a number"
    TYPES_OK=false
fi
# adr_0030_triggered should be a boolean
if ! echo "$JSON_OUT" | jq -e '.adr_0030_triggered | type == "boolean"' >/dev/null 2>&1; then
    fail "adr_0030_triggered should be a boolean"
    TYPES_OK=false
fi
if [[ "$TYPES_OK" == "true" ]]; then
    pass "JSON field types correct"
fi

# Test 11: Custom --days parameter works
echo "Test: --days 1 produces valid output"
SHORT_OUT="$("$SCRIPT" --days 1 --json --repo arcaven/ThreeDoors 2>&1)" || true
if echo "$SHORT_OUT" | jq -e '.period_days == 1' >/dev/null 2>&1; then
    pass "--days 1 reflected in JSON output"
else
    fail "--days 1 should set period_days to 1"
fi

# Test 12: Timestamp is present and ISO format
echo "Test: timestamp is ISO 8601 format"
TS="$(echo "$JSON_OUT" | jq -r '.timestamp')"
if [[ "$TS" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
    pass "timestamp is ISO 8601"
else
    fail "timestamp should be ISO 8601, got: $TS"
fi

echo ""
echo "=== Results ==="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total:  $((PASS + FAIL))"

if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi

echo "All tests passed."
exit 0
