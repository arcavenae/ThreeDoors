# Changelog

All notable changes to ThreeDoors, organized by date (most recent first).

---

## Highlights

- **488+ PRs merged** across 35+ completed epics
- **Core TUI** — Three-door task selection with themes, search, mood tracking, session metrics
- **Apple Notes** — Bidirectional sync with offline-first WAL
- **Obsidian** — Vault reader/writer with daily note integration
- **Jira** — Read-only and bidirectional sync with cache/retry
- **Apple Reminders** — Full CRUD via JXA scripts
- **GitHub Issues** — SDK client, TaskProvider, bidirectional sync, contract tests
- **Todoist** — HTTP client, read-only adapter, bidirectional sync, contract tests
- **Multi-Source Sync** — Scheduler, dashboard, cross-provider linking, duplicate detection
- **CLI Interface** — Cobra-based CLI with task management, config, stats, doors experience
- **MCP Server** — Resources, tools, prompts, security middleware, proposal review
- **Dev Dispatch** — Self-driving pipeline with queue, engine, TUI controls, safety guardrails
- **Door Themes** — Classic, Modern, Sci-Fi, Shoji, Seasonal variants with auto-switch
- **Keybinding System** — Registry, overlay, inline hints, command mode, autocomplete
- **Beautiful Stats** — Dashboard, sparklines, heatmap, counters, milestone celebrations
- **Charm Ecosystem** — Spinner, viewport, harmonica, layout utilities, color profiles
- **Security** — File permissions, credential protection, input size limits, keyring
- **Doctor Command** — Self-diagnosis with environment checks and data integrity
- **SLAES** — Retrospector agent, saga detection, doc audit, BOARD pipeline
- **Homebrew** — Dual-publish (stable + alpha), signing, notarization, retention cleanup
- **Testing** — E2E Docker harness, golden file snapshots, CI benchmarks, coverage gates
- **CI/CD** — GitHub Releases, code signing, notarization, coverage reporting, circuit breaker

---

## 2026-03-11

### Epic 49: Doctor Command (Stories 49.1-49.3)

