# Troubleshooting

Common issues, diagnostics, and FAQ for ThreeDoors.

---

## Health Check

Run `:health` in the TUI or `threedoors health` from the CLI to diagnose issues. The health checker verifies:

1. **Task file** — Exists, readable, writable (tests atomic write)
2. **Database** — Can load and parse tasks, shows task count
3. **Sync status** — Last sync time, warns if stale (>24 hours)
4. **Apple Notes** — AppleScript access, Full Disk Access permissions

Status levels:

| Level | Meaning |
|-------|---------|
| **OK** | Everything is working |
| **WARN** | Something needs attention but isn't broken |
| **FAIL** | Something is broken and needs fixing |

---

## Common Issues

### "No tasks found"

- Check that your task file exists: `ls ~/.threedoors/tasks.yaml`
- If using Apple Notes, verify the note title matches your config exactly
- If using Obsidian, verify `vault_path` and `tasks_folder` are correct
- If using a multi-provider config, check that at least one provider has tasks

### Apple Notes not syncing

- Grant Full Disk Access: **System Settings > Privacy & Security > Full Disk Access** > add your terminal app
- Verify the note exists with the exact title from your config
- Run `:health` to see specific error messages
- Read/write timeouts: 2 seconds for reads, 5 seconds for writes

### Obsidian tasks not appearing

- Confirm `vault_path` points to your vault root (the folder containing `.obsidian/`)
- Check `tasks_folder` is relative to vault root
- Verify your files match `file_pattern` (default: `*.md`)
- Ensure tasks use checkbox syntax: `- [ ] Task text`

### Jira tasks not loading

- Verify `url`, `email`, and `api_token` are correct
- Test your JQL query in the Jira web UI first
- Ensure the API token has read access to the target projects
- Run `:health` to check Jira connectivity
- The adapter uses a circuit breaker — if Jira is repeatedly unreachable, polling backs off automatically

### Todoist tasks not loading

- Verify `api_token` is set (in config or `TODOIST_API_TOKEN` env var)
- Check that `project_ids` and `filter` are not both set (they are mutually exclusive)
- Run `:health` to check Todoist connectivity

### Sync stuck in "Pending"

- Check `:health` for provider errors
- Verify the provider is accessible (Apple Notes open, vault mounted, network available)
- Pending operations replay automatically with backoff — wait a few seconds
- If stuck, check `~/.threedoors/sync-queue.jsonl` for error details
- Maximum 10,000 queued operations (oldest evicted first)

### Session data not showing in insights

- Insights require **3+ completed sessions**
- Run `:insights` to see how many more sessions are needed
- Each app launch and quit counts as one session

### Onboarding keeps appearing

- The `.onboarded` marker file may be missing from `~/.threedoors/`
- Run ThreeDoors and complete the onboarding wizard — it creates the marker automatically
- Or create it manually: `touch ~/.threedoors/.onboarded`

---

## Diagnostics Commands

### In-App

| Command | Description |
|---------|-------------|
| `:health` | Run system health checks |
| `:stats` | Show session metrics summary |
| `:synclog` | Show sync history for providers |
| `:dashboard` | View insights and patterns |

### CLI

```bash
threedoors health                    # System health check
threedoors stats                     # Session overview
threedoors stats --patterns          # Behavioral patterns
threedoors config show               # View configuration
threedoors --version                 # Version info
```

### Doctor Command

```bash
threedoors doctor                    # Comprehensive self-diagnosis
```

The doctor command runs environment checks, data integrity validation, and configuration verification in a single pass. It checks:

- Go runtime and system environment
- Data directory permissions and file integrity
- Provider connectivity and configuration
- Task data consistency (orphaned references, invalid states)

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

## FAQ

### Does ThreeDoors work on Linux?

Yes. The core TUI and all non-Apple providers work on Linux. Apple Notes and Apple Reminders providers require macOS.

### Can I use multiple providers at once?

Yes. Configure multiple providers in `config.yaml` and tasks are aggregated and deduplicated across all sources. See [Config File](configuration/config-file.md) for details.

### Where is my data stored?

Everything lives in `~/.threedoors/`. See [Data Directory](configuration/data-directory.md) for the complete layout.

### How do I back up my tasks?

Copy the `~/.threedoors/` directory. The most important files are `tasks.yaml` (your tasks), `sessions.jsonl` (your history), and `config.yaml` (your settings).

### Can I sync across machines?

ThreeDoors itself doesn't provide cloud sync, but you can:

- Use a cloud-synced provider (Jira, GitHub Issues, Todoist)
- Point the Obsidian provider at a vault synced via Obsidian Sync or iCloud
- Sync the `~/.threedoors/` directory with a file sync service (Dropbox, iCloud Drive)

### What happens if my provider goes offline?

ThreeDoors uses a write-ahead log (WAL) for offline-first operation. Changes queue locally and replay automatically when the provider comes back. See [Data Directory](configuration/data-directory.md) for details on the sync queue.
