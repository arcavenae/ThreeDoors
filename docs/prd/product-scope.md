# Product Scope

## Phase 1: Technical Demo & Validation (MVP)

**In Scope:**
- CLI/TUI application using Go and Bubbletea framework
- Three Doors interface displaying three randomly selected tasks
- Local text file storage (`~/.threedoors/tasks.txt`)
- Door selection via keyboard (A/W/D, arrow keys)
- Door refresh to generate new set of three tasks
- Expanded task detail view with status actions
- Task status management: complete, blocked, in progress, expand, fork, procrastinate, rework
- Quick search with live substring matching (/ key)
- Command palette with vi-style `:commands` (`:add`, `:edit`, `:mood`, `:stats`, `:help`, `:quit`)
- Mood tracking with predefined and custom options
- Silent session metrics collection (JSONL format)
- Completed task tracking with timestamps
- "Progress over perfection" messaging in interface
- macOS as primary target platform

**Out of Scope for Phase 1:**
- Apple Notes integration
- Bidirectional sync with any external system
- LLM-powered features
- Values/goals persistent display
- Automated tests (manual validation via daily use)
- Cross-computer sync
- Mobile interface
- Any cloud services or telemetry

---

## Phase 2: Growth (Post-Validation)

**In Scope:**
- Apple Notes integration with bidirectional sync
- Adapter pattern for pluggable backends
- Quick add mode and extended task capture with context
- Values/goals setup and persistent display
- Door feedback mechanisms (blocked, not now, needs breakdown)
- Learning and intelligent door selection based on session metrics
- macOS code signing, notarization, and Homebrew distribution
- Local enrichment storage (SQLite) for metadata
- Health check command for backend connectivity

- Door theme system with user-selectable themed door frames (onboarding picker, settings view, config.yaml)
- Platform readiness refactoring: core domain extraction, adapter hardening, config schema, regression test suite, session metrics reader, CI coverage floor (Epic 3.5)

**Out of Scope for Phase 2:**
- Third-party integrations beyond Apple Notes
- Cross-computer sync
- LLM task decomposition
- Calendar awareness
- Multi-source aggregation

---

## Phase 3: Vision (Post-MVP)

**In Scope:**
- Plugin/adapter SDK with registry and developer guide
- Obsidian vault integration
- Comprehensive testing infrastructure (integration, contract, E2E, CI gates)
- First-run onboarding experience
- Sync observability and offline-first operation
- Calendar awareness (local-first, no OAuth)
- Multi-source task aggregation with dedup
- LLM-powered task decomposition
- Psychology research validation
- Docker-based E2E and headless TUI testing infrastructure (teatest, golden file snapshots, workflow replay, CI integration)

**Out of Scope (Deferred Indefinitely):**
- Web interface
- Voice interface
- Gamification and trading mechanics
- Multi-user support

---

## Phase 3.5: Snooze/Defer as First-Class Action

**In Scope:**
- Z-key snooze action from doors view and detail view with quick options (Tomorrow, Next Week, Pick Date, Someday)
- `defer_until` timestamp field on Task model for date-based snooze
- Auto-return logic: deferred tasks automatically restore to todo when defer date passes (startup check + 1-minute tea.Tick)
- `:deferred` command showing snoozed tasks with un-snooze and edit-date actions
- Additional status transitions: in-progress/blocked to deferred
- Session metrics logging for snooze and auto-return events
- Integration with Daily Planning Mode (FR98 "snooze" option opens same SnoozeView)

**Out of Scope for this Phase:**
- Calendar date picker widget (v1 uses text input for Pick Date)
- CLI `threedoors task defer` command (deferred to Epic 23 extension)
- MCP `defer_task` tool (deferred to Epic 24 extension)

---

## Phase 3.5+: Task Dependencies & Blocked-Task Filtering

**In Scope:**
- `depends_on` field on Task model (list of task IDs)
- Automatic filtering of dependency-blocked tasks from door selection
- Pessimistic handling of orphaned dependencies (missing dep = still blocked)
- "Blocked by: [task]" indicator in doors view and detail view
- Auto-unblock check when dependencies complete, with door refresh
- Dependency management in detail view (+/- keys, task search/picker)
- Circular dependency detection and rejection
- Session metrics logging for dependency and unblock events
- DependencyResolver as standalone component in internal/core/

**Out of Scope for this Phase:**
- Cross-provider dependency resolution (deferred — requires enrichment DB extension)
- CLI `threedoors task deps` commands (deferred to Epic 23 extension)
- MCP `add_dependency` / `remove_dependency` tools (deferred to Epic 24 extension)
- Visual dependency graph rendering (text indicators sufficient for v1)
- Reverse dependency index (iterate pool at current scale)

---

## Phase 4: Task Source Integration & Sync Hardening

**In Scope:**
- Jira integration: read-only adapter (JQL search, status mapping, auth config), then bidirectional sync (MarkComplete via transitions API, WAL queuing)
- Apple Reminders integration: JXA-based adapter with full CRUD (read, create, update, complete, delete), configurable list filtering
- Sync protocol hardening: per-provider sync scheduler with adaptive intervals, circuit breaker per provider, canonical ID mapping via SourceRef
- Generic adapter patterns: rate limit handling, local cache with TTL, credential management via config.yaml/env vars
- Todoist integration: read-only adapter (REST API v1, API token auth, priority-to-effort mapping, project filtering), then bidirectional sync (complete tasks via API, WAL queuing)

**Out of Scope for Phase 4:**
- Linear, GitHub Issues, ClickUp integrations (deferred to Phase 5+)
- OAuth 2.0 flows (API token/PAT auth only for initial integrations)
- EventKit/cgo-based Apple Reminders (future optimization behind build tag)
- Property-level conflict resolution (deferred to Phase 5)
- Cross-computer sync

---

## Phase 5: Future Expansion (12+ months out)

**In Scope:**
- iPhone mobile app (SwiftUI) with Apple Notes sync and Three Doors card carousel
- Self-driving development pipeline (multiclaude worker dispatch from TUI)
- Additional integrations (Linear, GitHub Issues, ClickUp)
- Cross-computer sync

**Out of Scope (Deferred Indefinitely):**
- iPad app
- Apple Watch app
- Android app
- Multi-user support
- Web interface
- Voice interface

---
