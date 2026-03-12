# Obsidian

Read and write tasks from Markdown files in an Obsidian vault with real-time sync.

## Overview

The Obsidian provider integrates with your Obsidian vault, reading checkbox tasks from Markdown files and writing completions back. Changes sync bidirectionally in real time using filesystem notifications.

## Prerequisites

- An Obsidian vault (any version)
- Markdown files with checkbox task syntax

!!! note
    Obsidian itself doesn't need to be running — ThreeDoors reads the vault files directly. However, if both are open simultaneously, changes sync between them in real time.

## Setup

1. **Identify your vault path** — the folder containing the `.obsidian/` directory

2. **Add the Obsidian provider** to your configuration:

    ```yaml
    # ~/.threedoors/config.yaml
    providers:
      - name: obsidian
        settings:
          vault_path: /Users/you/Documents/MyVault
          tasks_folder: tasks
          file_pattern: "*.md"
    ```

3. Launch ThreeDoors — tasks from your vault appear as doors

## Configuration

```yaml
providers:
  - name: obsidian
    settings:
      vault_path: /Users/you/Documents/MyVault
      tasks_folder: tasks
      file_pattern: "*.md"
      daily_notes: true
      daily_notes_folder: Daily
      daily_notes_heading: "## Tasks"
      daily_notes_format: "2006-01-02.md"
```

### Settings

| Setting | Required | Default | Description |
|---------|----------|---------|-------------|
| `vault_path` | Yes | — | Absolute path to your Obsidian vault root |
| `tasks_folder` | No | — | Subfolder within the vault to scan for tasks |
| `file_pattern` | No | `"*.md"` | Glob pattern for files to scan |
| `daily_notes` | No | `false` | Enable daily notes integration |
| `daily_notes_folder` | No | `"Daily"` | Folder for daily notes (relative to vault root) |
| `daily_notes_heading` | No | `"## Tasks"` | Heading under which to add tasks in daily notes |
| `daily_notes_format` | No | `"2006-01-02.md"` | Date format for daily note filenames (Go time format) |

## Usage

### Task Format

ThreeDoors recognizes standard Markdown checkboxes:

```markdown
- [ ] Uncompleted task (imported as Todo)
- [x] Completed task (imported as Complete)
- [/] In-progress task (imported as In-Progress)
```

Both `-` and `*` list markers are supported.

### Metadata Parsing

ThreeDoors extracts metadata from Obsidian-style annotations:

| Annotation | Mapping |
|------------|---------|
| `📅 2026-03-15` | Due date |
| `⏫` | High priority → Deep Work effort |
| `🔼` | Medium priority → Medium effort |
| `🔽` | Low priority → Quick Win effort |
| `#project` `#urgent` | Stored in task context |

Example:

```markdown
- [ ] Review quarterly report 📅 2026-03-15 ⏫ #work
```

### Task ID Tracking

ThreeDoors embeds unique IDs in HTML comments to track tasks across syncs:

```markdown
- [ ] Buy groceries <!-- td:a1b2c3d4 -->
```

These are invisible in Obsidian's reading view but allow ThreeDoors to maintain identity across edits and syncs. IDs are added automatically when ThreeDoors first encounters a task.

### Daily Notes Integration

When enabled, ThreeDoors appends new tasks (added via `:add` or the CLI) to today's daily note under the configured heading:

```markdown
## Tasks
- [ ] New task from ThreeDoors <!-- td:uuid -->
```

The daily note file is created automatically if it doesn't exist (e.g., `Daily/2026-03-12.md`).

### Real-Time Sync

ThreeDoors watches your vault for changes using filesystem notifications:

- Changes in Obsidian appear in ThreeDoors within ~100ms
- Self-writes are tracked with a 2-second cooldown to avoid echo loops
- Both apps can be open simultaneously without conflicts

### Path Safety

The Obsidian adapter validates all paths:

- No `..` traversal allowed
- No null bytes
- No absolute paths within vault-relative settings

## Troubleshooting

**Obsidian tasks not appearing**

- Confirm `vault_path` points to your vault root (the folder containing `.obsidian/`)
- Check `tasks_folder` is relative to the vault root
- Verify your files match `file_pattern` (default: `*.md`)
- Ensure tasks use checkbox syntax: `- [ ] Task text`

**Tasks duplicating**

- This can happen if the same file is matched by multiple patterns. Ensure `tasks_folder` and `file_pattern` don't overlap with daily notes settings.

**Changes not syncing from Obsidian**

- Verify ThreeDoors has read access to the vault directory
- Check that filesystem notifications are working (some network drives don't support them)
- Restart ThreeDoors to force a full rescan

**Daily notes not working**

- Verify `daily_notes: true` is set
- Check that `daily_notes_folder` exists relative to your vault root
- Ensure the date format matches your daily notes plugin settings
