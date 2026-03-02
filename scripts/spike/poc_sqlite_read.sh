#!/bin/bash
# poc_sqlite_read.sh - Read tasks from Apple Notes SQLite database directly
# Requires: macOS, Full Disk Access permission (System Settings > Privacy > Full Disk Access)
# WARNING: This is READ-ONLY. Never write to this database.

set -euo pipefail

DB_PATH="$HOME/Library/Group Containers/group.com.apple.notes/NoteStore.sqlite"
NOTE_TITLE="ThreeDoors Tasks"

echo "=== SQLite Direct Read PoC ==="
echo "Database: $DB_PATH"
echo "---"

# Check database exists
if [ ! -f "$DB_PATH" ]; then
    echo "[FAIL] Apple Notes database not found at: $DB_PATH" >&2
    echo "[INFO] This may require Full Disk Access permission" >&2
    exit 1
fi

# Check readability
if [ ! -r "$DB_PATH" ]; then
    echo "[FAIL] Cannot read Apple Notes database (permission denied)" >&2
    echo "[INFO] Grant Full Disk Access: System Settings > Privacy & Security > Full Disk Access > Add Terminal/iTerm2" >&2
    exit 1
fi

# List all tables for schema exploration
echo "Schema exploration:"
echo "Tables:"
sqlite3 "$DB_PATH" ".tables" 2>&1 || {
    echo "[FAIL] sqlite3 query failed (database locked?)" >&2
    exit 1
}

echo ""
echo "---"
echo "Searching for note: $NOTE_TITLE"

# Query for the specific note
RESULT=$(sqlite3 -separator $'\t' "$DB_PATH" "
    SELECT
        ZICCLOUDSYNCINGOBJECT.Z_PK,
        ZICCLOUDSYNCINGOBJECT.ZTITLE1,
        ZICCLOUDSYNCINGOBJECT.ZSNIPPET,
        length(ZICCLOUDSYNCINGOBJECT.ZMERGEABLEDATA1) as DATA_LEN
    FROM ZICCLOUDSYNCINGOBJECT
    WHERE ZICCLOUDSYNCINGOBJECT.ZTITLE1 LIKE '%${NOTE_TITLE}%'
    LIMIT 5
" 2>&1) || {
    echo "[FAIL] Query failed: $RESULT" >&2
    exit 1
}

if [ -z "$RESULT" ]; then
    echo "[FAIL] Note '$NOTE_TITLE' not found in database" >&2
    echo "[INFO] Available notes (first 10):"
    sqlite3 "$DB_PATH" "
        SELECT ZTITLE1 FROM ZICCLOUDSYNCINGOBJECT
        WHERE ZTITLE1 IS NOT NULL AND ZTITLE1 != ''
        LIMIT 10
    " 2>&1
    exit 1
fi

echo "Found note(s):"
echo "$RESULT"

# Try to get plaintext content
echo ""
echo "---"
echo "Attempting to read note content..."

# Note: ZPLAINTEXT may or may not exist depending on macOS version
# The actual content is often in ZMERGEABLEDATA1 as protobuf
PLAINTEXT=$(sqlite3 "$DB_PATH" "
    SELECT ZSNIPPET FROM ZICCLOUDSYNCINGOBJECT
    WHERE ZTITLE1 LIKE '%${NOTE_TITLE}%'
    LIMIT 1
" 2>&1) || true

if [ -n "${PLAINTEXT:-}" ]; then
    echo "Snippet/preview:"
    echo "$PLAINTEXT"
else
    echo "[WARN] No plaintext content available via ZSNIPPET"
fi

echo ""
echo "---"
echo "=== SQLite Schema Observations ==="
echo "Key columns in ZICCLOUDSYNCINGOBJECT:"
sqlite3 "$DB_PATH" "PRAGMA table_info(ZICCLOUDSYNCINGOBJECT);" 2>&1 | head -30

echo ""
echo "[INFO] Note content is stored in ZMERGEABLEDATA1 as protobuf binary data"
echo "[INFO] Parsing this requires reverse-engineering Apple's protobuf schema"
echo "[INFO] This is fragile and undocumented — schema changes between macOS versions"

echo ""
echo "---"
echo "[OK] SQLite read PoC completed — database is accessible"
echo "[WARN] Full content extraction requires protobuf parsing (high complexity)"
