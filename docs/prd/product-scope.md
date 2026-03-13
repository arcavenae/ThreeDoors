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

## Phase 3.5+: Expand/Fork Key Implementations

**In Scope:**
- Enhanced Expand action: sequential subtask creation mode (E key in detail view)
- Native `parent_id` field on Task model for parent-child relationships
- Subtask list rendering in detail view with completion ratio
- Parent tasks excluded from door selection when they have children
- Enhanced Fork action: variant creation with field preservation/reset semantics
- `ForkTask` factory method copying Text/Context/Effort/Tags, resetting Status/Blocker/Notes
- Fork cross-references via enrichment DB (`forked-from` relationship)

**Out of Scope for this Phase:**
- Drag-and-drop subtask reordering (TUI limitation)
- Recursive subtask nesting (v1 supports single-level parent-child only)
- CLI `threedoors task expand/fork` commands (deferred to Epic 23 extension)
- MCP `expand_task` / `fork_task` tools (deferred to Epic 24 extension)

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

## Phase 4: Task Source Integration & Sync Hardening (Epics 19-21, 25-26, 30, 43-47, 63)

## Phase 4+: CLI/TUI Adapter Wiring Parity

**In Scope:**
- CLI adapter registry initialization: ensure `registerBuiltinAdapters()` runs before both CLI and TUI code paths
- ClickUp connect wiring: add ClickUp to CLI `knownProviderSpecs`, `ValidArgs`, and TUI `DefaultProviderSpecs()`
- Provider spec parity: all 9 registered adapters have complete flag specs in CLI connect command with required flag enforcement
- Parity test: build-time verification that adapter registry, CLI specs, CLI args, and TUI specs are in sync

**Out of Scope for this Phase:**
- Auto-generating provider specs from adapter metadata (premature abstraction)
- Interactive CLI wizard for providers without the TUI connect wizard (covered by Story 45.6)

---

## Phase 5: Future Expansion (12+ months out)

**In Scope:**
- Jira integration: read-only adapter (JQL search, status mapping, auth config), then bidirectional sync (Epic 19)
- Apple Reminders integration: JXA-based adapter with full CRUD (Epic 20)
- Sync protocol hardening: per-provider sync scheduler, circuit breaker, canonical ID mapping via SourceRef (Epic 21)
- Todoist integration: REST API v1 with priority-to-effort mapping, project filtering, bidirectional sync (Epic 25)
- GitHub Issues integration: go-github SDK with label/milestone mapping, bidirectional sync (Epic 26)
- Linear integration: GraphQL-based adapter with full field mapping, bidirectional sync (Epic 30)
- Connection Manager Infrastructure: 7-state machine, keychain credential storage, config schema v3, CRUD operations, sync event logging, adapter migration (Epic 43)
- Sources TUI: setup wizard (charmbracelet/huh), sources dashboard, detail view, sync log, status bar alerts, disconnect/re-auth flows (Epic 44)
- Sources CLI: `threedoors connect`, `threedoors sources` commands with `--json` output (Epic 45)
- OAuth Device Code Flow: generic RFC 8628 client, GitHub and Linear integrations, token refresh (Epic 46)
- Sync Lifecycle: conflict resolution (field-level strategy), orphaned task handling, tool auto-detection, predictive warnings (Epic 47)
- ClickUp integration: REST API v2 with token auth, standard adapter pattern (Epic 63)

**Out of Scope for Phase 4:**
- OAuth 2.0 authorization code flows (device code flow only)
- EventKit/cgo-based Apple Reminders (future optimization)
- Property-level conflict resolution beyond field-level strategy

---

## Phase 4.5: Self-Driving Development & CLI/MCP (Epics 22-24)

**In Scope:**
- Self-driving development pipeline: DevDispatch model, multiclaude CLI wrapper, TUI dispatch key ('x'), dev queue view, worker status polling, auto-generated follow-up tasks, safety guardrails (Epic 22)
- CLI Interface: Cobra-based `threedoors` command with `--json` flag, task CRUD, door commands, session/analytics commands, shell completions for 4 shells (Epic 23)
- MCP/LLM Integration Server: MCP server exposing task management tools, structured JSON responses, resource endpoints for LLM agent consumption (Epic 24)

**Out of Scope for this Phase:**
- GUI or web-based dispatch interface
- Real-time worker output streaming to TUI

---

## Phase 5: Feature Richness (Epics 27-32, 57)

**In Scope:**
- Daily Planning Mode: guided morning planning ritual with review/select/confirm steps, energy matching, focus-aware door scoring (Epic 27)
- Snooze/Defer: Z-key action, date-based snooze with auto-return, `:deferred` command, additional status transitions (Epic 28)
- Task Dependencies: `depends_on` field, automatic door filtering, blocked-by indicators, circular detection, DependencyResolver (Epic 29)
- Expand/Fork: E key subtask creation with parent_id, subtask rendering, parent exclusion from doors; F key variant creation with field copy/reset semantics (Epic 31)
- Undo Task Completion: complete→todo transition, dependency re-evaluation on undo (Epic 32)
- LLM CLI Services: CLIProvider wrapping Claude/Gemini/Ollama CLIs, TaskExtractor, TaskEnricher, TaskBreakdown, `threedoors llm status` (Epic 57)

