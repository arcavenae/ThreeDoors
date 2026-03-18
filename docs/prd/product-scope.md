---
title: Product Scope
section: Scope & Phasing
lastUpdated: '2026-03-06'
---

# Product Scope

## Phase 1: Foundation (Epics 1-2)

### Technical Demo & Validation (MVP)

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

### Apple Notes Integration & Adapter Pattern

**In Scope:**
- Apple Notes integration with bidirectional sync
- Adapter pattern for pluggable backends
- Health check command for backend connectivity

**Out of Scope for Phase 1:**
- Third-party integrations beyond Apple Notes
- LLM-powered features
- Values/goals persistent display
- Cross-computer sync
- Mobile interface
- Any cloud services or telemetry

---

## Phase 2: Growth (Epics 3-7, 17)

### Enhanced Interaction & Task Context

**In Scope:**
- Quick add mode and extended task capture with context
- Values/goals setup and persistent display
- Door feedback mechanisms (blocked, not now, needs breakdown)
- Learning and intelligent door selection based on session metrics

### Platform Readiness & Distribution

**In Scope:**
- Platform readiness refactoring: core domain extraction, adapter hardening, config schema, regression test suite, session metrics reader, CI coverage floor
- Local enrichment storage (SQLite) for metadata
- macOS code signing, notarization, and Homebrew distribution
- Plugin/adapter SDK with registry and developer guide

### Door Themes

**In Scope:**
- Door theme system with user-selectable themed door frames (onboarding picker, settings view, config.yaml)

**Out of Scope for Phase 2:**
- Third-party integrations beyond Apple Notes
- Cross-computer sync
- LLM task decomposition
- Calendar awareness
- Multi-source aggregation

---

## Phase 3: Platform Expansion (Epics 8-15, 18, 28-29, 31)

### Intelligence & Integration Foundations

**In Scope:**
- Obsidian vault integration
- Comprehensive testing infrastructure (integration, contract, E2E, CI gates)
- First-run onboarding experience
- Sync observability and offline-first operation
- Calendar awareness (local-first, no OAuth)
- Multi-source task aggregation with dedup
- LLM-powered task decomposition
- Psychology research validation
- Docker-based E2E and headless TUI testing infrastructure (teatest, golden file snapshots, workflow replay, CI integration)

### Task Management Extensions

**In Scope:**
- Snooze/Defer as first-class action: Z-key snooze action from doors view and detail view with quick options (Tomorrow, Next Week, Pick Date, Someday)
- `defer_until` timestamp field on Task model for date-based snooze
- Auto-return logic: deferred tasks automatically restore to todo when defer date passes (startup check + 1-minute tea.Tick)
- `:deferred` command showing snoozed tasks with un-snooze and edit-date actions
- Additional status transitions: in-progress/blocked to deferred
- Session metrics logging for snooze and auto-return events
- Integration with Daily Planning Mode (FR98 "snooze" option opens same SnoozeView)
- Task Dependencies: `depends_on` field on Task model (list of task IDs)
- Automatic filtering of dependency-blocked tasks from door selection
- Pessimistic handling of orphaned dependencies (missing dep = still blocked)
- "Blocked by: [task]" indicator in doors view and detail view
- Auto-unblock check when dependencies complete, with door refresh
- Dependency management in detail view (+/- keys, task search/picker)
- Circular dependency detection and rejection
- DependencyResolver as standalone component in internal/core/
- Enhanced Expand action: sequential subtask creation mode (E key in detail view)
- Native `parent_id` field on Task model for parent-child relationships
- Subtask list rendering in detail view with completion ratio
- Parent tasks excluded from door selection when they have children
- Enhanced Fork action: variant creation with field preservation/reset semantics
- `ForkTask` factory method copying Text/Context/Effort/Tags, resetting Status/Blocker/Notes
- Fork cross-references via enrichment DB (`forked-from` relationship)

**Out of Scope for Phase 3:**
- Calendar date picker widget (v1 uses text input for Pick Date)
- CLI `threedoors task defer` command (deferred to CLI extension)
- MCP `defer_task` tool (deferred to MCP extension)
- Cross-provider dependency resolution (deferred — requires enrichment DB extension)
- CLI `threedoors task deps` commands (deferred to CLI extension)
- MCP `add_dependency` / `remove_dependency` tools (deferred to MCP extension)
- Visual dependency graph rendering (text indicators sufficient for v1)
- Reverse dependency index (iterate pool at current scale)
- Drag-and-drop subtask reordering (TUI limitation)
- Recursive subtask nesting (v1 supports single-level parent-child only)
- CLI `threedoors task expand/fork` commands (deferred to CLI extension)
- MCP `expand_task` / `fork_task` tools (deferred to MCP extension)
- Web interface
- Voice interface
- Gamification and trading mechanics
- Multi-user support

