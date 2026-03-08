# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for the ThreeDoors project. ADRs capture significant architectural and design decisions made throughout the project's evolution from initial tech demo through 213+ merged PRs across 24 completed epics.

## What is an ADR?

An Architecture Decision Record (ADR) captures a single architectural decision, including the context that motivated it, the decision itself, and its consequences. ADRs provide a decision log that helps future contributors understand why the system is built the way it is.

## ADR Format

Each ADR follows a consistent structure:
- **Status** — Accepted, Deprecated, or Superseded
- **Date** — When the decision was made
- **Context** — The problem or situation that prompted the decision
- **Considered Options** — Alternatives evaluated (when applicable)
- **Decision** — What was decided
- **Consequences** — Positive and negative impacts

## ADR Index

### Foundation & Architecture

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0001](ADR-0001-go-as-primary-language.md) | Go as Primary Language | Accepted | 2025-11-07 |
| [ADR-0002](ADR-0002-bubbletea-tui-framework.md) | Bubbletea TUI Framework | Accepted | 2025-11-07 |
| [ADR-0003](ADR-0003-yaml-task-persistence.md) | YAML for Task Persistence | Accepted | 2025-11-07 |
| [ADR-0004](ADR-0004-monolithic-cli-architecture.md) | Monolithic CLI Architecture | Accepted | 2025-11-07 |
| [ADR-0005](ADR-0005-layered-architecture-evolution.md) | Layered Architecture Evolution | Accepted | 2025-11-07 |

### Domain Model & Data

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0006](ADR-0006-taskprovider-interface-pattern.md) | TaskProvider Interface Pattern | Accepted | 2025-11-10 |
| [ADR-0007](ADR-0007-compile-time-adapter-registration.md) | Compile-Time Adapter Registration | Accepted | 2026-01-15 |
| [ADR-0008](ADR-0008-atomic-file-writes.md) | Atomic File Writes for Persistence | Accepted | 2025-11-07 |
| [ADR-0009](ADR-0009-task-status-state-machine.md) | Task Status State Machine | Accepted | 2025-11-07 |
| [ADR-0010](ADR-0010-incremental-task-model-extension.md) | Incremental Task Model Extension | Accepted | 2026-01-20 |
| [ADR-0024](ADR-0024-jsonl-session-metrics.md) | JSONL Session Metrics Format | Accepted | 2025-11-08 |

### Sync & Multi-Source

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0011](ADR-0011-sync-scheduler-via-tea-cmd.md) | Sync Scheduler via tea.Cmd | Accepted | 2026-01-20 |
| [ADR-0012](ADR-0012-property-level-conflict-resolution.md) | Property-Level Conflict Resolution | Accepted | 2026-01-20 |
| [ADR-0013](ADR-0013-offline-first-local-change-queue.md) | Offline-First with Local Change Queue | Accepted | 2026-02-01 |
| [ADR-0014](ADR-0014-auto-migrate-schema-on-load.md) | Auto-Migrate Schema on Load | Accepted | 2026-02-15 |
| [ADR-0015](ADR-0015-multi-source-dedup-strategy.md) | Multi-Source Dedup Strategy | Accepted | 2026-02-01 |
| [ADR-0027](ADR-0027-multi-provider-integration-strategy.md) | Multi-Provider Integration Strategy | Accepted | 2026-02-01 |

### Features & UX

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0016](ADR-0016-cancel-sqlite-enrichment-layer.md) | Cancel SQLite Enrichment Layer | Accepted | 2026-01-20 |
| [ADR-0017](ADR-0017-local-first-calendar-integration.md) | Local-First Calendar Integration | Accepted | 2026-02-01 |
| [ADR-0020](ADR-0020-door-theme-system.md) | Door Theme System Architecture | Accepted | 2026-02-20 |
| [ADR-0021](ADR-0021-mcp-server-integration.md) | MCP Server Integration | Accepted | 2026-03-01 |
| [ADR-0022](ADR-0022-cli-interface-with-cobra.md) | CLI Interface with Cobra | Accepted | 2026-03-01 |
| [ADR-0023](ADR-0023-iphone-app-deferred.md) | iPhone App Deferred to Icebox | Accepted | 2026-03-07 |

### Infrastructure & Process

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0018](ADR-0018-macos-code-signing-notarization.md) | macOS Code Signing and Notarization | Accepted | 2025-11-09 |
| [ADR-0019](ADR-0019-docker-e2e-testing.md) | Docker E2E and Headless TUI Testing | Accepted | 2026-02-15 |
| [ADR-0025](ADR-0025-story-driven-development.md) | Story-Driven Development Process | Accepted | 2026-01-15 |
| [ADR-0026](ADR-0026-self-driving-development-pipeline.md) | Self-Driving Development Pipeline | Accepted | 2026-03-01 |
| [ADR-0028](ADR-0028-ci-quality-gates.md) | CI Quality Gates and Testing Strategy | Accepted | 2025-11-09 |

## Decision Dependency Map

Some ADRs are interconnected. Understanding these relationships helps when considering changes:

```
ADR-0003 (YAML) ──→ ADR-0008 (Atomic Writes)
                ──→ ADR-0010 (Incremental Model) ──→ ADR-0014 (Auto-Migrate)
                ──→ ADR-0016 (Cancel SQLite)

ADR-0006 (TaskProvider) ──→ ADR-0007 (Registration)
                         ──→ ADR-0015 (Dedup)
                         ──→ ADR-0027 (Integration Strategy)

ADR-0002 (Bubbletea) ──→ ADR-0011 (Sync via tea.Cmd)
                      ──→ ADR-0020 (Theme System)

ADR-0011 (Sync) ──→ ADR-0012 (Conflict Resolution)
                ──→ ADR-0013 (Offline-First)

ADR-0025 (Story-Driven) ──→ ADR-0026 (Self-Driving Pipeline)
ADR-0019 (Docker E2E) ──→ ADR-0028 (CI Quality Gates)
```

## Sources

ADRs were derived from:
- 213+ merged PRs in the ThreeDoors repository
- `docs/design-decisions-needed.md` — 42 design decisions with documented options and rationale
- `docs/architecture/` — Architecture documentation including high-level architecture, tech stack, and coding standards
- `docs/research/` — Research documents informing technology and design choices
- PR descriptions and commit history

## Decisions Board

For non-ADR decisions (tactical choices, rejected options, active research, pending recommendations), see the [Knowledge Decisions Board](../decisions/BOARD.md). The board is the living dashboard that tracks the full decision lifecycle; ADRs remain the permanent archive for significant architectural decisions.

## Contributing

When making a significant architectural decision:
1. Create a new ADR file: `ADR-NNNN-short-title.md`
2. Use the next available number
3. Follow the format described above
4. Add the new ADR to the index table in this README
5. Update the dependency map if the new ADR relates to existing ones
