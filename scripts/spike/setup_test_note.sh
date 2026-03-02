#!/bin/bash
# setup_test_note.sh - Creates a test note in Apple Notes for spike evaluation
# Requires: macOS with Automation permission for Notes.app

set -euo pipefail

NOTE_TITLE="ThreeDoors Tasks"
NOTE_BODY="- [ ] Buy groceries for the week
- [ ] Review pull request #42
- [x] Set up development environment
- [ ] Write architecture decision record
- [ ] Schedule team sync meeting
- [x] Update project README"

echo "Creating test note '$NOTE_TITLE' in Apple Notes..."

osascript <<EOF
tell application "Notes"
    tell account "iCloud"
        try
            set existingNote to first note whose name is "$NOTE_TITLE"
            set body of existingNote to "$NOTE_BODY"
            return "Updated existing note"
        on error
            make new note at folder "Notes" with properties {name:"$NOTE_TITLE", body:"$NOTE_BODY"}
            return "Created new note"
        end try
    end tell
end tell
EOF

if [ $? -eq 0 ]; then
    echo "[OK] Test note '$NOTE_TITLE' created/updated in Apple Notes"
    exit 0
else
    echo "[FAIL] Could not create test note. Check Automation permissions." >&2
    exit 1
fi
