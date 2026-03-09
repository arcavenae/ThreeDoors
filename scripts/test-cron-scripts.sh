#!/usr/bin/env bash
set -euo pipefail

# test-cron-scripts.sh — Tests for sm-sprint-health.sh and qa-coverage-audit.sh
# Run from the repository root: ./scripts/test-cron-scripts.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TESTDATA="$SCRIPT_DIR/testdata"
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

echo "=== SM Sprint Health Script Tests ==="

# Test 1: Help flag exits 0
echo "Test: --help exits successfully"
if "$SCRIPT_DIR/sm-sprint-health.sh" --help >/dev/null 2>&1; then
    pass "--help exits 0"
else
    fail "--help should exit 0"
fi

# Test 2: Detects blocked stories from test fixtures
echo "Test: detects blocked stories"
OUTPUT="$("$SCRIPT_DIR/sm-sprint-health.sh" --stories-dir "$TESTDATA/test-stories" --repo arcaven/ThreeDoors 2>&1)" || true
if echo "$OUTPUT" | grep -q "RISK.*1.2.*blocked"; then
    pass "detected blocked story 1.2"
else
    fail "should detect blocked story 1.2"
fi

# Test 3: Does not flag non-blocked stories
echo "Test: does not flag active stories as blocked"
if echo "$OUTPUT" | grep -q "RISK.*1.1"; then
    fail "should not flag story 1.1 as blocked"
else
    pass "story 1.1 not flagged"
fi

# Test 4: Report includes timestamp
echo "Test: report includes timestamp"
if echo "$OUTPUT" | grep -q "Timestamp:"; then
    pass "report has timestamp"
else
    fail "report should include timestamp"
fi

# Test 5: Invalid stories dir handled gracefully
echo "Test: missing stories dir handled gracefully"
OUTPUT2="$("$SCRIPT_DIR/sm-sprint-health.sh" --stories-dir "/nonexistent/dir" --repo arcaven/ThreeDoors 2>&1)" || true
if echo "$OUTPUT2" | grep -q "Warning.*not found"; then
    pass "missing dir produces warning"
else
    fail "should warn about missing stories dir"
fi

echo ""
echo "=== QA Coverage Audit Script Tests ==="

# Test 6: Help flag exits 0
echo "Test: --help exits successfully"
if "$SCRIPT_DIR/qa-coverage-audit.sh" --help >/dev/null 2>&1; then
    pass "--help exits 0"
else
    fail "--help should exit 0"
fi

# Test 7: First run without baseline
echo "Test: first run without baseline establishes initial baseline"
TMPDIR_TEST="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_TEST"' EXIT
OUTPUT3="$(cd "$REPO_ROOT" && BASELINE_FILE="$TMPDIR_TEST/nonexistent.json" "$SCRIPT_DIR/qa-coverage-audit.sh" 2>&1)" || true
if echo "$OUTPUT3" | grep -q "No baseline found\|initial baseline\|Establishing initial baseline"; then
    pass "first run recognizes no baseline"
else
    fail "should recognize first run with no baseline"
fi

# Test 8: Report includes package coverage
echo "Test: report includes coverage percentages"
if echo "$OUTPUT3" | grep -q "%"; then
    pass "report contains coverage percentages"
else
    fail "report should contain coverage percentages"
fi

# Test 9: Update flag creates baseline file
echo "Test: --update creates baseline file"
cd "$REPO_ROOT" && BASELINE_FILE="$TMPDIR_TEST/new-baseline.json" "$SCRIPT_DIR/qa-coverage-audit.sh" --update >/dev/null 2>&1 || true
if [[ -f "$TMPDIR_TEST/new-baseline.json" ]]; then
    pass "--update creates baseline file"
    # Verify it's valid JSON
    if jq empty "$TMPDIR_TEST/new-baseline.json" 2>/dev/null; then
        pass "baseline file is valid JSON"
    else
        fail "baseline file should be valid JSON"
    fi
else
    fail "--update should create baseline file"
fi

# Test 10: Baseline file has expected structure
echo "Test: baseline file has expected structure"
if jq -e '.updated and .packages' "$TMPDIR_TEST/new-baseline.json" >/dev/null 2>&1; then
    pass "baseline has 'updated' and 'packages' keys"
else
    fail "baseline should have 'updated' and 'packages' keys"
fi

echo ""
echo "=== Coverage Baseline JSON Tests ==="

# Test 11: Coverage baseline is valid JSON
echo "Test: coverage-baseline.json is valid JSON"
if jq empty "$REPO_ROOT/docs/quality/coverage-baseline.json" 2>/dev/null; then
    pass "coverage-baseline.json is valid JSON"
else
    fail "coverage-baseline.json should be valid JSON"
fi

# Test 12: Coverage baseline has expected structure
echo "Test: coverage-baseline.json has expected structure"
if jq -e '.updated and .packages' "$REPO_ROOT/docs/quality/coverage-baseline.json" >/dev/null 2>&1; then
    pass "baseline has correct structure"
else
    fail "baseline should have 'updated' and 'packages' keys"
fi

# Test 13: All packages have coverage and updated fields
echo "Test: all packages have coverage and updated fields"
INVALID="$(jq '[.packages | to_entries[] | select(.value.coverage == null or .value.updated == null)] | length' "$REPO_ROOT/docs/quality/coverage-baseline.json")"
if [[ "$INVALID" -eq 0 ]]; then
    pass "all packages have required fields"
else
    fail "$INVALID packages missing required fields"
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
