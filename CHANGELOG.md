# Changelog — ThreeDoors

All notable changes to this project, organized by date (most recent first).

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

### Epic 49: Doctor Command (Stories 49.1–49.3)

- feat: doctor command skeleton with health alias (Story 49.1) (#444)
- feat: comprehensive environment checks for Doctor (Story 49.2) (#473)
- feat: task data integrity checks for Doctor command (Story 49.3) (#471)

### Epic 50: In-App Bug Reporting (Story 50.1)

- feat: breadcrumb tracking system for in-app bug reporting (Story 50.1) (#478)

### Epic 51: SLAES (Stories 51.1, 51.4–51.6)

- feat: retrospector agent definition — Responsibility+WHY format (Story 51.1) (#461)
- feat: saga detection package for dispatch waste alerting (Story 51.4) (#464)
- feat: doc consistency audit package and CLI command (Story 51.5) (#465)
- feat: BOARD.md recommendation pipeline (Story 51.6) (#463)

### Epic 42: Application Security (Stories 42.3–42.4)

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

### Docs

- docs: Story 51.2 — rewrite operational agent definitions (Responsibility+WHY format) (#460)
- docs: PR failure analyses — #447, #438, #441 (#459, #458, #457)
- docs: CI churn reduction follow-up stories (Stories 0.36, 0.37) (#469)
- docs: GitHub community standards plan (#480)
- docs: README.md overhaul research and plan (#482)
- docs: resolve Q-003/Q-004 — planning doc ownership (D-161, D-162) (#484)
- docs: fix Epic Number Registry accuracy and enforcement (Story 0.38) (#485)
- docs: close BOARD P-004 — fork references already removed (#487)
- docs: envoy three-layer firewall stories (BOARD P-002) (#488)

---

## 2026-03-10

### Epic 42: Application Security (Story 42.1)

- feat: file permission standardization 0o700/0o600 (Story 42.1) (#437)

### Epic 43: Data Source Setup UX (Stories 43.1–43.2)

- feat: connection state machine and ConnectionManager type (Story 43.1) (#428)
- feat: keyring integration with environment variable fallback (Story 43.2) (#442)

### Bug Fixes

- fix: harden agent definitions and recover lost skills (#423)
- fix: remove destructive git sync overrides from agent definitions (#424)
- fix: resolve pre-existing SA5011 lint errors in test files (#449)

### Docs

- docs: Epic 42 — Application Security Hardening planning (#418)
- docs: Epic 49 — ThreeDoors Doctor self-diagnosis command planning (#430)
- docs: Epic 50 — In-App Bug Reporting stories and planning (P-006) (#445)
- docs: Epic 51 — SLAES stories and planning docs (#450)
- docs: renumber Door-Like Doors from Epic 42 → Epic 48 (#421)
- docs: incident reports — INC-001 pr-shepherd contamination (#422), INC-002 git sync (#426), INC-003 epic collision (#429)
- docs: governance syncs — Epic 25 Done (#427), Story 43.1 Done (#433)
- docs: SLAES party mode — retrospector agent design consensus (#434)
- docs: subagent abuse investigation and enforcement rule (#435)
- docs: remote collaboration research — investigation (#452), creative problem-solving (#453), feasibility analysis (#454), brainstorm (#455), web research (#456)

---

## 2026-03-09

The biggest day in the project — 13 epics completed, 150+ PRs merged across keybindings, stats, themes, integrations, and infrastructure.

### Epic 25: Todoist Integration (Stories 25.1–25.4, COMPLETE)

- feat: Todoist HTTP client & auth configuration (Story 25.1) (#308)
- feat: read-only Todoist adapter with field mapping (Story 25.2) (#321)
- feat: bidirectional Todoist sync & WAL integration (Story 25.3) (#336)
- feat: contract tests & integration testing (Story 25.4) (#353)

### Epic 27: Daily Planning Mode (Stories 27.1–27.5, COMPLETE)

- feat: planning data model & focus tag (Story 27.1) (#323)
- feat: review incomplete tasks flow (Story 27.2) (#339)
- feat: focus selection flow (Story 27.3) (#352)
- feat: energy level matching & time-of-day inference (Story 27.4) (#354)
- feat: CLI plan subcommand & TUI integration (Story 27.5) (#360)

### Epic 28: Snooze/Defer (Stories 28.1–28.4, COMPLETE)

- feat: DeferUntil field, status transitions, and auto-return logic (Story 28.1) (#310)
- feat: snooze TUI view and Z-key binding (Story 28.2) (#338)
- feat: deferred list view and :deferred command (Story 28.3) (#358)
- feat: session metrics logging for snooze events (Story 28.4) (#355)

### Epic 29: Task Dependencies (Stories 29.1–29.4, COMPLETE)

- feat: DependsOn field and DependencyResolver (Story 29.1) (#307)
- feat: door selection dependency filter and auto-unblock (Story 29.2) (#319)
- feat: TUI blocked-by indicator and dependency management (Story 29.3) (#340)
- feat: session metrics logging for dependency events (Story 29.4) (#356)

### Epic 32: Undo Task Completion (Stories 32.1–32.3, COMPLETE)

- feat: complete-to-todo status transition (Story 32.1) (#306)
- feat: session metrics undo complete event logging (Story 32.2) (#322)
- feat: TUI & CLI undo experience (Story 32.3) (#337)

### Epic 33: Seasonal Door Theme Variants (Stories 33.1–33.3)

- feat: seasonal theme metadata model and date-range resolver (Story 33.1) (#403)
- feat: four seasonal theme implementations (Story 33.2) (#409)
- feat: seasonal theme auto-switch integration (Story 33.3) (#410)

### Epic 36: Expand/Fork Key Implementations (Stories 36.1–36.4, COMPLETE)

- feat: enhanced door selection visual feedback (Story 36.1) (#277)
- feat: deselect toggle — press same key to unselect (Story 36.2) (#272)
- feat: universal quit — 'q' works from all screens (Story 36.3) (#276)
- feat: Space/Enter toggle to close door (Story 36.4) (#405)

### Epic 37: BMAD Agent Ecosystem (Stories 37.1–37.4, COMPLETE)

- feat: agent definitions — project-watchdog and arch-watchdog (Story 37.1) (#271)
- feat: cron configuration — SM sprint health & QA coverage audit (Story 37.2) (#280)
- docs: agent communication architecture documentation (Story 37.3) (#279)
- docs: monitoring, tuning, and Phase 1 evaluation (Story 37.4) (#281)

### Epic 38: Homebrew Dual Publish (Stories 38.1–38.6, COMPLETE)

- feat: alpha Homebrew formula — threedoors-a (Story 38.1) (#273)
- feat: alpha publishing toggle (Story 38.2) (#287)
- feat: stable release signing & notarization (Story 38.3) (#288)
- feat: alpha release verification (Story 38.4) (#295)
- feat: alpha release retention cleanup (Story 38.5) (#294)
- fix: alpha formula template DSL — use if/elsif conditionals (Story 38.6) (#312)

### Epic 39: Keybinding Display System (Stories 39.1–39.11, 39.13)

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

### Epic 40: Beautiful Stats Display (Stories 40.1–40.10, COMPLETE)

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

### Epic 41: Charm Ecosystem Adoption (Stories 41.1–41.6, COMPLETE)

- feat: spinner component for async provider operations (Story 41.1) (#372)
- feat: Lipgloss layout utilities adoption (Story 41.2) (#370)
- feat: viewport adoption for help view (Story 41.3) (#364)
- feat: viewport adoption for synclog and keybinding overlay (Story 41.4) (#379)
- feat: Harmonica door transition spike (Story 41.5) (#369)
- feat: adaptive color profile support (Story 41.6) (#373)

### Infrastructure

- feat: CI/security hardening — secrets, supply chain & reproducibility (Story 0.31) (#270)
- feat: CI churn reduction — path filtering & relaxed merge rules (Story 0.20) (#260)
- feat: Homebrew distribution via GoReleaser (Story 0.21) (#262)
- feat: cross-repo CI monitoring for shared homebrew tap (Story 0.22) (#263)
- feat: envoy operations guide (Story 0.29) (#259)
- feat: dedicated help view replacing flash message (Story 0.32) (#309)
- feat: Renovate + Dependabot automated dependency management (Story 0.24) (#402)
- feat: stabilize command mode input position (Story 0.35) (#401)
- build(deps): bump the actions group with 8 updates (#404)

### Bug Fixes

- fix: Homebrew formula template passes brew audit --strict (Story 0.23) (#261)
- fix: CI release duplicate asset error (#286)
- fix: scope q-quit to doors view only (Story 0.34) (#361)
- fix: stable search result ordering (Story 0.33) (#350)
- fix: add supervisor agent definition (#275)

### Docs

- docs: reconcile planning docs (Story 0.30) (#274)
- docs: governance syncs — epic/story status updates (#282, #285, #297, #302, #314, #316, #325, #327, #332, #351, #357, #363, #376, #377, #378, #382, #383, #385, #386, #389, #390, #394, #395, #397, #400, #406, #416)
- docs: Epic 38 — Homebrew Dual Publish planning (#284)
- docs: Epic 39 — Keybinding Display System planning (#292)
- docs: Epic 39 / 40 — Beautiful Stats Display planning (#299)
- docs: Epic 41 — Charm Ecosystem Adoption planning (#359)
- docs: Story 36.4 — Space/Enter toggle planning (#398)
- docs: Story 39.13 — unified door key indicator toggle planning (#399)
- docs: global command mode and autocomplete planning (Stories 39.7, 39.8) (#335)
- docs: inline tooltips/keybinding hints design (Stories 39.9–39.12) (#341)
- docs: help display UX redesign (Story 0.32) (#293)
- docs: dual Homebrew distribution design (#269)
- docs: Homebrew installation instructions in README (#289)
- docs: comprehensive README and user guide update (#345)
- docs: triage CI/security issues #244-248 — stories and artifacts (#264)
- docs: triage planning docs reconciliation issue #252 (#258)
- docs: triage issue #296 — alpha formula CI failures (Story 38.6) (#301)
- docs: triage issue #334 — stable search result ordering (#342)
- docs: triage issue #330 — dashboard q key regression (#344)
- docs: :dashboard exits-app bug investigation (#331)
- docs: governance quick wins + strategic decisions (ADRs 0029-0033) (#265)
- docs: scoped label taxonomy for human-LLM collaborative projects (#315)
- docs: dead research audit — surface untracked artifacts (#290)
- docs: dead research audit follow-up — scoped labels decisions (#317)
- docs: consolidate planning artifact sprawl into canonical directory (#324)
- docs: fix stale path references after artifact consolidation (#326)
- docs: research — open source license selection (#268)
- docs: research — spacebar action (full agent panel debate) (#298)
- docs: research — Beautiful Stats Display (#291)
- docs: research — full-terminal vertical layout & party mode (#329)
- docs: Bubbletea feature audit & Charm ecosystem UX roadmap (#347)
- docs: Renovate/Dependabot agent interaction risk analysis (#348)
- docs: state of testing audit report (#349)
- docs: persistent agent lockup diagnosis and restart protocol (#375)
- docs: persistent agent communication investigation (#267)
- docs: envoy agent messaging integration fix (#266)
- docs: application security audit (#411)
- docs: doors-more-doorlike party mode research (#412)
- docs: ThreeDoors doctor command research report (#413)
- docs: data source setup UX research — full lifecycle management (#415)
- docs: Epic 42 — Door-Like Doors planning (#417)
- docs: data source setup UX — Epics 43-47 planning (26 stories) (#420)
- docs: fix Epic 39/40 number collision residuals (#313)
- docs: ROADMAP.md status sync for Story 0.28 (#304)
- docs: research — in-app bug reporting via :bug command (#328)
- docs: Epic 39 governance sync investigation + epic number registry (#327)
- docs: resolve Q-001/Q-002, fix P-002/P-003 links, approve P-004 (#333)

---

## 2026-03-08

### BMAD Planning

- feat: BMAD planning — SOUL.md and Custom Multiclaude Skills epic and stories (#211)
- feat: BMAD planning — Seasonal Door Theme Variants epic and stories (#210)
- feat: BMAD planning — Undo Task Completion epic and stories (#209)
- feat: BMAD planning — Expand/Fork Key Implementations epic and stories (#208)
- feat: BMAD planning — Linear Integration epic and stories (#207)

### Docs

- docs: sync ROADMAP.md with current epic/story status (#212)

---

## 2026-03-07

A massive day of development — CLI, MCP, GitHub Issues integration, and theme polish all landed.

### Epic 26: GitHub Issues Integration (Stories 26.1–26.4)

- feat: GitHub SDK client & auth configuration (Story 26.1) (#201)
- feat: GitHub Issues TaskProvider with field mapping (Story 26.2) (#202)
- feat: GitHub Issues bidirectional sync with WAL & circuit breaker (Story 26.3) (#204)
- feat: GitHub Issues contract tests & edge case coverage (Story 26.4) (#205)

### Epic 23: CLI Interface (Stories 23.1–23.9)

- feat: add Cobra CLI scaffolding, root command, and output formatter (Story 23.1) (#170)
- feat: add task list and task show commands with prefix matching (Story 23.2) (#182)
- feat: add task add and task complete CLI commands (Story 23.3) (#171)
- feat: add doors command for CLI three doors experience (Story 23.4) (#173)
- feat: add health, version commands and exit code enforcement (Story 23.5) (#188)
- feat: task block, unblock, and status commands (Story 23.6) (#195)
- feat: add task edit, delete, note, and search CLI commands (Story 23.7) (#194)
- feat: add mood and stats CLI commands (Story 23.8) (#189)
- feat: add config commands and stdin/pipe support (Story 23.9) (#190)

### Epic 24: MCP/LLM Integration (Stories 24.1–24.8)

- feat: add MCP server scaffold with stdio and SSE transports (Story 24.1) (#177)
- feat: add read-only MCP resources and query tools (Story 24.2) (#180)
- feat: add security middleware for MCP server (Story 24.3) (#179)
- feat: proposal store and enrichment API (Story 24.4) (#185)
- feat: TUI proposal review view (Story 24.5) (#197)
- feat: add analytics resources, tools, and prompts for MCP (Story 24.6) (#184)
- feat: task relationship graph & cross-provider linking (Story 24.7) (#191)
- feat: MCP prompt templates & advanced interaction patterns (Story 24.8) (#196)

### Epic 17: Door Themes (Stories 17.7–17.9)

- feat: redesign shoji theme with large panes and thin frame (Story 17.8) (#186)
- feat: simplify sci-fi theme and improve modern theme contrast (Story 17.9) (#183)
- fix: replace countRunes with ansi.StringWidth for correct visual width (Story 17.7) (#181)

### Bug Fixes

- fix: eliminate CLI test race condition by removing global jsonOutput (#192)
- fix: remove duplicate shortID function and test (already in doors.go)

### BMAD Planning

- feat: BMAD planning — Todoist Integration epic and stories (#198)
- feat: BMAD planning — GitHub Issues Integration epic and stories (#199)
- feat: BMAD planning — Daily Planning Mode epic and stories (#200)
- feat: BMAD planning — Snooze/Defer as First-Class Action epic and stories (#203)
- feat: BMAD planning — Task Dependencies and Blocked-Task Filtering epic and stories (#206)

### Docs

- docs: add ROADMAP.md (#193)
- docs: add Epic 23 — CLI Interface with 10 stories (#168)
- docs: add Epic 24 — MCP/LLM Integration Server with 8 stories (#169)
- docs: add next-phase prioritization roadmap (CLI -> MCP -> iPhone) (#165)
- docs: add story files for Epic 16 (iPhone Mobile App, SwiftUI) (#163)
- docs: add theme rendering course correction and stories 17.7-17.9 (#178)
- docs: add UX target visuals and tighten theme story ACs
- docs: sync epics-and-stories.md with PRs #142-164 (#166)

---

## 2026-03-06

Major epic completions: Dev Dispatch (Epic 22), Reminders (Epic 20), and Jira sync (Epic 19).

### Epic 22: Dev Dispatch / Self-Driving Pipeline (Stories 22.1–22.8)

- feat: add dev dispatch data model and queue persistence (Story 22.1) (#149)
- feat: add dispatch engine with multiclaude CLI wrapper (Story 22.2) (#152)
- feat: Story 22.3 — TUI dispatch key binding and confirmation flow (#163)
- feat: Story 22.4 — Dev Queue View (List, Approve, Kill) (#162)
- feat: Story 22.5 — Worker status polling and task update loop (#161)
- feat: Story 22.6 — Auto-generated review and follow-up tasks (#164)
- feat: Story 22.7 — Optional story file generation via AgentService (#159)
- feat: Story 22.8 — Safety guardrails (rate limiting, cost caps, audit log) (#160)

### Epic 20: Apple Reminders Integration (Stories 20.2–20.4)

- feat: Story 20.2 — Reminders read-only TaskProvider (#148)
- feat: Story 20.3 — Reminders write support (SaveTask, DeleteTask, MarkComplete) (#155)
- feat: Story 20.4 — Reminders config, registration, and health check (#158)

### Epic 19: Jira Integration (Stories 19.3–19.4)

- feat: Story 19.3 — Jira bidirectional sync with MarkComplete, cache, and retry (#150)
- feat: Story 19.4 — Jira config parsing, validation, and registration (#153)

### Epic 21: Multi-Source Sync (Stories 21.3–21.4)

- feat: Story 21.3 — complete SourceRef TaskPool integration (#151)
- feat: Story 21.4 — sync dashboard enhancements (#157)

### Other Features

- feat: Story 13.2 — Wire duplicate detection & source attribution into TUI (#142)
- feat: Story 9.3 — CI benchmark job for performance regression detection (#143)

### Bug Fixes

- fix: make TestSyncSchedulerNoGoroutineLeaks resilient to parallel test interference
- fix: update DevBadge test to match merged QUEUED badge format

### Docs

- docs: comprehensive epic/story status sync — 144 merged PRs audit (#145)
- docs: PRD validation v1.7 — sync status, add success criteria, fix gaps (#144)
- docs: comprehensive LLM integration & MCP server research (#156)
- docs: add CLI interface design research (#154)
- docs: add baseline test phase to implement-story workflow (#146)

---

## 2026-03-04

### Epic 19: Jira Integration (Story 19.2)

- feat: Story 19.2 — Jira Read-Only Provider (#138)

### Epic 20: Apple Reminders (Story 20.1)

- feat: Story 20.1 — Reminders JXA Scripts and CommandExecutor (#137)

### Epic 21: Multi-Source Sync (Story 21.1)

- feat: Story 21.1 — Sync Scheduler with Per-Provider Loops (#139)

### Epic 1: Core TUI (Story 1.3b)

- feat: Implement Expand & Fork actions in detail view (Story 1.3b) (#134)

### Docs

- docs: self-driving development pipeline — PRD, architecture, epics, and stories (#141)
- docs: update PRD, architecture, and epic list for self-driving pipeline (#140)
- docs: self-driving development pipeline research (#135)
- docs: Makefile vs Justfile analysis for ThreeDoors (#136)
- docs: design decisions requiring maintainer input — party mode requested
- docs: add story-driven development rule to CLAUDE.md (#130)

---

## 2026-03-03

Massive implementation sprint — themes, Obsidian, testing infrastructure, and foundation hardening.

### Epic 17: Door Theme System (Stories 17.1–17.6)

- feat: Story 17.1 — Theme types, registry, and Classic theme wrapper
- feat: Story 17.2 — Modern, Sci-Fi, and Shoji theme implementations
- feat: Story 17.3 — DoorsView theme integration with config support
- feat: Story 17.4 — Theme Picker in Onboarding Flow (#123)
- feat: Story 17.5 — :theme command with ThemePicker and config persistence (#124)
- test: Story 17.6 — Golden file tests for all door themes (#131)

### Epic 8: Obsidian Integration (Stories 8.1–8.4)

- feat: Story 8.1 — Obsidian Vault Reader/Writer Adapter
- feat: Stories 8.2 & 8.3 — Obsidian Bidirectional Sync & Vault Configuration
- feat: Story 8.4 — Obsidian Daily Note Integration
- test: Story 8.1 AC-Q6 input sanitization tests for ObsidianAdapter

### Epic 18: Docker E2E Testing (Stories 18.2–18.5)

- feat: Story 18.2 — Golden File Snapshot Tests for TUI Views
- feat: Story 18.3 — Input Sequence Replay Tests for User Workflows (#116)
- feat: Story 18.4 — Docker Test Environment for Reproducible E2E (#117)
- feat: Story 18.5 — CI Integration for Docker E2E Tests (#118)

### Epic 3.5: Foundation Hardening (Stories 3.5.1–3.5.8)

- feat: Story 3.5.1 — Core Domain Extraction
- feat: Story 3.5.2 — TaskProvider Interface Hardening
- feat: Story 3.5.3 — Config.yaml Schema & Migration Spike
- feat: Story 3.5.4 — Apple Notes Adapter Hardening
- feat: Story 3.5.5 — Baseline Regression Test Suite
- feat: Story 3.5.6 — Session Metrics Reader Library
- feat: Story 3.5.7 — Adapter Test Scaffolding & CI Coverage Floor
- docs: Story 3.5.8 — Validation Gate Decision Documentation

### Epic 7: Adapter Registry (Stories 7.1–7.3)

- feat: Story 7.1 — Adapter Registry & Runtime Discovery
- feat: Story 7.2 — Config-Driven Provider Selection
- feat: Story 7.3 — Adapter Developer Guide & Contract Tests

### Epic 9: Testing & Quality (Stories 9.1–9.5)

- feat: Story 9.1 — Apple Notes Integration E2E Tests
- feat: Story 9.2 — Contract Tests for Adapter Compliance
- feat: Story 9.3 — Performance Benchmarks
- feat: Story 9.4 — Functional E2E Tests (#107)
- feat: Story 9.5 — CI Coverage Gates (#113)

### Epic 11: Sync UX

- feat: Story 11.2 — Sync Status Indicator
- feat: Story 11.3 — Conflict Visualization & Sync Log

### Epic 12: Calendar Integration

- feat: Story 12.1 — Local Calendar Source Reader
- feat: Story 12.2 — Time-Contextual Door Selection

### Epic 13: Multi-Source Aggregation

- feat: Story 13.1 — Cross-Provider Task Pool Aggregation
- feat: Story 13.2 — Duplicate Detection & Source Attribution

### Other Features

- feat: Story 4.6 — Better Than Yesterday Multi-Dimensional Tracking
- feat: Story 14.2 — Agent Action Queue Integration
- feat: Jira & task sync integration — full pipeline (Epics 19-21) (#132, #133)

### Bug Fixes

- fix: use ansi.StringWidth for consistent line width tests
- fix: make doors 60% terminal height and clarify onboarding values vs tasks (#115)
- fix: set CI=true in Dockerfile.test for flaky watcher test
- fix: skip flaky TestObsidianWatcher_IgnoresSelfWrites in CI
- fix: bump notarization timeout to 1800s for initial submissions (#67)
- fix: increase notarization timeout to 3600s (1 hour) (#76)
- fix: increase notarization timeout to 4 hours
- fix: grant pkgbuild access to installer certificate in CI keychain (#111)
- fix: install Apple Developer ID G2 intermediate for notarization (#101)
- fix: make coverage PR comment continue-on-error for fork PRs

### Testing

- test: improve unit test coverage from 75.9% to 82.4%
- test: additional TUI coverage for delegate functions and search commands

### Docs

- docs: comprehensive user guide for ThreeDoors
- docs: update README with all features since PR #11
- docs: restore door emojis to README
- docs: door theme system research with ANSI mockups
- docs: Door theme system — analyst review, party mode, PRD update
- docs: Create story files for Epic 17 — Door Theme System
- docs: Apple Reminders integration research
- docs: Jira integration research for ThreeDoors TaskProvider
- docs: Task source expansion research — API integration feasibility
- docs: Sync architecture scaling research for multi-source support
- docs: UX & workflow improvements research
- docs: PR-story gap analysis and Epic 0 backfill stories
- docs: add AC verification rule and auto-execution research
- docs: Apple code signing & notarization investigation
- docs: CI signing pipeline audit
- docs: Signing/notarization failure timeline analysis
- docs: Audit story statuses against merged PRs
- docs: fix PRD validation findings
- docs: add story 1.2.1 (door height) and update story 10.2, PRD

---

## 2026-03-02

Project inception day — core TUI, Apple Notes integration, and CI/CD pipeline all built.

### Epic 1: Core TUI (Stories 1.1–1.8)

- feat: Implement Story 1.1 — Project Setup & Basic Bubbletea App (#2)
- feat: Implement Story 1.2 — Display Three Doors from a Task File (#4)
- feat: Implement Story 1.3 — Door Selection & Task Status Management (#5, #6)
- feat: Implement Story 1.4 — Quick Search & Command Palette (#7)
- feat: Story 1.5 — Session Metrics Tracking (tests + analysis scripts) (#8)
- feat: Story 1.6 — Essential Polish (#9)
- feat: Implement Story 1.7 — CI/CD Pipeline & Alpha Release (#10)
- feat: Story 1.8 — CI Process Validation & Fixes (#11)

### Epic 2: Apple Notes Integration (Stories 2.1–2.6)

- feat: Story 2.1 — Add MarkComplete to TaskProvider interface (#12)
- feat: Story 2.2 — Apple Notes Integration Spike (#13)
- feat: Story 2.3 — Read Tasks from Apple Notes (AppleNotesProvider) (#15)
- feat: Story 2.4 — Write Task Updates to Apple Notes (#16)
- feat: Implement Story 2.5 — Bidirectional Sync Engine (#17)
- feat: Implement Story 2.6 — Health Check Command (#18)

### Epic 3: Task Engagement (Stories 3.1–3.7)

- feat: Story 3.1 — Quick Add Mode (#19)
- feat: Story 3.2 — Extended Task Capture with Context (#20)
- feat: Story 3.3 — Values & Goals Setup and Display (#21)
- feat: Story 3.4 — Door Feedback Options (#22)
- feat: Story 3.5 — Daily Completion Tracking & Comparison
- feat: Story 3.6 — Session Improvement Prompt
- feat: Story 3.7 — Enhanced Navigation & Messaging

### Epic 4: Learning & Insights (Stories 4.1–4.5)

- feat: Story 4.1 — Task Categorization & Diversity-Preferring Door Selection
- feat: Story 4.2 — Session Metrics Pattern Analysis & Avoidance Detection
- feat: Story 4.3 — Mood Correlation Analysis
- feat: Story 4.4 — Avoidance Detection & User Insights
- feat: Story 4.5 — User Insights Dashboard

### Epic 5: macOS Distribution

- feat: Story 5.1 — macOS Distribution & Packaging
- feat: Story 5.1 — SQLite Enrichment Database Setup

### Epic 10: Onboarding

- feat: Story 10.1 — First-Run Onboarding Experience
- feat: Story 10.2 — Values/Goals Setup & Task Import in Onboarding

### Epic 11: Sync UX

- feat: Story 11.1 — Offline-First Local Change Queue (WAL)

### Epic 14: LLM/Agent

- feat: Story 14.1 — LLM Task Decomposition Spike

### Epic 15: Research

- feat: Story 15.1 — Choice Architecture Literature Review
- feat: Story 15.2 — Mood-Task Correlation & Procrastination Research

### Epic 18: Testing

- feat: Story 18.1 — Headless TUI Test Harness with teatest
- feat: Epic 18 — Docker E2E & Headless TUI Testing Infrastructure

### Other

- feat: Story 6.2 — Cross-Reference Tracking
- feat: create GitHub Release with compiled binaries on merge to main (#61)
- feat: add test coverage reporting to CI pipeline
- feat: add /implement-story reusable workflow command
- feat: Add comprehensive CLAUDE.md with Go quality rules
- feat: Add Epic 16 — iPhone Mobile App (SwiftUI)
- feat: Create comprehensive epics and stories breakdown for all phases

### Bug Fixes

- fix: Address code review findings for Stories 1.3, 1.5, 1.6, 2.1–2.4, 2.6, 5.1
- fix: align CI secret names with configured GitHub secrets (#61)
- fix: resolve rebase conflicts and remove obsolete Story 1.1 tests
- fix: resolve duplicate imports after rebase onto Story 1.2
- fix: apply gofumpt formatting to detail_view_test.go
- fix: handle errcheck for f.Close() in metrics_writer_test.go
- fix: resolve golangci-lint issues in metrics_writer_test.go

### Docs

- docs: architecture v2.0 — update for 9 PRD party mode recommendations
- feat: integrate 9 party mode recommendations into PRD
- docs: PRD validation — add missing BMAD sections and fix quality issues
- docs: regenerate epics from PRD v2.0 + add bridging Epic 3.5
- docs: PR analysis-derived quality gates, NFRs, and coding standards
- docs: expand PR submission standards across all project documentation
- docs: add Quality Gate ACs to all unimplemented stories
- docs: add Pre-PR Submission Checklist to all story files
- docs: sprint status audit, story validation, and status fixes
- docs: AI tooling research — CLAUDE.md, SOUL.md, skills, quality improvements
- docs: Add macOS distribution & packaging to PRD (Epic 5)
- docs: Add Story 1.8 — CI Process Validation & Fixes
- docs: add install and usage documentation to README
- docs: code signing research findings

---

## 2026-03-01

- chore: Add BMAD method command files and project documentation (#1)

---

## 2025-11-11

- docs: Enhance PRD with mood tracking, search/command palette, and comprehensive README
- docs: Sync documentation with Epic 1 evolution and archive legacy files

---

## 2025-11-08

Initial development and documentation.

- feat: Initialize Go module and add build tools
- feat: Add internal tasks, scripts, and documentation
- feat: Implement task management and file operations
- feat: Implement UX enhancements and new task management key bindings
- feat: Add QA documentation and initial tests
- refactor: Update main application and test files
- docs: Add README.md
- docs: Add stories documentation
- docs: Update architecture, PRD, and story documentation

---

## 2025-11-07

Project inception.

- Initial commit: Migrate simple-todo to ThreeDoors repository
- add initial product brief
- initial PRD
- Expand PRD with requirements, UI design, and technical architecture
- Add comprehensive architecture document with task management
- Add Epic 1 story breakdown and implementation roadmap
- Reorganize documentation into modular structure
- Pivot PRD to phased approach with technical demo validation
- Optimize Epic 1 story sequence and reduce timeline
- Add AI implementation clarifications and split documentation
- Update .gitignore with comprehensive Go-specific rules
