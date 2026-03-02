# ThreeDoors Project Overview

## What Is ThreeDoors?

ThreeDoors is a terminal-based task management application that reduces decision friction by showing you only **three tasks at a time**. Instead of overwhelming you with an endless list, ThreeDoors presents three randomly selected "doors" — choose one, take action, and move forward.

**Core philosophy:** Progress over perfection.

## Project Identity

| Field | Value |
|---|---|
| **Name** | ThreeDoors |
| **Type** | CLI / TUI Application |
| **Language** | Go 1.25.4 |
| **Framework** | Bubbletea (Charm ecosystem) |
| **Architecture** | Monolith, Model-View-Update (MVU) |
| **License** | MIT |
| **Repository** | github.com/arcaven/ThreeDoors |
| **Data** | Local-only (`~/.threedoors/`) |
| **Platform** | macOS 14+ (Sonoma) |

## Current Status

**Phase:** Technical Demo & Validation (Epic 1)

| Story | Status |
|---|---|
| 1.1: Project Setup & Basic Bubbletea App | Completed |
| 1.2: Display Three Doors from Task File | Completed |
| 1.3: Door Selection & Task Status Management | In Progress |
| 1.3a: Quick Search & Command Palette | In Progress |
| 1.5: Session Metrics Tracking | Upcoming |
| 1.6: Essential Polish | Upcoming |

**Note:** Source code has not been committed to the `main` branch yet. The repository currently contains planning artifacts only.

## Key Features (Planned)

- Three doors display with random task selection
- Task detail view with status management (todo/blocked/in-progress/in-review/complete)
- Progress notes and blocker tracking
- Quick search (`/`) with live filtering
- Command palette (`:`) for vi-style commands
- Mood tracking and session metrics
- Local-first data storage (YAML)

## Roadmap

| Phase | Focus | Timeline |
|---|---|---|
| Phase 1 | Technical Demo & Validation (Epic 1) | Week 1 |
| Phase 2 | Apple Notes Integration | Post-validation |
| Phase 3 | Enhanced Interaction & Learning | Future |
| Phase 4 | Intelligent Door Selection | Future |

## Documentation Map

| Document | Purpose |
|---|---|
| [Architecture Summary](./architecture.md) | Consolidated architecture overview |
| [Architecture (detailed)](./architecture/index.md) | Full sharded architecture (19 files) |
| [PRD](./prd/index.md) | Product requirements (10 files) |
| [Product Brief](./brief.md) | Executive product brief |
| [Development Guide](./development-guide.md) | Setup, build, test, coding standards |
| [Source Tree Analysis](./source-tree-analysis.md) | Directory structure and critical folders |
| [Technology Stack](./technology-stack.md) | Full tech stack table |
| [Stories](./stories/) | Implementation stories (1.1, 1.2) |
