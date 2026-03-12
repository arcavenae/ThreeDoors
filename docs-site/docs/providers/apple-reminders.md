# Apple Reminders

Import tasks from macOS Reminders lists into ThreeDoors.

!!! warning "macOS Only"
    This provider requires macOS 12+ (Monterey or later) and uses JavaScript for Automation (JXA) via `osascript`. It is not available on Linux or Windows.

## Overview

The Apple Reminders provider imports tasks from one or more Reminders lists. Tasks from your selected lists appear as doors in ThreeDoors.

## Prerequisites

- **macOS 12+** (Monterey or later)
- **Reminders** app installed (ships with macOS)
- Reminders app access — you may be prompted to grant permission on first use

## Setup

1. Ensure you have at least one Reminders list with tasks

2. Add the Reminders provider to your configuration:

    ```yaml
    # ~/.threedoors/config.yaml
    providers:
      - name: reminders
        settings:
          lists: Work,Personal
          include_completed: false
    ```

3. Launch ThreeDoors — tasks from your specified lists appear as doors

!!! note
    On first use, macOS may prompt you to grant Reminders access to your terminal app. Accept the prompt to allow ThreeDoors to read your lists.

## Configuration

```yaml
providers:
  - name: reminders
    settings:
      lists: Work,Personal
      include_completed: false
```

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `lists` | *(required)* | Comma-separated list of Reminders list names to import from |
| `include_completed` | `false` | Whether to include completed reminders |

### List Selection

Specify one or more Reminders lists as a comma-separated string. Only tasks from those lists are imported.

```yaml
# Single list
lists: Work

# Multiple lists
lists: Work,Personal,Shopping
```

!!! tip
    List names must match exactly, including capitalization. Check the Reminders app for the exact names.

## Usage

### How Sync Works

- Tasks from your selected Reminders lists are imported into ThreeDoors
- Completing a task in ThreeDoors marks the corresponding reminder as complete
- New reminders added in the Reminders app appear in ThreeDoors on the next sync cycle

### Status Mapping

| Reminders State | ThreeDoors Status |
|-----------------|-------------------|
| Incomplete | Todo |
| Completed | Complete |

## Troubleshooting

**Tasks not appearing**

- Verify list names match exactly (case-sensitive)
- Check that the lists have incomplete reminders
- Ensure `include_completed` is set appropriately
- Run `:health` to check Reminders connectivity

**Permission denied**

- macOS may need you to grant automation access: System Settings → Privacy & Security → Automation → allow your terminal app to control Reminders
- Try removing and re-adding the permission if it was previously denied

**Slow performance**

- The JXA bridge can be slow with large numbers of reminders. Consider filtering to specific lists rather than importing all lists.
