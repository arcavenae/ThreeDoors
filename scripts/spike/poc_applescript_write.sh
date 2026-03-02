#!/bin/bash
# poc_applescript_write.sh - Test write capabilities to Apple Notes via osascript
# Requires: macOS, Automation permission for Notes.app
# WARNING: This modifies the test note! Run setup_test_note.sh after to restore.

set -euo pipefail

NOTE_TITLE="ThreeDoors Tasks"

echo "=== AppleScript Write PoC ==="
echo "Testing write capabilities on note: $NOTE_TITLE"
echo "---"

# Test 1: Read current body
echo "Test 1: Reading current note body..."
CURRENT_BODY=$(osascript -e "tell application \"Notes\" to get body of note \"$NOTE_TITLE\"" 2>&1) || {
    echo "[FAIL] Cannot read note body: $CURRENT_BODY" >&2
    exit 1
}
echo "[OK] Read body (${#CURRENT_BODY} chars)"

# Test 2: Append text via set body
echo ""
echo "Test 2: Attempting to append text via 'set body'..."
APPEND_RESULT=$(osascript <<'EOF' 2>&1
tell application "Notes"
    set theNote to first note whose name is "ThreeDoors Tasks"
    set currentBody to body of theNote
    set body of theNote to currentBody & "<br>- [ ] Spike test task (added by PoC)"
    return "success"
end tell
EOF
) || {
    echo "[FAIL] Append via set body failed: $APPEND_RESULT" >&2
    echo "[INFO] Write capability: APPEND NOT SUPPORTED"
}
if [ "${APPEND_RESULT:-}" = "success" ]; then
    echo "[OK] Append via set body succeeded"

    # Verify the append
    VERIFY=$(osascript -e "tell application \"Notes\" to get plaintext text of note \"$NOTE_TITLE\"" 2>&1)
    if echo "$VERIFY" | grep -q "Spike test task"; then
        echo "[OK] Verified: appended text is visible in plaintext"
    else
        echo "[WARN] Appended text not visible in plaintext output"
    fi
fi

# Test 3: Replace entire body
echo ""
echo "Test 3: Attempting to replace entire body..."
REPLACE_RESULT=$(osascript <<'EOF' 2>&1
tell application "Notes"
    set theNote to first note whose name is "ThreeDoors Tasks"
    set body of theNote to "- [ ] Replaced task 1
- [x] Replaced task 2
- [ ] Replaced task 3"
    return "success"
end tell
EOF
) || {
    echo "[FAIL] Replace body failed: $REPLACE_RESULT" >&2
}
if [ "${REPLACE_RESULT:-}" = "success" ]; then
    echo "[OK] Replace body succeeded"
    VERIFY2=$(osascript -e "tell application \"Notes\" to get plaintext text of note \"$NOTE_TITLE\"" 2>&1)
    echo "New plaintext content:"
    echo "$VERIFY2"
fi

# Test 4: Check if formatting is preserved
echo ""
echo "Test 4: Checking if checkbox formatting survives write..."
BODY_AFTER=$(osascript -e "tell application \"Notes\" to get body of note \"$NOTE_TITLE\"" 2>&1)
if echo "$BODY_AFTER" | grep -q "checkbox"; then
    echo "[OK] Checkbox HTML elements preserved after write"
else
    echo "[WARN] Checkbox formatting may be lost after write (plaintext only)"
    echo "[INFO] Body after write: ${BODY_AFTER:0:300}"
fi

# Restore original note
echo ""
echo "Restoring original test note..."
bash "$(dirname "$0")/setup_test_note.sh" > /dev/null 2>&1 && echo "[OK] Original note restored" || echo "[WARN] Could not restore note"

echo "---"
echo "=== Write Capability Summary ==="
echo "Append via set body: ${APPEND_RESULT:-FAILED}"
echo "Replace via set body: ${REPLACE_RESULT:-FAILED}"
echo "Checkbox preservation: Check output above"
