# ThreeDoors

[![Go Version](https://img.shields.io/badge/Go-1.25.4+-00ADD8?style=flat&logo=go)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Built with Bubbletea](https://img.shields.io/badge/Built%20with-Bubbletea-purple)](https://github.com/charmbracelet/bubbletea)

## What is ThreeDoors?

ThreeDoors is a **radical rethinking of task management** that reduces decision friction by showing you only **three tasks at a time**. Instead of overwhelming you with an endless list, ThreeDoors presents three carefully selected "doors" — choose one, take action, and move forward.

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

## Features

### Core Task Management
- **Three Doors Display** — View three randomly selected tasks, avoiding recently shown ones
- **Refresh Mechanism** — Re-roll doors when nothing appeals
- **Task Status Workflow** — Complete, block, defer, expand, fork, or flag tasks for rework
- **Quick Add** — Add tasks inline with `:add` or press `a`; supports context capture with `:add --why`
- **Inline Tagging** — Tag tasks as you add them: `Design homepage #creative #deep-work @work`
- **Task Categorization** — Classify by type (creative, technical, administrative, physical), effort (quick-win, medium, deep-work), and location (home, work, errands, anywhere)
- **Cross-Reference Linking** — Link related tasks together; browse and navigate links from detail view

### Search & Commands
- **Quick Search** — Press `/` for live task search with fuzzy filtering
- **Command Palette** — Press `:` for vi-style commands (`:add`, `:mood`, `:stats`, `:health`, `:dashboard`, `:goals`, `:help`)

### Analytics & Insights
- **Session Metrics** — Automatic tracking of door selections, bypasses, and timing data
- **Daily Completion Tracking** — Track completions per day with streak counting
- **Insights Dashboard** — View trends, streaks, mood correlations, and avoidance patterns (`:dashboard`)
- **Mood Correlation Analysis** — Discover how your emotional state affects task selection
- **Avoidance Detection** — Tasks bypassed 10+ times trigger an intervention prompt offering breakdown, deferral, or archival
- **Pattern Analysis** — Identifies door position bias, task type preferences, and procrastination patterns

### Apple Notes Integration
- **Bidirectional Sync** — Read and write tasks from Apple Notes
- **Provider Architecture** — Switch between text file and Apple Notes backends via `config.yaml`
- **Health Check** — Run `:health` to verify provider connectivity, file access, and disk space

### Sync & Offline-First
- **Write-Ahead Log (WAL)** — Crash-safe task persistence with atomic writes
- **Offline Queue** — Local change queue with replay when connectivity returns
- **Sync Status Indicator** — Visual sync state per provider in the TUI

### Calendar Awareness
- **Local Calendar Reader** — Reads from macOS Calendar.app (AppleScript), `.ics` files, and CalDAV caches
- **Free Block Detection** — Computes available time blocks between calendar events

### Enrichment Database
- **SQLite Storage** — Pure-Go SQLite (no CGO) for task metadata, cross-references, learning patterns, and feedback history
- **Cross-Reference Graph** — Track relationships between tasks across providers

### LLM Task Decomposition (Spike)
- **Task Breakdown** — Decompose complex tasks into stories using Claude or local Ollama
- **Git Integration** — Write generated story specs directly to git repos

### User Experience
- **First-Run Onboarding** — Guided welcome flow with keybinding tutorial, values/goals setup, and optional task import
- **Values & Goals Display** — Persistent footer showing your values as you work
- **Mood Logging** — Capture emotional state anytime with presets or custom text
- **Door Feedback** — Rate doors as blocked, not-now, or needs-breakdown to improve selection
- **Session Improvement Prompt** — On quit, optionally share improvement suggestions
- **Contextual Next Steps** — After completing or adding a task, see relevant next actions

### Distribution
- **Homebrew** — Install via `brew install arcaven/tap/threedoors`
- **Signed & Notarized** — macOS binaries are code-signed and Apple-notarized
- **Cross-Platform Binaries** — Pre-built for macOS (ARM & Intel) and Linux (x86_64)
- **GitHub Releases** — Automatic releases on every merge to main

## Key Bindings

### Three Doors View
| Key | Action |
|-----|--------|
| `a` / `Left` | Select left door |
| `w` / `Up` | Select center door |
| `d` / `Right` | Select right door |
| `s` / `Down` | Refresh doors (re-roll) |
| `n` | Send feedback on selected door |
| `/` | Open quick search |
| `:` | Open command palette |
| `m` | Log mood |
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
| `:stats` | Flash session statistics |
| `:health` | Run system health check |
| `:dashboard` | Open insights dashboard |
| `:insights` | Show full insights dashboard |
| `:insights mood` | Flash mood & productivity insights |
| `:insights avoidance` | Flash avoidance patterns |
| `:goals` | Open values & goals setup |
| `:goals edit` | Edit existing values & goals |
| `:help` | Show all commands |
| `:quit` | Exit application |

## Getting Started

### Option 1: Homebrew (macOS)

```bash
brew install arcaven/tap/threedoors
```

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

### Usage

1. **Launch** the app:
   ```bash
   threedoors
   ```
2. **First run** starts the onboarding wizard — learn key bindings, set your values/goals, and optionally import existing tasks.
3. **Select a door** with `a` (left), `w` (center), or `d` (right).
4. **Re-roll** doors with `s` if nothing appeals.
5. **Act on a task** using status keys: `c` (complete), `b` (blocked), `i` (in progress), `p` (procrastinate).
6. **Add tasks** with `:add Buy groceries #quick-win @errands`.
7. **Log your mood** with `m`.
8. **Search** with `/` to find a specific task.
9. **View insights** with `:dashboard` to see trends and patterns.

### Data Directory

All data is stored locally in `~/.threedoors/`:

```
~/.threedoors/
├── tasks.yaml          # Active tasks (YAML format)
├── config.yaml         # Provider configuration
├── values.json         # Your values & goals
├── completed.txt       # Completed task log
├── sessions.jsonl      # Session metrics (JSON Lines)
├── patterns.json       # Cached pattern analysis
├── enrichment.db       # SQLite enrichment database
├── improvements.txt    # Your improvement suggestions
└── onboarding.lock     # First-run marker
```

## Development

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
├── cmd/threedoors/           # Entry point
├── internal/
│   ├── tasks/                # Task domain: models, providers, sync, analytics
│   ├── tui/                  # Bubbletea views (13 views) and UI components
│   ├── calendar/             # Local calendar readers (AppleScript, ICS, CalDAV)
│   ├── enrichment/           # SQLite enrichment database
│   ├── intelligence/llm/     # LLM backends (Claude, Ollama) & task decomposition
│   ├── dist/                 # macOS code signing, notarization, pkg building
│   └── ci/                   # CI validation tests
├── Formula/                  # Homebrew formula
├── scripts/                  # Analysis & build scripts
├── docs/                     # PRD, architecture, stories, research
└── Makefile
```

### Make Targets

```bash
make build          # Build the application
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

## Philosophy

1. **Progress Over Perfection** — Taking action on imperfect tasks beats perfect planning
2. **Reduce Friction** — Every interaction should feel effortless
3. **Learn from Behavior** — Track patterns to help users understand their work habits
4. **Emotional Context Matters** — Mood affects productivity; acknowledge and track it
5. **Power Users Welcome** — Vi-style commands without sacrificing simplicity
6. **Local-First** — Your data stays on your machine, no accounts, no telemetry

## Data & Privacy

- **All data is local** — Stored in `~/.threedoors/`
- **No telemetry** — Session metrics stay on your machine
- **No accounts** — No sign-ups, no servers, no tracking
- **Offline-first** — Works without network; syncs when available

## Contributing

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

## Documentation

- **[Product Requirements (PRD)](docs/prd/index.md)** — Features, requirements, epics
- **[Architecture](docs/architecture/index.md)** — Technical design and patterns
- **[User Stories](docs/stories/)** — Story files with acceptance criteria
- **[Coding Standards](docs/architecture/coding-standards.md)** — Go best practices
- **[Research](docs/research/)** — Choice architecture, mood correlation, procrastination

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

**"Progress over perfection. Three doors. One choice. Move forward."**
