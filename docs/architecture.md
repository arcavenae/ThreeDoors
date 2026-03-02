# ThreeDoors Architecture Summary

> This is a consolidated architecture overview. For detailed specifications, see the sharded architecture docs in [docs/architecture/](./architecture/index.md).

## Executive Summary

ThreeDoors is a **monolithic CLI/TUI application** built in Go 1.25.4 using the Bubbletea framework. It implements a novel task management interface that presents three randomly selected tasks as "doors" to reduce decision friction. The architecture follows a **two-layer design** (TUI + Domain) using the **Model-View-Update (MVU)** pattern enforced by Bubbletea.

## Architecture Style

- **Type:** Simple Monolithic CLI Application
- **Pattern:** Layered Architecture (2 layers) + MVU (Elm Architecture)
- **No external services** — single Go binary, local file storage only

## Layer Diagram

```
┌─────────────────────────────────────────┐
│           cmd/threedoors/main.go        │  Entry Point
├─────────────────────────────────────────┤
│         TUI Layer (internal/tui)        │  Bubbletea MVU
│  ┌──────────┐ ┌────────────┐ ┌───────┐ │
│  │DoorsView │ │DetailView  │ │Status │ │
│  │          │ │            │ │Menu   │ │
│  └──────────┘ └────────────┘ └───────┘ │
├─────────────────────────────────────────┤
│       Domain Layer (internal/tasks)     │  Business Logic
│  ┌──────────┐ ┌────────────┐ ┌───────┐ │
│  │TaskPool  │ │FileManager │ │Status │ │
│  │          │ │ (YAML I/O) │ │Mgr    │ │
│  └──────────┘ └────────────┘ └───────┘ │
├─────────────────────────────────────────┤
│         ~/.threedoors/                  │  Filesystem
│    tasks.yaml    completed.txt          │
└─────────────────────────────────────────┘
```

## Technology Stack

| Category | Technology | Version |
|---|---|---|
| Language | Go | 1.25.4 |
| TUI Framework | Bubbletea | 1.2.4 |
| Styling | Lipgloss | 1.0.0 |
| Components | Bubbles | 0.20.0 |
| Data Format | YAML (gopkg.in/yaml.v3) | 3.0.1 |
| IDs | UUID v4 (google/uuid) | 1.6.0 |
| Build | Make | System |

## Key Components

### TUI Layer (`internal/tui`)
- **MainModel** — Root view router, global state orchestrator
- **DoorsView** — Three doors display, navigation (A/W/D or arrow keys)
- **TaskDetailView** — Full task details, status/notes management
- **StatusUpdateMenu** — Status transition selector with validation
- **NotesInputView** — Multi-line text input (Bubbles textarea)

### Domain Layer (`internal/tasks`)
- **TaskPool** — In-memory task collection with filtering and recently-shown tracking
- **FileManager** — YAML file I/O with atomic writes
- **StatusManager** — Task status state machine validation
- **DoorSelector** — Fisher-Yates shuffle selection algorithm

## Data Architecture

- **Storage:** YAML flat files at `~/.threedoors/tasks.yaml`
- **Completed log:** `~/.threedoors/completed.txt` (append-only)
- **No database** — intentionally simple for Technical Demo phase
- **Atomic writes:** All file operations use write-tmp/sync/rename pattern

## Task Status State Machine

```
todo → in-progress → in-review → complete
todo → blocked → in-progress
in-progress → blocked
Any state → complete (force)
```

## Key Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Architecture | Two-layer monolith | Minimal separation without over-engineering |
| UI Pattern | MVU (Elm Architecture) | Bubbletea framework requirement |
| Repository Pattern | Deferred to Epic 2 | YAGNI — only one data source now |
| DI | Constructor injection | Testability without framework overhead |
| Error Handling | Idiomatic Go (errors as values) | Standard Go practice |
| File Writes | Atomic write pattern | Prevent data corruption |

## Detailed Architecture Docs

For full specifications, see the sharded docs:

- [Introduction](./architecture/introduction.md)
- [High-Level Architecture](./architecture/high-level-architecture.md)
- [Tech Stack](./architecture/tech-stack.md)
- [Components](./architecture/components.md)
- [Core Workflows](./architecture/core-workflows.md) (6 mermaid sequence diagrams)
- [Data Models](./architecture/data-models.md)
- [Data Storage Schema](./architecture/data-storage-schema.md)
- [Source Tree](./architecture/source-tree.md)
- [Coding Standards](./architecture/coding-standards.md)
- [Test Strategy](./architecture/test-strategy-and-standards.md)
- [Error Handling](./architecture/error-handling-strategy.md)
- [Security](./architecture/security.md)
- [Infrastructure & Deployment](./architecture/infrastructure-and-deployment.md)
