# ROADMAP — ThreeDoors

> Source of truth for merge-queue scope checks and worker prioritization.
> Synced periodically by BMAD PM agent from `docs/prd/epics-and-stories.md`.
> Last updated: 2026-03-10

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

**Status:** Not Started.

Operationalize merge-queue emergency mode: proactively check push-to-main CI after each merge, halt merges if main is red. Prerequisite safety net for the relaxed up-to-date rule (Story 0.20).

### Story 0.37: CI Efficiency Metrics — Track Runs Per Merged PR (P2)

**Status:** Not Started.

Shell script to measure CI efficiency (runs per merged PR, churn ratio, docs-skip rate). Validates Story 0.20 improvements and monitors ADR-0030 re-entry gate for GitHub merge queue.

### Story 0.31: CI/Security Hardening — Secrets, Supply Chain & Reproducibility (P1)

**Status:** Done (PR #270).

Pin golangci-lint version, replace third-party release action with gh CLI, add protected environment for release secrets.

### Story 0.28: Issue Tracker & Authority Configuration (P1)

**Status:** Done (PR #255).

Local issue tracker file (`docs/issue-tracker.md`) with authority tier configuration for the envoy agent. Operationalizes party mode research from PRs #227, #232.

### Story 0.29: Envoy Operations Guide (P1)

**Status:** Story created. Implementation not started. Depends on Story 0.28.

Operations documentation for the envoy agent: patrol workflows, cross-agent protocols, staleness thresholds, SOUL.md alignment patterns.

### Story 0.32: Help Display UX — Dedicated Help View (P1)

**Status:** Done (PR #309).

Replace broken `:help` flash message with dedicated scrollable help view. Content runs off-screen and disappears after 3 seconds. Fix: new `ViewHelp` mode, categorized two-column layout, `?` global keybinding.

### Story 0.34: Fix 'q' Key in Sub-Views — Go Back Instead of Quit (P1)

**Status:** Done (PR #361).

Story 36.3 (PR #276) universal quit handler causes sub-views (dashboard, health, synclog, etc.) to exit on 'q' instead of going back. Fix: scope 'q' quit to doors view only; sub-views treat 'q' as go-back (D-128).

## Active Epics

### Epic 30: Linear Integration (P2) — 0/4 stories done

Linear as task source via GraphQL API. Best task model alignment of all evaluated services.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 30.1 | Linear GraphQL Client & Auth Configuration | Not Started | P2 | Epic 7 (done) |
| 30.2 | Read-Only Linear Provider with Field Mapping | Not Started | P2 | 30.1 |
| 30.3 | Bidirectional Sync & WAL Integration | Not Started | P2 | 30.2 |
| 30.4 | Contract Tests & Integration Testing | Not Started | P2 | 30.2 |

### Epic 31: Expand/Fork Key Implementations (P2) — 0/5 stories done

Complete Expand (manual sub-task creation) and Fork (variant creation) TUI features per Design Decision H9.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 31.1 | Task Model ParentID Extension | Not Started | P2 | None |
| 31.2 | Enhanced Expand — Sequential Subtask Creation | Not Started | P2 | 31.1 |
| 31.3 | Subtask List Rendering in Detail View | Not Started | P2 | 31.1, 31.2 |
| 31.4 | Enhanced Fork — Variant Creation with ForkTask Factory | Not Started | P2 | None |
| 31.5 | Design Decision H9 Status Update | Not Started | P2 | 31.1-31.4 |

### Epic 33: Seasonal Door Theme Variants (P2) — 3/4 stories done

Time-based seasonal theme variants that auto-switch based on current date. Extends Epic 17's theme infrastructure.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 33.1 | Seasonal Theme Metadata Model and Date-Range Resolver | Done (PR #403) | P2 | Epic 17 (done) |
| 33.2 | Four Seasonal Theme Implementations | Done (PR #409) | P2 | 33.1 |
| 33.3 | Auto-Switch Integration in DoorsView and Config | Done (PR #410) | P2 | 33.1 |
| 33.4 | Seasonal Theme Picker and `:seasonal` Command | Not Started | P2 | 33.2, 33.3 |

### Epic 48: Door-Like Doors — Visual Door Metaphor Enhancement (P2) — 0/4 stories done

Transform rectangular card/panel doors into visually convincing doors using side-mounted handles, hinge marks, threshold lines, crack-of-light selection feedback, and handle turn micro-animations. Based on 5-round party mode research with 7 agents.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 48.1 | Side-Mounted Handle + Hinge Marks | Not Started | P2 | Epic 35 (done), Epic 17 (done) |
| 48.2 | Continuous Threshold / Floor Line | Not Started | P2 | None |
| 48.3 | Crack of Light Effect on Selection | Not Started | P2 | 48.1 |
| 48.4 | Handle Turn Micro-Animation | Not Started | P2 | 48.1 |

**Dependency graph:** Stories 48.1 & 48.2 can parallelize. Stories 48.3 & 48.4 can parallelize after 48.1 completes.

### Epic 43: Connection Manager Infrastructure (P1) — 2/6 stories done

Connection lifecycle layer for data source integrations. State machine, credential storage (system keychain), config schema v3 (named connections), CRUD operations, sync event logging, and migration of existing adapters to the new pattern.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 43.1 | Connection State Machine and ConnectionManager Type | Done (PR #428) | P1 | None |
| 43.2 | Keyring Integration with Environment Variable Fallback | In Review | P1 | None |
| 43.3 | Config Schema v3 Migration with Connections Support | In Review (PR #441) | P1 | None |
| 43.4 | Connection CRUD Operations | Not Started | P1 | 43.1, 43.2, 43.3 |
| 43.5 | Sync Event Logging Infrastructure | Not Started | P1 | None |
| 43.6 | Migrate Existing Adapters to ConnectionManager Pattern | Not Started | P1 | 43.1-43.5 |

### Epic 44: Sources TUI (P1) — 0/7 stories done

TUI interfaces for data source management: setup wizard (`:connect`), sources dashboard (`:sources`), source detail view, sync log view, status bar health alerts, disconnection flow, and re-authentication flow. Uses `charmbracelet/huh` for wizard forms.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 44.1 | Setup Wizard with huh Forms | Not Started | P1 | Epic 43 |
| 44.2 | Sources Dashboard View | Not Started | P1 | Epic 43 |
| 44.3 | Source Detail View | Not Started | P1 | 44.2 |
| 44.4 | Sync Log View | Not Started | P1 | 43.5 |
| 44.5 | Status Bar Integration for Connection Health Alerts | Not Started | P1 | Epic 43 |
| 44.6 | Disconnection Flow with Task Preservation Options | Not Started | P1 | 44.2 |
| 44.7 | Re-Authentication Flow | Not Started | P1 | 44.3, Epic 46 |

### Epic 45: Sources CLI (P1) — 0/5 stories done

Non-interactive CLI commands for data source management: `threedoors connect`, `threedoors sources` (list/status/test/manage/log), and JSON output support for scripting and CI/automation.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 45.1 | `threedoors connect` Command (Non-Interactive) | Not Started | P1 | Epic 43 |
| 45.2 | `threedoors sources` List/Status/Test Commands | Not Started | P1 | Epic 43 |
| 45.3 | `threedoors sources` Management Commands | Not Started | P1 | Epic 43 |
| 45.4 | `threedoors sources log` Command | Not Started | P1 | 43.5 |
| 45.5 | JSON Output Support for All Sources Commands | Not Started | P1 | 45.1-45.4 |

### Epic 46: OAuth Device Code Flow (P2) — 0/4 stories done

Generic OAuth device code flow client for browser-based authentication. Provider-specific integrations for GitHub and Linear. Silent token refresh with explicit re-auth on expiry.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 46.1 | Generic Device Code Flow Client | Not Started | P2 | None |
| 46.2 | GitHub OAuth Integration | Not Started | P2 | 46.1 |
| 46.3 | Linear OAuth Integration | Not Started | P2 | 46.1 |
| 46.4 | Token Refresh Lifecycle | Not Started | P2 | 46.1 |

### Epic 47: Sync Lifecycle & Advanced Features (P2) — 0/4 stories done

Conflict resolution (last-writer-wins with field-level strategy), orphaned task handling, auto-detection of installed tools in setup wizard, and proactive connection health notifications.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 47.1 | Conflict Resolution Strategy with Logging | Not Started | P2 | Epic 43 |
| 47.2 | Orphaned Task Handling | Not Started | P2 | 47.1 |
| 47.3 | Auto-Detection of Existing Tools in Setup Wizard | Not Started | P2 | 44.1 |
| 47.4 | Proactive Connection Health Notifications | Not Started | P2 | 44.5 |

**Epic 43-47 dependency graph:** Epic 43 is the critical path — all other epics depend on it. Epics 44 (TUI) and 45 (CLI) can parallelize after Epic 43. Epic 46 (OAuth) is independent. Epic 47 (Advanced) depends on 43+44.

### Epic 49: ThreeDoors Doctor — Self-Diagnosis Command (P1) — 1/10 stories done

Comprehensive self-diagnosis command with flutter-style category-based output, conservative auto-repair, and channel-aware version checking. Supersedes existing `health` command.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 49.1 | Doctor Command Skeleton & Health Alias | In Review | P1 | None |
| 49.2 | Environment Checks | Not Started | P1 | 49.1 |
| 49.3 | Task Data Integrity Checks | Not Started | P1 | 49.1 |
| 49.4 | Provider Health Checks | Not Started | P1 | 49.1 |
| 49.5 | Session & Analytics Checks | Not Started | P1 | 49.1 |
| 49.6 | Sync & Offline Queue Checks | Not Started | P1 | 49.1 |
| 49.7 | Enrichment Database Checks | Not Started | P1 | 49.1 |
| 49.8 | Auto-Repair (`--fix` flag) | Not Started | P1 | 49.2-49.7 |
| 49.9 | Channel-Aware Version Checking | Not Started | P1 | 49.1 |
| 49.10 | Verbose Mode, Category Filter & Polish | Not Started | P1 | 49.2-49.9 |

### Epic 42: Application Security Hardening (P1) — 1/5 stories done

Remediate findings from the application security audit. Fix permissive file permissions, add symlink validation, enforce input size limits, protect credentials in config, and harden CI supply chain.

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 42.1 | File Permission Standardization (0o700/0o600) | Done (PR #437) | P1 | None |
| 42.2 | Symlink Validation for File Operations | Not Started | P1 | None |
| 42.3 | Input Size Limits for YAML and JSONL Readers | Not Started | P1 | None |
| 42.4 | Credential Protection in Config Files | Not Started | P2 | 42.1 |
| 42.5 | CI Supply Chain Hardening | Not Started | P1 | None |

### Epic 50: In-App Bug Reporting (P2) — 0/3 stories done

In-app `:bug` command for frictionless bug reporting without leaving the TUI. Breadcrumb navigation trail, environment data collection with strict privacy allowlist, mandatory preview, and tiered submission (browser URL, GitHub API, local file).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 50.1 | Breadcrumb Tracking System | Not Started | P2 | None |
| 50.2 | Bug Report View & Environment Collection | Not Started | P2 | 50.1 |
| 50.3 | Submission Methods (Browser, API, File) | Not Started | P2 | 50.2 |

**Dependency graph:** Linear chain: 50.1 → 50.2 → 50.3.

## Completed Epics

| Epic | Title | Stories |
|------|-------|---------|
| 0 | Infrastructure & Process (Backfill) | 12/16 |
| 1 | Three Doors Technical Demo | 7/7 |
| 2 | Apple Notes Integration | 6/6 |
| 3 | Enhanced Interaction | 7/7 |
| 3.5 | Platform Readiness & Tech Debt | 8/8 |
| 4 | Learning & Intelligent Door Selection | 6/6 |
| 5 | macOS Distribution & Packaging | 1/1 |
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
| 34 | SOUL.md + Custom Development Skills | 4/4 |
| 35 | Door Visual Appearance — Door-Like Proportions | 7/7 |
| 36 | Door Selection Interaction Feedback | 4/4 |
| 37 | Persistent BMAD Agent Infrastructure | 4/4 |
| 29 | Task Dependencies & Blocked-Task Filtering | 4/4 |
| 32 | Undo Task Completion | 3/3 |
| 38 | Dual Homebrew Distribution | 6/6 |
| 27 | Daily Planning Mode | 5/5 |
| 28 | Snooze/Defer as First-Class Action | 4/4 |
| 39 | Keybinding Display System | 12/13 (1 cancelled) |
| 40 | Beautiful Stats Display | 10/10 |
| 41 | Charm Ecosystem Adoption & TUI Polish | 6/6 |

### Epic 51: SLAES — Self-Learning Agentic Engineering System (P1) — 0/10 stories done

Continuous improvement meta-system with a persistent `retrospector` agent that monitors PR merges, detects process waste, audits doc consistency, and files improvement recommendations to BOARD.md. Dual-loop architecture: spec-chain quality (did we build the right thing?) and operational efficiency (are we building efficiently?).

| Story | Title | Status | Priority | Depends On |
|-------|-------|--------|----------|------------|
| 51.1 | Retrospector Agent Definition (Responsibility+WHY Format) | Not Started | P1 | None |
| 51.2 | Rewrite Operational Agent Definitions (Responsibility+WHY Format) | Not Started | P1 | None |
| 51.3 | JSONL Findings Log & Per-Merge Lightweight Retro | Not Started | P1 | 51.1 |
| 51.4 | Saga Detection (Dispatch Waste Alerting) | Not Started | P1 | 51.1 |
| 51.5 | Doc Consistency Audit (Periodic Cross-Check) | In Review | P1 | 51.1 |
| 51.6 | BOARD.md Recommendation Pipeline | Not Started | P1 | 51.3, 51.4, 51.5 |
| 51.7 | Merge Conflict Rate Analysis | Not Started | P2 | 51.3, 51.6 |
| 51.8 | CI Failure Rate Analysis & Coding Standard Proposals | Not Started | P2 | 51.3, 51.6 |
| 51.9 | Research Lifecycle Tracking | Not Started | P2 | 51.3, 51.6 |
| 51.10 | PR Creation Authority & Trend Reporting | Not Started | P2 | 51.1-51.9 |

**Phasing:** Phase 0 (stories 51.1-51.2): Bootstrap — rewrite agent definitions. Phase 1 (stories 51.3-51.6): MVP monitoring. Phase 2 (stories 51.7-51.10): Advanced analysis after 2 weeks of MVP validation.

**Dependency graph:** Stories 51.1 & 51.2 can parallelize. Stories 51.3-51.5 can parallelize after 51.1. Story 51.6 depends on 51.3-51.5. Phase 2 stories depend on Phase 1 validation.

## Icebox (Deferred Indefinitely)

| Epic | Title | Stories | Decision Date | Rationale |
|------|-------|---------|---------------|-----------|
| 16 | iPhone Mobile App (SwiftUI) | 0/7 | 2026-03-07 | No validated user demand; core user is CLI/TUI power user; MCP (Epic 24) may serve mobile-adjacent use cases via LLM agents; adds significant platform/build/distribution complexity |

**Re-entry gate for Epic 16:** Revisit if 5+ distinct user requests for mobile access, OR if MCP proves insufficient for on-the-go task management.

## Out of Scope

Work not listed above is out of scope. Merge-queue should reject PRs that introduce features or epics not on this roadmap without human approval.
