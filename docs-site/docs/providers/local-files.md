# Local Files

The local file provider stores tasks as YAML in a plain text file. This is the default provider — no external services or accounts required.

## Overview

Local file storage is the simplest way to use ThreeDoors. Tasks live in a YAML file on your filesystem, read and written atomically for crash safety.

## Prerequisites

- None — this provider works on any platform with no dependencies

## Setup

Local files are the default provider. On first launch, ThreeDoors creates `~/.threedoors/tasks.yaml` automatically.

No configuration is needed for basic usage. If you want to customize the file path, see [Configuration](#configuration) below.

## Configuration

**Simple mode** (single provider):

```yaml
# ~/.threedoors/config.yaml
provider: textfile
```

**Advanced mode** (multiple providers):

```yaml
providers:
  - name: textfile
    settings:
      task_file: ~/.threedoors/tasks.yaml
```

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `task_file` | `~/.threedoors/tasks.yaml` | Path to the YAML task file |

## Usage

### Task File Format

Tasks are stored as a YAML list. Each task has an ID, text, status, and optional metadata:

```yaml
- id: a1b2c3d4
  text: Buy groceries
  status: todo
  created: 2026-03-01T10:00:00Z

- id: e5f6g7h8
  text: Review pull request
  status: in-progress
  created: 2026-03-01T11:00:00Z
  type: technical
  effort: medium

- id: i9j0k1l2
  text: Schedule dentist appointment
  status: todo
  created: 2026-03-02T09:00:00Z
  type: administrative
  effort: quick-win
```

### Adding Tasks

You can add tasks through the TUI:

- **Quick add:** `:add Buy groceries`
- **Add with context:** `:add-ctx Refactor auth module` — prompts for why it matters

Or via the CLI:

```bash
threedoors task add "Buy groceries"
threedoors task add "Review PR" --type technical --effort medium
```

You can also edit `tasks.yaml` directly in any text editor. ThreeDoors picks up changes on the next launch.

### File Location

All data lives under `~/.threedoors/`:

```
~/.threedoors/
├── config.yaml         # Provider configuration
├── tasks.yaml          # Active tasks (this file)
├── completed.txt       # Completed task log (append-only)
└── ...
```

### Atomic Writes

The local file provider uses atomic writes for crash safety:

1. Write to a `.tmp` file
2. Call `fsync` to flush to disk
3. Rename `.tmp` to the final path

This ensures your task file is never corrupted, even if ThreeDoors crashes mid-write.

## Troubleshooting

**"No tasks found"**

- Check that your task file exists: `ls ~/.threedoors/tasks.yaml`
- Verify the YAML syntax is valid — a misplaced indent or missing dash can cause parse errors
- Run `:health` to see specific error messages

**Permission errors**

- Ensure your user has read/write access to `~/.threedoors/`
- Check that the directory exists: `ls -la ~/.threedoors/`

**Custom path not working**

- In advanced mode, ensure the path is absolute (starts with `/` or `~`)
- Verify the `task_file` setting is under the correct provider entry in `config.yaml`
