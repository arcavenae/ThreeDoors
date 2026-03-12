# CLI Commands

ThreeDoors includes a full CLI for headless and scripted usage. All commands support `--json` for machine-readable output.

---

## `task` — Task Management

```bash
threedoors task add <text>           # Add a task
  --type <type>                      #   creative, technical, administrative, physical
  --effort <effort>                  #   quick-win, medium, deep-work
  --context <text>                   #   Why this task matters
  --stdin                            #   Read task text from stdin

threedoors task list                 # List tasks
  --status <status>                  #   Filter: todo, in-progress, blocked, complete, deferred
  --type <type>                      #   Filter by type
  --effort <effort>                  #   Filter by effort

threedoors task show <id>            # Show full task details
threedoors task edit <id>            # Edit a task
  --text <text>                      #   New task text
  --context <text>                   #   New context
threedoors task complete <id>        # Mark task complete
threedoors task delete <id>          # Delete a task
threedoors task note <id> <text>     # Add a note to a task
threedoors task search <query>       # Search tasks by text
threedoors task unblock <id>         # Unblock a blocked task
```

### Examples

Add a task with categorization:

```bash
threedoors task add "Refactor auth module" --type technical --effort deep-work --context "Security audit flagged it"
```

List in-progress tasks as JSON:

```bash
threedoors task list --status in-progress --json
```

Pipe tasks from another tool:

```bash
echo "Review PR #42" | threedoors task add --stdin
```

---

## `doors` — Three Doors in the Terminal

```bash
threedoors doors                     # Show three random tasks
threedoors doors --pick 1            # Auto-select door 1 (1-3)
threedoors doors --interactive       # Prompted selection mode
```

The `doors` command brings the three-door experience to the terminal without launching the full TUI. Use `--pick` for scripting or `--interactive` for a guided flow.

---

## `mood` — Mood Tracking

```bash
threedoors mood set <mood>           # Record mood
threedoors mood history              # View mood entries
```

Available moods: `focused`, `tired`, `stressed`, `energized`, `distracted`, `calm`, or any custom string.

---

## `stats` — Productivity Analytics

```bash
threedoors stats                     # Session overview
threedoors stats --daily             # Daily breakdown
threedoors stats --weekly            # Weekly trends
threedoors stats --patterns          # Behavioral patterns
```

Stats are computed from your session data (`sessions.jsonl`). Requires at least 3 sessions for pattern analysis.

---

## `config` — Configuration

```bash
threedoors config show               # Display full configuration
threedoors config get <key>          # Get a single config value
threedoors config set <key> <value>  # Set a config value
```

### Examples

```bash
threedoors config set theme modern
threedoors config get provider
threedoors config show --json
```

---

## `health` — System Health Check

```bash
threedoors health                    # Check provider connectivity and data files
```

The health command verifies:

- Task file exists and is readable/writable
- Database can load and parse tasks
- Sync status for each provider
- Provider-specific connectivity (Apple Notes access, Jira API, etc.)

---

## `completion` — Shell Completions

```bash
threedoors completion bash           # Generate bash completions
threedoors completion zsh            # Generate zsh completions
threedoors completion fish           # Generate fish completions
```

### Setup

=== "Bash"

    ```bash
    threedoors completion bash > /etc/bash_completion.d/threedoors
    ```

=== "Zsh"

    ```bash
    threedoors completion zsh > "${fpath[1]}/_threedoors"
    ```

=== "Fish"

    ```bash
    threedoors completion fish > ~/.config/fish/completions/threedoors.fish
    ```

---

## `--version`

```bash
threedoors --version
```

Prints the version string and exits.

---

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format (machine-readable) |
| `--version` | Print version and exit |
| `--help` | Show help for any command |
