# ADR-0011: Sync Scheduler via tea.Cmd

- **Status:** Accepted
- **Date:** 2026-01-20
- **Decision Makers:** Design decision C2
- **Related PRs:** #139 (Story 21.1)
- **Related ADRs:** ADR-0002 (Bubbletea), ADR-0013 (Offline-First)

## Context

The sync scheduler needs to run periodic provider sync operations during TUI sessions. Three approaches were considered for integrating sync with the Bubbletea event loop.

## Considered Options

1. **Background goroutine + channel** — Standard Go concurrency; sync results sent via channel to TUI
2. **`tea.Cmd` integration** — Sync operations dispatched as Bubbletea commands returning `tea.Msg`
3. **Hybrid** — Goroutine manages scheduling; emits `tea.Cmd` for each sync cycle

## Decision

Use **`tea.Cmd` integration** (Option B). All sync operations are dispatched as Bubbletea commands that return `tea.Msg` results.

## Rationale

- Native to Bubbletea framework — consistent with existing patterns
- Better testability — `tea.Cmd` functions are pure and composable
- No channel bridging complexity between goroutines and the Bubbletea event loop
- `tea.Batch` handles concurrent sync operations naturally
- Sync status updates flow through normal `Update()` → `View()` cycle

## Consequences

### Positive
- Sync state is part of the Bubbletea model — single source of truth
- Testable with standard `tea.Cmd` testing patterns
- Sync status display updates automatically via `View()`
- No goroutine lifecycle management

### Negative
- `tea.Batch` management for concurrent provider syncs requires care
- Long-running sync operations must be broken into non-blocking chunks
- Sync only runs while TUI is active (acceptable for personal tool)