- feat: doctor command skeleton with health alias (Story 49.1) (#444)
- feat: comprehensive environment checks for Doctor (Story 49.2) (#473)
- feat: task data integrity checks for Doctor command (Story 49.3) (#471)

### Epic 50: In-App Bug Reporting (Story 50.1)

- feat: breadcrumb tracking system for in-app bug reporting (Story 50.1) (#478)

### Epic 51: SLAES (Stories 51.1, 51.4-51.6)

- feat: retrospector agent definition (Story 51.1) (#461)
- feat: saga detection package for dispatch waste alerting (Story 51.4) (#464)
- feat: doc consistency audit package and CLI command (Story 51.5) (#465)
- feat: BOARD.md recommendation pipeline (Story 51.6) (#463)

### Epic 42: Application Security (Stories 42.3-42.4)

- feat: input size limits for YAML and JSONL readers (Story 42.3) (#448)
- feat: credential protection in config files (Story 42.4) (#477)

### Epic 43: Data Source Setup UX (Story 43.5)

- feat: sync event logging infrastructure (Story 43.5) (#439)

### Epic 46: OAuth (Story 46.1)

- feat: generic device code flow client (Story 46.1) (#443)

### Infrastructure

- feat: post-merge CI circuit breaker (Story 0.36) (#479)

### Bug Fixes

- fix: staticcheck SA5011 in TestProviderConfig_GetConnection (#467)
- fix: remove t.Parallel() from config_paths tests sharing mutable state (Story 42.1) (#468)

---

## 2026-03-10

### Epic 42: Application Security (Story 42.1)

- feat: file permission standardization 0o700/0o600 (Story 42.1) (#437)

### Epic 43: Data Source Setup UX (Stories 43.1-43.2)

- feat: connection state machine and ConnectionManager type (Story 43.1) (#428)
- feat: keyring integration with environment variable fallback (Story 43.2) (#442)

### Bug Fixes

- fix: harden agent definitions and recover lost skills (#423)
- fix: remove destructive git sync overrides from agent definitions (#424)
- fix: resolve pre-existing SA5011 lint errors in test files (#449)

---

## 2026-03-09

The biggest day in the project — 13 epics completed, 150+ PRs merged across keybindings, stats, themes, integrations, and infrastructure.

### Epic 25: Todoist Integration (Stories 25.1-25.4, COMPLETE)

- feat: Todoist HTTP client & auth configuration (Story 25.1) (#308)
- feat: read-only Todoist adapter with field mapping (Story 25.2) (#321)
- feat: bidirectional Todoist sync & WAL integration (Story 25.3) (#336)
- feat: contract tests & integration testing (Story 25.4) (#353)

### Epic 27: Daily Planning Mode (Stories 27.1-27.5, COMPLETE)

- feat: planning data model & focus tag (Story 27.1) (#323)
- feat: review incomplete tasks flow (Story 27.2) (#339)
- feat: focus selection flow (Story 27.3) (#352)
- feat: energy level matching & time-of-day inference (Story 27.4) (#354)
- feat: CLI plan subcommand & TUI integration (Story 27.5) (#360)

### Epic 28: Snooze/Defer (Stories 28.1-28.4, COMPLETE)

- feat: DeferUntil field, status transitions, and auto-return logic (Story 28.1) (#310)
- feat: snooze TUI view and Z-key binding (Story 28.2) (#338)
- feat: deferred list view and :deferred command (Story 28.3) (#358)
- feat: session metrics logging for snooze events (Story 28.4) (#355)

### Epic 29: Task Dependencies (Stories 29.1-29.4, COMPLETE)

- feat: DependsOn field and DependencyResolver (Story 29.1) (#307)
- feat: door selection dependency filter and auto-unblock (Story 29.2) (#319)
- feat: TUI blocked-by indicator and dependency management (Story 29.3) (#340)
- feat: session metrics logging for dependency events (Story 29.4) (#356)

### Epic 32: Undo Task Completion (Stories 32.1-32.3, COMPLETE)

- feat: complete-to-todo status transition (Story 32.1) (#306)
- feat: session metrics undo complete event logging (Story 32.2) (#322)
- feat: TUI & CLI undo experience (Story 32.3) (#337)

### Epic 33: Seasonal Door Theme Variants (Stories 33.1-33.3)

- feat: seasonal theme metadata model and date-range resolver (Story 33.1) (#403)
- feat: four seasonal theme implementations (Story 33.2) (#409)
- feat: seasonal theme auto-switch integration (Story 33.3) (#410)

### Epic 36: Expand/Fork Key Implementations (Stories 36.1-36.4, COMPLETE)

- feat: enhanced door selection visual feedback (Story 36.1) (#277)
- feat: deselect toggle (Story 36.2) (#272)
- feat: universal quit (Story 36.3) (#276)
- feat: Space/Enter toggle to close door (Story 36.4) (#405)

### Epic 37: BMAD Agent Ecosystem (Stories 37.1-37.4, COMPLETE)

- feat: agent definitions — project-watchdog and arch-watchdog (Story 37.1) (#271)
- feat: cron configuration — SM sprint health & QA coverage audit (Story 37.2) (#280)
- docs: agent communication architecture documentation (Story 37.3) (#279)
- docs: monitoring, tuning, and Phase 1 evaluation (Story 37.4) (#281)

### Epic 38: Homebrew Dual Publish (Stories 38.1-38.6, COMPLETE)

- feat: alpha Homebrew formula (Story 38.1) (#273)
- feat: alpha publishing toggle (Story 38.2) (#287)
- feat: stable release signing & notarization (Story 38.3) (#288)
- feat: alpha release verification (Story 38.4) (#295)
- feat: alpha release retention cleanup (Story 38.5) (#294)
- fix: alpha formula template DSL (Story 38.6) (#312)

### Epic 39: Keybinding Display System (Stories 39.1-39.11, 39.13)

- feat: keybinding registry model (Story 39.1) (#305)
- feat: concise keybinding bar component (Story 39.2) (#318)
- feat: full keybinding overlay (Story 39.3) (#320)
- feat: toggle behavior, config persistence, MainModel integration (Story 39.4) (#346)
- feat: view-specific keybinding completeness and polish (Story 39.5) (#384)
- feat: spacebar as Enter alias in doors view (Story 39.6) (#303)
- feat: global ':' command mode (Story 39.7) (#365)
- feat: command autocomplete/completion (Story 39.8) (#381)
- feat: inline hint rendering infrastructure (Story 39.9) (#374)
- feat: door view inline hints (Story 39.10) (#387)
- feat: non-door view inline hints (Story 39.11) (#388)
- feat: unified door key indicator toggle (Story 39.13) (#407)

### Epic 40: Beautiful Stats Display (Stories 40.1-40.10, COMPLETE)

- feat: stats dashboard shell with Lipgloss panels (Story 40.1) (#343)
- feat: gradient sparkline with color-blind safe palette (Story 40.2) (#366)
- feat: fun facts engine (Story 40.3) (#371)
- feat: horizontal bar charts for mood correlation (Story 40.4) (#362)
- feat: GitHub-style activity heatmap (Story 40.5) (#391)
- feat: surface hidden session metrics (Story 40.6) (#368)
- feat: animated counter reveals (Story 40.7) (#392)
- feat: tab navigation for detail view (Story 40.8) (#367)
- feat: theme-matched stats color palettes (Story 40.9) (#380)
- feat: milestone celebrations (Story 40.10) (#393)

### Epic 41: Charm Ecosystem Adoption (Stories 41.1-41.6, COMPLETE)

- feat: spinner component for async provider operations (Story 41.1) (#372)
- feat: Lipgloss layout utilities adoption (Story 41.2) (#370)
- feat: viewport adoption for help view (Story 41.3) (#364)
- feat: viewport adoption for synclog and keybinding overlay (Story 41.4) (#379)
- feat: Harmonica door transition spike (Story 41.5) (#369)
- feat: adaptive color profile support (Story 41.6) (#373)

### Infrastructure

- feat: CI/security hardening (Story 0.31) (#270)
- feat: CI churn reduction (Story 0.20) (#260)
- feat: Homebrew distribution via GoReleaser (Story 0.21) (#262)
- feat: cross-repo CI monitoring (Story 0.22) (#263)
- feat: dedicated help view (Story 0.32) (#309)
- feat: Renovate + Dependabot dependency management (Story 0.24) (#402)
- feat: stabilize command mode input position (Story 0.35) (#401)

### Bug Fixes

- fix: Homebrew formula passes brew audit --strict (Story 0.23) (#261)
- fix: CI release duplicate asset error (#286)
- fix: scope q-quit to doors view only (Story 0.34) (#361)
- fix: stable search result ordering (Story 0.33) (#350)

---

## 2026-03-08

### BMAD Planning

- feat: BMAD planning — SOUL.md and Custom Multiclaude Skills epic and stories (#211)
- feat: BMAD planning — Seasonal Door Theme Variants epic and stories (#210)
- feat: BMAD planning — Undo Task Completion epic and stories (#209)
- feat: BMAD planning — Expand/Fork Key Implementations epic and stories (#208)
- feat: BMAD planning — Linear Integration epic and stories (#207)

---

## 2026-03-07

CLI, MCP, GitHub Issues integration, and theme polish all landed.

### Epic 26: GitHub Issues Integration (Stories 26.1-26.4)

- feat: GitHub SDK client & auth configuration (Story 26.1) (#201)
- feat: GitHub Issues TaskProvider with field mapping (Story 26.2) (#202)
- feat: GitHub Issues bidirectional sync with WAL & circuit breaker (Story 26.3) (#204)
- feat: GitHub Issues contract tests & edge case coverage (Story 26.4) (#205)

### Epic 23: CLI Interface (Stories 23.1-23.9)

- feat: Cobra CLI scaffolding, root command, output formatter (Story 23.1) (#170)
- feat: task list and task show commands (Story 23.2) (#182)
- feat: task add and task complete CLI commands (Story 23.3) (#171)
- feat: doors command for CLI experience (Story 23.4) (#173)
- feat: health, version commands and exit code enforcement (Story 23.5) (#188)
- feat: task block, unblock, and status commands (Story 23.6) (#195)
- feat: task edit, delete, note, and search CLI commands (Story 23.7) (#194)
- feat: mood and stats CLI commands (Story 23.8) (#189)
- feat: config commands and stdin/pipe support (Story 23.9) (#190)

### Epic 24: MCP/LLM Integration (Stories 24.1-24.8)

- feat: MCP server scaffold with stdio and SSE transports (Story 24.1) (#177)
- feat: read-only MCP resources and query tools (Story 24.2) (#180)
- feat: security middleware for MCP server (Story 24.3) (#179)
- feat: proposal store and enrichment API (Story 24.4) (#185)
- feat: TUI proposal review view (Story 24.5) (#197)
- feat: analytics resources, tools, and prompts (Story 24.6) (#184)
- feat: task relationship graph & cross-provider linking (Story 24.7) (#191)
- feat: MCP prompt templates & advanced interaction patterns (Story 24.8) (#196)

### Epic 17: Door Themes (Stories 17.7-17.9)

- feat: redesign shoji theme with large panes (Story 17.8) (#186)
- feat: simplify sci-fi theme, improve modern contrast (Story 17.9) (#183)
- fix: replace countRunes with ansi.StringWidth (Story 17.7) (#181)

---

## 2026-03-06

Major epic completions: Dev Dispatch (Epic 22), Reminders (Epic 20), and Jira sync (Epic 19).

### Epic 22: Dev Dispatch (Stories 22.1-22.8)

- feat: dev dispatch data model and queue persistence (Story 22.1) (#149)
- feat: dispatch engine with multiclaude CLI wrapper (Story 22.2) (#152)
- feat: TUI dispatch key binding and confirmation flow (Story 22.3) (#163)
- feat: Dev Queue View (Story 22.4) (#162)
- feat: Worker status polling and task update loop (Story 22.5) (#161)
- feat: Auto-generated review and follow-up tasks (Story 22.6) (#164)
- feat: Optional story file generation via AgentService (Story 22.7) (#159)
- feat: Safety guardrails (Story 22.8) (#160)

### Epic 20: Apple Reminders Integration (Stories 20.2-20.4)

- feat: Reminders read-only TaskProvider (Story 20.2) (#148)
- feat: Reminders write support (Story 20.3) (#155)
- feat: Reminders config, registration, and health check (Story 20.4) (#158)

### Epic 19: Jira Integration (Stories 19.3-19.4)

- feat: Jira bidirectional sync with cache and retry (Story 19.3) (#150)
- feat: Jira config parsing, validation, and registration (Story 19.4) (#153)

---

## 2026-03-04

- feat: Jira Read-Only Provider (Story 19.2) (#138)
- feat: Reminders JXA Scripts and CommandExecutor (Story 20.1) (#137)
- feat: Sync Scheduler with Per-Provider Loops (Story 21.1) (#139)
- feat: Expand & Fork actions in detail view (Story 1.3b) (#134)

---

## 2026-03-03

Massive implementation sprint — themes, Obsidian, testing infrastructure, and foundation hardening.

### Epic 17: Door Theme System (Stories 17.1-17.6)

- feat: Theme types, registry, and Classic theme wrapper (Story 17.1)
- feat: Modern, Sci-Fi, and Shoji theme implementations (Story 17.2)
- feat: DoorsView theme integration with config support (Story 17.3)
- feat: Theme Picker in Onboarding Flow (Story 17.4) (#123)
- feat: :theme command with ThemePicker (Story 17.5) (#124)
- test: Golden file tests for all door themes (Story 17.6) (#131)

### Epic 8: Obsidian Integration (Stories 8.1-8.4)

- feat: Obsidian Vault Reader/Writer Adapter (Story 8.1)
- feat: Obsidian Bidirectional Sync & Vault Configuration (Stories 8.2 & 8.3)
- feat: Obsidian Daily Note Integration (Story 8.4)

### Epic 18: Docker E2E Testing (Stories 18.2-18.5)

- feat: Golden File Snapshot Tests (Story 18.2)
- feat: Input Sequence Replay Tests (Story 18.3) (#116)
- feat: Docker Test Environment (Story 18.4) (#117)
- feat: CI Integration for Docker E2E Tests (Story 18.5) (#118)

---

## 2026-03-02

Project inception day — core TUI, Apple Notes integration, and CI/CD pipeline all built.

### Epic 1: Core TUI (Stories 1.1-1.8)

- feat: Project Setup & Basic Bubbletea App (Story 1.1) (#2)
- feat: Display Three Doors from a Task File (Story 1.2) (#4)
- feat: Door Selection & Task Status Management (Story 1.3) (#5, #6)
- feat: Quick Search & Command Palette (Story 1.4) (#7)
- feat: Session Metrics Tracking (Story 1.5) (#8)
- feat: Essential Polish (Story 1.6) (#9)
- feat: CI/CD Pipeline & Alpha Release (Story 1.7) (#10)
- feat: CI Process Validation & Fixes (Story 1.8) (#11)

### Epic 2: Apple Notes Integration (Stories 2.1-2.6)

- feat: MarkComplete in TaskProvider interface (Story 2.1) (#12)
- feat: Apple Notes Integration Spike (Story 2.2) (#13)
- feat: Read Tasks from Apple Notes (Story 2.3) (#15)
- feat: Write Task Updates to Apple Notes (Story 2.4) (#16)
- feat: Bidirectional Sync Engine (Story 2.5) (#17)
- feat: Health Check Command (Story 2.6) (#18)

### Epic 3: Task Engagement (Stories 3.1-3.7)

- feat: Quick Add Mode (Story 3.1) (#19)
- feat: Extended Task Capture with Context (Story 3.2) (#20)
- feat: Values & Goals Setup and Display (Story 3.3) (#21)
- feat: Door Feedback Options (Story 3.4) (#22)
- feat: Daily Completion Tracking (Story 3.5)
- feat: Session Improvement Prompt (Story 3.6)
- feat: Enhanced Navigation & Messaging (Story 3.7)

---

## 2026-03-01

- Initial BMAD method setup and project documentation (#1)

---

## 2025-11-11

- docs: PRD enhancements — mood tracking, search/command palette

---

## 2025-11-08

- Initial Go module, task management, and UX implementation
- QA documentation and initial tests

---

## 2025-11-07

- Project inception: Migrate simple-todo to ThreeDoors
- Initial PRD, architecture, and epic planning
