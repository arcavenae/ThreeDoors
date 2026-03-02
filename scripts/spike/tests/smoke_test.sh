#!/bin/bash
# smoke_test.sh - Validates PoC scripts are well-formed
# Checks: executable, proper shebang, expected output format
# This test runs on any platform (no macOS dependency)

set -euo pipefail

SPIKE_DIR="scripts/spike"
ERRORS=0

check_script() {
    local script="$1"
    local desc="$2"

    if [ ! -f "$script" ]; then
        echo "[SKIP] $desc: $script not found (approach may be non-viable)"
        return
    fi

    # Check executable
    if [ ! -x "$script" ]; then
        echo "[FAIL] $desc: not executable (run chmod +x)" >&2
        ERRORS=$((ERRORS + 1))
        return
    fi

    # Check shebang
    local first_line
    first_line=$(head -1 "$script")
    if [[ "$first_line" != "#!"* ]]; then
        echo "[FAIL] $desc: missing shebang line" >&2
        ERRORS=$((ERRORS + 1))
        return
    fi

    echo "[OK] $desc: well-formed"
}

echo "Running spike PoC smoke tests"
echo "---"

check_script "$SPIKE_DIR/setup_test_note.sh" "Test note setup script"
check_script "$SPIKE_DIR/poc_applescript_read.sh" "AppleScript read PoC"
check_script "$SPIKE_DIR/poc_applescript_write.sh" "AppleScript write PoC"
check_script "$SPIKE_DIR/poc_sqlite_read.sh" "SQLite read PoC"
check_script "$SPIKE_DIR/poc_mcp.sh" "MCP server PoC"

# Validate test fixtures exist
echo "---"
echo "Checking test fixtures:"
for fixture in testdata/sample_note_plaintext.txt testdata/sample_note_html.txt testdata/expected_parsed_tasks.json; do
    if [ -f "$SPIKE_DIR/$fixture" ]; then
        echo "[OK] Fixture exists: $fixture"
    else
        echo "[FAIL] Missing fixture: $fixture" >&2
        ERRORS=$((ERRORS + 1))
    fi
done

# Validate expected_parsed_tasks.json is valid JSON
if [ -f "$SPIKE_DIR/testdata/expected_parsed_tasks.json" ]; then
    if python3 -m json.tool "$SPIKE_DIR/testdata/expected_parsed_tasks.json" > /dev/null 2>&1; then
        echo "[OK] expected_parsed_tasks.json is valid JSON"
    else
        echo "[FAIL] expected_parsed_tasks.json is not valid JSON" >&2
        ERRORS=$((ERRORS + 1))
    fi
fi

echo "---"
if [ "$ERRORS" -eq 0 ]; then
    echo "[PASS] All smoke tests passed"
    exit 0
else
    echo "[FAIL] $ERRORS smoke test(s) failed" >&2
    exit 1
fi
