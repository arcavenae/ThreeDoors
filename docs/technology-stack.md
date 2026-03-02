# ThreeDoors Technology Stack

## Stack Summary

- **Language:** Go 1.25.4
- **Architecture Pattern:** Model-View-Update (MVU / Elm Architecture)
- **Application Type:** CLI / Terminal UI (TUI)
- **Infrastructure:** Local-only, no cloud services

## Technology Table

| Category | Technology | Version | Purpose |
|---|---|---|---|
| Language | Go | 1.25.4 | Primary development language |
| TUI Framework | Bubbletea | 1.2.4 | Elm-architecture terminal UI framework |
| TUI Styling | Lipgloss | 1.0.0 | ANSI terminal styling and layout |
| TUI Components | Bubbles | 0.20.0 | Pre-built TUI components (text input, list) |
| Terminal Utilities | golang.org/x/term | 0.26.0 | Terminal size detection |
| YAML Parser | gopkg.in/yaml.v3 | 3.0.1 | Task file parsing |
| UUID Generator | github.com/google/uuid | 1.6.0 | Unique task IDs |
| Code Formatting | gofumpt | 0.7.0 | Strict code formatting |
| Linting | golangci-lint | 1.61.0 | Static analysis |
| Build System | Make | System default | Build automation |
| Dependency Mgmt | Go Modules | 1.25.4 | Package management |
| Testing | Go testing (stdlib) | 1.25.4 | Unit testing |
| Storage | YAML flat files | N/A | ~/.threedoors/tasks.yaml |
| Platform | macOS | 14+ (Sonoma) | Target OS |
| Version Control | Git | 2.40+ | Source control |

## Architecture Pattern

**Model-View-Update (MVU)** — The Elm Architecture adapted for Go via Bubbletea:

- **Model**: Application state structs (task list, current view, selections)
- **Update**: Message-driven state transitions (key presses, file I/O results)
- **View**: Pure functions rendering state to terminal output via Lipgloss

## Source Code Status

**No source code files currently exist in the repository.** The project has detailed planning artifacts (PRD, architecture, stories) but Go source code has not been committed to the `main` branch. The planned structure is:

```
cmd/threedoors/main.go          — Entry point
internal/tasks/task.go          — Task model
internal/tasks/file_manager.go  — File I/O
internal/tasks/session_tracker.go — Metrics
```

## Data Architecture

- **Task Storage:** YAML files at `~/.threedoors/tasks.yaml`
- **Session Metrics:** JSONL at `~/.threedoors/sessions.jsonl`
- **No database** — flat file storage only
