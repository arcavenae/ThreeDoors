# ROADMAP — ThreeDoors

> Source of truth for merge-queue scope checks and worker prioritization.
> Synced periodically by BMAD PM agent from `docs/prd/epics-and-stories.md`.
> Last updated: 2026-03-15 (batch-767)

## Priority Legend

- **P0** — Must ship. Blocks other work or users.
- **P1** — Should ship. High value, no blockers.
- **P2** — Nice to have. Lower urgency.

## Open Issues

### Issue #219: Door Selection UX Improvements

**Status:** Resolved. Epic 36 fully implemented (PRs #272, #276, #277 merged). Issue #219 closed.

Door selection lacks tactile feedback and intuitive interaction patterns.

## Infrastructure Backlog

### Story 0.24: Renovate + Dependabot Automated Dependency Management (P1)

**Status:** Done (PR #402).

Renovate manages Go module dependencies (grouping, auto-merge, OSV vulnerability scanning). Dependabot manages GitHub Actions version pinning. Merge-queue integration via `dependencies` label.

### Story 0.20: CI Churn Reduction (P1)

**Status:** Done (PR #260).

Branch protection & merge queue optimization to reduce cascading CI reruns. Relaxed up-to-date requirement, added path filtering for docs-only PRs, deferred GitHub merge queue (ADR-0030).

### Story 0.36: CI Circuit Breaker — Post-Merge Main Branch Monitoring (P1)

**Status:** Done (PR #479).

Operationalize merge-queue emergency mode: proactively check push-to-main CI after each merge, halt merges if main is red. Prerequisite safety net for the relaxed up-to-date rule (Story 0.20).

### Story 0.37: CI Efficiency Metrics — Track Runs Per Merged PR (P2)

**Status:** Done (PR #494).

Shell script to measure CI efficiency (runs per merged PR, churn ratio, docs-skip rate). Validates Story 0.20 improvements and monitors ADR-0030 re-entry gate for GitHub merge queue.

### Story 0.50: Remove Session Reflection Quit Intercept (P0)

**Status:** Done (PR #556).

Remove the ImprovementView ("Session Reflection") that intercepts quit after productive sessions. Quit must be immediate — no confirmation, no prompt, no second keypress. Reverses Story 3.6. SOUL.md violation (D-165).

### Story 0.47: Alpha Version Chronological Sorting (P1)

**Status:** Done (PR #523).

Insert UTC HHMMSS timecode into alpha version format so same-day releases sort chronologically. Changes `0.1.0-alpha.YYYYMMDD.SHA` to `0.1.0-alpha.YYYYMMDD.HHMMSS.SHA`. One-line CI fix, SemVer 2.0 compliant. Unblocks Story 49.9 (channel-aware version checking).

### Story 0.31: CI/Security Hardening — Secrets, Supply Chain & Reproducibility (P1)

**Status:** Done (PR #270).

Pin golangci-lint version, replace third-party release action with gh CLI, add protected environment for release secrets.

### Story 0.28: Issue Tracker & Authority Configuration (P1)

**Status:** Done (PR #255).

Local issue tracker file (`docs/issue-tracker.md`) with authority tier configuration for the envoy agent. Operationalizes party mode research from PRs #227, #232.

### Story 0.29: Envoy Operations Guide (P1)

**Status:** Done (PR #259).

Operations documentation for the envoy agent: patrol workflows, cross-agent protocols, staleness thresholds, SOUL.md alignment patterns.

### Story 0.32: Help Display UX — Dedicated Help View (P1)

**Status:** Done (PR #309).

Replace broken `:help` flash message with dedicated scrollable help view. Content runs off-screen and disappears after 3 seconds. Fix: new `ViewHelp` mode, categorized two-column layout, `?` global keybinding.

### Story 0.34: Fix 'q' Key in Sub-Views — Go Back Instead of Quit (P1)

**Status:** Done (PR #361).

Story 36.3 (PR #276) universal quit handler causes sub-views (dashboard, health, synclog, etc.) to exit on 'q' instead of going back. Fix: scope 'q' quit to doors view only; sub-views treat 'q' as go-back (D-128).

### Story 0.38: Epic Number Registry Accuracy & Enforcement (P1)

**Status:** Done (PR #485).

Fix data accuracy in BOARD.md Epic Number Registry (wrong epic mappings, missing epics), backfill all active epics 39-51, and strengthen enforcement rules to reference project-watchdog as MUTEX. Implements D-112.

### Story 0.44: Scoped Label Migration — Rename and Create GitHub Labels (P1)

**Status:** Done (PR #513).

Migrate all 21 GitHub labels to scoped `.` separator format per finalized 27-label taxonomy (D-106, D-107). Rename-first strategy preserves label-issue associations (D-110). Create 6 new labels, delete 2 obsolete ones.

### Story 0.45: Agent Definition Updates for Scoped Labels (P1)

**Status:** Done (PR #520).

Update all agent definition files (envoy, merge-queue, pr-shepherd) to reference new scoped label names. Text-only changes — agents must be restarted after merge.

### Story 0.46: Label Authority & Triage Flow Documentation (P1)

**Status:** Done (PR #519).

Document label authority matrix (who sets/removes each label) and end-to-end triage flow. Covers BOARD recommendations P-003 and P-005. Consolidates party mode consensus into operational reference.

### Story 0.51: TUI View Rendering Benchmarks (P2)

**Status:** Done (PR #623).

Add benchmarks for `View()` rendering in complex TUI views (doors, dashboard, stats, sources). Captures baseline for regression detection. Driven by TEA audit R-001 finding that benchmark coverage is limited to core/textfile packages.

### Story 0.52: Multi-Adapter Integration Tests (P2)

**Status:** Done (PR #619).

Integration tests for sync conflict resolution across simulated adapter pairs. Covers last-writer-wins, orphaned task detection, field-level conflicts. Currently only unit-tested. Driven by TEA audit R-001.

### Story 0.53: Docker E2E Scenario Expansion (P2)

**Status:** Done (PR #621).

Audit and expand Docker E2E test scenarios to cover all primary user workflows (task completion, blocking, daily planning, source connection). Three-tier TUI testing infrastructure (ADR-0019) is fully built but scenario coverage gaps exist. Driven by TEA audit R-001.

### Story 0.57: t.Helper() Audit (P1)

**Status:** Done (PR #737).

Add `t.Helper()` to all test helper functions across the codebase. Currently only 23% of test files use it. Low effort, high payoff for debugging test failures. Driven by TEA audit R-001.

### Story 0.58: Fix govulncheck Vulnerabilities (P0)

**Status:** Done (PR #761).

Bump Go toolchain to 1.26.1 to resolve stdlib vulnerabilities (GO-2026-4599 through GO-2026-4602) reported by `govulncheck ./...`. Resolves issue #592.

### Story 0.59: Migrate from Makefile to Justfile (P2)

**Status:** Done (PR #768).

Replace Makefile with Justfile for better error messages, simpler syntax, and cross-platform support. Updated CLAUDE.md, CI workflows, agent definitions, and README.

### Story 0.60: PRD Forensic Reconstruction — Complete Coverage (P1)

**Status:** Done (PR #771).

Complete PRD forensic reconstruction to ensure full epic and story coverage across planning docs.

### Story 0.61: PRD Cleanup — Stale Content & Structural Fixes (P1)

**Status:** Done (PR #794).

Remove stale content and fix structural issues in PRD shards identified by forensic reconstruction.

### Story 0.62: Phase Numbering Consolidation in product-scope.md (P1)

**Status:** Done (PR #792).

Consolidate inconsistent phase numbering in product-scope.md.

### Story 0.63: User Journey Expansion for Post-Phase-2 Features (P2)

**Status:** Done (PR #791).

Expand user journey documentation to cover features implemented after Phase 2.

## Active Epics

### Epic 66: CLI/TUI Adapter Wiring Parity (P0) — 3/3 stories done — COMPLETE

Fix three gaps where implemented adapter code is not properly connected to CLI and TUI entry points: CLI adapter registration bug (critical), ClickUp connect wiring, and provider spec parity.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 66.1 | CLI Adapter Registration Fix | Done (PR #749) | P0 | None |
| 66.2 | ClickUp Connect Wiring | Done (PR #751) | P1 | 66.1 |
| 66.3 | Provider Spec Parity & Validation | Done (PR #750) | P1 | 66.2 |

**Dependency graph:** Linear chain: 66.1 → 66.2 → 66.3. Story 66.1 is the critical bug fix.

### Epic 65: CLI Test Coverage Hardening (P0) — 3/3 stories done — COMPLETE

Increase `internal/cli` from 34.8% to ≥70% coverage. Only critical gap from TEA audit. All three stories are independent and can parallelize.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 65.1 | Core CLI Path Tests — Bootstrap, Root, Doors, TaskPool Loading | Done (PR #740) | P0 | None |
| 65.2 | Subcommand Test Coverage — Config, Mood, Health, Stats, Plan | Done (PR #739) | P0 | None |
| 65.3 | Remaining Command Coverage — Task, Sources, Connect, Extract | Done (PR #736) | P1 | None |

**Dependency graph:** All three stories are fully independent and can be implemented in parallel.

### Epic 29: Task Dependencies & Blocked-Task Filtering (P1) — 4/4 stories done — COMPLETE

Native dependency graph support. Blocks tasks with unmet dependencies from door selection.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 29.1 | DependsOn Field, DependencyResolver, and YAML Persistence | Done (PR #307) | P1 | None |
| 29.2 | Door Selection Filter and Auto-Unblock on Completion | Done (PR #319) | P1 | 29.1 |
| 29.3 | TUI Blocked-By Indicator and Dependency Management | Done (PR #340) | P1 | 29.1 |
| 29.4 | Session Metrics Logging for Dependency Events | Done (PR #356) | P1 | 29.1 |

### Epic 30: Linear Integration (P2) — 4/4 stories done — COMPLETE

Linear as task source via GraphQL API. Best task model alignment of all evaluated services.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 30.1 | Linear GraphQL Client & Auth Configuration | Done (PR #446) | P2 | Epic 7 (done) |
| 30.2 | Read-Only Linear Provider with Field Mapping | Done (PR #699) | P2 | 30.1 |
| 30.3 | Bidirectional Sync & WAL Integration | Done (PR #709) | P2 | 30.2 |
| 30.4 | Contract Tests & Integration Testing | Done (PR #705) | P2 | 30.2 |

### Epic 31: Expand/Fork Key Implementations (P2) — 5/5 stories done — COMPLETE

Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features per Design Decision H9.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 31.1 | Task Model ParentID Extension | Done (PR #698) | P2 | None |
| 31.2 | Enhanced Expand — Sequential Subtask Creation | Done (PR #708) | P2 | 31.1 |
| 31.3 | Subtask List Rendering in Detail View | Done (PR #714) | P2 | 31.1, 31.2 |
| 31.4 | Enhanced Fork — Variant Creation with ForkTask Factory | Done (PR #701) | P2 | None |
| 31.5 | Design Decision H9 Status Update | Done (PR #718) | P2 | 31.1-31.4 |

### Epic 48: Door-Like Doors — Visual Door Metaphor Enhancement (P2) — 4/4 stories done — COMPLETE

Transform rectangular card/panel doors into visually convincing doors using side-mounted handles, hinge marks, threshold lines, crack-of-light selection feedback, and handle turn micro-animations. Based on 5-round party mode research with 7 agents.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 48.1 | Side-Mounted Handle + Hinge Marks | Done (PR #451) | P2 | Epic 35 (done), Epic 17 (done) |
| 48.2 | Continuous Threshold / Floor Line | Done (PR #483) | P2 | None |
| 48.3 | Crack of Light Effect on Selection | Done (PR #572) | P2 | 48.1 |
| 48.4 | Handle Turn Micro-Animation | Done (PR #588) | P2 | 48.1 |

**Dependency graph:** Stories 48.1 & 48.2 can parallelize. Stories 48.3 & 48.4 can parallelize after 48.1 completes.

### Epic 44: Sources TUI (P1) — 7/7 stories done — COMPLETE

TUI interfaces for data source management: setup wizard (`:connect`), sources dashboard (`:sources`), source detail view, sync log view, status bar health alerts, disconnection flow, and re-authentication flow. Uses `charmbracelet/huh` for wizard forms.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 44.1 | Setup Wizard with huh Forms | Done (PR #574) | P1 | Epic 43 |
| 44.2 | Sources Dashboard View | Done (PR #553) | P1 | Epic 43 |
| 44.3 | Source Detail View | Done (PR #563) | P1 | 44.2 |
| 44.4 | Sync Log View | Done (PR #582) | P1 | 43.5 |
| 44.5 | Status Bar Integration for Connection Health Alerts | Done (PR #562) | P1 | Epic 43 |
| 44.6 | Disconnection Flow with Task Preservation Options | Done (PR #581) | P1 | 44.2 |
| 44.7 | Re-Authentication Flow | Done (PR #654) | P1 | 44.3, Epic 46 |

### Epic 45: Sources CLI (P1) — 6/6 stories done — COMPLETE

Non-interactive CLI commands for data source management: `threedoors connect`, `threedoors sources` (list/status/test/manage/log), and JSON output support for scripting and CI/automation.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 45.1 | `threedoors connect` Command (Non-Interactive) | Done (PR #573) | P1 | Epic 43 |
| 45.2 | `threedoors sources` List/Status/Test Commands | Done (PR #550) | P1 | Epic 43 |
| 45.3 | `threedoors sources` Management Commands | Done (PR #587) | P1 | Epic 43 |
| 45.4 | `threedoors sources log` Command | Done (PR #565) | P1 | 43.5 |
| 45.5 | JSON Output Support for All Sources Commands | Done (PR #589) | P1 | 45.1-45.4 |
| 45.6 | Wire Connect Wizard into CLI Entry Point | Done (PR #732) | P1 | 44.1 |

### Epic 46: OAuth Device Code Flow (P2) — 4/4 stories done — COMPLETE

Generic OAuth device code flow client for browser-based authentication. Provider-specific integrations for GitHub and Linear. Silent token refresh with explicit re-auth on expiry.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 46.1 | Generic Device Code Flow Client | Done (PR #443) | P2 | None |
| 46.2 | GitHub OAuth Integration | Done (PR #636) | P2 | 46.1 |
| 46.3 | Linear OAuth Integration | Done (PR #671) | P2 | 46.1 |
| 46.4 | Token Refresh Lifecycle | Done (PR #631) | P2 | 46.1 |

### Epic 47: Sync Lifecycle & Advanced Features (P2) — 4/4 stories done — COMPLETE

Conflict resolution (last-writer-wins with field-level strategy), orphaned task handling, auto-detection of installed tools in setup wizard, and proactive connection health notifications.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 47.1 | Conflict Resolution Strategy with Logging | Done (PR #632) | P2 | Epic 43 |
| 47.2 | Orphaned Task Handling | Done (PR #650) | P2 | 47.1 |
| 47.3 | Auto-Detection of Existing Tools in Setup Wizard | Done (PR #635) | P2 | 44.1 |
| 47.4 | Proactive Connection Health Notifications | Done (PR #630) | P2 | 44.5 |

**Epic 43-47 dependency graph:** Epic 43 is the critical path — all other epics depend on it. Epics 44 (TUI) and 45 (CLI) can parallelize after Epic 43. Epic 46 (OAuth) is independent. Epic 47 (Advanced) depends on 43+44.

### Epic 42: Application Security Hardening (P1) — 5/5 stories done — COMPLETE

Remediate findings from the application security audit. Fix permissive file permissions, add symlink validation, enforce input size limits, protect credentials in config, and harden CI supply chain.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 42.1 | File Permission Standardization (0o700/0o600) | Done (PR #437) | P1 | None |
| 42.2 | Symlink Validation for File Operations | Done (PR #440) | P1 | None |
| 42.3 | Input Size Limits for YAML and JSONL Readers | Done (PR #448) | P1 | None |
| 42.4 | Credential Protection in Config Files | Done (PR #477) | P2 | 42.1 |
| 42.5 | CI Supply Chain Hardening | Done (PR #607) | P1 | None |

### Epic 50: In-App Bug Reporting (P2) — 3/3 stories done — COMPLETE

In-app `:bug` command for frictionless bug reporting without leaving the TUI. Breadcrumb navigation trail, environment data collection with strict privacy allowlist, mandatory preview, and tiered submission (browser URL, GitHub API, local file).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 50.1 | Breadcrumb Tracking System | Done (PR #478) | P2 | None |
| 50.2 | Bug Report View & Environment Collection | Done (PR #624) | P2 | 50.1 |
| 50.3 | Submission Methods (Browser, API, File) | Done (PR #649) | P2 | 50.2 |

**Dependency graph:** Linear chain: 50.1 → 50.2 → 50.3.

## Completed Epics

| Epic | Title | Stories |
|------|-------|---------|
| 0 | Infrastructure & Process (Backfill) | 35/35 |
| 1 | Three Doors Technical Demo | 7/7 |
| 2 | Apple Notes Integration | 6/6 |
| 3 | Enhanced Interaction | 7/7 |
| 3.5 | Platform Readiness & Tech Debt | 8/8 |
| 4 | Learning & Intelligent Door Selection | 6/6 |
| 5 | macOS Distribution & Packaging | 2/2 (COMPLETE) |
| 6 | Data Layer & Enrichment | 2/2 |
| 7 | Plugin/Adapter SDK & Registry | 3/3 |
| 8 | Obsidian Integration | 4/4 |
| 9 | Testing Strategy & Quality Gates | 5/5 |
| 10 | First-Run Onboarding | 2/2 |
| 11 | Sync Observability & Offline-First | 3/3 |
| 12 | Calendar Awareness | 2/2 |
| 13 | Multi-Source Aggregation | 2/2 |
| 14 | LLM Task Decomposition | 2/2 |
| 15 | Psychology Research & Validation | 2/2 |
| 17 | Door Theme System | 6/6 |
| 18 | Docker E2E & Headless TUI Testing | 5/5 |
| 19 | Jira Integration | 4/4 |
| 20 | Apple Reminders Integration | 4/4 |
| 21 | Sync Protocol Hardening | 4/4 |
| 22 | Self-Driving Development Pipeline | 8/8 |
| 23 | CLI Interface | 11/11 |
| 24 | MCP/LLM Integration Server | 8/8 |
| 25 | Todoist Integration | 4/4 |
| 26 | GitHub Issues Integration | 4/4 |
| 27 | Daily Planning Mode | 5/5 |
| 28 | Snooze/Defer as First-Class Action | 4/4 |
| 32 | Undo Task Completion | 3/3 |
| 33 | Seasonal Door Theme Variants | 4/4 |
| 34 | SOUL.md + Custom Development Skills | 4/4 |
| 35 | Door Visual Appearance — Door-Like Proportions | 7/7 |
| 36 | Door Selection Interaction Feedback | 4/4 |
| 37 | Persistent BMAD Agent Infrastructure | 4/4 |
| 38 | Dual Homebrew Distribution | 6/6 |
| 39 | Keybinding Display System | 12/13 (1 cancelled) |
| 40 | Beautiful Stats Display | 10/10 |
| 41 | Charm Ecosystem Adoption & TUI Polish | 6/6 |
| 42 | Application Security Hardening | 5/5 |
| 43 | Connection Manager Infrastructure | 6/6 |
| 44 | Sources TUI | 7/7 |
| 45 | Sources CLI | 5/5 |
| 46 | OAuth Device Code Flow | 4/4 |
| 47 | Sync Lifecycle & Advanced Features | 4/4 |
| 48 | Door-Like Doors | 4/4 |
| 49 | ThreeDoors Doctor | 10/10 |
| 52 | Envoy Three-Layer Firewall | 4/4 |
| 55 | CI Optimization Phase 1 | 3/3 |
| 56 | Door Visual Redesign — Three-Layer Depth System | 5/5 |
| 57 | LLM CLI Services | 8/8 |
| 58 | Supervisor Shift Handover | 7/7 |
| 59 | Full-Terminal Vertical Layout | 2/2 |
| 60 | README Overhaul | 5/5 |
| 61 | GitHub Pages User Guide | 4/4 |
| 62 | Retrospector Agent Reliability | 3/3 |

### Epic 53: Remote Collaboration — multiclaude Cross-Machine Access (P2) — 5/5 stories done — COMPLETE

Document and enable remote collaboration with multiclaude via SSH, with future MCP bridge support.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 53.1 | SSH Tunnel Setup & Documentation | Done (PR #613) | P2 | None |
| 53.2 | Remote Agent Attachment | Done (PR #615) | P2 | 53.1 |
| 53.3 | Cross-Machine Message Routing | Done (PR #665) | P2 | 53.1 |
| 53.4 | Remote Worker Dispatch | Done (PR #691) | P2 | 53.2, 53.3 |
| 53.5 | MCP Bridge Prototype | Done (PR #693) | P2 | 53.1 |

### Epic 51: SLAES — Self-Learning Agentic Engineering System (P1) — 11/11 stories done — COMPLETE

Continuous improvement meta-system with a persistent `retrospector` agent that monitors PR merges, detects process waste, audits doc consistency, and files improvement recommendations to BOARD.md. Dual-loop architecture: spec-chain quality (did we build the right thing?) and operational efficiency (are we building efficiently?).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 51.1 | Retrospector Agent Definition (Responsibility+WHY Format) | Done (PR #461) | P1 | None |
| 51.2 | Rewrite Operational Agent Definitions (Responsibility+WHY Format) | Done (PR #460) | P1 | None |
| 51.3 | JSONL Findings Log & Per-Merge Lightweight Retro | Done (PR #462) | P1 | 51.1 |
| 51.4 | Saga Detection (Dispatch Waste Alerting) | Done (PR #464) | P1 | 51.1 |
| 51.5 | Doc Consistency Audit (Periodic Cross-Check) | Done (PR #465) | P1 | 51.1 |
| 51.6 | BOARD.md Recommendation Pipeline | Done (PR #463) | P1 | 51.3, 51.4, 51.5 |
| 51.7 | Merge Conflict Rate Analysis | Done (PR #506) | P2 | 51.3, 51.6 |
| 51.8 | CI Failure Rate Analysis & Coding Standard Proposals | Done (PR #505) | P2 | 51.3, 51.6 |
| 51.9 | Research Lifecycle Tracking | Done (PR #507) | P2 | 51.3, 51.6 |
| 51.10 | PR Creation Authority & Trend Reporting | Done (PR #509) | P2 | 51.1-51.9 |
| 51.11 | Retrospector Autonomy Fixes — Agent Definition Rewrite | Done (PR #608) | P1 | 51.1 |

**Phasing:** Phase 0 (stories 51.1-51.2): Bootstrap — rewrite agent definitions. Phase 1 (stories 51.3-51.6): MVP monitoring. Phase 2 (stories 51.7-51.10): Advanced analysis after 2 weeks of MVP validation.

**Dependency graph:** Stories 51.1 & 51.2 can parallelize. Stories 51.3-51.5 can parallelize after 51.1. Story 51.6 depends on 51.3-51.5. Phase 2 stories depend on Phase 1 validation.

### Epic 54: Gemini Research Supervisor — Deep Research Agent Infrastructure (Rearchitected) (P2) — 5/5 stories done — COMPLETE

Persistent research-supervisor agent that wraps the official Gemini CLI (`@google/gemini-cli`) with OAuth authentication for web-grounded research. **Rearchitected** from the original Python/API-key approach (D-154 → D-164). Uses free tier: 50 Pro/day + 1,000 Flash/day. Features context packaging (8 bundles, `--include-directories`), three-layer result shielding, and dual-tier budget management. No Python, no API key, no third-party tools.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 54.1 | Research-Supervisor Agent Definition (Gemini CLI + OAuth) | Done (PR #537) | P2 | None |
| 54.2 | Gemini CLI Installation, OAuth Setup & Wrapper Script | Done (PR #538) | P2 | None |
| 54.3 | Context Packaging & Prompt Engineering (Gemini CLI) | Done (PR #664) | P2 | 54.1 |
| 54.4 | Result Shielding & Artifact Storage (Gemini CLI JSON) | Done (PR #690) | P2 | 54.1, 54.2 |
| 54.5 | Rate Limiting, Budget Management & Model Selection | Done (PR #689) | P2 | 54.1, 54.2 |

**Dependency graph:** Stories 54.1 & 54.2 can parallelize. Stories 54.3, 54.4, 54.5 can parallelize after 54.1+54.2 complete.

### Epic 55: CI Optimization Phase 1 (P1) — 3/3 stories done — COMPLETE

Reduce PR CI wall clock time from 3m33s to ~2m08s through CI configuration changes only. Docker E2E push-only, benchmark path filtering, local dev acceleration. No test code modifications.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 55.1 | Docker E2E Push-Only + Lint Version Fix | Done (PR #578) | P1 | None |
| 55.2 | Benchmark Path Filtering | Done (PR #579) | P1 | None |
| 55.3 | Local Dev Acceleration (make test-fast + CI Cache) | Done (PR #580) | P1 | None |

**Dependency graph:** All three stories are fully independent and can be implemented in parallel.

### Epic 57: LLM CLI Services (P1) — 8/8 stories done — COMPLETE

Enable ThreeDoors to invoke LLM CLI tools (Claude CLI, Gemini CLI, Ollama CLI) as subprocess-based service providers for intelligent task operations. ThreeDoors as CLIENT calling LLMs (Direction 1), complementing Epic 24's MCP server (Direction 2). Extends existing `LLMBackend` interface with CLI-based implementations via `os/exec`. Two-layer architecture: Services (what: extract, enrich, breakdown) + Backends (how: which CLI tool). Auto-discovery with fallback chain, privacy-tiered model (local default, cloud opt-in).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 57.1 | CLIProvider + CLISpec + CommandRunner Abstraction | Done (PR #610) | P0 | None |
| 57.2 | Auto-Discovery and Fallback Chain | Done (PR #629) | P0 | 57.1 |
| 57.3 | TaskExtractor Service + Extraction Prompt | Done (PR #628) | P0 | 57.1 |
| 57.4 | Extraction TUI (`:extract` + review screen) | Done (PR #652) | P0 | 57.3 |
| 57.5 | Extraction CLI (`threedoors extract`) | Done (PR #647) | P0 | 57.3 |
| 57.6 | TaskEnricher Service + Enrichment TUI | Done (PR #638) | P1 | 57.1 |
| 57.7 | TaskBreakdown Service (extend Epic 14) | Done (PR #633) | P1 | 57.1 |
| 57.8 | `threedoors llm status` Command | Done (PR #651) | P1 | 57.1, 57.2 |

**Dependency graph:** 57.1 is foundation. 57.2-57.7 parallelize after 57.1 (except 57.4/57.5 depend on 57.3). 57.8 depends on 57.1 + 57.2.

### Epic 58: Supervisor Shift Handover — Context-Aware Supervisor Rotation (P2) — 7/7 stories done — COMPLETE

Detect supervisor context window degradation via daemon monitoring, serialize operational state, and transfer control to a fresh supervisor instance while workers continue uninterrupted. Two phases: MVP (shift clock, snapshot, orchestrator, startup) and Hardening (emergency protocol, audit, manual trigger).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 58.1 | Shift Clock — Transcript Monitoring in Daemon Refresh Loop | Done (PR #627) | P2 | None |
| 58.2 | Rolling State Snapshot Generator | Done (PR #618) | P2 | None |
| 58.3 | Handover Orchestrator — Daemon Coordination Logic | Done (PR #645) | P2 | 58.1, 58.2 |
| 58.4 | Supervisor Startup with State File | Done (PR #656) | P2 | 58.2, 58.3 |
| 58.5 | Emergency Handover Protocol | Done (PR #657) | P2 | 58.3, 58.4 |
| 58.6 | Handover History & Audit Trail | Done (PR #658) | P2 | 58.3 |
| 58.7 | Manual Handover Trigger & User Notification | Done (PR #659) | P2 | 58.1, 58.3 |

**Dependency graph:** Stories 58.1 & 58.2 can parallelize. Story 58.3 depends on both. Story 58.4 depends on 58.2 & 58.3. Phase 2 stories (58.5-58.7) can parallelize after Phase 1 completes.

### Epic 59: Full-Terminal Vertical Layout (P1) — 2/2 stories done — COMPLETE

Transform ThreeDoors from a content-driven partial-terminal app into a full-terminal experience. AltScreen, fixed-header/flex-content/fixed-footer layout engine, capped door height with perceptual centering, and graceful degradation across terminal sizes. Based on full-terminal layout research (PR #329) and party mode consensus.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 59.1 | AltScreen + Layout Engine + Door Height Cap | Done (PR #655) | P1 | None |
| 59.2 | Header/Footer Extraction + Graceful Degradation + Secondary Views Fill Height | Done (PR #667) | P1 | 59.1 |

**Dependency graph:** Linear chain: 59.1 → 59.2. Story 59.1 is the foundation (AltScreen, layout engine, door cap). Story 59.2 is the follow-up (header/footer extraction, degradation breakpoints, all views fill height).

**Note:** Story 59.1 is a prerequisite for Story 39.2 (keybinding bar footer slot) per D-121.

### Epic 60: README Overhaul (P2) — 5/5 stories done — COMPLETE

Polish the README with centered badge clusters, table of contents, foldable reference sections, updated feature list reflecting 35+ completed epics, and visual demo section with ASCII art mockup.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 60.1 | README Badges & Header Polish | Done (PR #604) | P2 | None |
| 60.2 | README Table of Contents | Done (PR #612) | P2 | None |
| 60.3 | Foldable Reference Sections | Done (PR #620) | P2 | None |
| 60.4 | Feature List Audit & Update | Done (PR #617) | P2 | None |
| 60.5 | Visual Demo Section | Done (PR #616) | P2 | 60.4 |

**Dependency graph:** Stories 60.1-60.4 can parallelize. Story 60.5 depends on 60.4.

### Epic 62: Retrospector Agent Reliability — Messaging, BOARD.md Access, and Context Resilience (P1) — 3/3 stories done — COMPLETE

Fix three infrastructure reliability issues preventing the retrospector agent (SLAES) from operating as designed: broken messaging identity registration, inability to persist BOARD.md recommendations, and context exhaustion with state loss. All changes are to agent definitions and operational files — no application code.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 62.1 | Messaging Identity Verification + File-Based Fallback | Done (PR #675) | P1 | None |
| 62.2 | Recommendation Queue File + BOARD.md Batch Pipeline | Done (PR #676) | P1 | None |
| 62.3 | Structured Checkpointing + Context Budget Optimization | Done (PR #677) | P1 | None |

**Dependency graph:** All three stories are fully independent and can be implemented in parallel.

### Epic 67: Retrospector Operational Data Pipeline (P1) — 1/1 stories done — COMPLETE

Automate periodic sync of retrospector operational data (`docs/operations/`) to git via cron-triggered project-watchdog pipeline. Workers in isolated worktrees currently cannot see retrospector data. The project-watchdog handler already exists (Epic 62); this epic adds the cron trigger and supervisor verification.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 67.1 | Cron-Triggered Operational Data Sync Pipeline | Done (PR #757) | P1 | Epic 62 (done) |

### Epic 61: GitHub Pages User Guide (P2) — 4/4 stories done — COMPLETE

Publish ThreeDoors documentation as a professional GitHub Pages site using MkDocs + Material for MkDocs, making the user guide discoverable via search engines and accessible without cloning the repo.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 61.1 | MkDocs Infrastructure + Landing Page | Done (PR #609) | P2 | None |
| 61.2 | Getting Started + Core Workflow Guide | Done (PR #643) | P2 | 61.1 |
| 61.3 | Integrations Guide | Done (PR #644) | P2 | 61.1 |
| 61.4 | CLI, Config & Advanced Guide | Done (PR #646) | P2 | 61.1 |

### Epic 56: Door Visual Redesign — Three-Layer Depth System (P1) — 5/5 stories done — COMPLETE

Transform door rendering from imperceptible wireframe shadows into solid, 3D-feeling surfaces using a three-layer approach: background fill for visual mass, bevel lighting for raised-surface perception, and gradient shadow for spatial depth.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 56.1 | ThemeColors Extension + Background Fill | Done (PR #653) | P1 | None |
| 56.2 | Bevel Lighting (3D Raised-Surface Borders) | Done (PR #678) | P1 | None |
| 56.3 | Shadow Overhaul (Gradient Multi-Column) | Done (PR #668) | P1 | 56.1 |
| 56.4 | Panel Zone Shading | Done (PR #666) | P1 | 56.1 |
| 56.5 | Width-Adaptive Shadow Tuning | Done (PR #673) | P1 | 56.3 |

**Dependency graph:** 56.1 & 56.2 can parallelize. 56.3 & 56.4 depend on 56.1. 56.5 depends on 56.3.

## Icebox (Deferred Indefinitely)

| Epic | Title | Stories | Decision Date | Rationale |
|------|-------|---------|---------------|-----------|
| 16 | iPhone Mobile App (SwiftUI) | 0/7 | 2026-03-07 | No validated user demand; core user is CLI/TUI power user; MCP (Epic 24) may serve mobile-adjacent use cases via LLM agents; adds significant platform/build/distribution complexity |

**Re-entry gate for Epic 16:** Revisit if 5+ distinct user requests for mobile access, OR if MCP proves insufficient for on-the-go task management.

### Epic 63: ClickUp Integration (P2) — 4/4 stories done — COMPLETE

ClickUp as task source via REST API v2. Follows the established 4-story integration adapter pattern (Jira, Todoist, GitHub Issues, Linear).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 63.1 | ClickUp REST API Client & Auth Configuration | Done (PR #706) | P2 | Epic 7 (done) |
| 63.2 | Read-Only ClickUp Provider with Field Mapping | Done (PR #719) | P2 | 63.1 |
| 63.3 | Bidirectional Sync & WAL Integration | Done (PR #728) | P2 | 63.2 |
| 63.4 | Contract Tests & Integration Testing | Done (PR #727) | P2 | 63.2 |

**Dependency graph:** Stories 63.3 & 63.4 can parallelize after 63.2 completes.

### Epic 64: Cross-Computer Sync (P2) — 6/6 stories done — COMPLETE

Task data synchronization across multiple computers. Architecturally distinct from provider sync. Requires research spike before implementation.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 64.1 | Architecture Research Spike | Done (PR #715) | P2 | None |
| 64.2 | Device Identity & Registration | Done (PR #721) | P2 | 64.1 |
| 64.3 | Sync Transport Layer | Done (PR #734) | P2 | 64.1, 64.2 |
| 64.4 | Cross-Machine Conflict Resolution | Done (PR #731) | P2 | 64.1, 64.2 |
| 64.5 | Offline Queue & Reconciliation | Done (PR #743) | P2 | 64.3, 64.4 |
| 64.6 | Cross-Computer Sync E2E Tests | Done (PR #748) | P2 | 64.3, 64.4, 64.5 |

**Note:** Stories 64.2-64.6 are provisional — acceptance criteria will be refined after the research spike (64.1) completes.

### Story 5.3: DMG/pkg Installer for macOS (P2)

**Status:** Done (PR #707).

Single story under Epic 5 (macOS Distribution). CI generates signed, notarized .pkg installer uploaded to GitHub Releases alongside binaries. Reopens Epic 5 from COMPLETE to 1/2.

### Epic 69: TUI MainModel Decomposition (P1) — 4/4 stories done — COMPLETE

Refactor `internal/tui/main_model.go` (2991 lines) into focused files. Extract view transition/navigation logic, source/sync view controllers, planning/task management view controllers, and auxiliary view controllers into separate files.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 69.1 | Extract View Transition & Navigation Logic | Done (PR #767) | P1 | None |
| 69.2 | Extract Source/Sync View Controllers | Done (PR #786) | P1 | 69.1 |
| 69.3 | Extract Planning & Task Management View Controllers | Done (PR #779) | P1 | 69.1 |
| 69.4 | Extract Auxiliary View Controllers & Command Dispatch | Done (PR #789) | P1 | 69.2, 69.3 |

**Dependency graph:** 69.1 first, then 69.2 & 69.3 can parallelize, then 69.4 last.

### Epic 70: Completion History & Progress View (P1) — 3/3 stories done — COMPLETE

New `:history` TUI view and `threedoors history` CLI command for browsing completed tasks with aggregated stats.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 70.1 | Completion Data Reader & Aggregator | Done (PR #766) | P1 | None |
| 70.2 | History TUI View (`:history`) | Done (PR #780) | P1 | 70.1 |
| 70.3 | History CLI Command (`threedoors history`) | Done (PR #777) | P1 | 70.1 |

**Dependency graph:** 70.1 first, then 70.2 & 70.3 can parallelize.

### Epic 68: BOARD.md Redesign (P2) — 3/3 stories done — COMPLETE

Split BOARD.md into a focused active dashboard (<120 lines) and a complete decision archive, extract Epic Number Registry to its own file, fix duplicate IDs, and update all supporting docs and agent definitions.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 68.1 | Create Decision Archive and Extract Epic Registry | Done (PR #793) | P2 | None |
| 68.2 | Restructure Active BOARD.md | Done (PR #799) | P2 | 68.1 |
| 68.3 | Update Supporting Documentation for Board Redesign | Done (PR #800) | P2 | 68.2 |

**Dependency graph:** Linear chain: 68.1 → 68.2 → 68.3.

### Epic 71: Drop Apple Intel (darwin/amd64) Builds (P1) — 3/3 stories done — COMPLETE

Remove darwin/amd64 build targets from CI workflows, release builds, Homebrew formula, and installer/packaging to save CI runner minutes and focus on Apple Silicon (darwin/arm64). Reference: Issue #803.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 71.1 | Remove darwin/amd64 from CI Build and Alpha Release Pipeline | Done (PR #810) | P1 | None |
| 71.2 | Remove darwin/amd64 from Stable Release Workflow | Done (PR #815) | P1 | 71.1 |
| 71.3 | Update Docs, Tests, and Agent Definitions for Intel Removal | Done (PR #817) | P1 | 71.1, 71.2 |

**Dependency graph:** Linear chain: 71.1 → 71.2 → 71.3. Story 71.1 is the foundation (CI + GoReleaser). Story 71.2 handles the tag-triggered release workflow. Story 71.3 cleans up docs and tests.

### Epic 72: Operationalize GitHub Label Usage (P1) — 4/4 stories done — COMPLETE

Wire GitHub label application into agent workflows so that PRs are routinely labeled, issue labeling is resilient to envoy downtime, and mutual exclusivity is enforced. Agent definition and operational doc changes only — no application code. Based on label usage gap analysis (PR #806).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 72.1 | Merge-Queue PR Labeling | Done (PR #809) | P1 | None |
| 72.2 | Envoy Label Resilience | Done (PR #808) | P1 | None |
| 72.3 | Supervisor Label Discipline & Missing Label | Done (PR #813) | P1 | None |
| 72.4 | Retroactive Label Cleanup | Done (PR #814) | P2 | 72.3 |

**Dependency graph:** Stories 72.1, 72.2, and 72.3 are independent and can parallelize. Story 72.4 depends on 72.3 (needs `resolution.wontfix` label to exist).

---

## Dark Factory Phase 1: Stabilize & Harden

### Epic 73: Operational Foundation — Agent Reliability & Operator UX (P1) — 1/6 stories

Stabilize multiclaude operator experience and agent lifecycle. Fix operator UX (workspace-as-primary), remove redundant heartbeats, add hook-enforced git safety, design session handoff, quota monitoring, daemon-native heartbeats. Research: R-007, R-010, R-004. Decisions: Q-C-005, Q-C-010, Q-C-011.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 73.1 | Workspace-as-Primary Operator Pattern | Not Started | P1 | None |
| 73.2 | Remove CronCreate Heartbeats | Not Started | P1 | None |
| 73.3 | Hook-Enforced Git Safety for Workers | Done (PR #840) | P0 | None |
| 73.4 | Session Handoff Protocol for Persistent Agents | Not Started | P1 | None |
| 73.5 | Passive Quota Monitoring | Not Started | P2 | None |
| 73.6 | Daemon-Native Heartbeats | Not Started | P2 | 73.2 |

**Dependency graph:** 73.1, 73.2, 73.3, 73.4, 73.5 are independent. 73.6 depends on 73.2 (heartbeat removal before replacement). 73.3 is P0 — implement first.

### Epic 74: Golden Repo Hardening — CODEOWNERS, CI Gates & Provenance (P1) — 1/5 stories

Protect governance files via CODEOWNERS, enforce commit conventions via CI, add provenance tracking, define .dfcp.yaml, introduce typed story comments. Research: R-005, R-003, R-010. Decisions: Q-C-001, Q-C-002, Q-C-007, Q-C-012.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 74.1 | CODEOWNERS for ThreeDoors Golden Repo | Done (PR #839) | P0 | None |
| 74.2 | CI Scope-Check Workflow | Not Started | P1 | None |
| 74.3 | Provenance Tagging (L0-L4) | Not Started | P1 | None |
| 74.4 | DFCP Configuration File (.dfcp.yaml) | Not Started | P2 | 74.1, 74.3 |
| 74.5 | Typed Comments on Story Files | Not Started | P2 | None |

**Dependency graph:** 74.1, 74.2, 74.3, 74.5 are independent. 74.4 depends on 74.1 (CODEOWNERS) and 74.3 (provenance) as it consolidates both into a machine-readable format. 74.1 is P0 — implement first.

### Epic 75: Perplexity MCP Integration (P2) — 0/1 stories

Install Perplexity MCP server, disabled by default with per-session toggle. Parallel track. Research: R-008.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 75.1 | Install Perplexity MCP Server with Per-Session Toggle | Not Started | P2 | None |

**Dependency graph:** Independent — can be implemented at any time.

## Out of Scope

Work not listed above is out of scope. Merge-queue should reject PRs that introduce features or epics not on this roadmap without human approval.