**Out of Scope for Phase 5:**
- Calendar date picker widget (text input for snooze dates)
- Recursive subtask nesting (single-level parent-child only)
- Drag-and-drop subtask reordering

---

## Phase 5.5: Visual Polish & UX Enhancement (Epics 33, 35-36, 39-41, 48, 56, 59)

**In Scope:**
- Seasonal Door Theme Variants: time-based auto-switching by calendar date (Epic 33)
- Door Visual Appearance: portrait orientation, panel dividers, asymmetric handles, threshold lines, compact mode fallback, shadow/depth effects (Epic 35)
- Door Selection Feedback: high-contrast selection, deselect toggle, universal quit, selection animation (Epic 36)
- Keybinding Display System: compile-time registry, bottom bar, full overlay (`?`), `h` toggle unifying door indicators and bar, `:hints` command (Epic 39)
- Beautiful Stats Display: Lipgloss panels, gradient sparklines, fun facts, bar charts, heatmap, animated counters, tab navigation, theme-matched colors, milestone celebrations (Epic 40)
- Charm Ecosystem Adoption: bubbles/spinner, lipgloss layout, bubbles/viewport, harmonica spring-physics, adaptive color profiles (Epic 41)
- Door-Like Doors: side handles, hinge marks, threshold line, crack-of-light selection, handle turn animation (Epic 48)
- Door Visual Redesign: three-layer depth (background fill, bevel lighting, gradient shadow), panel zone shading, width-adaptive shadow (Epic 56)
- Full-Terminal Vertical Layout: AltScreen, layout engine, door height cap, perceptual centering, graceful degradation breakpoints (Epic 59)

**Out of Scope for this Phase:**
- Voice interface
- Apple Watch/iPad apps

---

## Phase 6: Developer Experience & Tooling (Epics 34, 42, 49-50, 55, 65)

**In Scope:**
- SOUL.md + Custom Dev Skills: project philosophy document, `/pre-pr`, `/validate-adapter`, `/check-patterns`, `/new-story` slash commands (Epic 34)
- Application Security Hardening: file permissions, symlink validation, input size limits, credential protection, CI supply chain (Epic 42)
- ThreeDoors Doctor: `threedoors doctor` command with 6 check categories, flutter-style output, auto-repair, channel-aware version checking (Epic 49)
- In-App Bug Reporting: `:bug` command, breadcrumb tracking, tiered submission (browser/API/file), privacy allowlist (Epic 50)
- CI Optimization: Docker E2E push-only, benchmark path filtering, `make test-fast` (Epic 55)
- CLI Test Coverage Hardening: coverage from 34.8% to ≥70% (Epic 65)

**Out of Scope for this Phase:**
- Automated security scanning beyond govulncheck
- Performance profiling tools

---

## Phase 6+: Autonomous Project Governance (Epics 37-38, 51-53, 58, 62, 67)

**In Scope:**
- Persistent BMAD Agent Infrastructure: project-watchdog, arch-watchdog agent definitions, SM/QA cron jobs, agent communication architecture (Epic 37)
- Dual Homebrew Distribution: stable + alpha channels, signing parity, publishing toggle, retention management (Epic 38)
- SLAES: retrospector agent, saga detection, doc consistency audit, BOARD.md recommendations, CI failure taxonomy, weekly trends, 5 Watchmen safeguards (Epic 51)
- Envoy Three-Layer Firewall: structured issue screening with syntax/scope/impact layers (Epic 52)
- Remote Collaboration: SSH access, tmux management, MCP bridge design, security hardening (Epic 53)
- Supervisor Shift Handover: transcript monitoring, rolling state snapshot, 5-step handover protocol, emergency handover, history logging (Epic 58)
- Retrospector Agent Reliability: file-based inbox, recommendation queue, checkpoint persistence (Epic 62)
- Retrospector Operational Data Pipeline: cron-triggered project-watchdog sync of `docs/operations/` to git (Epic 67)

**Out of Scope for this Phase:**
- Tech Writer persistent agent
- Analyst persistent agent
- Webhook-based event triggers
- Adaptive polling intervals

---

## Phase 7: Documentation & Distribution (Epics 60-61)

**In Scope:**
- README Overhaul: badges, table of contents, foldable sections, feature audit, visual demo (Epic 60)
- GitHub Pages User Guide: MkDocs + Material, CI deployment, guides for all features and integrations (Epic 61)

---

## Phase 8: Future Expansion

**In Scope:**
- iPhone mobile app (SwiftUI) — deferred indefinitely (Epic 16, Icebox)
- Cross-computer sync (Epic 64, In Progress)
- Gemini Research Supervisor (Epic 54, In Progress)

**Out of Scope (Deferred Indefinitely):**
- iPad app
- Apple Watch app
- Android app
- Multi-user support
- Web interface
- Voice interface

---