---

## Phase 4: Task Source Integration (Epics 19-21, 25-26, 30, 43-47, 63, 66)

### Provider Integrations

**In Scope:**
- Jira integration: read-only adapter (JQL search, status mapping, auth config), then bidirectional sync (MarkComplete via transitions API, WAL queuing)
- Apple Reminders integration: JXA-based adapter with full CRUD (read, create, update, complete, delete), configurable list filtering
- Todoist integration: read-only adapter (REST API v1, API token auth, priority-to-effort mapping, project filtering), then bidirectional sync (complete tasks via API, WAL queuing)
- GitHub Issues integration: go-github SDK with label/milestone mapping, bidirectional sync
- Linear integration: GraphQL-based adapter with full field mapping, bidirectional sync
- ClickUp integration: REST API v2 with token auth, standard adapter pattern

### Sync Architecture

**In Scope:**
- Sync protocol hardening: per-provider sync scheduler with adaptive intervals, circuit breaker per provider, canonical ID mapping via SourceRef
- Generic adapter patterns: rate limit handling, local cache with TTL, credential management via config.yaml/env vars

### Connection Management

**In Scope:**
- Connection Manager Infrastructure: 7-state machine, keychain credential storage, config schema v3, CRUD operations, sync event logging, adapter migration
- Sources TUI: setup wizard (charmbracelet/huh), sources dashboard, detail view, sync log, status bar alerts, disconnect/re-auth flows
- Sources CLI: `threedoors connect`, `threedoors sources` commands with `--json` output
- OAuth Device Code Flow: generic RFC 8628 client, GitHub and Linear integrations, token refresh
- Sync Lifecycle: conflict resolution (field-level strategy), orphaned task handling, tool auto-detection, predictive warnings

### CLI/TUI Adapter Wiring Parity

**In Scope:**
- CLI adapter registry initialization: ensure `registerBuiltinAdapters()` runs before both CLI and TUI code paths
- ClickUp connect wiring: add ClickUp to CLI `knownProviderSpecs`, `ValidArgs`, and TUI `DefaultProviderSpecs()`
- Provider spec parity: all registered adapters have complete flag specs in CLI connect command with required flag enforcement
- Parity test: build-time verification that adapter registry, CLI specs, CLI args, and TUI specs are in sync

**Out of Scope for Phase 4:**
- OAuth 2.0 authorization code flows (device code flow only)
- EventKit/cgo-based Apple Reminders (future optimization behind build tag)
- Property-level conflict resolution beyond field-level strategy
- Auto-generating provider specs from adapter metadata (premature abstraction)

---

## Phase 5: User Experience & Polish (Epics 27, 32-33, 35-36, 39-41, 48, 56-57, 59, 69-70)

### Feature Richness

**In Scope:**
- Daily Planning Mode: guided morning planning ritual with review/select/confirm steps, energy matching, focus-aware door scoring
- Undo Task Completion: complete→todo transition, dependency re-evaluation on undo
- LLM CLI Services: CLIProvider wrapping Claude/Gemini/Ollama CLIs, TaskExtractor, TaskEnricher, TaskBreakdown, `threedoors llm status`

### Visual Polish & UX Enhancement

**In Scope:**
- Seasonal Door Theme Variants: time-based auto-switching by calendar date
- Door Visual Appearance: portrait orientation, panel dividers, asymmetric handles, threshold lines, compact mode fallback, shadow/depth effects
- Door Selection Feedback: high-contrast selection, deselect toggle, universal quit, selection animation
- Keybinding Display System: compile-time registry, bottom bar, full overlay (`?`), `h` toggle unifying door indicators and bar, `:hints` command
- Beautiful Stats Display: Lipgloss panels, gradient sparklines, fun facts, bar charts, heatmap, animated counters, tab navigation, theme-matched colors, milestone celebrations
- Charm Ecosystem Adoption: bubbles/spinner, lipgloss layout, bubbles/viewport, harmonica spring-physics, adaptive color profiles
- Door-Like Doors: side handles, hinge marks, threshold line, crack-of-light selection, handle turn animation
- Door Visual Redesign: three-layer depth (background fill, bevel lighting, gradient shadow), panel zone shading, width-adaptive shadow
- Full-Terminal Vertical Layout: AltScreen, layout engine, door height cap, perceptual centering, graceful degradation breakpoints

