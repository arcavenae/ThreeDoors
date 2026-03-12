# Data Directory

All ThreeDoors data lives in `~/.threedoors/`. This directory is created automatically on first launch.

---

## Directory Layout

```
~/.threedoors/
├── config.yaml         # Provider and theme configuration
├── tasks.yaml          # Active tasks (text file provider)
├── completed.txt       # Append-only completed task log
├── values.yaml         # Your personal values and goals
├── sessions.jsonl      # Session metrics (one JSON object per line)
├── patterns.json       # Cached pattern analysis report
├── enrichment.db       # SQLite database for cross-references and linking
├── improvements.txt    # Exit survey responses
├── sync-queue.jsonl    # Write-ahead log for pending sync operations
└── .onboarded          # First-run marker (empty file)
```

---

## File Descriptions

### `config.yaml`

Provider configuration and settings. See [Config File](config-file.md) for the complete schema reference.

### `tasks.yaml`

Active tasks stored by the text file provider. Each task includes an ID, text, status, timestamps, and optional metadata (type, effort, location, context). This file is read and written atomically — ThreeDoors writes to a `.tmp` file, syncs to disk, then renames to prevent corruption.

### `completed.txt`

Append-only log of completed tasks with timestamps. Used for daily completion tracking and the "better than yesterday" comparison. Each line records the task ID, text, and completion time.

### `values.yaml`

Up to 5 personal values or goals configured during onboarding or via `:goals edit`. These display in the footer as a reminder while you work.

### `sessions.jsonl`

One JSON object per line, one line per session. Records:

- Session start/end times and duration
- Tasks completed, doors viewed, refreshes used
- Door selections (position and task ID)
- Task bypasses, mood entries, door feedback
- Time to first door selection

### `patterns.json`

Cached analysis report generated from session data. Includes:

- Door position bias (left/center/right preferences)
- Task type completion rates
- Time-of-day productivity patterns
- Mood-task correlations

Automatically regenerated when new sessions arrive. Delete this file to force reanalysis.

### `enrichment.db`

SQLite database storing cross-reference links between tasks. Created when you first use the link feature (`l` in detail view). Links are bidirectional and persist across sessions.

### `improvements.txt`

Responses from the exit survey that appears when you quit. Helps track feature requests and pain points over time.

### `sync-queue.jsonl`

Write-ahead log (WAL) for offline-first operation. When a provider is temporarily unavailable, write operations queue here and replay automatically with exponential backoff when the provider comes back. Maximum 10,000 queued operations (oldest evicted first).

### `.onboarded`

Empty marker file created after completing the onboarding wizard. Delete this file to re-trigger onboarding on next launch.

---

## Log Files

Session data is stored in `sessions.jsonl` within the data directory. ThreeDoors does not write separate log files — diagnostic output goes to stderr when needed.

Run `make analyze` from the project source to execute analysis scripts against your session data, or use `:stats` in the TUI for a quick overview.

---

## Resetting Data

To start completely fresh:

```bash
rm -rf ~/.threedoors
```

ThreeDoors will recreate the directory and show onboarding on next launch.

To reset specific data:

| Action | Command |
|--------|---------|
| Clear session history | `rm ~/.threedoors/sessions.jsonl` |
| Force pattern reanalysis | `rm ~/.threedoors/patterns.json` |
| Reset goals | `rm ~/.threedoors/values.yaml` |
| Re-trigger onboarding | `rm ~/.threedoors/.onboarded` |
| Clear pending syncs | `rm ~/.threedoors/sync-queue.jsonl` |

---

## File Permissions

ThreeDoors creates the data directory with `0700` permissions (owner-only access) and individual files with `0600` permissions. This prevents other users on the system from reading your task data or credentials.
