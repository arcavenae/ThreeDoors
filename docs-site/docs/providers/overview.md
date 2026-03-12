# Task Sources Overview

ThreeDoors supports multiple task storage backends called **providers**. You can use one or several simultaneously — tasks are aggregated and deduplicated across all sources.

## Multi-Source Architecture

At the core, ThreeDoors uses the **TaskProvider** interface to abstract storage backends. Each provider implements a common contract for loading, saving, and completing tasks. This means you can mix local files with cloud services seamlessly.

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│  Local YAML  │   │ Apple Notes  │   │    Jira     │
└──────┬───────┘   └──────┬───────┘   └──────┬──────┘
       │                  │                  │
       └──────────┬───────┴──────────────────┘
                  │
          ┌───────▼────────┐
          │  Connection     │
          │  Manager        │
          │  (aggregation,  │
          │   dedup, WAL)   │
          └───────┬─────────┘
                  │
          ┌───────▼─────────┐
          │   ThreeDoors    │
          │   Task Pool     │
          └─────────────────┘
```

### How Providers Work

1. **Registration** — Each provider registers with the adapter registry at startup
2. **Configuration** — You declare active providers in `~/.threedoors/config.yaml`
3. **Loading** — On launch, ThreeDoors calls `LoadTasks()` on each configured provider
4. **Aggregation** — Tasks from all providers merge into a single pool for door selection
5. **Write-back** — Changes (completions, status updates) route back to the originating provider

### Connection Manager

The connection manager handles:

- **Aggregation** — Merges tasks from all configured providers into one pool
- **Deduplication** — Prevents the same task from appearing twice if it exists in multiple sources
- **Fallback** — If a provider fails (e.g., API timeout), ThreeDoors falls back gracefully to the next available provider
- **Write-Ahead Log (WAL)** — All writes go through a crash-safe queue (`sync-queue.jsonl`) before being applied to the provider

## Available Providers

| Provider | Platform | Direction | Description |
|----------|----------|-----------|-------------|
| [Local Files](local-files.md) | Any | Read/Write | YAML task file (default) |
| [Apple Notes](apple-notes.md) | macOS | Bidirectional | Sync with Apple Notes checkboxes |
| [Apple Reminders](apple-reminders.md) | macOS | Read/Write | Import from Reminders lists |
| [Jira](jira.md) | Any | Read | Pull issues via JQL filtering |
| [GitHub Issues](github-issues.md) | Any | Read | Import repository issues |
| [Todoist](todoist.md) | Any | Read | Sync via REST API |
| [Obsidian](obsidian.md) | Any | Bidirectional | Read/write Markdown checkboxes in vaults |

## Mixing Sources

Configure multiple providers in `config.yaml` to pull tasks from several sources at once:

```yaml
providers:
  - name: applenotes
    settings:
      note_title: ThreeDoors Tasks
  - name: textfile
    settings:
      task_file: ~/.threedoors/tasks.yaml
  - name: jira
    settings:
      url: https://company.atlassian.net
      auth_type: basic
      email: user@example.com
      api_token: your-api-token
      jql: "assignee = currentUser() AND statusCategory != Done"
```

Tasks from all providers appear together in your three doors. Each task remembers which provider it came from, so completions and status changes route back to the correct source.

## Offline & Sync

All providers are wrapped with a **Write-Ahead Log (WAL)** for crash-safe, offline-first operation:

- Operations queue locally in `sync-queue.jsonl` when a provider is unavailable
- Automatic replay with exponential backoff when the provider reconnects
- Maximum 10,000 queued operations (oldest evicted first)

The footer shows sync status for each provider:

| Status | Meaning |
|--------|---------|
| **Synced** | Up to date |
| **Syncing** | Sync in progress |
| **Pending (N)** | N operations waiting to sync |
| **Error** | Last sync failed — run `:health` for details |

For bidirectional providers (Apple Notes, Obsidian), ThreeDoors uses a three-way sync engine. When conflicts occur (same task edited in both places), the most recent change wins.

## Health Check

Run `:health` in the TUI or `threedoors health` from the CLI to verify provider connectivity and diagnose issues.
