# ThreeDoors 🚪🚪🚪

<p align="center">
  <a href="https://github.com/arcaven/ThreeDoors/actions/workflows/ci.yml"><img src="https://github.com/arcaven/ThreeDoors/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI"></a>
  <a href="https://github.com/arcaven/ThreeDoors/releases/latest"><img src="https://img.shields.io/github/v/release/arcaven/ThreeDoors?style=flat&label=release&color=green" alt="Latest Release"></a>
  <a href="https://goreportcard.com/report/github.com/arcaven/ThreeDoors"><img src="https://goreportcard.com/badge/github.com/arcaven/ThreeDoors" alt="Go Report Card"></a>
  <a href="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux-lightgrey"><img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux-lightgrey" alt="Platform"></a>
  <a href="https://golang.org/doc/devel/release.html"><img src="https://img.shields.io/badge/Go-1.25.4+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/Built%20with-Bubbletea-purple" alt="Built with Bubbletea"></a>
</p>

<p align="center"><em>Progress over perfection. Three doors. One choice. Move forward.</em></p>

## 📑 Table of Contents

- [📦 Installation](#-installation)
- [🚀 Quick Start](#-quick-start)
- [✨ Features](#-features)
- [📖 User Guide](#-user-guide)
- [⌨️ Key Bindings](#%EF%B8%8F-key-bindings)
- [💻 CLI Reference](#-cli-reference)
- [🤖 MCP Server](#-mcp-server)
- [📁 Data Directory](#-data-directory)
- [🔒 Data & Privacy](#-data--privacy)
- [🧭 Philosophy](#-philosophy)
- [🛠️ Development](#%EF%B8%8F-development)
- [🤝 Contributing](#-contributing)
- [📚 Documentation](#-documentation)
- [License](#license)

---

## What is ThreeDoors?

ThreeDoors is a **radical rethinking of task management** that reduces decision friction by showing you only **three tasks at a time**. Instead of overwhelming you with an endless list, ThreeDoors presents three carefully selected "doors" — choose one, take action, and move forward.

It ships as both an **interactive TUI** (terminal user interface) and a **headless CLI** for scripting, plus an **MCP server** for LLM agent integration.

### The Problem

Traditional task lists create **choice paralysis**. Staring at 50+ tasks makes it hard to start anything. You spend more time reorganizing and re-prioritizing than actually doing the work.

### The ThreeDoors Solution

- **Three doors, one choice** — Reduces cognitive load by limiting options
- **Refresh when needed** — Don't like your options? Roll again
- **Quick search** — Press `/` to find something specific
- **Mood-aware tracking** — Log your emotional state to understand work patterns
- **Pattern learning** — Over time, learn which tasks you avoid and why
- **Avoidance detection** — Automatically surfaces tasks you keep skipping
- **Values alignment** — Keep your goals front-and-center while working
- **Multi-source aggregation** — Pull tasks from local files, Jira, GitHub Issues, Apple Notes, Apple Reminders, and Obsidian
- **CLI + TUI + MCP** — Three interfaces for different workflows

---

## 📸 Screenshots

<p align="center"><em>Three doors. Pick one. Move forward.</em></p>

```
  ╭────────────────────────╮ ╭────────────────────────╮ ╭────────────────────────╮
  │                        │ │                        │ │                        │
  │  [todo]                │ │  [todo]                │ │  [todo]                │
  │                        │ │                        │ │                        │
  │  Buy groceries         │ │  Read Go book          │ │  Exercise for 30 min   │
  │                        │ │                        │ │                        │
  ╰────────────────────────╯ ╰────────────────────────╯ ╰────────────────────────╯

  a/left, w/up, d/right to select │ s/down to re-roll │ Enter/Space to open
```

<!-- To capture an actual terminal recording, use charmbracelet/vhs:
     https://github.com/charmbracelet/vhs
     Store recordings and screenshots in docs/assets/ -->

<details>
<summary>More screenshots</summary>

| View | Screenshot |
|------|-----------|
| Three Doors | *Coming soon — door selection flow* |
| Task Detail | *Coming soon — task detail with actions* |
| Dashboard | *Coming soon — insights and analytics* |
| Themes | *Coming soon — classic, modern, scifi, shoji* |
| Search | *Coming soon — quick search with fuzzy filtering* |
| Onboarding | *Coming soon — first-run wizard* |

Screenshots and GIFs will be stored in [`docs/assets/`](docs/assets/).

</details>

---

## 📦 Installation

### Option 1: Homebrew (recommended)

```bash
brew install arcaven/tap/threedoors
```

**Alpha channel** — latest development builds from `main`:

```bash
brew install arcaven/tap/threedoors-a
```

Both can be installed side-by-side. Stable runs as `threedoors`, alpha runs as `threedoors-a`.

### Option 2: Download Pre-built Binary

Download the latest release from [GitHub Releases](https://github.com/arcaven/ThreeDoors/releases). Binaries are available for:

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `threedoors-darwin-arm64` |
| macOS (Intel) | `threedoors-darwin-amd64` |
| Linux (x86_64) | `threedoors-linux-amd64` |

```bash
chmod +x threedoors-*
mv threedoors-darwin-arm64 /usr/local/bin/threedoors   # adjust for your platform
```

All macOS binaries are **code-signed and Apple-notarized**.

### Option 3: Install with `go install`

```bash
go install github.com/arcaven/ThreeDoors/cmd/threedoors@latest
```

### Option 4: Build from Source

**Prerequisites:** Go 1.25.4+, Git, Make (optional)

```bash
git clone https://github.com/arcaven/ThreeDoors.git
cd ThreeDoors
make build
# Binary at bin/threedoors
```

---

## 🚀 Quick Start

1. **Launch** the TUI:
   ```bash
   threedoors
   ```
2. **First run** starts the onboarding wizard — learn key bindings, set your values/goals, and optionally import existing tasks.
3. **Select a door** with `a` (left), `w` (center), or `d` (right).
4. **Re-roll** doors with `s` if nothing appeals.
5. **Act on a task** — `c` (complete), `b` (blocked), `i` (in progress), `p` (procrastinate/defer).
6. **Add tasks** — `:add Buy groceries #quick-win @errands`
7. **Log your mood** — `m` anytime, or `:mood focused`
8. **Search** — `/` to find a specific task.
9. **View insights** — `:dashboard` to see trends and patterns.

Or use the **CLI** without launching the TUI:

```bash
threedoors task add "Buy groceries" --type administrative --effort quick-win
threedoors task list --status todo
threedoors doors                    # Show three doors in the terminal
threedoors stats --daily
```

---

## ✨ Features

### Core Task Management
- 🚪 **Three Doors Display** — View three randomly selected tasks, avoiding recently shown ones
- 🔄 **Refresh Mechanism** — Re-roll doors when nothing appeals
- ✅ **Task Status Workflow** — Seven states: `todo` → `in-progress` → `in-review` → `complete`, plus `blocked`, `deferred`, and `archived`
- ➕ **Quick Add** — Add tasks inline with `:add` or from the CLI; supports context capture with `:add --why`
- 🏷️ **Inline Tagging** — Tag tasks as you add them: `Design homepage #creative #deep-work @work`
- 📂 **Task Categorization** — Classify by type (creative, technical, administrative, physical), effort (quick-win, medium, deep-work), and location (home, work, errands, anywhere)
- 🔗 **Cross-Reference Linking** — Link related tasks together; browse and navigate links from detail view

### Search & Commands
- 🔍 **Quick Search** — Press `/` for live task search with fuzzy filtering
- ⌨️ **Command Palette** — Press `:` for vi-style commands (see [full list below](#command-palette))

### Analytics & Insights
- 📊 **Session Metrics** — Automatic tracking of door selections, bypasses, and timing data
- 📈 **Daily Completion Tracking** — Track completions per day with streak counting
- 📋 **Insights Dashboard** — View trends, streaks, mood correlations, and avoidance patterns (`:dashboard`)
- 😊 **Mood Correlation Analysis** — Discover how your emotional state affects task selection
- 🚨 **Avoidance Detection** — Tasks bypassed 10+ times trigger an intervention prompt offering breakdown, deferral, or archival
- 🧠 **Pattern Analysis** — Identifies door position bias, task type preferences, and procrastination patterns

### Integrations & Providers
- 📄 **Text File** (default) — YAML-based local task storage
- 🍎 **Apple Notes** — Bidirectional sync with Apple Notes
- 📋 **Apple Reminders** — Sync tasks from macOS Reminders lists (macOS only)
- 🔵 **Jira** — Pull tasks from Jira Cloud/Server via REST API with JQL filtering
- 🐙 **GitHub Issues** — Import issues from GitHub repositories
- 💎 **Obsidian** — Read tasks from Obsidian vault markdown files with daily notes support
- ✅ **Todoist** — Sync tasks from Todoist via REST API with project and filter support
- 🔌 **Multi-Provider Aggregation** — Run multiple providers simultaneously; tasks merge across sources
- 🩺 **Health Check** — Run `:health` or `threedoors health` to verify provider connectivity

### Sync & Offline-First
- 💾 **Write-Ahead Log (WAL)** — Crash-safe task persistence with atomic writes
- 📡 **Offline Queue** — Local change queue with replay when connectivity returns
- 🔄 **Sync Status Indicator** — Visual sync state per provider in the TUI
- 🔀 **Conflict Resolution** — Duplicate detection and merge UI for multi-provider conflicts

### Calendar Awareness
- 📅 **Local Calendar Reader** — Reads from macOS Calendar.app (AppleScript), `.ics` files, and CalDAV caches
- ⏰ **Free Block Detection** — Computes available time blocks between calendar events

### Enrichment Database
- 🗃️ **SQLite Storage** — Pure-Go SQLite (no CGO) for task metadata, cross-references, learning patterns, and feedback history
- 🕸️ **Cross-Reference Graph** — Track relationships between tasks across providers

### LLM Task Decomposition
- 🤖 **Task Breakdown** — Decompose complex tasks into stories using Claude or local Ollama
- 📝 **Git Integration** — Write generated story specs directly to git repos
- 💡 **Suggestions View** — Browse LLM-generated task proposals in the TUI (`:suggestions`)

### Themes
- 🎨 **Door Themes** — Four built-in themes: `classic`, `modern`, `scifi`, `shoji`
- 🖌️ **Theme Picker** — Switch themes live with `:theme`
- ⚙️ **Persistent Selection** — Chosen theme saved in `config.yaml`

### Task Workflow
- ⏰ **Snooze / Defer** — Snooze a task until a specific date; auto-returns to the pool when due
- 🔗 **Task Dependencies** — Define `depends_on` relationships; blocked tasks are filtered from door selection and auto-unblock on completion
- ↩️ **Undo Completion** — Reverse accidental completions (`complete → todo`)

### User Experience
- 👋 **First-Run Onboarding** — Guided welcome flow with keybinding tutorial, values/goals setup, and optional task import
- 🎯 **Values & Goals Display** — Persistent footer showing your values as you work
- 😊 **Mood Logging** — Capture emotional state anytime with presets: focused, energized, tired, stressed, neutral, calm, distracted
- 💬 **Door Feedback** — Rate doors as blocked, not-now, or needs-breakdown to improve selection
- 💡 **Session Improvement Prompt** — On quit, optionally share improvement suggestions
- ➡️ **Contextual Next Steps** — After completing or adding a task, see relevant next actions
- ❓ **Keybinding Display** — Context-sensitive keybinding bar at the bottom of the screen; press `?` to open full keybinding overlay

### MCP Server 🤖
- 🔌 **Model Context Protocol** — Expose tasks and analytics to LLM agents via MCP
- 📡 **Dual Transport** — stdio (default) or SSE (`--transport sse --port 8080`)
- 🛠️ **15 MCP Tools** — Query, search, analyze tasks, traverse dependency graphs, assess burnout risk, and more (see [MCP section](#mcp-server))

### Distribution
- 🍺 **Homebrew** — Install via `brew install arcaven/tap/threedoors`
- 🔏 **Signed & Notarized** — macOS binaries are code-signed and Apple-notarized
- 💻 **Cross-Platform Binaries** — Pre-built for macOS (ARM & Intel) and Linux (x86_64)
- 🚀 **GitHub Releases** — Automatic releases on every merge to main

---

## 📖 User Guide

### The Three Doors Concept

When you launch ThreeDoors, you see three randomly selected tasks presented as "doors." This constraint is intentional — research in choice architecture shows that limiting options reduces decision paralysis and increases follow-through.

**Your workflow:**
1. Look at your three doors
2. Pick the one that feels right — or press `s` to re-roll
3. Take action on the task (complete it, mark it blocked, start working on it, etc.)
4. Return to three new doors

### Adding Tasks

**In the TUI:**
```
:add Buy groceries #quick-win @errands
:add Design new landing page #creative #deep-work @work
:add --why Review Q3 budget                          # prompts for why it matters
```

**From the CLI:**
```bash
threedoors task add "Buy groceries" --type administrative --effort quick-win
threedoors task add "Deploy v2.0" --context "Blocked until staging tests pass"
cat tasks.txt | threedoors task add --stdin            # bulk import from stdin
```

Tags in the task text are parsed automatically:
- `#tag` → maps to type or effort categories (e.g., `#creative`, `#deep-work`, `#quick-win`)
- `@location` → maps to location (e.g., `@home`, `@work`, `@errands`, `@anywhere`)

### Managing Tasks

**From the task detail view** (after selecting a door):

| Key | Action | Description |
|-----|--------|-------------|
| `c` | Complete | Mark the task done |
| `i` | In Progress | Start working on it |
| `b` | Block | Mark blocked (prompts for reason) |
| `p` | Procrastinate | Defer the task for later |
| `e` | Expand | Break into subtasks |
| `f` | Fork | Clone/split into a variant |
| `r` | Rework | Flag for rework |
| `l` | Link | Link to another task |
| `x` | Cross-refs | Browse cross-references |

**From the CLI:**
```bash
threedoors task complete <id>
threedoors task edit <id> --text "Updated title"
threedoors task note <id> "Added deployment instructions"
threedoors task delete <id>
threedoors task unblock <id>
```

### Task Status Workflow

```
          ┌──────────┐
          │   todo   │
          └────┬─────┘
               │
    ┌──────────┼──────────┬──────────┐
    ▼          ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│blocked │ │  in    │ │deferred│ │archived│
│        │ │progress│ │        │ │        │
└───┬────┘ └───┬────┘ └───┬────┘ └────────┘
    │          │          │
    │          ▼          │
    │     ┌────────┐      │
    │     │  in    │      │
    │     │ review │      │
    │     └───┬────┘      │
    │         │           │
    ▼         ▼           ▼
    └────► complete ◄─────┘
```

### Mood Tracking

Log your mood anytime with `m` in the TUI or via the CLI:

```bash
threedoors mood set focused
threedoors mood history
```

**Available moods:** focused, energized, tired, stressed, neutral, calm, distracted

Mood data feeds into the insights dashboard, showing correlations between your emotional state and task completion patterns.

### Insights & Analytics

Access the analytics dashboard with `:dashboard` or `:insights` in the TUI, or from the CLI:

```bash
threedoors stats              # Session overview
threedoors stats --daily      # Daily breakdown
threedoors stats --weekly     # Weekly trends
threedoors stats --patterns   # Behavioral patterns
```

The dashboard shows:
- **Completion streaks** — Consecutive days with at least one completion
- **Mood correlations** — Which moods lead to the most productive sessions
- **Avoidance patterns** — Tasks you keep bypassing (10+ bypasses triggers an intervention)
- **Door position bias** — Whether you favor left, center, or right doors
- **Task type preferences** — Which categories you gravitate toward

### Configuring Providers

Edit `~/.threedoors/config.yaml` to configure task sources:

```yaml
schema_version: 2
theme: classic

providers:
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
      max_results: "50"
      poll_interval: 30s

  - name: github
    settings:
      owner: your-username
      repo: your-repo
      token: ghp_your_token

  - name: obsidian
    settings:
      vault_path: /path/to/vault
      tasks_folder: tasks
      file_pattern: "*.md"
      daily_notes: true
      daily_notes_folder: Daily
      daily_notes_heading: "## Tasks"

  - name: applenotes
    settings:
      note_title: ThreeDoors Tasks

  - name: reminders
    settings:
      lists: Work,Personal
      include_completed: false

  - name: todoist
    settings:
      api_token: your-todoist-api-token    # or set TODOIST_API_TOKEN env var
      project_ids: "111222333, 444555666"  # optional; filter to specific projects
      filter: "today | overdue"            # optional; Todoist filter query (mutually exclusive with project_ids)
      poll_interval: 30s                   # optional; default 30s
```

You can also manage config from the CLI:

```bash
threedoors config show
threedoors config get theme
threedoors config set theme modern
```

Multiple providers can run simultaneously — tasks are aggregated and deduplicated across all sources.

### Themes

Switch door themes with `:theme` in the TUI or by setting the `theme` key in config:

| Theme | Description |
|-------|-------------|
| `classic` | Traditional door styling |
| `modern` | Contemporary, clean design |
| `scifi` | Sci-fi / cyberpunk aesthetic |
| `shoji` | Japanese minimalist sliding doors |

---

## ⌨️ Key Bindings

### Three Doors View
| Key | Action |
|-----|--------|
| `a` / `Left` | Select left door |
| `w` / `Up` | Select center door |
| `d` / `Right` | Select right door |
| `Space` / `Enter` | Open the selected door |
| `s` / `Down` | Refresh doors (re-roll) |
| `n` | Send feedback on selected door |
| `/` | Open quick search |
| `:` | Open command palette |
| `m` | Log mood |
| `?` | Open keybinding overlay |
| `q` / `Ctrl+C` | Quit |

### Task Detail View
| Key | Action |
|-----|--------|
| `c` | Mark complete |
| `i` | Mark in progress |
| `b` | Mark blocked (prompts for reason) |
| `e` | Expand task (break down) |
| `f` | Fork task (clone/split) |
| `p` | Procrastinate (defer) |
| `r` | Flag for rework |
| `l` | Link to another task |
| `x` | Browse cross-references |
| `m` | Log mood |
| `z` | Snooze / defer task |
| `?` | Open keybinding overlay |
| `Esc` | Return to doors |

### Search Mode
| Key | Action |
|-----|--------|
| Type | Live filter tasks |
| `j` / `Down` | Next result |
| `k` / `Up` | Previous result |
| `Enter` | Open selected task |
| `Esc` | Exit search |

### Command Palette

| Command | Action |
|---------|--------|
| `:add <task>` | Add a new task |
| `:add --why` | Add task with context (why it matters) |
| `:mood [mood]` | Log mood (or open selector) |
| `:tag` | Open task categorization editor |
| `:theme` | Open theme picker |
| `:stats` | Flash session statistics |
| `:health` | Run system health check |
| `:dashboard` | Open insights dashboard |
| `:insights` | Show full insights dashboard |
| `:insights mood` | Flash mood & productivity insights |
| `:insights avoidance` | Flash avoidance patterns |
| `:goals` | Open values & goals setup |
| `:goals edit` | Edit existing values & goals |
| `:synclog` | Show sync history |
| `:suggestions` | Browse LLM task proposals |
| `:deferred` | Show deferred/snoozed tasks |
| `:devqueue` | Open dev dispatch queue |
| `:help` | Show all commands |
| `:quit` / `:exit` | Exit application |

---

## 💻 CLI Reference

ThreeDoors includes a full CLI for headless/scripted usage. All commands support `--json` for machine-readable output.

```
threedoors [command]
threedoors --version
```

### `task` — Task Management

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

### `doors` — Three Doors in the Terminal

```bash
threedoors doors                     # Show three random tasks
threedoors doors --pick 1            # Auto-select door 1 (1-3)
threedoors doors --interactive       # Prompted selection mode
```

### `mood` — Mood Tracking

```bash
threedoors mood set <mood>           # Record mood (focused, energized, tired, etc.)
threedoors mood history              # View mood entries
```

### `stats` — Productivity Analytics

```bash
threedoors stats                     # Session overview
threedoors stats --daily             # Daily breakdown
threedoors stats --weekly            # Weekly trends
threedoors stats --patterns          # Behavioral patterns
```

### `config` — Configuration

```bash
threedoors config show               # Display full configuration
threedoors config get <key>          # Get a single config value
threedoors config set <key> <value>  # Set a config value
```

### `health` — System Health Check

```bash
threedoors health                    # Check provider connectivity, file access, disk space
```

### `completion` — Shell Completions

```bash
threedoors completion bash           # Generate bash completions
threedoors completion zsh            # Generate zsh completions
threedoors completion fish           # Generate fish completions
```

---

## 🤖 MCP Server

ThreeDoors ships a separate MCP (Model Context Protocol) server binary that exposes tasks and analytics to LLM agents like Claude.

### Running the MCP Server

```bash
# stdio transport (default — for Claude Desktop, Cursor, etc.)
threedoors-mcp

# SSE transport (for web-based clients)
threedoors-mcp --transport sse --port 8080
```

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `query_tasks` | Query tasks with filters (status, type, effort, provider, text, date range) |
| `get_task` | Get full task details with enrichment data |
| `search_tasks` | Full-text search with relevance scoring |
| `list_providers` | List configured providers with health/sync status |
| `get_session` | Current or historical session metrics |
| `get_mood_correlation` | Mood vs. productivity correlation analysis |
| `get_productivity_profile` | Time-of-day productivity analysis |
| `burnout_risk` | Burnout risk assessment (0–1 score) |
| `walk_graph` | Traverse task relationship graph (BFS) |
| `find_paths` | Find paths between two tasks in the graph |
| `get_critical_path` | Longest dependency chain |
| `get_orphans` | Find tasks with no relationships |
| `get_completions` | Completion data with grouping options |
| `get_clusters` | Discover related task groups |
| `get_provider_overlap` | Find tasks shared between providers |

### Claude Desktop Configuration

Add to your Claude Desktop `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "threedoors": {
      "command": "threedoors-mcp",
      "args": []
    }
  }
}
```

---

## 📁 Data Directory

All data is stored locally in `~/.threedoors/`:

```
~/.threedoors/
├── tasks.yaml          # Active tasks (YAML format)
├── config.yaml         # Provider & theme configuration
├── values.json         # Your values & goals
├── completed.txt       # Completed task log
├── sessions.jsonl      # Session metrics (JSON Lines)
├── patterns.json       # Cached pattern analysis
├── enrichment.db       # SQLite enrichment database
├── proposals.jsonl     # LLM task decomposition proposals
├── synclog.jsonl       # Sync history log
├── improvements.txt    # Your improvement suggestions
└── onboarding.lock     # First-run marker
```

---

## 🔒 Data & Privacy

- **All data is local** — Stored in `~/.threedoors/`
- **No telemetry** — Session metrics stay on your machine
- **No accounts** — No sign-ups, no servers, no tracking
- **Offline-first** — Works without network; syncs when available
- **Your API tokens stay local** — Provider credentials in `config.yaml` are never transmitted beyond the configured service

---

## 🧭 Philosophy

1. **Progress Over Perfection** — Taking action on imperfect tasks beats perfect planning
2. **Reduce Friction** — Every interaction should feel effortless
3. **Learn from Behavior** — Track patterns to help users understand their work habits
4. **Emotional Context Matters** — Mood affects productivity; acknowledge and track it
5. **Power Users Welcome** — Vi-style commands without sacrificing simplicity
6. **Local-First** — Your data stays on your machine, no accounts, no telemetry

---

## 🛠️ Development

### Tech Stack

- **Language:** Go 1.25.4+
- **TUI Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Database:** [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO)
- **Architecture:** Model-View-Update (MVU) with provider pattern
- **Build System:** Make
- **CI/CD:** GitHub Actions (lint, test, build, sign, notarize, release, Homebrew update)

### Project Structure

```
ThreeDoors/
├── cmd/
│   ├── threedoors/              # TUI + CLI entry point
│   └── threedoors-mcp/          # MCP server entry point
├── internal/
│   ├── adapters/                # Provider implementations
│   │   ├── applenotes/          #   Apple Notes adapter
│   │   ├── github/              #   GitHub Issues adapter
│   │   ├── jira/                #   Jira adapter
│   │   ├── obsidian/            #   Obsidian adapter
│   │   ├── reminders/           #   Apple Reminders adapter
│   │   ├── textfile/            #   YAML file adapter
│   │   └── todoist/             #   Todoist adapter
│   ├── cli/                     # CLI command handling
│   ├── core/                    # Task domain: models, status, config, registry
│   ├── tui/                     # Bubbletea views (20 views) and UI components
│   │   └── themes/              # Door theme implementations
│   ├── calendar/                # Calendar readers (AppleScript, ICS, CalDAV)
│   ├── dispatch/                # Dev dispatch system
│   ├── enrichment/              # SQLite enrichment database
│   ├── intelligence/            # LLM backends (Claude, Ollama)
│   ├── mcp/                     # MCP protocol and tools
│   ├── dist/                    # macOS code signing, notarization
│   └── ci/                      # CI validation tests
├── Formula/                     # Homebrew formula
├── scripts/                     # Analysis & build scripts
├── docs/                        # PRD, architecture, stories, research
└── Makefile
```

### Make Targets

```bash
make build          # Build the application (TUI + MCP)
make run            # Build and run
make test           # Run tests
make lint           # Run golangci-lint
make fmt            # Format with gofumpt
make clean          # Remove build artifacts
make sign           # Code-sign binary (requires APPLE_SIGNING_IDENTITY)
make pkg            # Build macOS .pkg installer
make release-local  # Build + sign + pkg
```

### Code Style

We use `gofumpt` (stricter than `gofmt`) and `golangci-lint`. See [CLAUDE.md](CLAUDE.md) for full coding standards.

```bash
make fmt    # Format code
make lint   # Run linter (must pass with zero warnings)
```

---

## 🤝 Contributing

**Before contributing:**
1. Read the [PRD](docs/prd/index.md) and [Architecture](docs/architecture/index.md) docs
2. Check current status in the [epic list](docs/prd/epic-list.md)
3. Open an issue to discuss significant changes

**To contribute:**
1. Fork the project
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Follow coding standards (`make lint && make fmt`)
4. Write tests for new functionality
5. Commit your changes
6. Push and open a Pull Request

**Code Quality Requirements:**
- `gofumpt` formatting
- `golangci-lint` passes with zero warnings
- Unit tests for new logic
- No `//nolint` without justification

---

## 📚 Documentation

- **[Product Requirements (PRD)](docs/prd/index.md)** — Features, requirements, epics
- **[Architecture](docs/architecture/index.md)** — Technical design and patterns
- **[User Stories](docs/stories/)** — Story files with acceptance criteria
- **[Coding Standards](docs/architecture/coding-standards.md)** — Go best practices
- **[Research](_bmad-output/planning-artifacts/)** — Choice architecture, mood correlation, procrastination

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Acknowledgments

Built with the [Charm](https://charm.sh/) ecosystem:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components

## Links

- **Repository:** [github.com/arcaven/ThreeDoors](https://github.com/arcaven/ThreeDoors)
- **Issues:** [github.com/arcaven/ThreeDoors/issues](https://github.com/arcaven/ThreeDoors/issues)
- **Releases:** [github.com/arcaven/ThreeDoors/releases](https://github.com/arcaven/ThreeDoors/releases)

---

**"Progress over perfection. Three doors. One choice. Move forward."** 🚪✨