### TUI Refactoring & User Progress

**In Scope:**
- TUI MainModel Decomposition: extract view navigation, source/sync controllers, planning controllers, auxiliary views from monolithic main_model.go
- Completion History & Progress View: `:history` TUI view, `threedoors history` CLI command, completion data aggregator, streak/stats display

**Out of Scope for Phase 5:**
- Voice interface
- Apple Watch/iPad apps
- Calendar date picker widget (text input for snooze dates)
- Recursive subtask nesting (single-level parent-child only)
- Drag-and-drop subtask reordering
- Recursive decomposition of individual view files
- Historical data migration from legacy formats

---

## Phase 6: Developer Experience & Governance (Epics 22-24, 34, 37-38, 42, 49-53, 55, 58, 60-62, 65, 67-68)

### Self-Driving Development & CLI/MCP

**In Scope:**
- Self-driving development pipeline: DevDispatch model, multiclaude CLI wrapper, TUI dispatch key ('x'), dev queue view, worker status polling, auto-generated follow-up tasks, safety guardrails
- CLI Interface: Cobra-based `threedoors` command with `--json` flag, task CRUD, door commands, session/analytics commands, shell completions for 4 shells
- MCP/LLM Integration Server: MCP server exposing task management tools, structured JSON responses, resource endpoints for LLM agent consumption

### Developer Tooling & Quality

**In Scope:**
- SOUL.md + Custom Dev Skills: project philosophy document, `/pre-pr`, `/validate-adapter`, `/check-patterns`, `/new-story` slash commands
- Application Security Hardening: file permissions, symlink validation, input size limits, credential protection, CI supply chain
- ThreeDoors Doctor: `threedoors doctor` command with 6 check categories, flutter-style output, auto-repair, channel-aware version checking
- In-App Bug Reporting: `:bug` command, breadcrumb tracking, tiered submission (browser/API/file), privacy allowlist
- CI Optimization: Docker E2E push-only, benchmark path filtering, `make test-fast`
- CLI Test Coverage Hardening: coverage from 34.8% to ≥70%

### Autonomous Project Governance

**In Scope:**
- Persistent BMAD Agent Infrastructure: project-watchdog, arch-watchdog agent definitions, SM/QA cron jobs, agent communication architecture
- Dual Homebrew Distribution: stable + alpha channels, signing parity, publishing toggle, retention management
- SLAES: retrospector agent, saga detection, doc consistency audit, BOARD.md recommendations, CI failure taxonomy, weekly trends, 5 Watchmen safeguards
- Envoy Three-Layer Firewall: structured issue screening with syntax/scope/impact layers
- Remote Collaboration: SSH access, tmux management, MCP bridge design, security hardening
- Supervisor Shift Handover: transcript monitoring, rolling state snapshot, 5-step handover protocol, emergency handover, history logging
- Retrospector Agent Reliability: file-based inbox, recommendation queue, checkpoint persistence
- Retrospector Operational Data Pipeline: cron-triggered project-watchdog sync of `docs/operations/` to git

### Documentation & Distribution

**In Scope:**
- README Overhaul: badges, table of contents, foldable sections, feature audit, visual demo
- GitHub Pages User Guide: MkDocs + Material, CI deployment, guides for all features and integrations
- BOARD.md Redesign: split into active dashboard + decision archive, extract Epic Number Registry, fix duplicate IDs
- GitHub Label Operationalization: wire label application into agent workflows (merge-queue PR labeling, envoy startup catch-up, supervisor label discipline, mutual exclusivity enforcement)

**Out of Scope for Phase 6:**
- Tech Writer persistent agent
- Analyst persistent agent
- Webhook-based event triggers
- Adaptive polling intervals
- GUI or web-based dispatch interface
- Real-time worker output streaming to TUI
- Automated security scanning beyond govulncheck
- Performance profiling tools

---

## Phase 7: Future Expansion (Epics 16, 54, 64)

**In Scope:**
- iPhone mobile app (SwiftUI) — deferred indefinitely (Epic 16, Icebox)
- Gemini Research Supervisor: persistent research-supervisor agent wrapping Gemini CLI with OAuth, context packaging, result shielding, dual-tier budget management (Epic 54)
- Cross-computer sync: device identity, sync transport, cross-machine conflict resolution, offline queue (Epic 64)

**Out of Scope (Deferred Indefinitely):**
- iPad app
- Apple Watch app
- Android app
- Multi-user support
- Web interface
- Voice interface
- Gamification and trading mechanics

---
