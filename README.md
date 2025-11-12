# ThreeDoors 🚪🚪🚪

[![Go Version](https://img.shields.io/badge/Go-1.25.4+-00ADD8?style=flat&logo=go)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Built with Bubbletea](https://img.shields.io/badge/Built%20with-Bubbletea-purple)](https://github.com/charmbracelet/bubbletea)

## What is ThreeDoors?

ThreeDoors is a **radical rethinking of task management** that reduces decision friction by showing you only **three tasks at a time**. Instead of overwhelming you with an endless list, ThreeDoors presents three carefully selected "doors"—choose one, take action, and move forward. It's built on the philosophy that **progress over perfection** matters more than perfect planning.

### The Problem

Traditional task lists create **choice paralysis**. Staring at 50+ tasks makes it hard to start anything. You spend more time reorganizing and re-prioritizing than actually doing the work.

### The ThreeDoors Solution

- **Three doors, one choice** - Reduces cognitive load by limiting options
- **Refresh when needed** - Don't like your options? Roll again (press `s`)
- **Quick search** - Need something specific? Press `/` to search
- **Mood-aware tracking** - Log your emotional state to understand work patterns
- **Pattern learning** - Over time, learn which tasks you avoid and why

## Project Status

**Phase:** Technical Demo & Validation (Epic 1)
**Current Milestone:** Story 1.2 - Display Three Doors from Task File
**Goal:** Build and validate core Three Doors UX concept within 1 week

### Completed
- ✅ Story 1.1: Project Setup & Basic Bubbletea App
- ✅ Story 1.2: File I/O and Three Doors Display

### In Progress
- 🔄 Story 1.3: Door Selection & Task Status Management
- 🔄 Story 1.3a: Quick Search & Command Palette

### Upcoming
- Story 1.5: Session Metrics Tracking
- Story 1.6: Essential Polish

## Features

### Core Functionality
- 🚪 **Three Doors Display** - View three randomly selected tasks at once
- 🔄 **Refresh Mechanism** - Re-roll doors when nothing appeals (press `s` or down arrow)
- ✅ **Task Status Management** - Complete, block, defer, expand, fork, or flag tasks for rework
- 🔍 **Quick Search** - Press `/` for live task search with bottom-up results
- ⌨️ **Command Palette** - Press `:` for vi-style commands (`:add`, `:mood`, `:stats`, `:help`, etc.)
- 😊 **Mood Tracking** - Log your emotional state anytime (press `m`)
- 📊 **Session Metrics** - Automatic tracking of door selections, bypasses, and patterns
- 🎨 **Beautiful TUI** - Built with Bubbletea and Lipgloss for a polished terminal experience

### Key Bindings

#### Three Doors View
| Key | Action |
|-----|--------|
| `a` / `←` | Select left door |
| `w` / `↑` | Select center door |
| `d` / `→` | Select right door |
| `s` / `↓` | Refresh doors (re-roll) |
| `/` | Open quick search |
| `m` | Log mood/context |
| `q` / `Ctrl+C` | Quit application |

#### Task Detail View (After Selecting Door)
| Key | Action |
|-----|--------|
| `c` | Mark as complete |
| `b` | Mark as blocked |
| `i` | Mark as in progress |
| `e` | Expand task (break down) |
| `f` | Fork task (clone/split) |
| `p` | Procrastinate (defer) |
| `r` | Flag for rework |
| `Esc` | Return to previous screen |

#### Search Mode (`/`)
| Key | Action |
|-----|--------|
| Type | Live filter tasks |
| `j` / `↓` / `s` | Move down in results |
| `k` / `↑` / `w` | Move up in results |
| `Enter` | Open selected task |
| `Esc` | Exit search (return to doors) |

#### Command Mode (`:`)
| Command | Action |
|---------|--------|
| `:add <task>` | Add new task |
| `:mood [mood]` | Quick mood log |
| `:stats` | Show session statistics |
| `:help` | Display all commands |
| `:quit` | Exit application |

## Getting Started

### Prerequisites

* **Go** 1.25.4 or higher ([installation guide](https://golang.org/doc/install))
* **Git**
* **Make** (optional, for build automation)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/arcaven/ThreeDoors.git
   cd ThreeDoors
   ```

2. Build the application:
   ```bash
   make build
   # or without Make:
   go build -o threedoors ./cmd/threedoors
   ```

3. Run the application:
   ```bash
   make run
   # or directly:
   ./threedoors
   ```

### Quick Start Usage

1. **First run** creates `~/.threedoors/tasks.txt` with sample tasks
2. **Add your tasks** by editing `~/.threedoors/tasks.txt` (one task per line)
3. **Launch ThreeDoors** and see three tasks displayed as doors
4. **Select a door** (a/w/d or arrow keys) to view task details
5. **Take action** on the task (complete, block, defer, etc.)
6. **Refresh** (press `s`) if you don't like your options
7. **Search** (press `/`) when you need a specific task

## Development

### Tech Stack

- **Language:** Go 1.25.4
- **TUI Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) (The Elm Architecture for Go)
- **Styling:** [Lipgloss](https://github.com/charmbracelet/lipgloss) (Style definitions for TUI)
- **Architecture:** Model-View-Update (MVU) pattern
- **Build System:** Make

### Project Structure

```
ThreeDoors/
├── cmd/threedoors/          # Application entry point
│   ├── main.go
│   └── main_test.go
├── internal/
│   └── tasks/               # Task domain logic
│       ├── task.go          # Task model
│       ├── file_manager.go  # File I/O
│       └── session_tracker.go  # Metrics tracking
├── docs/
│   ├── prd/                 # Product Requirements
│   ├── architecture/        # Architecture docs
│   └── stories/             # User stories
└── Makefile                 # Build automation
```

### Running Tests

```bash
make test
# or
go test ./...
```

### Code Style and Linting

We use `gofumpt` (stricter than `go fmt`) and `golangci-lint`:

```bash
# Format code
gofumpt -w .

# Run linter
golangci-lint run ./...

# Or use Make targets
make lint
make fmt
```

### Make Targets

```bash
make build    # Build the application
make run      # Run the application
make test     # Run tests
make clean    # Remove build artifacts
make lint     # Run linter
make fmt      # Format code
```

## Philosophy & Design Principles

ThreeDoors is built on several core principles:

1. **Progress Over Perfection** - Taking action on imperfect tasks beats perfect planning
2. **Reduce Friction** - Every interaction should feel effortless and natural
3. **Learn from Behavior** - Track patterns to help users understand their work habits
4. **Emotional Context Matters** - Mood affects productivity; acknowledge and track it
5. **Power Users Welcome** - Vi-style commands for efficiency without sacrificing simplicity
6. **Local-First** - Your data stays on your machine (`~/.threedoors/`)

## Data & Privacy

- **All data is local** - Tasks stored in `~/.threedoors/tasks.txt`
- **No telemetry** - Session metrics stay on your machine (`~/.threedoors/sessions.jsonl`)
- **No accounts** - No sign-ups, no servers, no tracking
- **Plain text** - Your tasks are readable and portable

## Future Roadmap

### Phase 2: Apple Notes Integration (Post-Validation)
- Sync tasks with Apple Notes
- Bidirectional updates (edit in Notes or ThreeDoors)
- iCloud sync across devices

### Phase 3: Enhanced Interaction & Learning
- Interactive prompts and guidance
- Values and goals display
- Daily completion tracking
- Continuous improvement prompts

### Phase 4: Intelligent Door Selection
- **Mood correlation analysis** - "When stressed, you avoid complex tasks"
- **Pattern recognition** - Identify tasks you consistently skip
- **Adaptive selection** - Show appropriate tasks based on current mood
- **Goal re-evaluation** - Prompt when persistent avoidance detected

## Contributing

We welcome contributions! ThreeDoors is currently in **Technical Demo phase** (Epic 1).

**Before contributing:**
1. Read the [PRD](docs/prd/index.md) and [Architecture](docs/architecture/index.md) docs
2. Check current milestone in [Project Status](#project-status)
3. Open an issue to discuss significant changes

**To contribute:**
1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Follow coding standards (run `make lint` and `make fmt`)
4. Write tests for new functionality
5. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
6. Push to the Branch (`git push origin feature/AmazingFeature`)
7. Open a Pull Request

**Code Quality Requirements:**
- `gofumpt` formatting
- `golangci-lint` passes
- Unit tests for new logic
- Documentation for public APIs

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Product Requirements Document (PRD)](docs/prd/index.md)** - Features, requirements, and epic details
- **[Architecture Documentation](docs/architecture/index.md)** - Technical design, patterns, and standards
- **[User Stories](docs/stories/)** - Detailed story files with acceptance criteria
- **[Coding Standards](docs/architecture/coding-standards.md)** - Go best practices for this project
- **[Tech Stack](docs/architecture/tech-stack.md)** - Dependencies and tooling

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Acknowledgments

Built with the excellent [Charm](https://charm.sh/) ecosystem:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## Project Links

- **Repository:** [https://github.com/arcaven/ThreeDoors](https://github.com/arcaven/ThreeDoors)
- **Issues:** [https://github.com/arcaven/ThreeDoors/issues](https://github.com/arcaven/ThreeDoors/issues)
- **Discussions:** [https://github.com/arcaven/ThreeDoors/discussions](https://github.com/arcaven/ThreeDoors/discussions)

---

**"Progress over perfection. Three doors. One choice. Move forward."** 🚪✨
