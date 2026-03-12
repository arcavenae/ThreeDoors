# Apple Notes

Sync tasks bidirectionally with Apple Notes using checkbox syntax. Changes in Apple Notes appear in ThreeDoors, and completions in ThreeDoors update your note.

!!! warning "macOS Only"
    This provider requires macOS and uses AppleScript via `osascript` to communicate with Apple Notes. It is not available on Linux or Windows.

## Overview

The Apple Notes provider reads and writes tasks from a specific note in Apple Notes. Tasks use standard checkbox syntax (`- [ ]` / `- [x]`), making them easy to manage from either ThreeDoors or the Notes app.

## Prerequisites

- **macOS** (any recent version)
- **Apple Notes** app installed (ships with macOS)
- **Full Disk Access** may be required — System Settings → Privacy & Security → Full Disk Access → add your terminal app
- A note in Apple Notes to use as your task list

## Setup

1. Create a note in Apple Notes with the title you want to use (e.g., "ThreeDoors Tasks")

2. Add tasks using checkbox syntax:
    ```
    - [ ] Task one
    - [ ] Task two
    - [x] Already done
    ```

3. Configure ThreeDoors to use Apple Notes:

    ```yaml
    # ~/.threedoors/config.yaml
    provider: applenotes
    note_title: ThreeDoors Tasks
    ```

4. Launch ThreeDoors — your Apple Notes tasks appear as doors

## Configuration

**Simple mode:**

```yaml
provider: applenotes
note_title: ThreeDoors Tasks
```

**Advanced mode** (multiple providers):

```yaml
providers:
  - name: applenotes
    settings:
      note_title: ThreeDoors Tasks
```

### Settings

| Setting | Required | Description |
|---------|----------|-------------|
| `note_title` | Yes | Exact title of the Apple Notes note to sync with |

!!! tip
    The note title must match exactly, including capitalization and spacing.

## Usage

### Bidirectional Sync

Changes flow both ways:

- **Notes → ThreeDoors:** Tasks added or edited in Apple Notes appear in ThreeDoors
- **ThreeDoors → Notes:** Tasks completed or updated in ThreeDoors are reflected in Apple Notes

### Task ID Tracking

ThreeDoors embeds unique IDs as HTML comments to track tasks across syncs:

```
- [ ] Buy groceries <!-- td:a1b2c3d4 -->
```

These comments are invisible in the Apple Notes UI but allow ThreeDoors to maintain task identity across edits and syncs.

### Timeouts

The Apple Notes provider enforces timeouts for AppleScript operations:

- **Read timeout:** 2 seconds
- **Write timeout:** 5 seconds

If Apple Notes is slow to respond (e.g., during iCloud sync), operations queue locally and replay automatically.

## Troubleshooting

**Apple Notes not syncing**

- Grant Full Disk Access: System Settings → Privacy & Security → Full Disk Access → add your terminal app
- Verify the note exists with the exact title from your config
- Run `:health` to see specific error messages

**Permission dialog keeps appearing**

- This usually means Full Disk Access hasn't been granted. Add your terminal application (Terminal.app, iTerm2, Alacritty, etc.) to the Full Disk Access list in System Settings.

**Tasks not appearing**

- Confirm the note title in `config.yaml` matches the Apple Notes title exactly
- Ensure tasks use checkbox syntax: `- [ ] Task text`
- Run `:health` to check AppleScript execution access

**Health check**

Run `:health` to verify Apple Notes connectivity. The health checker tests:

- AppleScript execution access
- Note accessibility (can the note be found and read?)
- Full Disk Access permissions
